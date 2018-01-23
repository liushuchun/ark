package filelog

import (
	"testing"
	"time"
)

func TestFileName(t *testing.T) {
	d := time.Date(2014, 7, 7, 6, 16, 24, 0, time.Now().Location())

	{
		name := genFileName(now(d, 1))
		if name != "20140707061624" {
			t.Fatal("expected name:20140707061624, actual:", name)
		}
	}

	{
		name := genFileName(now(d, 60))
		if name != "20140707061600" {
			t.Fatal("expected name:20140707061600, actual:", name)
		}
	}

	{
		name := genFileName(now(d, 300))
		if name != "20140707061500" {
			t.Fatal("expected name:20140707061500, actual:", name)
		}
	}

	{
		name := genFileName(now(d, 600))
		if name != "20140707061000" {
			t.Fatal("expected name:20140707061000, actual:", name)
		}
	}

	{
		name := genFileName(now(d, 3600))
		if name != "20140707060000" {
			t.Fatal("expected name:20140707060000, actual:", name)
		}
	}

	{
		name := genFileName(now(d, 2*3600))
		if name != "20140707060000" {
			t.Fatal("expected name:20140707060000, actual:", name)
		}
	}

	{
		name := genFileName(now(d, 4*3600))
		if name != "20140707040000" {
			t.Fatal("expected name:20140707060000, actual:", name)
		}
	}

	{
		name := genFileName(now(d, 8*3600))
		if name != "20140707000000" {
			t.Fatal("expected name:20140707000000, actual:", name)
		}
	}

	{
		name := genFileName(now(d, 24*3600))
		if name != "20140707000000" {
			t.Fatal("expected name:20140707000000, actual:", name)
		}
	}
}
