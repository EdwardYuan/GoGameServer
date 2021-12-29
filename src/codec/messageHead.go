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
	//*in = (*in)[n:]
	return
}

func (in *inBuffer) read(begin, end int) (buf []byte, err error) {
	if begin*end <= 0 {
		return nil, errors.New("negative index")
	}
	if end <= begin {
		return nil, errors.New("end of buffer less than begin")
	}
	if end > len(*in) {
		return nil, errors.New("exceeding buffer length")
	}
	buf = (*in)[begin:end]
	return
}

func (sh *ServerMessageHead) Decode(buf []byte) {
	sh.Sign = binary.LittleEndian.Uint32(buf[:4])
	sh.PieceFlag = buf[4]
	sh.Flag = buf[5]
	sh.Cmd = buf[6]
	sh.DataLength = int(binary.LittleEndian.Uint64(buf[7:15]))
	sh.SocketHandle = int(binary.LittleEndian.Uint64(buf[15:23]))
	sh.OnLineIdx = int(binary.LittleEndian.Uint64(buf[23:31]))
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
