package lbnsq

import (
	"bytes"
	"fmt"
)

func (c *Client) pub(host string, topic string, body []byte) (err error) {
	return c.nsqdCli.CallWith(nil,
		nil,
		fmt.Sprintf(host+"/pub?topic=%s", topic),
		"application/octet-stream",
		bytes.NewReader(body),
		len(body))
}
