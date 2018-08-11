package main

import (
	"bufio"
	"errors"
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
		return nil, errors.New("fromHexString: invalid input length")
	}
	n := length / 2
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		a := byteFromHex(s[2*i])
		b := byteFromHex(s[2*i+1])
		if a > 15 || b > 15 {
			return nil, errors.New("fromHexString: invalid input char")
		}
		bytes[i] = (a << 4) | b
	}
	return bytes, nil
}

func fromBase64String(s base64) ([]byte, error) {
	// TODO Support unpadded input and check that padding only occurs at the end.
	length := len(s)
	if length%4 != 0 {
		return nil, errors.New("fromBase64String: invalid input length")
	}
	padding := 0
	switch {
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
			return nil, errors.New("fromBase64String: invalid input char")
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
		return nil, errors.New("xor: byte arrays have different length")
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = x[i] ^ y[i]
	}
	return buf, nil
}

func xorHex(x, y hex) (hex, error) {
	b1, err1 := fromHexString(x)
	b2, err2 := fromHexString(y)
	if err1 != nil || err2 != nil {
		return "", errors.New("xorHex: invalid input")
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

func main() {
}
