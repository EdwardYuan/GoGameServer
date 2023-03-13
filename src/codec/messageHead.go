package codec

import (
	"encoding/binary"
	"errors"
)

const (
	ServerMaxReceiveLength = 255 * 1024
	MessageHeadLength      = 19
)

type ServerMessageHead struct {
	Flag       byte
	PieceFlag  byte
	Cmd        uint8
	DataLength int
	OnLineIdx  int
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
	// *in = (*in)[n:]
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

func (in *inBuffer) ShiftN(n int) {
	if n < 0 || n >= len(*in) {
		return
	}
	*in = (*in)[n:]
}

func (sh *ServerMessageHead) Decode(buf []byte) {
	sh.Flag = buf[0]
	sh.PieceFlag = buf[1]
	sh.Cmd = buf[2]
	sh.DataLength = int(binary.LittleEndian.Uint64(buf[3:11]))
	sh.OnLineIdx = int(binary.LittleEndian.Uint64(buf[11:19]))
}

func (sh *ServerMessageHead) Encode(buf []byte) {
	buf[0] = sh.Flag
	buf[1] = sh.PieceFlag
	buf[2] = sh.Cmd
	binary.BigEndian.PutUint32(buf[3:11], uint32(sh.DataLength))
	binary.BigEndian.PutUint32(buf[11:19], uint32(sh.OnLineIdx))
}

func (sh *ServerMessageHead) Check() (finished bool, err error) {
	return true, nil
}
