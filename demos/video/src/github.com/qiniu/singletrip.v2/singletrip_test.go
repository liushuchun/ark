package singletrip

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"qiniupkg.com/x/log.v7"

	"github.com/stretchr/testify.v2/require"
)

func TestError(t *testing.T) {
	f, err := os.Create(filepath.Join(tempDir, "file"))
	require.NoError(t, err)
	f.Close()
	g, err := New(Config{TempDirs: []string{f.Name()}})
	require.Error(t, err)
	os.Remove(f.Name())

	g, _ = New(Config{})
	req, _ := http.NewRequest("GET", "invalid_host", nil)
	_, err = g.Do("key", req)
	require.Error(t, err)
	require.False(t, g.Has("key"))
}

func TestDo_NoSuppress(t *testing.T) {
	g, _ := New(Config{TempDirs: tempDirs, MaxMemory: 5})
	c := make(chan []byte, 1)
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b := <-c
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.Write(b)
	}))
	srvURL := srv.URL

	srv304 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(304)
	}))
	srv304URL := srv304.URL

	req, _ := http.NewRequest("GET", srv304URL, nil)
	resp, err := g.Do("key", req)
	require.NoError(t, err)
	require.Equal(t, 304, resp.StatusCode)
	b, _ := ioutil.ReadAll(resp.Body)
	require.EqualValues(t, "", string(b))
	resp.Body.Close()
	require.False(t, g.Has("key"))
	require.EqualValues(t, 1, atomic.LoadInt32(&calls))

	c <- []byte("hello")
	req, _ = http.NewRequest("GET", srvURL, nil)
	resp, err = g.Do("key", req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	b, _ = ioutil.ReadAll(resp.Body)
	require.EqualValues(t, "hello", string(b))
	resp.Body.Close()
	require.False(t, g.Has("key"))
	require.EqualValues(t, 2, atomic.LoadInt32(&calls))

	var tempFilename string
	g.CreateTempFile = func(dir, prefix string) (f *os.File, err error) {
		f, err = ioutil.TempFile(dir, prefix)
		if err == nil {
			tempFilename = f.Name()
		}
		return
	}
	c <- []byte("helloworld")
	resp, err = g.Do("key", req)
	require.NoError(t, err)
	b, _ = ioutil.ReadAll(resp.Body)
	require.EqualValues(t, "helloworld", string(b))
	resp.Body.Close()
	require.False(t, g.Has("key"))
	_, err = os.Stat(tempFilename)
	require.True(t, os.IsNotExist(err))
	require.EqualValues(t, 3, atomic.LoadInt32(&calls))

	cerr := errors.New("failed to create temp file")
	g.CreateTempFile = func(dir, prefix string) (f *os.File, err error) {
		return nil, cerr
	}
	c <- []byte("helloworld")
	_, err = g.Do("key", req)
	require.Equal(t, cerr, err)
	require.False(t, g.Has("key"))
	require.EqualValues(t, 4, atomic.LoadInt32(&calls))
}

func TestDo_SuppressSeq(t *testing.T) {
	g, _ := New(Config{TempDirs: tempDirs, MaxMemory: 5})
	c := make(chan []byte, 100)
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b := <-c
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.Write(b)
	}))
	srvURL := srv.URL

	c <- []byte("hello")
	req, _ := http.NewRequest("GET", srvURL, nil)
	resp, err := g.Do("key", req)
	require.NoError(t, err)
	b, _ := ioutil.ReadAll(resp.Body)
	require.EqualValues(t, "hello", string(b))
	require.True(t, g.Has("key"))
	require.EqualValues(t, 1, atomic.LoadInt32(&calls))

	nresp, err := g.Do("key", req)
	require.NoError(t, err)
	b, _ = ioutil.ReadAll(nresp.Body)
	require.EqualValues(t, "hello", string(b))
	nresp.Body.Close()
	require.True(t, g.Has("key"))
	require.EqualValues(t, 1, atomic.LoadInt32(&calls))

	nresp, err = g.Do("key", req)
	require.NoError(t, err)
	b, _ = ioutil.ReadAll(nresp.Body)
	require.EqualValues(t, "hello", string(b))
	resp.Body.Close()
	require.True(t, g.Has("key"))
	require.EqualValues(t, 1, atomic.LoadInt32(&calls))

	nresp.Body.Close()
	require.False(t, g.Has("key"))

	c <- []byte("hello")
	resp, err = g.Do("key", req)
	b, _ = ioutil.ReadAll(resp.Body)
	require.EqualValues(t, "hello", string(b))
	require.NoError(t, err)

	resp.Body.Close()
	require.EqualValues(t, 2, atomic.LoadInt32(&calls))

	cerr := errors.New("failed to create temp file")
	g.CreateTempFile = func(dir, prefix string) (f *os.File, err error) {
		return nil, cerr
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		c <- []byte("helloworld")
		_, err = g.Do("key", req)
		require.Equal(t, cerr, err)
	}()
	_, err = g.Do("key", req)
	require.Equal(t, cerr, err)
	require.False(t, g.Has("key"))
	require.EqualValues(t, 3, atomic.LoadInt32(&calls))
}

func TestDo_SuppressConcurrent(t *testing.T) {
	c := make(chan string)
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s := <-c
		atomic.AddInt32(&calls, 1)
		switch req.URL.Path {
		case "/good":
			w.Header().Set("Content-Length", strconv.Itoa(len(s)))
			w.Write([]byte(s))
		case "/unexpected_eof":
			w.Header().Set("Content-Length", strconv.Itoa(len(s)+1))
			w.Write([]byte(s))
		case "/code304":
			w.WriteHeader(304)
		}
		return
	}))
	srvURL := srv.URL

	concurrentFn := func(g *Group) {
		log.Infof("Conf: %+v", g.Config)

		atomic.StoreInt32(&calls, 0)
		const n = 10
		var wg sync.WaitGroup

		// trip error.
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				req, _ := http.NewRequest("GET", "invalidHost", nil)
				resp, err := g.Do("key", req)
				require.Error(t, err)
				require.Nil(t, resp)
				wg.Done()
			}()
		}
		time.Sleep(100 * time.Millisecond)
		wg.Wait()

		// code 200, body: helloworld.
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				req, _ := http.NewRequest("GET", srvURL+"/good", nil)
				resp, err := g.Do("key", req)
				require.NoError(t, err)
				require.Equal(t, 200, resp.StatusCode)
				defer resp.Body.Close()
				b, err := ioutil.ReadAll(resp.Body)
				require.EqualValues(t, "helloworld", string(b))
				wg.Done()
			}()
		}
		time.Sleep(100 * time.Millisecond)
		c <- "helloworld"
		wg.Wait()
		require.EqualValues(t, 1, atomic.LoadInt32(&calls))

		// code 200, no body.
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				req, _ := http.NewRequest("GET", srvURL+"/good", nil)
				resp, err := g.Do("key", req)
				require.NoError(t, err)
				require.Equal(t, 200, resp.StatusCode)
				defer resp.Body.Close()
				b, err := ioutil.ReadAll(resp.Body)
				require.EqualValues(t, "", string(b))
				wg.Done()
			}()
		}
		time.Sleep(100 * time.Millisecond)
		c <- ""
		wg.Wait()
		require.EqualValues(t, 2, atomic.LoadInt32(&calls))

		// code 304, no body.
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				req, _ := http.NewRequest("GET", srvURL+"/code304", nil)
				resp, err := g.Do("key", req)
				require.NoError(t, err)
				require.Equal(t, 304, resp.StatusCode)
				defer resp.Body.Close()
				b, err := ioutil.ReadAll(resp.Body)
				require.EqualValues(t, "", string(b))
				wg.Done()
			}()
		}
		time.Sleep(100 * time.Millisecond)
		c <- ""
		wg.Wait()
		require.EqualValues(t, 3, atomic.LoadInt32(&calls))

		// copy data error.
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				req, _ := http.NewRequest("GET", srvURL+"/unexpected_eof", nil)
				resp, err := g.Do("key", req)
				require.NoError(t, err)
				defer resp.Body.Close()
				b, err := ioutil.ReadAll(resp.Body)
				require.Equal(t, "helloworld", string(b))
				require.EqualValues(t, io.ErrUnexpectedEOF, err)
				wg.Done()
			}()
		}
		time.Sleep(100 * time.Millisecond)
		c <- "helloworld"
		wg.Wait()
		require.EqualValues(t, 4, atomic.LoadInt32(&calls))
	}

	// in mem.
	g1, _ := New(Config{MaxMemory: 10})
	concurrentFn(g1)

	// in file.
	g2, _ := New(Config{MaxMemory: 5, TempDirs: tempDirs})
	concurrentFn(g2)
}
