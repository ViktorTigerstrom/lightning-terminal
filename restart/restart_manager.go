package restart

import (
	"errors"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/signal"
)

var (
	interceptor signal.Interceptor
)

// Function signature for the start function of each start process
type StartFunc func() error

// StartProcess struct represents each start process in the tree
type StartProcess struct {
	Name            string
	Start           StartFunc
	Children        []*StartProcess
	cancel          chan struct{}
	errCallBack     func(error)
	successCallBack func()

	// restartMu ensures that only one restart is happening at the same time
	// for a start process.
	// Note that the lock order for the start processes is the following
	// traversal order of the tree:
	// 1. Lock the start process's restartMu.
	// 2. Lock all immediate children of the start process.
	// 3. Recursively apply the same lock order to the leftmost child and
	// its entire subtree.
	// 4. Once the leftmost branch is fully traversed, repeat the process
	// for the next child to the right, and so on.
	restartMu sync.Mutex
}

// NewStartProcess creates a new StartProcess
func NewStartProcess(name string, start StartFunc, successCallBack func(),
	errCallBack func(error)) *StartProcess {

	return &StartProcess{
		Name:            name,
		Start:           start,
		Children:        []*StartProcess{},
		cancel:          make(chan struct{}, 1),
		errCallBack:     errCallBack,
		successCallBack: successCallBack,
	}
}

// Manager manages the start processes and their start functions
type Manager struct {
	Root              *StartProcess
	StartProcesses    map[string]*StartProcess
	restartMultiplier float64

	// cancelMu ensures that only one cancel is happening at the same time
	// for the tree.
	cancelMu sync.Mutex
}

// A compile-time check to ensure that RestartManager implements RManager.
var _ RManager = (*Manager)(nil)

// NewManager creates a new RestartManager
func NewManager(restartMultiplier float64,
	intercept signal.Interceptor) *Manager {

	interceptor = intercept

	return &Manager{
		StartProcesses:    make(map[string]*StartProcess),
		restartMultiplier: restartMultiplier,
	}
}

// Add function to add a start process to the tree
func (m *Manager) Add(parentName string, s *StartProcess) error {
	parentStartProcess, exists := m.StartProcesses[parentName]
	if !exists {
		return errors.New("parent start process not found")
	}

	parentStartProcess.restartMu.Lock()
	defer parentStartProcess.restartMu.Unlock()

	m.StartProcesses[s.Name] = s

	parentStartProcess.Children = append(parentStartProcess.Children, s)

	return nil
}

// Add function to add a start process to the tree
func (m *Manager) AddNewRoot(s *StartProcess) error {
	if m.Root != nil {
		currentRoot := m.Root

		currentRoot.restartMu.Lock()
		defer currentRoot.restartMu.Unlock()

		// TODO: ALSO CANCEL CURRENT ROOTS CHILDREN, we need to hold the
		// root + all of the recursive children's restartMu throughout
		// the add of the new root, as they will depend on the root
		// start process.

		m.Root = s
		m.StartProcesses[s.Name] = s

		s.Children = append(s.Children, currentRoot)
	} else {
		m.Root = s
	}

	return nil
}

// Start function to start all start processes
func (m *Manager) Start(block bool) {
	m.cancelBranch(m.Root)

	var wg sync.WaitGroup

	wg.Add(1)

	go m.startProcess(m.Root, &wg)

	if block {
		wg.Wait()
	}
}

func (m *Manager) startProcess(s *StartProcess, wg *sync.WaitGroup) {
	s.restartMu.Lock()

	m.startProcessInternal(s, wg)
}

// NOTE: The start process's restartMu must be held when calling this function.
func (m *Manager) startProcessInternal(s *StartProcess, wg *sync.WaitGroup) {
	defer wg.Done()

	// In case the start process was cancelled du
	select {
	case <-s.cancel:
		s.restartMu.Unlock()

		return
	default:
	}

	restartWaitTime := time.Second
	restartTicker := time.NewTicker(restartWaitTime)

	for {
		err := s.Start()
		if err == nil {
			// Send a success signal to the success callback.
			s.successCallBack()

			// As the lock order for start processes is the breadth
			// first order of the tree, we can lock all children's
			// restart mutexes before attempting to run their start
			// function.
			for _, child := range s.Children {
				child.restartMu.Lock()
			}

			// When we hold all children's restart mutexes, we can
			// release the parent start process's mutex. This
			// ensures that cancelBranch is guaranteed to cancel any
			// children in case they should be cancelled.
			s.restartMu.Unlock()

			// Now start run all start functions of all child start
			// processes and their children recursively. This is
			// done concurrently, as the start functions of the
			// children are not dependent on each other.
			var childWG sync.WaitGroup

			for _, child := range s.Children {
				childWG.Add(1)
				go m.startProcessInternal(child, &childWG)
			}

			childWG.Wait()

			return
		} else {
			// Now send the start error over the start process's
			// error callback.
			s.errCallBack(err)

			// Now wait for the restart timeout to expire before we
			// attempt to run the start process's start function
			// again.
			select {
			case <-interceptor.ShutdownChannel():
				s.restartMu.Unlock()

				return
			case <-s.cancel:
				s.restartMu.Unlock()

				// Incase the start process is cancelled, we
				// also cancel the attempt to successfully run
				// the start process's start function.
				return
			case <-restartTicker.C:
				// Increase the restart timeout exponentially.
				restartWaitTime = time.Duration(
					float64(restartWaitTime) *
						m.restartMultiplier,
				)

				restartTicker.Reset(restartWaitTime)
			}
		}
	}
}

func (m *Manager) FailedAt(sName string, err error, block bool) {
	startProcess, exists := m.StartProcesses[sName]
	if !exists {
		return
	}

	m.errorBranch(startProcess, err)

	m.cancelBranch(startProcess)

	var wg sync.WaitGroup
	wg.Add(1)

	go m.startProcess(startProcess, &wg)

	if block {
		wg.Wait()
	}
}

// RestartAt function to restart at a specific start process
func (m *Manager) RestartAt(sName string, block bool) {
	s, exists := m.StartProcesses[sName]
	if !exists {
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go m.startProcess(s, &wg)

	if block {
		wg.Wait()
	}
}

// cancelBranch cancels all current ongoing restarts
func (m *Manager) cancelBranch(s *StartProcess) {
	m.cancelMu.Lock()
	defer m.cancelMu.Unlock()

	// First we send a cancel signal to all start processes in the tree.
	m.cancelAllChildren(s)

	// When we hold all children's restart mutexes, we are sure that any
	// child in the branch that should have been cancelled by the previous
	// call to cancelAllChildren is now cancelled.
	m.lockAllChildren(s, true)

	// We can therefore drain the cancel channel for all start processes in
	// the branch as any start process that should have been cancelled is
	// now cancelled.
	m.drainBranchCancel(s)

	// Then release all children's restart mutexes.
	m.lockAllChildren(s, false)
}

// cancelAllChildren sends a cancel signal to all recursive children of the
// start process.
func (m *Manager) cancelAllChildren(s *StartProcess) {
	// Drain the cancel channel incase the start process has already been
	// cancelled without listening to the channel.
	select {
	case <-s.cancel:
	default:
	}

	// Send a cancel signal to the start process which should stop any
	// ongoing attempts to tun the start processes start function.
	s.cancel <- struct{}{}

	// Cancel all children of the start process and their children
	// recursively.
	for _, child := range s.Children {
		m.cancelAllChildren(child)
	}
}

func (m *Manager) drainBranchCancel(s *StartProcess) {
	// Drain the cancel channel for the start process.
	select {
	case <-s.cancel:
	default:
	}

	// Then drain the cancel channel for all children of the start process
	// and their children recursively.
	for _, child := range s.Children {
		m.drainBranchCancel(child)
	}
}

func (m *Manager) lockAllChildren(s *StartProcess, lock bool) {
	if lock {
		s.restartMu.Lock()
	} else {
		s.restartMu.Unlock()
	}

	for _, child := range s.Children {
		m.lockAllChildren(child, lock)
	}
}

// errorBranch sends an error over the error channel for all of the start
// processes in the branch.
func (m *Manager) errorBranch(s *StartProcess, err error) {
	// Send an error signal over the start process error callback.
	s.errCallBack(err)

	// Also send an error signal over all of the children's error channels.
	for _, child := range s.Children {
		m.errorBranch(child, err)
	}
}
