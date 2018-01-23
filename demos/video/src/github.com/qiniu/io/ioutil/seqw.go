package ioutil

import (
	"io"
	"io/ioutil"
)

// ---------------------------------------------------

type seqWitem struct {
	w io.Writer
	n int64
}

type seqWriter struct {
	items []seqWitem
}

// w := SeqWriter(w1, n1, w2, n2, ...)
//
func SeqWriter(args ...interface{}) *seqWriter {

	n := len(args)
	if (n & 1) != 0 {
		panic("usage: SeqWriter(w1, n1, w2, n2, ...)")
	}

	items := make([]seqWitem, (n>>1)+1)

	for i := 0; i+1 < n; i += 2 {
		w1, ok1 := args[i].(io.Writer)
		n2, ok2 := args[i+1].(int64)
		if !ok2 {
			if n3, ok3 := args[i+1].(int); ok3 {
				n2, ok2 = int64(n3), true
			}
		}
		if !ok1 || !ok2 {
			panic("usage: SeqWriter(w1, n1, w2, n2, ...)")
		}
		items[i>>1] = seqWitem{w1, n2}
	}
	items[n>>1] = seqWitem{ioutil.Discard, (1<<63 - 1)}

	return &seqWriter{items: items}
}

func (sw *seqWriter) Write(buf []byte) (written int, err error) {

	for len(sw.items) > 0 {
		item := &sw.items[0]
		n1 := len(buf)
		if item.n < int64(n1) {
			n1 = int(item.n)
		}
		n1, err = item.w.Write(buf[:n1])
		written += n1
		item.n -= int64(n1)
		if item.n == 0 {
			sw.items = sw.items[1:]
		}
		if err != nil || n1 == len(buf) {
			return
		}
		buf = buf[n1:]
	}
	return
}

// ---------------------------------------------------
