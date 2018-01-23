package trace

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testDir = "/tmp/trace-test"
var testPath = "/tmp/trace-test/collector-dir"

func TestFileCollector(t *testing.T) {
	err := os.RemoveAll(testDir)
	assert.Equal(t, nil, err, "clean dir fail")

	fc, err := NewFileCollector(&FileCollectorConfig{
		LogDir:    testPath,
		ChunkBits: 10,
	})
	assert.Equal(t, nil, err, "new file collector fail")

	err = fc.Collect(&Span{
		SpanID: NewRootSpanID(),
	})
	assert.Equal(t, nil, err, "collect fail")

	err = fc.Close()
	assert.Equal(t, nil, err, "collect close fail")

	fi, err := os.Stat(testPath + "/0")
	assert.Equal(t, nil, err, "stat log file fail")
	assert.NotEmpty(t, fi.Size(), "log file size not right")
}

var testServiceName = "trace-test-demo"

func TestServiceCollector(t *testing.T) {

	err := os.RemoveAll(testDir)
	assert.Equal(t, nil, err, "clean dir fail")

	DefaultCollectRoot = testDir

	sc, err := NewServiceCollector(testDir, testServiceName)
	assert.Equal(t, nil, err, "new service collector fail")

	err = sc.Collect(&Span{
		SpanID: NewRootSpanID(),
	})
	assert.Equal(t, nil, err, "collect fail")

	err = sc.Close()
	assert.Equal(t, nil, err, "collect close fail")

	fi, err := os.Stat(getLogPath(testDir, testServiceName) + "/0")
	assert.Equal(t, nil, err, "stat log file fail")
	assert.NotEmpty(t, fi.Size(), "log file size not right")
}

func TestServiceCollectorCleanHistory(t *testing.T) {

	err := os.RemoveAll(testDir)
	assert.Equal(t, nil, err, "clean dir fail")

	// create fake log history dir
	history1 := path.Join(testDir, testServiceName+".12345")
	err = os.MkdirAll(history1, os.ModePerm)
	assert.Equal(t, nil, err, "mkdir fail")

	history2 := path.Join(testDir, "hahahaha.12345")
	err = os.MkdirAll(history2, os.ModePerm)
	assert.Equal(t, nil, err, "mkdir fail")

	history3 := path.Join(testDir, ".12345")
	err = os.MkdirAll(history3, os.ModePerm)
	assert.Equal(t, nil, err, "mkdir fail")

	// start collector
	DefaultCollectRoot = testDir
	historyCleanDuration = time.Nanosecond

	sc, err := NewServiceCollector(testDir, testServiceName)
	assert.Equal(t, nil, err, "new service collector fail")

	err = sc.Collect(&Span{
		SpanID: NewRootSpanID(),
	})
	assert.Equal(t, nil, err, "collect fail")

	err = sc.Close()
	assert.Equal(t, nil, err, "collect close fail")

	fi, err := os.Stat(getLogPath(testDir, testServiceName) + "/0")
	assert.Equal(t, nil, err, "stat log file fail")
	assert.NotEmpty(t, fi.Size(), "log file size not right")

	// check history
	_, err = os.Stat(history1)
	assert.Equal(t, true, os.IsNotExist(err), "history still exist: %v", err)

	_, err = os.Stat(history2)
	assert.Equal(t, nil, err, "history not exist")

	_, err = os.Stat(history3)
	assert.Equal(t, nil, err, "history not exist")
}
