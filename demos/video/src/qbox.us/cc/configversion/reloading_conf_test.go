package configversion

import (
	"errors"
	"testing"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/stretchr/testify.v1/require"
	"github.com/stretchr/testify/assert"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestReloading(t *testing.T) {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		t.Fatal(err)
	}
	//	defer session.Close()
	c := session.DB("test").C("test_configversion")
	c.RemoveAll(bson.M{})

	cfg1 := &ReloadingConfig{
		Id:       "haha",
		ReloadMs: 1000,
		C:        c,
	}
	cfg2 := &ReloadingConfig{
		Id:       "haha",
		ReloadMs: 1000,
		C:        c,
	}
	count := 0
	onReload := func(l rpc.Logger) (err error) {
		count++
		return
	}
	doNothing := func(l rpc.Logger) error { return nil }
	advance, err := StartReloading(cfg1, func(l rpc.Logger) (err error) { return onReload(l) })
	require.NoError(t, err)
	advance, err = StartReloading(cfg1, func(l rpc.Logger) (err error) { return onReload(l) })
	require.NoError(t, err)
	anotherAdvance, err := StartReloading(cfg2, doNothing)
	require.NoError(t, err)
	anotherAdvance, err = StartReloading(cfg2, doNothing)
	require.NoError(t, err)

	assert.Equal(t, 1, count)
	time.Sleep(.5e9)
	assert.Equal(t, 1, count)
	advance()
	time.Sleep(1e9)
	assert.Equal(t, 2, count)
	time.Sleep(1e9)
	assert.Equal(t, 2, count)
	advance()
	time.Sleep(1e9)
	assert.Equal(t, 3, count)

	anotherAdvance()
	time.Sleep(1e9)
	assert.Equal(t, 4, count)
	assert.Equal(t, 4, cfg1.ver)
	assert.Equal(t, 4, cfg2.ver)

	advance()
	onReload = func(l rpc.Logger) (err error) { return errors.New("fail") }
	time.Sleep(1e9)
	assert.Equal(t, 4, count)
	assert.Equal(t, 4, cfg1.ver)
	assert.Equal(t, 5, cfg2.ver)

	onReload = func(l rpc.Logger) (err error) {
		count++
		return
	}
	time.Sleep(1e9)
	assert.Equal(t, 5, count)
	assert.Equal(t, 5, cfg1.ver)
	assert.Equal(t, 5, cfg2.ver)
}
