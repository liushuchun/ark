package trace

import (
	"testing"
)

func TestNewRootSpanID(t *testing.T) {
	id := NewRootSpanID()
	if id.Parent != 0 {
		t.Errorf("unexpected parent: %+v", id)
	}
	if id.Span == 0 {
		t.Errorf("zero Span: %+v", id)
	}
	if id.Trace == 0 {
		t.Errorf("zero root: %+v", id)
	}
	if id.Trace == id.Span {
		t.Errorf("duplicate IDs: %+v", id)
	}
}

func TestNewSpanID(t *testing.T) {
	root := NewRootSpanID()
	id := NewSpanID(root)
	if id.Parent != root.Span {
		t.Errorf("unexpected parent: %+v", id)
	}
	if id.Span == 0 {
		t.Errorf("zero Span: %+v", id)
	}
	if id.Trace != root.Trace {
		t.Errorf("mismatched root: %+v", id)
	}
}

func TestNewRootSpan(t *testing.T) {
	span := NewRootSpan()
	if !span.IsRoot() {
		t.Errorf("non root Span: %+v", span)
	}
	if span.Tag == nil {
		t.Errorf("Span use nil tag map")
	}
}

func TestNewSpan(t *testing.T) {
	span := NewRootSpan()
	span1 := NewSpan(span)
	if span1.IsRoot() {
		t.Errorf("root Span: %+v", span)
	}
	if span1.Tag == nil {
		t.Errorf("Span use nil tag map")
	}
}

func TestSpanContextToken(t *testing.T) {
	id := SpanID{
		spanID: spanID{
			Trace: 100,
			Span:  300,
		},
	}
	got := id.ContextToken()
	want := "64/12c|0"
	if got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestSpanContextTokenWithParent(t *testing.T) {
	id := SpanID{
		spanID: spanID{
			Trace:  100,
			Parent: 200,
			Span:   300,
		},
	}
	actual := id.ContextToken()
	expected := "64/12c/c8|0"
	if actual != expected {
		t.Errorf("Was %#v, but expected %#v", actual, expected)
	}
}

func TestParseContextToken(t *testing.T) {
	id, err := ParseContextToken("64/12c|1")
	if err != nil {
		t.Fatal(err)
	}
	if id.Trace != 100 || id.Span != 300 || !id.sampled {
		t.Errorf("unexpected ID: %+v", id)
	}
}

func TestParseContextTokenWithParent(t *testing.T) {
	id, err := ParseContextToken("64/12c/96|0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.Trace != 100 || id.Parent != 150 || id.Span != 300 || id.sampled {
		t.Errorf("unexpected event ID: %+v", id)
	}
}

func TestParseContextTokenMalformed(t *testing.T) {
	id, err := ParseContextToken(`6412c`)
	if id != nil {
		t.Errorf("unexpected ID: %+v", id)
	}
	if err != ErrBadSpanContext {
		t.Error(err)
	}
}

func TestParseContextTokenBadID(t *testing.T) {
	id, err := ParseContextToken("64/g00012c")
	if id != nil {
		t.Errorf("unexpected ID: %+v", id)
	}
	if err != ErrBadSpanContext {
		t.Error(err)
	}
}

func TestParseContextTokenBadParent(t *testing.T) {
	id, err := ParseContextToken("64/12c/g0096|1")
	if id != nil {
		t.Errorf("unexpected event ID: %+v", id)
	}
	if err != ErrBadSpanContext {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseContextTokenBadSampleFlag(t *testing.T) {
	id, err := ParseContextToken("64/12c/96")
	if id != nil {
		t.Errorf("unexpected event ID: %+v", id)
	}
	if err != ErrBadSpanContext {
		t.Errorf("unexpected error: %v", err)
	}
}

func BenchmarkNewRootSpanID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRootSpanID()
	}
}

func BenchmarkNewSpanID(b *testing.B) {
	root := NewRootSpanID()
	for i := 0; i < b.N; i++ {
		NewSpanID(root)
	}
}

func BenchmarkSpanContextToken(b *testing.B) {
	id := SpanID{
		spanID: spanID{
			Trace:  100,
			Parent: 200,
			Span:   300,
		},
	}
	for i := 0; i < b.N; i++ {
		id.ContextToken()
	}
}

func BenchmarkParseContextToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParseContextToken("64/12c")
		if err != nil {
			b.Fatal(err)
		}
	}
}
