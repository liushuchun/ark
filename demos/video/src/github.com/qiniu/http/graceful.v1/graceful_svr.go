package graceful

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/qiniu/log.v1"
)

const (
	closeWaitTick = 2e8
)

// --------------------------------------------------------------------

type realService struct {
	http.Handler
	processNum int32
}

func (h *realService) closeAndWait(timeout int64) (err error) {

	from := time.Now().UnixNano()
	for {
		processNum := atomic.LoadInt32(&h.processNum)
		if processNum == 0 {
			break
		}
		if timeout != 0 {
			duration := time.Now().UnixNano() - from
			if duration > timeout {
				return fmt.Errorf("close wait timeout after %d ms, %d process remained", duration/1e6, processNum)
			}
		}
		time.Sleep(time.Duration(closeWaitTick))
	}
	return
}

// --------------------------------------------------------------------

type Service struct {
	*realService
	creator func() http.Handler
	closed  bool
}

func New(creator func() http.Handler) *Service {

	h := &realService{
		Handler: creator(),
	}
	return &Service{realService: h, creator: creator}
}

func (p *Service) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if p.closed {
		w.WriteHeader(570)
		return
	}

	h := (*realService)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&p.realService))))

	atomic.AddInt32(&h.processNum, 1)
	h.ServeHTTP(w, req)
	atomic.AddInt32(&h.processNum, -1)
}

func (p *Service) Quit(code int, timeout int64) {

	p.closed = true

	h := (*realService)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&p.realService))))
	err := h.closeAndWait(timeout)
	if err != nil {
		log.Warn("Graceful-quit failed:", err)
		os.Exit(254)
	}
	os.Exit(code)
}

func (p *Service) Reload(timeout int64) {

	h := &realService{
		Handler: p.creator(),
	}
	h = (*realService)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&p.realService)), unsafe.Pointer(h)))
	err := h.closeAndWait(timeout)
	if err != nil {
		log.Info("Graceful-reload failed:", err)
	}
}

func (p *Service) ProcessSignals(timeout int64) {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGUSR2, os.Interrupt, os.Kill)

	for {
		sig := <-c
		switch sig {
		case syscall.SIGUSR2:
			log.Info("Receive reload signal:", sig.String())
			p.Reload(timeout)
		default:
			log.Info("Receive graceful-close signal:", sig.String())
			p.Quit(exitCodeOf(sig), timeout)
		}
	}
}

func exitCodeOf(sig os.Signal) int {

	if v, ok := sig.(syscall.Signal); ok {
		return 128 + int(v)
	}
	return 255
}

// --------------------------------------------------------------------
