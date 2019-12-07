package middleware

import (
	"errors"
)

// Sema the semaphore object used internally
type Sema struct {
	Size int
	Ch   chan int
}

// NewSema creates a new semaphore
func NewSema(Length int) Sema {
	obj := Sema{}
	InitChan := make(chan int, Length)
	obj.Size = Length
	obj.Ch = InitChan
	return obj
}

// Acquire aquires a semaphore lock
func (s *Sema) Acquire() error {
	select {
	case s.Ch <- 1:
		return nil
	default:
		return errors.New("all semaphore buffer full")
	}
}

// Release release a semaphore lock
func (s *Sema) Release() error {
	select {
	case <-s.Ch:
		return nil
	default:
		return errors.New("all semaphore buffer empty")
	}
}
