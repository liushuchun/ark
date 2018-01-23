package authutil

import (
	. "code.google.com/p/go.net/context"

	authp "qiniu.com/auth/proto.v1"
)

type key int

const (
	sudoerInfoKey key = 0
)

func NewContext(ctx Context, info *authp.SudoerInfo) Context {

	return WithValue(ctx, sudoerInfoKey, info)
}

func FromContext(ctx Context) (info *authp.SudoerInfo, ok bool) {

	info, ok = ctx.Value(sudoerInfoKey).(*authp.SudoerInfo)
	return
}
