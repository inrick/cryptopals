package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
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

func TestBase64FromHex(t *testing.T) {
	tests := []struct {
		input    hex
		expected base64
		experr   error
	}{
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
		input, expected, experr := test.input, test.expected, test.experr
		res, err := base64FromHex(input)
		if experr != err {
			t.Errorf("base64FromHex(%s) expected error '%v', got '%v'\n", input, experr, err)
		} else if res != expected {
			t.Errorf("base64FromHex(%s) expected '%v', got '%v'\n", input, expected, res)
		}
	}
}

func TestHexFromBase64(t *testing.T) {
	tests := []struct {
		input    base64
		expected hex
		experr   error
	}{
		{"TWFu", "4d616e", nil},
		{"SSdt", "49276d", nil},
		{"EA==", "10", nil},
		{"SSdtIA==", "49276d20", nil},
		{"TWF-", "", ErrInvalidB64Char},
		{"TWF", "", ErrInvalidB64Len},
	}
	for _, test := range tests {
		input, expected, experr := test.input, test.expected, test.experr
		res, err := hexFromBase64(input)
		if experr != err {
			t.Errorf("hexFromBase64(%s) expected error '%v', got '%v'\n", input, experr, err)
		} else if res != expected {
			t.Errorf("hexFromBase64(%s) expected '%v', got '%v'\n", input, expected, res)
		}
	}
}

func TestXorHex(t *testing.T) {
	tests := []struct {
		x1, x2   hex
		expected hex
		experr   error
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
		x1, x2, expected, experr := test.x1, test.x2, test.expected, test.experr
		res, err := xorHex(x1, x2)
		if experr != err {
			t.Errorf("xorHex(%s, %s) expected error '%v', got '%v'\n", x1, x2, experr, err)
		} else if res != expected {
			t.Errorf("xorHex(%s, %s) expected '%v', got '%v'\n", x1, x2, expected, res)
		}
	}
}

func TestChallenge1_3(t *testing.T) {
	input := []hex{"1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736"}
	res, err := crackSingleCharXorHexs(input)
	if err != nil || res != "Cooking MC's like a pound of bacon" {
		t.Fail()
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
	res, err := crackSingleCharXorHexs(hexs)
	if err != nil || res != "Now that the party is jumping\n" {
		t.Error("Challenge 1.4 failed")
	}
}

func TestChallenge1_5(t *testing.T) {
	key := []byte("ICE")
	input := []byte("Burning 'em, if you ain't quick and nimble\nI go crazy when I hear a cymbal")
	expected := hex("0b3637272a2b2e63622c2e69692a23693a2a3c6324202d623d63343c2a26226324272765272a282b2f20430a652e2c652a3124333a653e2b2027630c692b20283165286326302e27282f")
	res := toHexString(repeatingKeyXor(input, key))
	if res != expected {
		t.Error("Challenge 1.5 failed")
	}
}

func TestHammingDistance(t *testing.T) {
	x := []byte("this is a test")
	y := []byte("wokka wokka!!!")
	dist, err := hammingDistance(x, y)
	if err != nil || dist != 37 {
		t.Fail()
	}
}

func TestChallenge1_6(t *testing.T) {
	input, err := readBase64File("data/6.txt")
	if err != nil {
		t.Fail()
	}
	keys := findRepeatingKeyXorCandidates(input)
	results := make([][]byte, len(keys))
	for i, key := range keys {
		results[i] = repeatingKeyXor(input, key)
	}
	// Introspected data through debugger to find second entry
	key, res := keys[2], results[2]
	expkey := []byte("Terminator X: Bring the noise")
	expected, err := ioutil.ReadFile("data/expected1_6.txt")
	if err != nil {
		t.Fail()
	}
	if !bytes.Equal(res, expected) || !bytes.Equal(key, expkey) {
		t.Error("Challenge 1.6 failed")
	}
}
