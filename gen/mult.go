// +build generate

package main

import "fmt"

func main() {
	output("mult2", fieldMultX)
	output("mult3", fieldMult3)
	output("mult9", fieldMult9)
	output("mult11", fieldMult11)
	output("mult13", fieldMult13)
	output("mult14", fieldMult14)
}

func output(name string, fn func(byte) byte) {
	fmt.Printf("var %s = [256]byte{\n", name)
	for i := 0; i < 16; i++ {
		fmt.Print(" ")
		for j := 0; j < 16; j++ {
			b := byte(16*i + j)
			fmt.Printf(" %#02x,", fn(b))
		}
		fmt.Println()
	}
	fmt.Println("}")
	fmt.Println()
}

// Multiply an element in our field by x = 0b00000010.
// Remember: F = Z2[x]/(x^8 + x^4 + x^3 + x + 1),
// meaning x^8 = x^4 + x^3 + x + 1, hence the left shift and potential xor.
func fieldMultX(b byte) byte {
	b7 := (b >> 7) & 1
	mask := b7<<4 | b7<<3 | b7<<1 | b7
	return b<<1 ^ mask
}

// Naively multiply by x^n.
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
