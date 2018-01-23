package signal

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForInterrupt() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-c
}
