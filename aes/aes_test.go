package aes

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestShiftRows(t *testing.T) {
	tests := []struct{ input, want []byte }{{
		[]byte{0, 4, 8, 12, 1, 5, 9, 13, 2, 6, 10, 14, 3, 7, 11, 15},
		[]byte{0, 5, 10, 15, 1, 6, 11, 12, 2, 7, 8, 13, 3, 4, 9, 14},
	}, {
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{0, 5, 10, 15, 4, 9, 14, 3, 8, 13, 2, 7, 12, 1, 6, 11},
	}}

	for _, test := range tests {
		input, want := test.input, test.want
		shiftRows(input)
		if !bytes.Equal(input, want) {
			t.Errorf("ShiftRows failed: got %v, want %v", input, want)
		}
	}
}

func TestShiftRowsInv(t *testing.T) {
	input := []byte{0, 4, 8, 12, 1, 5, 9, 13, 2, 6, 10, 14, 3, 7, 11, 15}
	want := []byte{0, 7, 10, 13, 1, 4, 11, 14, 2, 5, 8, 15, 3, 6, 9, 12}

	shiftRowsInv(input)
	if !bytes.Equal(input, want) {
		t.Errorf("ShiftRowsInv failed: got %v, want %v", input, want)
	}
}

func TestMixColumns(t *testing.T) {
	//input := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	//want := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	for i := 0; i < 128; i++ {
		input := make([]byte, 16)
		want := make([]byte, 16)
		rand.Read(input)
		copy(want, input)

		mixColumns(input) // = u
		if bytes.Equal(want, input) {
			t.Errorf("MixColumns failed: MixColumns(%v) didn't do anything", input)
		}
		mixColumnsInv(input) // = got
		if !bytes.Equal(want, input) {
			t.Errorf("MixColumns failed: wanted MixColumnsInv(MixColumns(%v)) = %v, got %v", want, want, input)
		}
	}
}

func TestAes128(t *testing.T) {
	key := []byte("YELLOW SUBMARINE")
	txt := []byte("I'm back and I'm")
	enc := EncryptBlock128(key, txt)
	dec := DecryptBlock128(key, enc)
	if !bytes.Equal(dec, txt) {
		t.Errorf("AES128 failed: key=%d, enc=%d, dec=%d", key, enc, dec)
	}
}
