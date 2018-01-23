package trace

import (
	"reflect"
	"testing"
	"time"
)

func TestFlattenBools(t *testing.T) {
	type T struct {
		Bool  bool `trace:"bool"`
		Uname bool
	}
	e := T{
		Bool:  true,
		Uname: true,
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"bool": "true",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenStrings(t *testing.T) {
	type T struct {
		Value string `trace:"value"`
	}
	e := T{
		Value: "bar",
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"value": "bar",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenNamedValues(t *testing.T) {
	type T struct {
		Value string `trace:"foo"`
	}
	e := T{
		Value: "bar",
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"foo": "bar",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenTime(t *testing.T) {
	type T struct {
		Value time.Time `trace:"time"`
	}
	ti := time.Date(2014, 5, 16, 12, 28, 38, 400, time.UTC)
	e := T{
		Value: ti,
	}

	got := make(map[string]time.Time)
	flattenValue("", reflect.ValueOf(e), nil, func(t time.Time, v string) {
		got[v] = t
	})

	want := map[string]time.Time{
		"time": ti,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenTimeUname(t *testing.T) {
	type T struct {
		Value time.Time
	}
	ti := time.Date(2014, 5, 16, 12, 28, 38, 400, time.UTC)
	e := T{
		Value: ti,
	}

	got := make(map[string]time.Time)
	flattenValue("", reflect.ValueOf(e), nil, func(t time.Time, v string) {
		got[v] = t
	})

	want := map[string]time.Time{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenFloats(t *testing.T) {
	type T struct {
		A float32 `trace:"a"`
		B float64 `trace:"b"`
	}
	e := T{
		A: 3,
		B: 500.3,
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"a": "3",
		"b": "500.3",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenInts(t *testing.T) {
	type T struct {
		A int8  `trace:"a"`
		B int16 `trace:"b"`
		C int32 `trace:"c"`
		D int64 `trace:"d"`
		E int   `trace:"e"`
	}
	e := T{
		A: 1,
		B: 2,
		C: 3,
		D: 4,
		E: 5,
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
		"e": "5",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenUints(t *testing.T) {
	type T struct {
		A uint8  `trace:"a"`
		B uint16 `trace:"b"`
		C uint32 `trace:"c"`
		D uint64 `trace:"d"`
		E uint   `trace:"e"`
	}
	e := T{
		A: 1,
		B: 2,
		C: 3,
		D: 4,
		E: 5,
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
		"e": "5",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenMaps(t *testing.T) {
	type T struct {
		Value map[string]int `trace:"value"`
	}
	e := T{
		Value: map[string]int{
			"one": 1,
			"two": 2,
		},
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"value.one": "1",
		"value.two": "2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenSlices(t *testing.T) {
	type T struct {
		Value []int `trace:"value"`
	}
	e := T{
		Value: []int{1, 2, 3},
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"value.0": "1",
		"value.1": "2",
		"value.2": "3",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenArrays(t *testing.T) {
	type T struct {
		Value [3]int `trace:"value"`
	}
	e := T{
		Value: [3]int{1, 2, 3},
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"value.0": "1",
		"value.1": "2",
		"value.2": "3",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

type stringer byte

func (stringer) String() string {
	return "stringer"
}

func TestFlattenStringers(t *testing.T) {
	type T struct {
		Value stringer `trace:"value"`
	}
	e := T{
		Value: 30,
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"value": "stringer",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenArbitraryTypes(t *testing.T) {
	type T struct {
		Value complex64 `trace:"value"`
	}
	e := T{
		Value: complex(17, 4),
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"value": "(17+4i)",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenUnexportedFields(t *testing.T) {
	type T struct {
		value string
	}
	e := T{
		value: "bar",
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenDuration(t *testing.T) {
	type T struct {
		Value time.Duration `trace:"value"`
	}
	e := T{
		Value: 500 * time.Microsecond,
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"value": "0.5",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFlattenPointers(t *testing.T) {
	type T struct {
		S *string `trace:"s"`
		I *int    `trace:"i"`
	}

	s := "bar"
	i := 7
	e := T{
		S: &s,
		I: &i,
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"s": "bar",
		"i": "7",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

type testInnerEvent struct {
	Days  map[string]int `trace:"days"`
	Other []bool         `trace:"other"`
}

type testEvent struct {
	Name   string         `trace:"name"`
	Age    int            `trace:"age"`
	Inner  testInnerEvent `trace:"inner"`
	Weight float64        `trace:"weight"`
	Count  uint           `trace:"count"`
	turds  *byte          `trace:"turds"`
}

func BenchmarkFlatten(b *testing.B) {
	e := testEvent{
		Name: "hello",
		Age:  400,
		Inner: testInnerEvent{
			Days: map[string]int{
				"Sunday": 1,
			},
			Other: []bool{true, false},
		},
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		flattenValue("", reflect.ValueOf(e), func(k, v string) {}, nil)
	}
}

func TestFlattenOmitempty(t *testing.T) {
	type T struct {
		A   float32 `trace:"a,omitempty"`
		B   float64 `trace:"b"`
		CCC int     `trace:",omitempty"`
		DDD int     `trace:",omitempty"`
	}
	e := T{
		B:   500.3,
		CCC: 120,
	}

	got := make(map[string]string)
	flattenValue("", reflect.ValueOf(e), func(k, v string) {
		got[k] = v
	}, nil)

	want := map[string]string{
		"b":   "500.3",
		"ccc": "120",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}
