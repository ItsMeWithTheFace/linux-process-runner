package core

import "fmt"

type ErrNotFound struct{}

type ErrIllegalStateChange struct{}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("asset not found")
}

func (e *ErrIllegalStateChange) Error() string {
	return fmt.Sprintf("attempted to change from terminal state")
}
