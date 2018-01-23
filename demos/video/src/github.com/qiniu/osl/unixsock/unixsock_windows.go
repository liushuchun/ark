package unixsock

import (
	"net"
	"syscall"
)

// --------------------------------------------------------------------

func Listen(nett, laddr string) (l net.Listener, err error) {

	panic("notimpl")
	return nil, syscall.EACCES
}

func Dial(nett, addr string) (c net.Conn, err error) {

	panic("notimpl")
	return nil, syscall.ENOENT
}

func Remove(addr string) error {

	panic("notimpl")
	return nil
}

// --------------------------------------------------------------------

