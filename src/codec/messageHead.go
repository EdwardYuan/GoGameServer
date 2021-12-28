package codec

import (
	"encoding/binary"
	"errors"
)

const (
	ServerMaxReceiveLength = 255 * 1024
	MessageHeadLength      = 32
)

type ServerMessageHead struct {
	Sign         uint32
	PieceFlag    byte
	Flag         byte
	Cmd          uint8
	DataLength   int
	SocketHandle int // 兼容 不使用
	OnLineIdx    int
}

type inBuffer []byte

func (in *inBuffer) readN(n int) (buf []byte, err error) {
	if n == 0 {
		return nil, nil
	}

	if n < 0 {
		return nil, errors.New("negative length is invalid")
	} else if n > len(*in) {
		return nil, errors.New("exceeding buffer length")
	}
	buf = (*in)[:n]
	*in = (*in)[n:]
	return
}

func (sh *ServerMessageHead) Decode(buf []byte) {
	sh.Sign = binary.LittleEndian.Uint32(buf[:4])
	sh.PieceFlag = buf[5]
	sh.Flag = buf[6]
	sh.Cmd = buf[7]
	sh.DataLength = int(binary.LittleEndian.Uint64(buf[8:16]))
	sh.SocketHandle = int(binary.LittleEndian.Uint64(buf[16:24]))
	sh.OnLineIdx = int(binary.LittleEndian.Uint64(buf[24:32]))
}

func (sh *ServerMessageHead) Encode(buf []byte) {
	binary.LittleEndian.PutUint32(buf[:4], sh.Sign)
	buf[5] = sh.PieceFlag
	buf[6] = sh.Flag
	buf[7] = sh.Cmd
	binary.BigEndian.PutUint32(buf[8:12], uint32(sh.DataLength))
	binary.BigEndian.PutUint32(buf[12:16], uint32(sh.OnLineIdx))
}

func (sh *ServerMessageHead) Check() (finished bool, err error) {
	return true, nil
}
