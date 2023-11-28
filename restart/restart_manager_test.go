package restart

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/signal"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	interceptor := signal.Interceptor{}
	rm := NewManager(2.0, interceptor)

	if rm.restartMultiplier != 2.0 {
		t.Errorf("Expected restartMultiplier to be 2.0, got %f", rm.restartMultiplier)
	}

	if len(rm.StartProcesses) != 0 {
		t.Errorf("Expected StartProcesses to be initialized empty")
	}
}

func TestNewStartProcess(t *testing.T) {
	startFunc := func() error { return nil }
	node := NewStartProcess("testNode", startFunc, nil, nil)

	if node.Name != "testNode" {
		t.Errorf("Expected node name to be 'testNode', got %s", node.Name)
	}

	if node.Start == nil {
		t.Errorf("Expected start function to be initialized")
	}
}

func TestAddNode(t *testing.T) {
	interceptor := signal.Interceptor{}
	rm := NewManager(2.0, interceptor)

	parentNode := NewStartProcess(
		"parent",
		func() error { return nil },
		func() {},
		func(err error) {},
	)
	rm.StartProcesses["parent"] = parentNode

	childNode := NewStartProcess(
		"child",
		func() error { return nil },
		func() {},
		func(err error) {},
	)

	err := rm.Add("parent", childNode)

	if err != nil {
		t.Errorf("Failed to add child node: %v", err)
	}

	if len(parentNode.Children) != 1 {
		t.Errorf("Expected 1 child, found %d", len(parentNode.Children))
	}

	if parentNode.Children[0].Name != "child" {
		t.Errorf("Expected child node to be 'child', found %s", parentNode.Children[0].Name)
	}
}

func TestStartNode(t *testing.T) {
	startCalled := make(chan struct{})
	node := NewStartProcess(
		"testNode",
		func() error {
			startCalled <- struct{}{}
			return nil
		},
		func() {},
		func(err error) {},
	)

	interceptor := signal.Interceptor{}
	rm := NewManager(2.0, interceptor)
	rm.Root = node

	var wg sync.WaitGroup
	wg.Add(1)
	go rm.Start(true)

	select {
	case <-startCalled:
		wg.Done()
		// start function called successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Start function was not called in time")
	}

	wg.Wait() // Ensure goroutine finishes
}

func TestNodeFailureAndRetry(t *testing.T) {
	attempt := 0
	maxAttempts := 3
	node := NewStartProcess(
		"testNode",
		func() error {
			if attempt < maxAttempts {
				attempt++
				return errors.New("error")
			}
			return nil
		},
		func() {},
		func(err error) {},
	)

	interceptor := signal.Interceptor{}
	rm := NewManager(1.0, interceptor)
	rm.Root = node

	rm.Start(true)

	require.Equal(t, maxAttempts, attempt)
}

func TestFailedAtFunction(t *testing.T) {
	startCalled := make(chan struct{})
	node := NewStartProcess(
		"testNode", func() error {
			startCalled <- struct{}{}
			return errors.New("error")
		},
		func() {},
		func(err error) {},
	)

	interceptor := signal.Interceptor{}
	rm := NewManager(1.0, interceptor)
	rm.StartProcesses["testNode"] = node

	go rm.FailedAt("testNode", errors.New("failed"), false)

	select {
	case <-startCalled:
		// Start function was called
	case <-time.After(1 * time.Second):
		t.Fatal("Start function was not called in time")
	}
}

func TestRestartAtFunction(t *testing.T) {
	restarted := make(chan struct{})
	node := NewStartProcess(
		"testNode",
		func() error {
			restarted <- struct{}{}
			return nil
		},
		func() {},
		func(err error) {},
	)

	interceptor := signal.Interceptor{}
	rm := NewManager(1.0, interceptor)
	rm.StartProcesses["testNode"] = node

	go rm.RestartAt("testNode", false)

	select {
	case <-restarted:
		// Node was restarted
	case <-time.After(1 * time.Second):
		t.Fatal("Node was not restarted in time")
	}
}

func TestShutdownBehaviour(t *testing.T) {
	node := NewStartProcess(
		"testNode",
		func() error {
			return errors.New("error")
		},
		func() {},
		func(err error) {},
	)

	interceptor := signal.Interceptor{}
	rm := NewManager(1.0, interceptor)
	rm.Root = node

	// Send shutdown signal
	interceptor.RequestShutdown()

	// Ensure that the goroutine finishes without getting stuck
	done := make(chan struct{})
	go func() {
		rm.Start(true)
		close(done)
	}()

	select {
	case <-done:
		// Shutdown was successful
	case <-time.After(3 * time.Second):
		t.Fatal("Shutdown did not complete in time")
	}
}

func TestSuccessCallBackManagement(t *testing.T) {
	successChannel := make(chan struct{}, 1)

	node := NewStartProcess(
		"testNode",
		func() error {
			return nil
		},
		func() {
			successChannel <- struct{}{}
		},
		func(err error) {},
	)

	interceptor := signal.Interceptor{}
	rm := NewManager(1.0, interceptor)
	rm.Root = node

	rm.Start(false)

	select {
	case <-successChannel:
		// Success callback was called.
	case <-time.After(1 * time.Second):
		t.Fatal("Success channel was not notified")
	}
}
