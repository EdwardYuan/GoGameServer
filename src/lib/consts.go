package lib

const (
	MaxReceiveBufCap = 255 * 1024              // 服务端接收包大小上限
	MinPieceBufSize  = MaxReceiveBufCap - 1024 // 最小分片大小
)
