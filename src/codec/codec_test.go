package codec

import "testing"

func TestMsgCodecZeroValue(t *testing.T) {
	var mc MsgCodec
	if mc.Offset != 0 {
		t.Errorf("default offset should be 0")
	}
}
