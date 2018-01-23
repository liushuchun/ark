package flag_test

import (
	"fmt"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/flag.v1"
	"github.com/qiniu/ts"
	"testing"
)

type imageViewArgs struct {
	Mode   int    `flag:"_"`
	W      int    `flag:"w"`
	H      int    `flag:"h"`
	Format string `flag:"format"`
}

type watermarkArgs struct {
	Mode     int    `flag:"_"`
	Image    string `flag:"image,base64,has"`
	Image2   []byte `flag:"image2,base64"`
	Image3   string `flag:"image3,base64,default"`
	Image4   string `flag:"image4,base64"`
	HasImage bool
}

func Test(t *testing.T) {

	var args imageViewArgs

	err := flag.Parse(&args, "imageView/2/w/200/h/300/format/jpg")
	if err != nil {
		ts.Fatal(t, "Parse failed:", err)
	}
	if args.Mode != 2 || args.W != 200 || args.H != 300 || args.Format != "jpg" {
		ts.Fatal(t, "Parse failed:", args)
	}

	err = flag.Parse(&args, "imageView/2/w/200/h/a300/format/jpg")
	if err == nil {
		ts.Fatal(t, "Parse failed:", err)
	}
	fmt.Println(errors.Detail(err))

	var wargs watermarkArgs

	err = flag.Parse(&wargs, "watermark/1/image/aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw")
	if err != nil || wargs.Mode != 1 || wargs.Image != "http://www.b1.qiniudn.com/images/logo-2.png" || !wargs.HasImage {
		ts.Fatal(t, "Parse failed:", wargs, err)
	}

	err = flag.Parse(&wargs, "watermark/2/image/aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw==")
	if err != nil || wargs.Mode != 2 || wargs.Image != "http://www.b1.qiniudn.com/images/logo-2.png" || !wargs.HasImage {
		ts.Fatal(t, "Parse failed:", wargs, err)
	}

	err = flag.Parse(&wargs, "watermark/2/image2/aHR0cDovL3d3dy5iMS5xaW5pdWRuLmNvbS9pbWFnZXMvbG9nby0yLnBuZw==")
	if err != nil || wargs.Mode != 2 || string(wargs.Image2) != "http://www.b1.qiniudn.com/images/logo-2.png" {
		ts.Fatal(t, "Parse failed:", wargs, err)
	}

	wargs.Image3 = "test"
	wargs.Image4 = "test"
	err = flag.Parse(&wargs, "watermark/2")
	if err != nil || wargs.Mode != 2 || wargs.Image != "" || wargs.HasImage ||
		wargs.Image2 != nil || wargs.Image3 != "test" || wargs.Image4 != "" {
		ts.Fatal(t, "Parse failed:", wargs, err)
	}
}
