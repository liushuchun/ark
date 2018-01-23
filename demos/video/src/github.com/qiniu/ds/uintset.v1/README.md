github.com/qiniu/ds/uintset.v1
======

uintset.Type is a set of uint, implemented via `map[uint]struct{}` for minimal memory consumption.

Here is an example:

```go
import (
	"github.com/qiniu/ds/uintset.v1"
)

set := uintset.New(1, 2)
set.Add(3)

if set.Has(4) {
	println("set has item `4`")
}
```
