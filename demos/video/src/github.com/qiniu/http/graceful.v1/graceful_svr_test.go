package graceful

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {

	hasReq := make(chan bool, 10)
	shouldRT := make(chan bool, 1)
	times := 0
	creator := func() http.Handler {

		times += 1
		myTimes := times
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			hasReq <- true
			<-shouldRT
			w.Header().Set("times", strconv.Itoa(myTimes))
			w.WriteHeader(200)
		})
	}

	srv := New(creator)
	timeout := int64(time.Millisecond * 20)
	go srv.ProcessSignals(timeout)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// test normal
	{
		shouldRT <- true
		testGet(t, ts.URL, "1", 200)
		<-hasReq
	}

	// test reload
	{
		go func() {
			<-hasReq
			syscall.Kill(os.Getpid(), syscall.SIGUSR2)
			time.Sleep(time.Millisecond * 10)
			shouldRT <- true
		}()
		// get before reload, served by previous srv
		testGet(t, ts.URL, "1", 200)

		// get after reload, served by current srv
		shouldRT <- true
		testGet(t, ts.URL, "2", 200)
		<-hasReq
	}

	// test quit
	{
		go func() {
			<-hasReq
			shouldRT <- true
			srv.Quit(0, int64(time.Millisecond*40))
		}()

		// get before quit, should be served
		testGet(t, ts.URL, "2", 200)

		// get after quit, should get 570
		time.Sleep(time.Millisecond * 10)
		shouldRT <- true
		testGet(t, ts.URL, "", 570)
		<-hasReq
	}

}

func testGet(t *testing.T, u, times string, code int) {

	ast := assert.New(t)

	resp, err := http.Get(u)
	ast.Nil(err)
	ast.Equal(code, resp.StatusCode)
	if code == 200 {
		ast.Equal(times, resp.Header.Get("times"))
	}
	resp.Body.Close()
}
