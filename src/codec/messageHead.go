package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
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

func (sh *ServerMessageHead) Encode(buf []byte) {
	_ = sh.EncodeTo(buf)
}

func (sh *ServerMessageHead) Check() (finished bool, err error) {
	if sh.DataLength < 0 {
		return false, errors.New("negative data length is invalid")
	}
	if sh.DataLength > ServerMaxReceiveLength {
		return false, fmt.Errorf("data length %d exceeds max receive length %d", sh.DataLength, ServerMaxReceiveLength)
	}
	return true, nil
}
