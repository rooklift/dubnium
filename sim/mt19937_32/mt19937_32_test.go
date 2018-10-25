package mt19937_32

import (
	"testing"
)

func TestUint32(t *testing.T) {

	var z uint32

	for n := 0; n < 10000; n++ {
		z = Uint32()
	}

	if z != 4123659995 {
		t.Errorf("Expected 4123659995 after 10000 calls on unseeded RNG, got %v\n", z)
	}
}
