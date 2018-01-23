webroute
=======================


自动路由
--------

webroute 简化了 Go 标准库的 net/http 模块路由。示意如下：

	package main

	import (
		"io"
		"net/http"
		"github.com/qiniu/webroute"
	)

	type Service struct {
	}

	func (r *Service) Do_(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /")
	}

	func (r *Service) DoFooBar(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /foo/bar")
	}

	func (r *Service) DoFooBar_(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /foo/bar/")
	}

	func main() {
		service := &Service{}
		webroute.ListenAndServe(":8080", service)
	}

这是一个最简单的 webroute 使用示意。规则是这样的：

* 所有路由处理函数以 Do 开头。
* 由大写字母分割分割路由的路径。比如 DoFooBar 代表 /foo/bar 这个路由 pattern；DoABC 代表 /a/b/c 这个路由 pattern。
* 如果路由处理函数以 _ 结尾，则表示一种通配路由。比如 DoFooBar_ 表示 /foo/bar/ 这个路由 pattern；而 Do_ 表示 / 这个路由 pattern。所以在没有其他规则的情况下，所有的请求都由 Do_ 接管。

理解以上规则后，你就非常轻松就写出 web service 了。


路由风格
--------

在 Web 服务的路由路径中，有时候会出现 - 分隔符。比如：

	/save-as?a=1&b=2

对于这种路由路径，我们其实默认已经支持，如下：

	func (r *Service) DoSave_as(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /save-as")
	}

	func (r *Service) DoSave_as_(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /save-as/")
	}

有时我们也会出现 _ 分隔符。比如：

	/save_as?a=2&b=345

为了支持这种风格的路由路径，我们需要将

	webroute.ListenAndServe(":8080", service)

改写为：

	router := webroute.Router{Style: '_', Mux: http.DefaultServeMux}
	router.Register(service)
	http.ListenAndServe(":8080", nil)

有时候我们我们也会采用大小写敏感的路由路径，比如：

	/saveAs?a=3&b=456
	/SaveAs?a=4&b=5678

对于这种风格的路由路径，只需要将 webroute.Router.Style 设为 '/' 而不是 '_' 就行了:

	package main

	import (
		"io"
		"net/http"
		"qbox.us/net/webroute"
	)

	type Service struct {
	}

	func (r *Service) Do_(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /")
	}

	func (r *Service) Do_foo_bar(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /foo/bar")
	}

	func (r *Service) Do_foo_bar_(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /foo/bar/")
	}

	func (r *Service) Do_saveAs(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /saveAs")
	}

	func (r *Service) Do_saveAs_(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "handle /saveAs/")
	}

	func main() {
		service := &Service{}
		router := webroute.Router{Style: '/', Mux: http.DefaultServeMux}
		router.Register(service)
		http.ListenAndServe(":8080", nil)
	}

这些风格相互排斥，只能选择一种。
