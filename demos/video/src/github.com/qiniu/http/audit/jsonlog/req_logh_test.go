package jsonlog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	qrpc "github.com/qiniu/rpc.v1"
	"qbox.us/net/httputil"
	"qbox.us/servestk"
	"qiniupkg.com/x/errors.v8"
)

type testlog struct {
}

func (l testlog) Log(msg []byte) error {
	return errors.New(string(msg))
}

func TestServeStack(t *testing.T) {

	logf := testlog{}
	var dec Decoder
	al := New("FOO", logf, dec, 512)
	ss := servestk.New(http.NewServeMux(), al.Handler)

	// test xBody works(no panic)
	ss.HandleFunc("/xbody", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("call /xbody")
		fmt.Println("content-length:", req.ContentLength)
		req.Body = ioutil.NopCloser(req.Body)
		httputil.Reply(w, 200, map[string]string{"foo": "bar"})
	})

	svr := httptest.NewServer(ss)
	svrUrl := svr.URL
	defer svr.Close()

	rpc := qrpc.Client{http.DefaultClient}
	req2Body := bytes.NewReader([]byte{1, 2})
	req2, err := http.NewRequest("POST", svrUrl+"/xbody", req2Body)
	if err != nil {
		fmt.Println(err)
	}
	req2.ContentLength = -1
	rpc.Do(nil, req2)
}
