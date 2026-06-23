// messageHead.go 定义服务间消息帧的固定长度包头。
//
// 当前包头长度为 19 字节，采用小端序编码：
// 1 字节 Flag、1 字节 PieceFlag、1 字节 Cmd、8 字节 DataLength、8 字节 OnLineIdx。
package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// ServerMaxReceiveLength 限制单个消息体的最大长度，防止异常包占用过多内存。
	ServerMaxReceiveLength = 255 * 1024

	// MessageHeadLength 是服务间消息头的固定字节数。
	MessageHeadLength = 19
)

// ServerMessageHead 表示服务间传输帧的固定包头。
//
// Flag 和 PieceFlag 当前保留给协议标记或分片能力；Cmd 表示业务命令；
// DataLength 表示后续消息体长度；OnLineIdx 表示连接或在线索引。
type ServerMessageHead struct {
	Flag       byte
	PieceFlag  byte
	Cmd        uint8
	DataLength int
	OnLineIdx  int
}

// Decode 从 19 字节包头缓冲区中解析 ServerMessageHead。
//
// 调用方必须保证 buf 长度至少为 MessageHeadLength。
func (sh *ServerMessageHead) Decode(buf []byte) {
	sh.Flag = buf[0]
	sh.PieceFlag = buf[1]
	sh.Cmd = buf[2]
	sh.DataLength = int(binary.LittleEndian.Uint64(buf[3:11]))
	sh.OnLineIdx = int(binary.LittleEndian.Uint64(buf[11:19]))
}

// EncodeTo 将 ServerMessageHead 编码到调用方提供的缓冲区。
//
// 该方法会校验缓冲区长度、消息体长度和在线索引范围。
func (sh *ServerMessageHead) EncodeTo(buf []byte) error {
	if len(buf) < MessageHeadLength {
		return fmt.Errorf("message head buffer too short: %d", len(buf))
	}
	if sh.DataLength < 0 {
		return errors.New("negative data length is invalid")
	}
	if sh.DataLength > ServerMaxReceiveLength {
		return fmt.Errorf("data length %d exceeds max receive length %d", sh.DataLength, ServerMaxReceiveLength)
	}
	if sh.OnLineIdx < 0 {
		return errors.New("negative online index is invalid")
	}
	buf[0] = sh.Flag
	buf[1] = sh.PieceFlag
	buf[2] = sh.Cmd
	binary.LittleEndian.PutUint64(buf[3:11], uint64(sh.DataLength))
	binary.LittleEndian.PutUint64(buf[11:19], uint64(sh.OnLineIdx))
	return nil
}

// Encode 将 ServerMessageHead 编码到缓冲区，并忽略校验错误。
//
// 该方法保留给旧调用点兼容；新代码应优先使用 EncodeTo 获取错误。
func (sh *ServerMessageHead) Encode(buf []byte) {
	_ = sh.EncodeTo(buf)
}

// Check 校验包头中的长度字段是否在协议允许范围内。
//
// finished 目前固定表示校验完成；err 非空表示包头不可接受。
func (sh *ServerMessageHead) Check() (finished bool, err error) {
	if sh.DataLength < 0 {
		return false, errors.New("negative data length is invalid")
	}
	if sh.DataLength > ServerMaxReceiveLength {
		return false, fmt.Errorf("data length %d exceeds max receive length %d", sh.DataLength, ServerMaxReceiveLength)
	}
	return true, nil
}
