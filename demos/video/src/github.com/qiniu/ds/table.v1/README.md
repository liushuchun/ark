github.com/qiniu/ds/table.v1
======================

这是一个内存中的 Table 实现。相比于 Dictionary（在 Go 里面也就是 map），Table 的特点是有多个索引。

这个 Table 的基础规则如下：

* 首先行数据（row）是强 schema 的，类似于 MySQL，每行有哪些 field 非常明确，甚至数据类型（row type）也有明确的要求。
* 查询条件（query）也是强 schema，需要预先明确哪些是 unique 索引，哪些是普通 index；每个查询条件的类型（query type）同样也有明确要求。

### 实例

首先第一步我们定义 row schema：

```go
type Row struct {
	Id   int
	Pid  int
	Name string
	File string
}
```

然后，我们再定义所有可能的 query schema：

```go
type ById struct {
	Id int
}

type ByPid struct {
	Pid int
}

type ByPidName struct {
	Pid  int
	Name string
}
```

有了这些，我们就可以开始创建 Table：

```go
import (
	"github.com/qiniu/ds/table.v1"
)

cfg := table.NewConfig(new(Row)).
	WithUniques(new(ById), new(ByPidName)).
	WithIndexes(new(ByPid))

coll := table.New(cfg)
```

接着插入数据：

```go
err := coll.Insert(
	&Row{Id: 1, Pid: 0, Name: "abc.doc", File: "hello"},
	&Row{Id: 2, Pid: 0, Name: "ttt"},
	&Row{Id: 3, Pid: 2, Name: "1.txt", File: "qiniu"},
)
```

旁白：好吧，你看出来了，这是定义了一个目录树，根目录下有 abc.doc 和 ttt 目录；然后 ttt 目录下有个 1.txt 文件。

然后就可以查询：

```go
var rows []*Row
err = coll.FindAll(&rows, ByPid{0})
if err != nil {
	log.Fatal("Find by pid=0 failed:", err)
}
```

这是查询根目录下有哪些东西，显然我们将得到：

```go
&Row{Id: 1, Pid: 0, Name: "abc.doc", File: "hello"}
&Row{Id: 2, Pid: 0, Name: "ttt"}
```

如果是 unique 索引，查询结果应该是唯一的，此时用 FindOne。例如：

```go
var row *Row
err = coll.FindOne(&row, &ByPidName{0, "abc.doc"})
if err != nil {
	log.Fatal("Find by pid=0, name=abc.doc failed:", err)
}
```

这将得到：

```go
&Row{Id: 1, Pid: 0, Name: "abc.doc", File: "hello"}
```

最后，你还可以根据条件删除一些行：

```go
coll.RemoveAll(ById{2})
coll.RemoveAll(ByPid{1})
```

详细样例参考：

* [github.com/qiniu/ds/table.v1/table_test.go](https://github.com/qbox/base/blob/develop/qiniu/src/github.com/qiniu/ds/table.v1/table_test.go)
