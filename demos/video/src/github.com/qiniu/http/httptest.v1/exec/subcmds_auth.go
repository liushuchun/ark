package exec

import (
	"net/http"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/http/httptest.v1"
)

// ---------------------------------------------------------------------------

type authTransportComposer struct {
	mac *digest.Mac
}

func (p authTransportComposer) Compose(base http.RoundTripper) http.RoundTripper {
	return digest.NewTransport(p.mac, base)
}

// ---------------------------------------------------------------------------

type qboxArgs struct {
	AK string `arg:"access-key"`
	SK string `arg:"secret-key"`
}

func (p *subContext) Eval_qbox(ctx *httptest.Context, args *qboxArgs) (httptest.TransportComposer, error) {

	mac := &digest.Mac{
		AccessKey: args.AK,
		SecretKey: []byte(args.SK),
	}
	return authTransportComposer{mac}, nil
}

// ---------------------------------------------------------------------------

