package ui

// ExecutionMode controls whether tasks run concurrently or one at a time.
type ExecutionMode int

const (
	ExecutionModeAsync ExecutionMode = iota
	ExecutionModeSync
)

func (m ExecutionMode) String() string {
	switch m {
	case ExecutionModeSync:
		return "sync"
	default:
		return "async"
	}
}

func (m ExecutionMode) Toggle() ExecutionMode {
	if m == ExecutionModeSync {
		return ExecutionModeAsync
	}
	return ExecutionModeSync
}
