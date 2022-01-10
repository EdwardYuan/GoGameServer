package lib

const (
	MaxReceiveBufCap = 255 * 1024              // 服务端接收包大小上限
	MinPieceBufSize  = MaxReceiveBufCap - 1024 // 最小分片大小

	MaxGameServerCount = 255 // 一个代理上连接的游戏服上限
)
