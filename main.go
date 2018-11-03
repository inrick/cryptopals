package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

type (
	hex    string
	base64 string
)

const (
	hexAlphabet    = "0123456789abcdef"
	base64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

var (
	ErrInvalidHexLen  = errors.New("fromHexString: invalid input length")
	ErrInvalidHexChar = errors.New("fromHexString: invalid input char")
	ErrInvalidB64Len  = errors.New("fromBase64String: invalid input length")
	ErrInvalidB64Char = errors.New("fromBase64String: invalid input char")
	ErrDiffInputLen   = errors.New("different length of args")
)

// Entry point
func main() {
	challenge1_7()
}

func absFloat32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// Valid max output is 15, returns 255 on error.
func byteFromHex(x byte) byte {
	i := strings.IndexByte(hexAlphabet, x)
	if i == -1 {
		return 255
	}
	return byte(i)
}

// Valid max output is 63, returns 255 on error.
func byteFromBase64(x byte) byte {
	switch x {
	case '=':
		return 0
	default:
		i := strings.IndexByte(base64Alphabet, x)
		if i == -1 {
			return 255
		}
		return byte(i)
	}
}

func fromHexString(s hex) ([]byte, error) {
	length := len(s)
	if length%2 != 0 {
		return nil, ErrInvalidHexLen
	}
	n := length / 2
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		a := byteFromHex(s[2*i])
		b := byteFromHex(s[2*i+1])
		if a > 15 || b > 15 {
			return nil, ErrInvalidHexChar
		}
		bytes[i] = (a << 4) | b
	}
	return bytes, nil
}

func fromBase64String(s base64) ([]byte, error) {
	// TODO Support unpadded input and check that padding only occurs at the end.
	length := len(s)
	if length%4 != 0 {
		return nil, ErrInvalidB64Len
	}
	padding := 0
	switch {
	case length < 4:
		// do nothing
	case s[length-2] == '=':
		padding = 2
	case s[length-1] == '=':
		padding = 1
	}
	chunks := length / 4
	nbytes := 3 * chunks
	bytes := make([]byte, nbytes)
	for i := 0; i < chunks; i++ {
		a := byteFromBase64(s[4*i])
		b := byteFromBase64(s[4*i+1])
		c := byteFromBase64(s[4*i+2])
		d := byteFromBase64(s[4*i+3])
		if a > 63 || b > 63 || c > 63 || d > 63 {
			return nil, ErrInvalidB64Char
		}
		packed := (int32(a) << 18) | (int32(b) << 12) | (int32(c) << 6) | int32(d)
		bytes[3*i] = byte(packed >> 16)
		bytes[3*i+1] = byte(packed >> 8)
		bytes[3*i+2] = byte(packed)
	}
	return bytes[:nbytes-padding], nil // TODO returned slice will occupy padding
}

func toHexString(bytes []byte) hex {
	length := len(bytes)
	buf := make([]byte, 2*length)
	for i := 0; i < length; i++ {
		a := bytes[i] >> 4
		b := bytes[i] & 0xf
		buf[2*i] = hexAlphabet[a]
		buf[2*i+1] = hexAlphabet[b]
	}
	return hex(buf)
}

func toBase64String(bytes []byte) base64 {
	length := len(bytes)
	n := length / 3
	rem := length % 3
	extra := 0
	if rem > 0 {
		extra = 4
	}
	buf := make([]byte, 4*n+extra)
	for i := 0; i < n; i++ {
		a := int32(bytes[3*i]) << 16
		b := int32(bytes[3*i+1]) << 8
		c := int32(bytes[3*i+2])
		packed := a | b | c
		buf[4*i] = base64Alphabet[(packed>>18)&0x3f]
		buf[4*i+1] = base64Alphabet[(packed>>12)&0x3f]
		buf[4*i+2] = base64Alphabet[(packed>>6)&0x3f]
		buf[4*i+3] = base64Alphabet[packed&0x3f]
	}
	switch rem {
	case 0:
		break
	case 1:
		packed := int32(bytes[3*n]) << 16
		buf[4*n] = base64Alphabet[(packed>>18)&0x3f]
		buf[4*n+1] = base64Alphabet[(packed>>12)&0x3f]
		buf[4*n+2] = '='
		buf[4*n+3] = '='
	case 2:
		packed := (int32(bytes[3*n]) << 16) | (int32(bytes[3*n+1]) << 8)
		buf[4*n] = base64Alphabet[(packed>>18)&0x3f]
		buf[4*n+1] = base64Alphabet[(packed>>12)&0x3f]
		buf[4*n+2] = base64Alphabet[(packed>>6)&0x3f]
		buf[4*n+3] = '='
	default:
		panic("unreachable")
	}
	return base64(buf)
}

func xor(x, y []byte) ([]byte, error) {
	length := len(x)
	if length != len(y) {
		return nil, ErrDiffInputLen
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = x[i] ^ y[i]
	}
	return buf, nil
}

func xorHex(x, y hex) (hex, error) {
	b1, err := fromHexString(x)
	if err != nil {
		return "", err
	}
	b2, err := fromHexString(y)
	if err != nil {
		return "", err
	}
	xored, err := xor(b1, b2)
	if err != nil {
		return "", err
	}
	return toHexString(xored), nil
}

// Frequency table taken from Wikipedia
var letterFreq = map[byte]float32{
	'e': 0.12702,
	't': 0.09056,
	'a': 0.08167,
	'o': 0.07507,
	'i': 0.06966,
	'n': 0.06749,
	's': 0.06327,
	'h': 0.06094,
	'r': 0.05987,
	'd': 0.04253,
	'l': 0.04025,
	'c': 0.02782,
	'u': 0.02758,
	'm': 0.02406,
	'w': 0.02361,
	'f': 0.02228,
	'g': 0.02015,
	'y': 0.01974,
	'p': 0.01929,
	'b': 0.01492,
	'v': 0.00978,
	'k': 0.00772,
	'j': 0.00153,
	'x': 0.00150,
	'q': 0.00095,
	'z': 0.00074,
}

func byteFrequency(bytes []byte, c byte) float32 {
	count := 0
	for _, x := range bytes {
		if c == x {
			count++
		}
	}
	return float32(count) / float32(len(bytes))
}

func calcScore(bytes []byte) float32 {
	score := float32(0)
	for c, f1 := range letterFreq {
		f2 := byteFrequency(bytes, c)
		score += absFloat32(f2 - f1) // L1 norm
	}
	return score
}

func crackSingleCharXorHexs(hexs []hex) (string, error) {
	var bestMatch string
	var bestScore float32 = math.MaxFloat32
	for _, x := range hexs {
		bytes, err := fromHexString(x)
		if err != nil {
			return "", err
		}
		match, score, _ := crackSingleCharXor(bytes)
		if score < bestScore {
			bestMatch = string(match)
			bestScore = score
		}
	}
	return bestMatch, nil
}

func crackSingleCharXor(bytes []byte) (string, float32, byte) {
	var bestMatch string
	var bestScore float32 = math.MaxFloat32
	var bestKey byte
	buf := make([]byte, len(bytes))
	for i := 0; i < 256; i++ {
		for j := range buf {
			buf[j] = bytes[j] ^ byte(i)
		}
		score := calcScore(buf)
		if score < bestScore {
			bestMatch = string(buf)
			bestScore = score
			bestKey = byte(i)
		}
	}
	return bestMatch, bestScore, bestKey
}

func repeatingKeyXor(input, key []byte) []byte {
	keyLen := len(key)
	buf := make([]byte, len(input))
	for i, c := range input {
		buf[i] = c ^ key[i%keyLen]
	}
	return buf
}

func countOnes(bytes []byte) int {
	count := 0
	for _, b := range bytes {
		for x := b; x > 0; x &= x - 1 {
			count++
		}
	}
	return count
}

func hammingDistance(x, y []byte) (int, error) {
	xored, err := xor(x, y)
	if err != nil {
		return -1, err
	}
	return countOnes(xored), nil
}

func findRepeatingKeyXorCandidates(input []byte) [][]byte {
	type info struct {
		keySize int
		norm    float32
	}
	blocks := 4 // Number of blocks to average distance between
	norms := make([]info, 0)
	for keySize := 2; keySize < 40; keySize++ {
		var dist int
		for i := 0; i < blocks; i++ {
			slice1 := input[i*keySize : (i+1)*keySize]
			slice2 := input[(i+1)*keySize : (i+2)*keySize]
			hamming, err := hammingDistance(slice1, slice2)
			if err != nil {
				panic(err) // TODO
			}
			dist += hamming
		}
		norms = append(norms, info{keySize, float32(dist) / float32(blocks*keySize)})
	}
	sort.Slice(norms, func(i, j int) bool {
		return norms[i].norm < norms[j].norm
	})
	n := 3 // Number of candidates
	candidates := make([][]byte, n)
	for i, c := range norms[:n] {
		candidates[i] = crackKeyAssumingKeySize(input, c.keySize)
	}
	return candidates
}

func crackKeyAssumingKeySize(input []byte, keySize int) []byte {
	length := len(input)
	chunks := length / keySize
	// TODO Ignoring the remainder at the end. Maybe irrelevant anyway?
	transpose := make([][]byte, keySize)
	for i := 0; i < keySize; i++ {
		transpose[i] = make([]byte, chunks)
		for j := 0; j < chunks; j++ {
			transpose[i][j] = input[j*keySize+i]
		}
	}
	key := make([]byte, keySize)
	for i, chunk := range transpose {
		_, _, char := crackSingleCharXor(chunk)
		key[i] = char
	}
	return key
}

func readBase64File(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var input []byte
	for scanner.Scan() {
		bytes, err := fromBase64String(base64(scanner.Text()))
		if err != nil {
			return nil, err
		}
		input = append(input, bytes...)
	}
	return input, nil
}

// AES

var sbox = [256]byte{
	0x63, 0x7c, 0x77, 0x7b, 0xf2, 0x6b, 0x6f, 0xc5, 0x30, 0x01, 0x67, 0x2b, 0xfe, 0xd7, 0xab, 0x76,
	0xca, 0x82, 0xc9, 0x7d, 0xfa, 0x59, 0x47, 0xf0, 0xad, 0xd4, 0xa2, 0xaf, 0x9c, 0xa4, 0x72, 0xc0,
	0xb7, 0xfd, 0x93, 0x26, 0x36, 0x3f, 0xf7, 0xcc, 0x34, 0xa5, 0xe5, 0xf1, 0x71, 0xd8, 0x31, 0x15,
	0x04, 0xc7, 0x23, 0xc3, 0x18, 0x96, 0x05, 0x9a, 0x07, 0x12, 0x80, 0xe2, 0xeb, 0x27, 0xb2, 0x75,
	0x09, 0x83, 0x2c, 0x1a, 0x1b, 0x6e, 0x5a, 0xa0, 0x52, 0x3b, 0xd6, 0xb3, 0x29, 0xe3, 0x2f, 0x84,
	0x53, 0xd1, 0x00, 0xed, 0x20, 0xfc, 0xb1, 0x5b, 0x6a, 0xcb, 0xbe, 0x39, 0x4a, 0x4c, 0x58, 0xcf,
	0xd0, 0xef, 0xaa, 0xfb, 0x43, 0x4d, 0x33, 0x85, 0x45, 0xf9, 0x02, 0x7f, 0x50, 0x3c, 0x9f, 0xa8,
	0x51, 0xa3, 0x40, 0x8f, 0x92, 0x9d, 0x38, 0xf5, 0xbc, 0xb6, 0xda, 0x21, 0x10, 0xff, 0xf3, 0xd2,
	0xcd, 0x0c, 0x13, 0xec, 0x5f, 0x97, 0x44, 0x17, 0xc4, 0xa7, 0x7e, 0x3d, 0x64, 0x5d, 0x19, 0x73,
	0x60, 0x81, 0x4f, 0xdc, 0x22, 0x2a, 0x90, 0x88, 0x46, 0xee, 0xb8, 0x14, 0xde, 0x5e, 0x0b, 0xdb,
	0xe0, 0x32, 0x3a, 0x0a, 0x49, 0x06, 0x24, 0x5c, 0xc2, 0xd3, 0xac, 0x62, 0x91, 0x95, 0xe4, 0x79,
	0xe7, 0xc8, 0x37, 0x6d, 0x8d, 0xd5, 0x4e, 0xa9, 0x6c, 0x56, 0xf4, 0xea, 0x65, 0x7a, 0xae, 0x08,
	0xba, 0x78, 0x25, 0x2e, 0x1c, 0xa6, 0xb4, 0xc6, 0xe8, 0xdd, 0x74, 0x1f, 0x4b, 0xbd, 0x8b, 0x8a,
	0x70, 0x3e, 0xb5, 0x66, 0x48, 0x03, 0xf6, 0x0e, 0x61, 0x35, 0x57, 0xb9, 0x86, 0xc1, 0x1d, 0x9e,
	0xe1, 0xf8, 0x98, 0x11, 0x69, 0xd9, 0x8e, 0x94, 0x9b, 0x1e, 0x87, 0xe9, 0xce, 0x55, 0x28, 0xdf,
	0x8c, 0xa1, 0x89, 0x0d, 0xbf, 0xe6, 0x42, 0x68, 0x41, 0x99, 0x2d, 0x0f, 0xb0, 0x54, 0xbb, 0x16,
}

var sboxInv = [256]byte{
	0x52, 0x09, 0x6a, 0xd5, 0x30, 0x36, 0xa5, 0x38, 0xbf, 0x40, 0xa3, 0x9e, 0x81, 0xf3, 0xd7, 0xfb,
	0x7c, 0xe3, 0x39, 0x82, 0x9b, 0x2f, 0xff, 0x87, 0x34, 0x8e, 0x43, 0x44, 0xc4, 0xde, 0xe9, 0xcb,
	0x54, 0x7b, 0x94, 0x32, 0xa6, 0xc2, 0x23, 0x3d, 0xee, 0x4c, 0x95, 0x0b, 0x42, 0xfa, 0xc3, 0x4e,
	0x08, 0x2e, 0xa1, 0x66, 0x28, 0xd9, 0x24, 0xb2, 0x76, 0x5b, 0xa2, 0x49, 0x6d, 0x8b, 0xd1, 0x25,
	0x72, 0xf8, 0xf6, 0x64, 0x86, 0x68, 0x98, 0x16, 0xd4, 0xa4, 0x5c, 0xcc, 0x5d, 0x65, 0xb6, 0x92,
	0x6c, 0x70, 0x48, 0x50, 0xfd, 0xed, 0xb9, 0xda, 0x5e, 0x15, 0x46, 0x57, 0xa7, 0x8d, 0x9d, 0x84,
	0x90, 0xd8, 0xab, 0x00, 0x8c, 0xbc, 0xd3, 0x0a, 0xf7, 0xe4, 0x58, 0x05, 0xb8, 0xb3, 0x45, 0x06,
	0xd0, 0x2c, 0x1e, 0x8f, 0xca, 0x3f, 0x0f, 0x02, 0xc1, 0xaf, 0xbd, 0x03, 0x01, 0x13, 0x8a, 0x6b,
	0x3a, 0x91, 0x11, 0x41, 0x4f, 0x67, 0xdc, 0xea, 0x97, 0xf2, 0xcf, 0xce, 0xf0, 0xb4, 0xe6, 0x73,
	0x96, 0xac, 0x74, 0x22, 0xe7, 0xad, 0x35, 0x85, 0xe2, 0xf9, 0x37, 0xe8, 0x1c, 0x75, 0xdf, 0x6e,
	0x47, 0xf1, 0x1a, 0x71, 0x1d, 0x29, 0xc5, 0x89, 0x6f, 0xb7, 0x62, 0x0e, 0xaa, 0x18, 0xbe, 0x1b,
	0xfc, 0x56, 0x3e, 0x4b, 0xc6, 0xd2, 0x79, 0x20, 0x9a, 0xdb, 0xc0, 0xfe, 0x78, 0xcd, 0x5a, 0xf4,
	0x1f, 0xdd, 0xa8, 0x33, 0x88, 0x07, 0xc7, 0x31, 0xb1, 0x12, 0x10, 0x59, 0x27, 0x80, 0xec, 0x5f,
	0x60, 0x51, 0x7f, 0xa9, 0x19, 0xb5, 0x4a, 0x0d, 0x2d, 0xe5, 0x7a, 0x9f, 0x93, 0xc9, 0x9c, 0xef,
	0xa0, 0xe0, 0x3b, 0x4d, 0xae, 0x2a, 0xf5, 0xb0, 0xc8, 0xeb, 0xbb, 0x3c, 0x83, 0x53, 0x99, 0x61,
	0x17, 0x2b, 0x04, 0x7e, 0xba, 0x77, 0xd6, 0x26, 0xe1, 0x69, 0x14, 0x63, 0x55, 0x21, 0x0c, 0x7d,
}

func subBytes(state []byte) {
	for i := range state {
		state[i] = sbox[state[i]]
	}
}

func subBytesInv(state []byte) {
	for i := range state {
		state[i] = sboxInv[state[i]]
	}
}

func shiftRows(state []byte) {
	// Store old state by row in r
	var r [4][4]byte
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			r[i][j] = state[4*j+i]
		}
	}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state[4*j+i] = r[i][(j+i)%4]
		}
	}
}

func shiftRowsInv(state []byte) {
	// Store old state by row in r
	var r [4][4]byte
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			r[i][j] = state[4*j+i]
		}
	}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state[4*j+i] = r[i][(j+3*i)%4]
		}
	}
}

// t ↦ u
func mixColumns(state []byte) {
	var t, u [4][4]byte

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			t[i][j] = state[4*i+j]
		}
	}

	for j := 0; j < 4; j++ {
		u[j][0] = fieldMultX(t[j][0]) ^ fieldMult3(t[j][1]) ^ t[j][2] ^ t[j][3]
		u[j][1] = fieldMultX(t[j][1]) ^ fieldMult3(t[j][2]) ^ t[j][3] ^ t[j][0]
		u[j][2] = fieldMultX(t[j][2]) ^ fieldMult3(t[j][3]) ^ t[j][0] ^ t[j][1]
		u[j][3] = fieldMultX(t[j][3]) ^ fieldMult3(t[j][0]) ^ t[j][1] ^ t[j][2]
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state[4*i+j] = u[i][j]
		}
	}
}

// u ↦ t
func mixColumnsInv(state []byte) {
	var t, u [4][4]byte

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			u[i][j] = state[4*i+j]
		}
	}

	// Coefficients looked up on Wikipedia because of laziness
	for j := 0; j < 4; j++ {
		t[j][0] = fieldMult14(u[j][0]) ^ fieldMult11(u[j][1]) ^ fieldMult13(u[j][2]) ^ fieldMult9(u[j][3])
		t[j][1] = fieldMult14(u[j][1]) ^ fieldMult11(u[j][2]) ^ fieldMult13(u[j][3]) ^ fieldMult9(u[j][0])
		t[j][2] = fieldMult14(u[j][2]) ^ fieldMult11(u[j][3]) ^ fieldMult13(u[j][0]) ^ fieldMult9(u[j][1])
		t[j][3] = fieldMult14(u[j][3]) ^ fieldMult11(u[j][0]) ^ fieldMult13(u[j][1]) ^ fieldMult9(u[j][2])
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state[4*i+j] = t[i][j]
		}
	}
}

var rcon = [10]uint32{
	0x01000000, 0x02000000, 0x04000000, 0x08000000, 0x10000000,
	0x20000000, 0x40000000, 0x80000000, 0x1b000000, 0x36000000,
}

func subWord(w uint32) uint32 {
	return uint32(sbox[byte(w>>24)])<<24 |
		uint32(sbox[byte(w>>16)])<<16 |
		uint32(sbox[byte(w>>8)])<<8 |
		uint32(sbox[byte(w)])
}

func rotWord(w uint32) uint32 {
	return w<<8 | w>>24
}

func keyExpansion(key []byte) []uint32 {
	w := make([]uint32, 44)
	for i := 0; i < 4; i++ {
		w[i] = uint32(key[4*i])<<24 | uint32(key[4*i+1])<<16 | uint32(key[4*i+2])<<8 | uint32(key[4*i+3])
	}
	for i := 4; i < 44; i++ {
		tmp := w[i-1]
		if (i % 4) == 0 {
			tmp = subWord(rotWord(tmp)) ^ rcon[i/4-1]
		}
		w[i] = w[i-4] ^ tmp
	}
	return w
}

func addRoundKey(state []byte, n int, w []uint32) {
	key := w[4*n : 4*(n+1)]
	for i := 0; i < 4; i++ {
		state[4*i+0] ^= byte(key[i] >> 24)
		state[4*i+1] ^= byte(key[i] >> 16)
		state[4*i+2] ^= byte(key[i] >> 8)
		state[4*i+3] ^= byte(key[i])
	}
}

// Multiply an element in our field by x = 0b00000010.
// Remember: F = Z2[x]/(x^8 + x^4 + x^3 + x + 1),
// meaning x^8 = x^4 + x^3 + x + 1, hence the left shift and potential xor.
func fieldMultX(b byte) byte {
	b7 := (b >> 7) & 1
	mask := b7<<4 | b7<<3 | b7<<1 | b7
	return b<<1 ^ mask
}

// Multiply by x^n.
// Yeah, it's naive. Sue me.
func fieldMultXN(b byte, n int) byte {
	for i := 0; i < n; i++ {
		b = fieldMultX(b)
	}
	return b
}

// Multiply by (x+1) = 0b00000011
func fieldMult3(b byte) byte {
	return fieldMultX(b) ^ b
}

// x^3 + x^2 + x = 0b00001110
func fieldMult14(b byte) byte {
	return fieldMultXN(b, 3) ^ fieldMultXN(b, 2) ^ fieldMultXN(b, 1)
}

// x^3 + x^2 + 1 = 0b00001101
func fieldMult13(b byte) byte {
	return fieldMultXN(b, 3) ^ fieldMultXN(b, 2) ^ b
}

// x^3 + x + 1 = 0b00001011
func fieldMult11(b byte) byte {
	return fieldMultXN(b, 3) ^ fieldMultXN(b, 1) ^ b
}

// x^3 + 1 = 0b0001001
func fieldMult9(b byte) byte {
	return fieldMultXN(b, 3) ^ b
}

func aes128EncryptBlock(key, block []byte) []byte {
	// Elements are stored in column-major order
	state := make([]byte, 16)
	copy(state, block)
	w := keyExpansion(key)
	addRoundKey(state, 0, w)
	for i := 1; i < 10; i++ {
		subBytes(state)
		shiftRows(state)
		mixColumns(state)
		addRoundKey(state, i, w)
	}
	subBytes(state)
	shiftRows(state)
	addRoundKey(state, 10, w)
	return state
}

func aes128DecryptBlock(key, block []byte) []byte {
	// Elements are stored in column-major order
	state := make([]byte, 16)
	copy(state, block)
	w := keyExpansion(key)
	addRoundKey(state, 10, w)
	shiftRowsInv(state)
	subBytesInv(state)
	for i := 1; i < 10; i++ {
		addRoundKey(state, 10-i, w)
		mixColumnsInv(state)
		shiftRowsInv(state)
		subBytesInv(state)
	}
	addRoundKey(state, 0, w)
	return state
}

func aes128EcbEncrypt(key, input []byte) ([]byte, error) {
	if len(input)%16 != 0 {
		return nil, errors.New("aes128EcbEncrypt: input len must be multiple of 16")
	}
	out := make([]byte, len(input))
	for i := 0; i < len(input); i += 16 {
		copy(out[i:i+16], aes128EncryptBlock(key, input[i:i+16]))
	}
	return out, nil
}

func aes128EcbDecrypt(key, input []byte) ([]byte, error) {
	if len(input)%16 != 0 {
		return nil, errors.New("aes128EcbDecrypt: input len must be multiple of 16")
	}
	out := make([]byte, len(input))
	for i := 0; i < len(input); i += 16 {
		copy(out[i:i+16], aes128DecryptBlock(key, input[i:i+16]))
	}
	return out, nil
}

func challenge1_7() {
	input, err := readBase64File("data/7.txt")
	if err != nil {
		panic(err)
	}

	key := []byte("YELLOW SUBMARINE")
	decrypted, err := aes128EcbDecrypt(key, input)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(decrypted))
}
