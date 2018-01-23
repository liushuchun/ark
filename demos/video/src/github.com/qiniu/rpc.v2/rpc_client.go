package rpc

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"qiniupkg.com/trace.v1"
)

var UserAgent = "Golang qiniu/rpc package"

// --------------------------------------------------------------------

type Client struct {
	*http.Client
}

var DefaultClient = Client{&http.Client{Transport: http.DefaultTransport}}

// --------------------------------------------------------------------

type Logger interface {
	ReqId() string
	Xput(logs []string)
}

// --------------------------------------------------------------------

func (r Client) DoRequest(
	l Logger, method, url string) (resp *http.Response, err error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}
	return r.Do(l, req)
}

func (r Client) DoRequestWith(
	l Logger, method, url1, bodyType string, body io.Reader, bodyLength int) (resp *http.Response, err error) {

	req, err := http.NewRequest(method, url1, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = int64(bodyLength)
	return r.Do(l, req)
}

func (r Client) DoRequestWith64(
	l Logger, method, url1, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {

	req, err := http.NewRequest(method, url1, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = bodyLength
	return r.Do(l, req)
}

func (r Client) DoRequestWithForm(
	l Logger, method, url1 string, data map[string][]string) (resp *http.Response, err error) {

	msg := url.Values(data).Encode()
	if method == "GET" || method == "HEAD" || method == "DELETE" {
		if strings.ContainsRune(url1, '?') {
			url1 += "&"
		} else {
			url1 += "?"
		}
		return r.DoRequest(l, method, url1+msg)
	}
	return r.DoRequestWith(
		l, method, url1, "application/x-www-form-urlencoded", strings.NewReader(msg), len(msg))
}

func (r Client) DoRequestWithJson(
	l Logger, method, url1 string, data interface{}) (resp *http.Response, err error) {

	msg, err := json.Marshal(data)
	if err != nil {
		return
	}
	return r.DoRequestWith(
		l, method, url1, "application/json", bytes.NewReader(msg), len(msg))
}

func (r Client) Do(l Logger, req *http.Request) (resp *http.Response, err error) {

	t := trace.SafeRecorder(l).Child().Client()
	t.Inject(req)

	e := trace.NewClientEvent(t, req)
	defer func() { t.FlattenKV("http", e.LogResponse(resp, err)).Finish() }()

	if l != nil {
		t.Tag("reqid", l.ReqId())
		req.Header.Set("X-Reqid", l.ReqId())
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", UserAgent)
	}
	resp, err = r.Client.Do(req)
	if err != nil {
		return
	}
	if l != nil {
		details := resp.Header["X-Log"]
		if len(details) > 0 {
			l.Xput(details)
		}
	}
	return
}

// --------------------------------------------------------------------

type ErrorInfo struct {
	Err     string   `json:"error"`
	Reqid   string   `json:"reqid"`
	Details []string `json:"details"`
	Code    int      `json:"code"`
}

func (r *ErrorInfo) ErrorDetail() string {
	msg, _ := json.Marshal(r)
	return string(msg)
}

func (r *ErrorInfo) Error() string {
	if r.Err != "" {
		return r.Err
	}
	return http.StatusText(r.Code)
}

func (r *ErrorInfo) HttpCode() int {
	return r.Code
}

// --------------------------------------------------

type httpCoder interface {
	HttpCode() int
}

func HttpCodeOf(err error) int {
	if hc, ok := err.(httpCoder); ok {
		return hc.HttpCode()
	}
	return 0
}

// --------------------------------------------------------------------

type errorRet struct {
	Error string `json:"error"`
}

func parseError(r io.Reader) (err string) {

	body, err1 := ioutil.ReadAll(r)
	if err1 != nil {
		return err1.Error()
	}

	m := make(map[string]interface{})
	json.Unmarshal(body, &m)
	if e, ok := m["error"]; ok {
		if err, ok = e.(string); ok {
			// qiniu error msg style returns here
			return
		}
	}
	return string(body)
}

func ResponseError(resp *http.Response) (err error) {

	e := &ErrorInfo{
		Details: resp.Header["X-Log"],
		Reqid:   resp.Header.Get("X-Reqid"),
		Code:    resp.StatusCode,
	}
	if resp.StatusCode > 299 {
		if resp.ContentLength != 0 {
			if ct := resp.Header.Get("Content-Type"); strings.TrimSpace(strings.SplitN(ct, ";", 2)[0]) == "application/json" {
				e.Err = parseError(resp.Body)
			}
		}
	}
	return e
}

func CallRet(l Logger, ret interface{}, resp *http.Response) (err error) {

	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = json.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				return
			}
		}
		if resp.StatusCode == 200 || resp.StatusCode == 204 {
			return nil
		}
	}
	return ResponseError(resp)
}

func (r Client) CallWithForm(
	l Logger, ret interface{}, method, url1 string, param map[string][]string) (err error) {

	resp, err := r.DoRequestWithForm(l, method, url1, param)
	if err != nil {
		return err
	}
	return CallRet(l, ret, resp)
}

func (r Client) CallWithJson(
	l Logger, ret interface{}, method, url1 string, param interface{}) (err error) {

	resp, err := r.DoRequestWithJson(l, method, url1, param)
	if err != nil {
		return err
	}
	return CallRet(l, ret, resp)
}

func (r Client) CallWith(
	l Logger, ret interface{}, method, url1, bodyType string, body io.Reader, bodyLength int) (err error) {

	resp, err := r.DoRequestWith(l, method, url1, bodyType, body, bodyLength)
	if err != nil {
		return err
	}
	return CallRet(l, ret, resp)
}

func (r Client) CallWith64(
	l Logger, ret interface{}, method, url1, bodyType string, body io.Reader, bodyLength int64) (err error) {

	resp, err := r.DoRequestWith64(l, method, url1, bodyType, body, bodyLength)
	if err != nil {
		return err
	}
	return CallRet(l, ret, resp)
}

func (r Client) Call(
	l Logger, ret interface{}, method, url1 string) (err error) {

	resp, err := r.DoRequestWith(l, method, url1, "application/x-www-form-urlencoded", nil, 0)
	if err != nil {
		return err
	}
	return CallRet(l, ret, resp)
}

// --------------------------------------------------------------------
