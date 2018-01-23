package unixsock

import (
	"net"
	"syscall"
)

var (
	pipeBasePath = "/var/tmp/qiniu.unixsock."
)

// --------------------------------------------------------------------

func Listen(nett, laddr string) (l net.Listener, err error) {

	fulladdr := pipeBasePath + laddr
	return net.Listen(nett, fulladdr)
}

func Dial(nett, addr string) (c net.Conn, err error) {

	fulladdr := pipeBasePath + addr
	c, err = net.Dial(nett, fulladdr)
	if err == nil {
		if c == nil {
			err = syscall.ENOENT
		}
	} else if err2, ok := err.(*net.OpError); ok {
		err = err2.Err
	}
	return
}

func Remove(addr string) error {

	fulladdr := pipeBasePath + addr
	return syscall.Unlink(fulladdr)
}

// --------------------------------------------------------------------

