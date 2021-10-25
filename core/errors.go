package core

import "fmt"

type ErrNotFound struct{}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("asset not found")
}
