package brpc

import (
	"bytes"
	"github.com/qiniu/rpc.v1"
	"io/ioutil"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

// --------------------------------------------------------------------

func bsonDecode(v interface{}, resp *http.Response) error {

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return bson.Unmarshal(b, v)
}

func callRet(l rpc.Logger, ret interface{}, resp *http.Response) (err error) {

	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = bsonDecode(ret, resp)
			if err != nil {
				return
			}
		}
		if resp.StatusCode == 200 {
			return nil
		}
	}
	return rpc.ResponseError(resp)
}

// --------------------------------------------------------------------

type Client struct {
	*http.Client
}

var DefaultClient = Client{http.DefaultClient}

// --------------------------------------------------------------------

func (r Client) PostWithBson(l rpc.Logger, url1 string, data interface{}) (resp *http.Response, err error) {

	msg, err := bson.Marshal(data)
	if err != nil {
		return
	}
	return rpc.Client{r.Client}.PostWith(l, url1, "application/bson", bytes.NewReader(msg), len(msg))
}

func (r Client) CallWithForm(l rpc.Logger, ret interface{}, url1 string, param map[string][]string) (err error) {

	resp, err := rpc.Client{r.Client}.PostWithForm(l, url1, param)
	if err != nil {
		return err
	}
	return callRet(l, ret, resp)
}

func (r Client) CallWithBson(l rpc.Logger, ret interface{}, url1 string, param interface{}) (err error) {

	resp, err := r.PostWithBson(l, url1, param)
	if err != nil {
		return err
	}
	return callRet(l, ret, resp)
}

func (r Client) Call(l rpc.Logger, ret interface{}, url1 string) (err error) {

	resp, err := rpc.Client{r.Client}.PostWith(l, url1, "application/x-www-form-urlencoded", nil, 0)
	if err != nil {
		return err
	}
	return callRet(l, ret, resp)
}

// --------------------------------------------------------------------
