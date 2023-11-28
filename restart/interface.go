package restart

// SubServer defines an interface that should be implemented by any sub-server
// that the subServer manager should manage. A sub-server can be run in either
// integrated or remote mode. A sub-server is considered non-fatal to LiT
// meaning that if a sub-server fails to start, LiT can safely continue with its
// operations and other sub-servers can too.
type RManager interface {
	Add(string, *StartProcess) error

	AddNewRoot(s *StartProcess) error

	Start(bool)

	FailedAt(string, error, bool)

	RestartAt(string, bool)
}
