package crypto

import (
	"encoding/base64"
	"testing"
)

func TestDes(t *testing.T) {
	key := []byte("sfe023f_")
	result, err := DesEncrypt([]byte("polaris@studygolang"), key)
	if err != nil {
		t.Error(err)
	}
	t.Log(base64.StdEncoding.EncodeToString(result))
	origData, err := DesDecrypt(result, key)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(origData))
}

func Test3Des(t *testing.T) {
	key := []byte("sfe023f_sefiel#fi32lf3e!")
	result, err := TripleDesEncrypt([]byte("polaris@studygol"), key)
	if err != nil {
		t.Error(err)
	}
	t.Log(base64.StdEncoding.EncodeToString(result))
	origData, err := TripleDesDecrypt(result, key)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(origData))
}
