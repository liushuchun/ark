package urlfetch

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
)

var ErrExpired = errors.New("local file expired")

// --------------------------------------------------------------------

type LocalCache struct {
	Root       string
	ExpireTime time.Duration
}

var gTmpFid uint64

func NewLocalCache(root string, expireTime time.Duration) LocalCache {

	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	return LocalCache{root, expireTime}
}

func (r LocalCache) Get(key string) (file string, err error) {

	file = r.Root + key
	fi, err := os.Lstat(file)
	if err != nil {
		return
	}
	duration := time.Since(fi.ModTime())
	if r.ExpireTime != 0 && duration > r.ExpireTime {
		os.Remove(file)
		err = ErrExpired
		log.Infof("key: %s, %v", key, err)
	}
	return
}

func (r LocalCache) Set(key string, f io.Reader) (file string, err error) {

	fid := strconv.FormatUint(atomic.AddUint64(&gTmpFid, 1), 36)
	tmpfile := r.Root + "~" + fid

	w, err := os.Create(tmpfile)
	if err != nil {
		err = errors.Info(err, "urlfetch: os.Create failed", tmpfile).Detail(err)
		return
	}

	_, err = io.Copy(w, f)
	w.Close()
	if err != nil {
		os.Remove(tmpfile)
		err = errors.Info(err, "urlfetch: io.Copy failed").Detail(err)
		return
	}

	file = r.Root + key
	os.Remove(file)

	err = os.Rename(tmpfile, file)
	if err != nil {
		err = errors.Info(err, "urlfetch: os.Rename failed", tmpfile, file).Detail(err)
	}
	return
}

func (r LocalCache) Delete(key string) {

	os.Remove(r.Root + key)
}

// --------------------------------------------------------------------

type Client struct {
	Cache LocalCache
	*http.Client
}

func (r Client) Fetch(url string, ext string) (file string, err error) {

	h := md5.New()
	h.Write([]byte(url))
	md5val := h.Sum(nil)
	key := base64.URLEncoding.EncodeToString(md5val)[:22] + ext

	file, err = r.Cache.Get(key)
	if err == nil {
		return
	}

	if r.Client == nil {
		r.Client = http.DefaultClient
	}

	log.Info("urlfetch: http.Get", url)

	resp, err := r.Client.Get(url)
	if err != nil {
		err = errors.Info(err, "urlfetch: http.Get failed", url).Detail(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New("urlfetch: http.Get status -" + strconv.Itoa(resp.StatusCode))
		return
	}

	file, err = r.Cache.Set(key, resp.Body)
	return
}

// --------------------------------------------------------------------
