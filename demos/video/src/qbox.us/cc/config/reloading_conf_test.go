package config

import (
	"github.com/qiniu/rpc.v1"
	// "fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	g_confName = "testconf.qboxtest"
	g_lockName = "testconf.lock.qboxtest"
)

func TestReloading(t *testing.T) {
	os.Remove(g_confName)
	defer os.Remove(g_confName)
	os.Remove(g_lockName)
	defer os.Remove(g_lockName)

	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		w.Write([]byte("1111"))
	}))
	defer ts1.Close()
	md5sum1 := calcMd5sum([]byte("1111"))

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		w.Write([]byte("222"))
	}))
	defer ts2.Close()
	md5sum2 := calcMd5sum([]byte("222"))

	cfg := &ReloadingConfig{
		ConfName:   g_confName,
		RemoteLock: g_lockName,
		ReloadMs:   1000,
		RemoteURL:  ts1.URL,
	}

	var realMd5 []byte
	onReload := func(l rpc.Logger, data []byte) (err error) {
		// fmt.Println(confName)
		realMd5 = calcMd5sum(data)
		return
	}

	err := StartReloading(cfg, onReload)
	if err != nil {
		t.Fatal(err)
	}

	//check

	// 1. starting
	time.Sleep(.5e9)
	b, err := ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum1)
	assert.Equal(t, cfg.md5sum, md5sum1)
	assert.Equal(t, realMd5, md5sum1)

	// 2. After 1 second, nothing happened because the remote config is not modified.
	time.Sleep(1e9)
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum1)
	assert.Equal(t, cfg.md5sum, md5sum1)
	assert.Equal(t, realMd5, md5sum1)

	// 3. change the remote url. After 1 second, the local config file is changed.
	cfg.RemoteURL = ts2.URL
	time.Sleep(1e9)
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum2)
	assert.Equal(t, cfg.md5sum, md5sum2)
	assert.Equal(t, realMd5, md5sum2)

	// 4. close the remote connection
	ts2.Close()
	time.Sleep(1e9)
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum2)
	assert.Equal(t, cfg.md5sum, md5sum2)
	assert.Equal(t, realMd5, md5sum2)

	ts2 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		w.Write([]byte("222"))
	}))

	// 5. change the remote url again. After 1 second, the local config file is changed.
	cfg.RemoteURL = ts1.URL
	time.Sleep(1e9)
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum1)
	assert.Equal(t, cfg.md5sum, md5sum1)
	assert.Equal(t, realMd5, md5sum1)

	// 6. change the remote url again, but set the lock file. After 1 second, nothing should happen.
	cfg.RemoteURL = ts2.URL
	os.Create(g_lockName)
	time.Sleep(1e9)
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum1)
	assert.Equal(t, cfg.md5sum, md5sum1)
	assert.Equal(t, realMd5, md5sum1)

	// 7. change the local file. The realMd5 should change following that because the lock file.
	ioutil.WriteFile(g_confName, []byte("hello world"), 666)
	md5sum3 := calcMd5sum([]byte("hello world"))
	time.Sleep(1e9)
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum3)
	assert.Equal(t, cfg.md5sum, md5sum3)
	assert.Equal(t, realMd5, md5sum3)

	// 8. remove the lock file. the local file should be the same as the remote file.
	os.Remove(g_lockName)
	time.Sleep(1e9)
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum2)
	assert.Equal(t, cfg.md5sum, md5sum2)
	assert.Equal(t, realMd5, md5sum2)

	// 9. start to force remote reload but fail
	ts1.Close()
	cfg.RemoteURL = ts1.URL
	err = StartReloading(cfg, onReload)
	if err != nil {
		t.Fatal(err)
	}
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum2)
	assert.Equal(t, cfg.md5sum, md5sum2)
	assert.Equal(t, realMd5, md5sum2)

	// 10. start to force remote reload
	ts1 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		w.Write([]byte("1111"))
	}))
	defer ts1.Close()
	cfg.RemoteURL = ts1.URL
	err = StartReloading(cfg, onReload)
	if err != nil {
		t.Fatal(err)
	}
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum1)
	assert.Equal(t, cfg.md5sum, md5sum1)
	assert.Equal(t, realMd5, md5sum1)

	// 11. start on lock
	os.Create(g_lockName)
	ioutil.WriteFile(g_confName, []byte("222"), 0666)
	cfg.RemoteURL = ts1.URL
	err = StartReloading(cfg, onReload)
	if err != nil {
		t.Fatal(err)
	}
	b, err = ioutil.ReadFile(g_confName)
	assert.Nil(t, err)
	assert.Equal(t, calcMd5sum(b), md5sum2)
	assert.Equal(t, cfg.md5sum, md5sum2)
	assert.Equal(t, realMd5, md5sum2)
}
