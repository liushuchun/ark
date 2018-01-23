// +build windows

package sync

import (
	"log"
	"syscall"
	"unsafe"
)

// --------------------------------------------------------------------

var g_createMutex uintptr

func init() {

	libkernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		log.Fatal(err)
	}

	g_createMutex, err = syscall.GetProcAddress(libkernel32, "CreateMutexW")
	if err != nil {
		log.Fatal(err)
	}
}

// --------------------------------------------------------------------

type Mutex struct {
}

func CreateMutex(name string) (mutex *Mutex, err error) {

	namew, _ := syscall.UTF16PtrFromString(name)
	ret, _, errno := syscall.Syscall(g_createMutex, 3,
		0, 0, uintptr(unsafe.Pointer(namew)))
	if errno != 0 {
		if errno == syscall.ERROR_ALREADY_EXISTS {
			return nil, syscall.EAGAIN
		}
		return nil, errno
	}

	return (*Mutex)(unsafe.Pointer(ret)), nil
}

func (mutex *Mutex) Close() (err error) {

	return syscall.CloseHandle(syscall.Handle(unsafe.Pointer(mutex)))
}

// --------------------------------------------------------------------
