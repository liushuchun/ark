package rs

import (
	"encoding/json"
	gourl "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/api/url"
)

// ----------------------------------------------------------

type GetPolicy struct {
	Expires uint32
}

func (r GetPolicy) MakeRequest(baseUrl string, mac *digest.Mac) (privateUrl string) {

	expires := r.Expires
	if expires == 0 {
		expires = 3600
	}
	deadline := time.Now().Unix() + int64(expires)

	if strings.Contains(baseUrl, "?") {
		baseUrl += "&e="
	} else {
		baseUrl += "?e="
	}
	baseUrl += strconv.FormatInt(deadline, 10)

	token := digest.Sign(mac, []byte(baseUrl))
	return baseUrl + "&token=" + token
}

func (r GetPolicy) MakeRequestUrl(baseUrl *gourl.URL, mac *digest.Mac) (privateUrl *gourl.URL) {

	copy := *baseUrl
	privateUrl = &copy

	expires := r.Expires
	if expires == 0 {
		expires = 3600
	}
	deadline := time.Now().Unix() + int64(expires)

	if privateUrl.RawQuery != "" {
		privateUrl.RawQuery += "&e="
	} else {
		privateUrl.RawQuery = "e="
	}
	privateUrl.RawQuery += strconv.FormatInt(deadline, 10)

	token := digest.Sign(mac, []byte(privateUrl.String()))
	privateUrl.RawQuery += "&token=" + token
	return
}

func MakeBaseUrl(domain, key string) (baseUrl string) {
	return "http://" + domain + "/" + url.Escape(key)
}

// --------------------------------------------------------------------------------

type PutPolicy struct {
	Scope               string   `json:"scope"`
	Expires             uint32   `json:"deadline"`             // 截止时间（以秒为单位）
	InsertOnly          uint16   `json:"exclusive,omitempty"`  // 若非0, 即使Scope为 Bucket:Key 的形式也是insert only
	DetectMime          uint16   `json:"detectMime,omitempty"` // 若非0, 则服务端根据内容自动确定 MimeType
	FsizeLimit          int64    `json:"fsizeLimit,omitempty"`
	SaveKey             string   `json:"saveKey,omitempty"`
	CallbackUrl         string   `json:"callbackUrl,omitempty"`
	CallbackBody        string   `json:"callbackBody,omitempty"`
	ReturnUrl           string   `json:"returnUrl,omitempty"`
	ReturnBody          string   `json:"returnBody,omitempty"`
	PersistentOps       string   `json:"persistentOps,omitempty"`
	PersistentNotifyUrl string   `json:"persistentNotifyUrl,omitempty"`
	PersistentPipeline  string   `json:"persistentPipeline,omitempty"`
	EndUser             string   `json:"endUser,omitempty"`
	DeleteAfterDays     int      `json:"deleteAfterDays,omitempty"`
	FileType            FileType `json:"fileType,omitempty"`
	NotifyQueue         string   `json:"notifyQueue,omitempty"`
	NotifyMessage       string   `json:"notifyMessage,omitempty"`
	NotifyMessageType   string   `json:"notifyMessageType,omitempty"`
}

func (r *PutPolicy) Token(mac *digest.Mac) string {

	var rr = *r
	if rr.Expires == 0 {
		rr.Expires = 3600
	}
	rr.Expires += uint32(time.Now().Unix())
	b, _ := json.Marshal(&rr)
	return digest.SignWithData(mac, b)
}

// ----------------------------------------------------------
