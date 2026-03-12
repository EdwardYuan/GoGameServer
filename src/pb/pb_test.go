package pb

import "testing"

func TestInternalCmd(t *testing.T) {
	if InternalGateToProxy != 0 {
		t.Errorf("expected zero")
	}
}
