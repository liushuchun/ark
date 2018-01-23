// +build !windows

package sync

import (
	"os"
	"syscall"
)

// --------------------------------------------------------------------

type Mutex os.File

var mutexBase = os.TempDir() + "/"

func CreateMutex(name string) (mutex *Mutex, err error) {

	f, err := os.Create(mutexBase + name)
	if err != nil {
		return
	}

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		f.Close()
		return
	}

	return (*Mutex)(f), nil
}

func (mutex *Mutex) Close() (err error) {

	f := (*os.File)(mutex)
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN|syscall.LOCK_NB)
	return f.Close()
}

// --------------------------------------------------------------------
