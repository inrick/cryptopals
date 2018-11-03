package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

func base64FromHex(x hex) (base64, error) {
	bytes, err := fromHexString(x)
	if err != nil {
		return "", err
	}
	return toBase64String(bytes), nil
}

func hexFromBase64(x base64) (hex, error) {
	bytes, err := fromBase64String(x)
	if err != nil {
		return "", err
	}
	return toHexString(bytes), nil
}

func TestHexAndB64Inverses(t *testing.T) {
	for i := 0; i < 1000; i++ {
		l := rand.Intn(1024)
		bs := make([]byte, l)
		for j := range bs {
			bs[j] = byte(rand.Intn(256))
		}
		hex := toHexString(bs)
		b64 := toBase64String(bs)

		b1, err := fromHexString(hex)
		if err != nil {
			t.Errorf("Invalid hex: '%v'", hex)
		}
		b2, err := fromBase64String(b64)
		if err != nil {
			t.Errorf("Invalid b64: '%v'", b64)
		}
		if !bytes.Equal(b1, bs) {
			t.Errorf("Invalid hex conversion: want '%v', got '%v'", bs, b1)
		}
		if !bytes.Equal(b2, bs) {
			t.Errorf("Invalid base64 conversion: want '%v', got '%v'", bs, b2)
		}
	}
}

func TestBase64FromHex(t *testing.T) {
	tests := []struct {
		input   hex
		want    base64
		wanterr error
	}{
		{"", "", nil},
		{"4d616e", "TWFu", nil},
		{"49276d", "SSdt", nil},
		{"11xcv6", "", ErrInvalidHexChar},
		{"11c", "", ErrInvalidHexLen},
		{"10", "EA==", nil},
		{"49276d20", "SSdtIA==", nil},
		// Challenge 1.1
		{"49276d206b696c6c696e6720796f757220627261696e206c696b65206120706f69736f6e6f7573206d757368726f6f6d",
			"SSdtIGtpbGxpbmcgeW91ciBicmFpbiBsaWtlIGEgcG9pc29ub3VzIG11c2hyb29t",
			nil},
	}
	for _, test := range tests {
		input, want, wanterr := test.input, test.want, test.wanterr
		got, err := base64FromHex(input)
		if wanterr != err {
			t.Errorf("base64FromHex(%q) = '%v', want '%v'\n", input, err, wanterr)
		} else if got != want {
			t.Errorf("base64FromHex(%q) = '%v', want '%v'\n", input, got, want)
		}
	}
}

func TestHexFromBase64(t *testing.T) {
	tests := []struct {
		input   base64
		want    hex
		wanterr error
	}{
		{"", "", nil},
		{"TWFu", "4d616e", nil},
		{"SSdt", "49276d", nil},
		{"EA==", "10", nil},
		{"SSdtIA==", "49276d20", nil},
		{"TWF-", "", ErrInvalidB64Char},
		{"TWF", "", ErrInvalidB64Len},
	}
	for _, test := range tests {
		input, want, wanterr := test.input, test.want, test.wanterr
		got, err := hexFromBase64(input)
		if err != wanterr {
			t.Errorf("hexFromBase64(%q) = '%v', want '%v'\n", input, err, wanterr)
		} else if got != want {
			t.Errorf("hexFromBase64(%q) = '%v', want '%v'\n", input, got, want)
		}
	}
}

func TestXorHex(t *testing.T) {
	tests := []struct {
		x1, x2  hex
		want    hex
		wanterr error
	}{
		// Uneven input 1
		{"1c0111001f010100061a024b53535009181c",
			"686974207468652062756c6c27732065796", "", ErrInvalidHexLen},
		// Uneven input 2
		{"1c0111001f010100061a024b53535009181",
			"686974207468652062756c6c277320657965", "", ErrInvalidHexLen},
		// Illegal input
		{"1345", "what", "", ErrInvalidHexChar},
		// Different input length
		{"1c01", "686974", "", ErrDiffInputLen},
		// Challenge 1.2
		{"1c0111001f010100061a024b53535009181c",
			"686974207468652062756c6c277320657965",
			"746865206b696420646f6e277420706c6179", nil},
	}
	for _, test := range tests {
		x1, x2, want, wanterr := test.x1, test.x2, test.want, test.wanterr
		got, err := xorHex(x1, x2)
		if err != wanterr {
			t.Errorf("xorHex(%s, %s) = '%v', want '%v'\n", x1, x2, err, wanterr)
		} else if got != want {
			t.Errorf("xorHex(%s, %s) = '%v', want '%v'\n", x1, x2, got, want)
		}
	}
}

func TestChallenge1_3(t *testing.T) {
	input := []hex{"1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736"}
	got, err := crackSingleCharXorHexs(input)
	if err != nil || got != "Cooking MC's like a pound of bacon" {
		t.Error()
	}
}

func TestChallenge1_4(t *testing.T) {
	file, err := os.Open("data/4.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	hexs := make([]hex, 0, 100)
	for scanner.Scan() {
		hexs = append(hexs, hex(scanner.Text()))
	}
	got, err := crackSingleCharXorHexs(hexs)
	if err != nil || got != "Now that the party is jumping\n" {
		t.Error("Challenge 1.4 failed")
	}
}

func TestChallenge1_5(t *testing.T) {
	key := []byte("ICE")
	input := []byte("Burning 'em, if you ain't quick and nimble\nI go crazy when I hear a cymbal")
	want := hex("0b3637272a2b2e63622c2e69692a23693a2a3c6324202d623d63343c2a26226324272765272a282b2f20430a652e2c652a3124333a653e2b2027630c692b20283165286326302e27282f")
	got := toHexString(repeatingKeyXor(input, key))
	if got != want {
		t.Error("Challenge 1.5 failed")
	}
}

func TestHammingDistance(t *testing.T) {
	x := []byte("this is a test")
	y := []byte("wokka wokka!!!")
	dist, err := hammingDistance(x, y)
	if err != nil || dist != 37 {
		t.Error()
	}
}

func TestChallenge1_6(t *testing.T) {
	input, err := readBase64File("data/6.txt")
	if err != nil {
		t.Fatal(err)
	}
	keys := findRepeatingKeyXorCandidates(input)
	results := make([][]byte, len(keys))
	for i, key := range keys {
		results[i] = repeatingKeyXor(input, key)
	}
	// Introspected data through debugger to find second entry
	key, got := keys[2], results[2]
	wantkey := []byte("Terminator X: Bring the noise")
	want, err := ioutil.ReadFile("data/want1_6.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) || !bytes.Equal(key, wantkey) {
		t.Error("Challenge 1.6 failed")
	}
}

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
	enc := aes128EncryptBlock(key, txt)
	dec := aes128DecryptBlock(key, enc)
	if !bytes.Equal(dec, txt) {
		t.Errorf("AES128 failed: key=%d, enc=%d, dec=%d", key, enc, dec)
	}
}
