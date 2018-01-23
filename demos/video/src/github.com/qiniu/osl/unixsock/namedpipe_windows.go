package unixsock

/*
import (
	"unsafe"
	"syscall"
)

// --------------------------------------------------------------------

var (
	modkernel32, _          = syscall.LoadLibrary("kernel32.dll")
	procCreateNamedPipeW, _ = syscall.GetProcAddress(modkernel32, "CreateNamedPipeW")
	procConnectNamedPipe, _ = syscall.GetProcAddress(modkernel32, "ConnectNamedPipe")
	procWaitNamedPipeW, _   = syscall.GetProcAddress(modkernel32, "WaitNamedPipeW")
	procCallNamedPipeW, _   = syscall.GetProcAddress(modkernel32, "CallNamedPipeW")
)

// HANDLE WINAPI CreateNamedPipe(
//  __in      LPCTSTR lpName,
//   __in      DWORD dwOpenMode,
//   __in      DWORD dwPipeMode,
//   __in      DWORD nMaxInstances,
//   __in      DWORD nOutBufferSize,
//   __in      DWORD nInBufferSize,
//   __in      DWORD nDefaultTimeOut,
//   __in_opt  LPSECURITY_ATTRIBUTES lpSecurityAttributes
// )
func CreateNamedPipe(
	name string, openMode uint32, pipeMode uint32, maxInstances uint32,
	outBufSize uint32, inBufSize uint32, defaultTimeout uint32) (fd int, err error) {

	h, _, err := syscall.Syscall9(
		uintptr(procCreateNamedPipeW), 8,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
		uintptr(openMode), uintptr(pipeMode), uintptr(maxInstances), uintptr(outBufSize), uintptr(inBufSize),
		uintptr(defaultTimeout), uintptr(0), uintptr(0))

	return int(h), err
}

// BOOL WINAPI ConnectNamedPipe(
//   __in         HANDLE hNamedPipe,
//   __inout_opt  LPOVERLAPPED lpOverlapped
// )
func ConnectNamedPipe(fd int) (b bool, err error) {

	r, _, err := syscall.Syscall(uintptr(procConnectNamedPipe), 2, uintptr(fd), uintptr(0), uintptr(0))
	return (r != 0), err
}

// BOOL WINAPI WaitNamedPipe(
//   __in  LPCTSTR lpNamedPipeName,
//   __in  DWORD nTimeOut
// )
const (
	NMPWAIT_USE_DEFAULT_WAIT = 0
	NMPWAIT_WAIT_FOREVER     = 0xffffffff
)

func WaitNamedPipe(name string, timeout uint32) (b bool, err error) {

	r, _, err := syscall.Syscall(
		uintptr(procWaitNamedPipeW), 2,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))), uintptr(timeout), uintptr(0))
	return (r != 0), err
}

// BOOL WINAPI CallNamedPipe(
//   __in   LPCTSTR lpNamedPipeName,
//   __in   LPVOID lpInBuffer,
//   __in   DWORD nInBufferSize,
//   __out  LPVOID lpOutBuffer,
//   __in   DWORD nOutBufferSize,
//   __out  LPDWORD lpBytesRead,
//   __in   DWORD nTimeOut
// )
func CallNamedPipe(
	name string, inBuf []byte, outBuf []byte, bytesRead *uint32, timeout uint32) (b bool, err error) {

	r, _, err := syscall.Syscall9(uintptr(procCallNamedPipeW), 7,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
		uintptr(unsafe.Pointer(&inBuf[0])), uintptr(len(inBuf)),
		uintptr(unsafe.Pointer(&outBuf[0])), uintptr(len(outBuf)),
		uintptr(unsafe.Pointer(bytesRead)), uintptr(timeout),
		uintptr(0), uintptr(0))
	return (r != 0), err
}

func CreatePipeFile(name string) (fd int32, err error) {

	return syscall.CreateFile(
		syscall.StringToUTF16Ptr(name),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0, nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL, 0) // | C.FILE_FLAG_OVERLAPPED, 0)
}
*/
// --------------------------------------------------------------------

