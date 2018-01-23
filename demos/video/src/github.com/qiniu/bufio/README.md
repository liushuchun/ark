# 七牛 bufio

七牛 bufio 是 golang 官方的 bufio 的增强：

* 增加 Reader.Next() ([]byte, error)
* 增加 Reader.Underlayer() io.Reader // 暂无人使用
* 增加 Reader.BufferLen() int // 获得缓冲区大小
* 升级到 go1.1

