package helm

type Helm interface {
	Init() //初始化helm客户端, 检测helm二进制程序是否存在等等

}
