github.com/qiniu/ds/stringset.v1
======

stringset.Type is a set of string, implemented via `map[string]struct{}` for minimal memory consumption.

Here is an example:

```go
import (
	"github.com/qiniu/ds/stringset.v1"
)

set := stringset.New("a", "b")
set.Add("c")

if set.Has("d") {
	println("set has item `"d"`")
}
```
