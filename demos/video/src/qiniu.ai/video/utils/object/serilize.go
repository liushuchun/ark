package object

import (
	"bytes"
	"encoding/gob"
)

func Serilize(obj interface{}) (str string, err error) {
	buf := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buf)
	err = encoder.Encode(obj)
	if err != nil {
		return
	}
	str = buf.String()
	return
}

func Unserilize(str string, obj interface{}) (err error) {
	buf := bytes.NewBufferString(str)
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(obj)
	return
}
