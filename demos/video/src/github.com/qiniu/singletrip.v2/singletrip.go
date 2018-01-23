package singletrip

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
)

type Config struct {
	MaxMemory   int      `json:"max_memory"`
	ReadTimeout int      `json:"read_timeout"`
	TempDirs    []string `json:"temp_dirs"`
}

type Group struct {
	mu    sync.Mutex
	calls map[string]*call

	Transport      http.RoundTripper
	CreateTempFile func(dir, prefix string) (*os.File, error)

	tempDirIdx uint64

	Config
	readTimeoutDur time.Duration
}

type call struct {
	wg    sync.WaitGroup
	resp  *http.Response
	err   error
	nproc int64
	body  buffer
}

func New(conf Config) (*Group, error) {
	if conf.MaxMemory == 0 {
		conf.MaxMemory = 4 << 20 // 4M.
	}
	if conf.ReadTimeout == 0 {
		conf.ReadTimeout = 20 // 20s.
	}
	if len(conf.TempDirs) == 0 {
		conf.TempDirs = []string{os.TempDir()}
	}
	for _, tempDir := range conf.TempDirs {
		if err := os.MkdirAll(tempDir, 0700); err != nil {
			return nil, err
		}
		cleanTempFiles(tempDir)
	}

	g := &Group{
		calls:          make(map[string]*call),
		Transport:      http.DefaultTransport,
		CreateTempFile: ioutil.TempFile,
		Config:         conf,
		readTimeoutDur: time.Duration(conf.ReadTimeout) * time.Second,
	}
	return g, nil
}

func cleanTempFiles(dir string) {
	f, err := os.Open(dir)
	if err != nil {
		log.Warnf("Failed to open dir: %s, err: %v", dir, err)
		return
	}
	defer f.Close()

	fis, err := f.Readdir(-1)
	if err != nil {
		log.Warnf("Failed to readdir: %s, err: %v", dir, err)
		return
	}
	for _, fi := range fis {
		if !fi.IsDir() && strings.HasPrefix(fi.Name(), "singletrip") {
			os.Remove(filepath.Join(dir, fi.Name()))
		}
	}
	return
}

func (g *Group) createTempFile(xl *xlog.Logger, prefix string) (f *os.File, err error) {
	for i := 0; i < len(g.TempDirs); i++ {
		idx := atomic.AddUint64(&g.tempDirIdx, 1) % uint64(len(g.TempDirs))
		dir := g.TempDirs[idx]
		f, err = g.CreateTempFile(dir, prefix)
		if err == nil {
			return
		}
		xl.Warnf("Failed to create temp file, i: %d, err: %v", i, err)
	}
	return
}

func (g *Group) Has(key string) bool {
	g.mu.Lock()
	_, ok := g.calls[key]
	g.mu.Unlock()
	return ok
}

func (g *Group) Do(key string, req *http.Request) (*http.Response, error) {
	xl := xlog.NewWithReq(req)

	g.mu.Lock()
	if c, ok := g.calls[key]; ok {
		c.nproc++
		g.mu.Unlock()
		c.wg.Wait()

		if c.err != nil {
			xl.Warnf("Failed to wait round trip, err: %v", c.err)
			return nil, c.err
		}
		xl.Debugf("Wait done in singletrip, key: %s", key)
		return g.newCachedResp(key, c), nil
	}

	c := new(call)
	c.nproc++
	c.wg.Add(1)
	g.calls[key] = c

	g.mu.Unlock()

	c.resp, c.err = g.Transport.RoundTrip(req)
	if c.err != nil {
		xl.Warnf("Failed to round trip, err: %v", c.err)
		g.mu.Lock()
		delete(g.calls, key)
		g.mu.Unlock()
		c.wg.Done()
		return nil, c.err
	}

	resp := c.resp

	// 缓存响应体。
	var buf buffer
	cl := resp.ContentLength
	if 0 <= cl && cl <= int64(g.MaxMemory) {
		buf = &byteBuffer{}
	} else {
		f, err := g.createTempFile(xl, "singletrip")
		if err == nil {
			buf = &fileBuffer{f: f}
		} else {
			xl.Warnf("Failed to create cache file, err: %v", err)
			c.err = err
			g.mu.Lock()
			delete(g.calls, key)
			g.mu.Unlock()
			c.wg.Done()
			resp.Body.Close()
			return nil, err
		}
	}
	c.body = buf

	go func() {
		defer resp.Body.Close()
		n, err := io.Copy(buf, resp.Body)
		if err != nil {
			xl.Errorf("Failed to copy to singletrip body buffer, n: %d/%d, err: %v", n, resp.ContentLength, err)
		}
		buf.FinishWrite(err)
	}()

	c.wg.Done()

	return g.newCachedResp(key, c), nil
}

func (g *Group) newCachedResp(key string, c *call) *http.Response {
	brClose := func() {
		g.mu.Lock()
		c.nproc--
		if c.nproc == 0 {
			delete(g.calls, key)
			c.body.Close()
		}
		g.mu.Unlock()
	}
	br := &bufferReader{
		buf:         c.body,
		readTimeout: g.readTimeoutDur,
		closeFn:     brClose,
	}

	nresp := cloneResp(c.resp)
	nresp.Body = br
	return nresp
}

func cloneResp(resp *http.Response) *http.Response {
	nresp := new(http.Response)
	*nresp = *resp
	nresp.Header = cloneHeader(resp.Header)
	return nresp
}

func cloneHeader(src http.Header) http.Header {
	dst := http.Header{}
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
	return dst
}
