package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"
)

func askYesNo(showQuestion func(), defaultValue *bool) bool {
	for {
		showQuestion()

		if defaultValue != nil {
			if *defaultValue {
				fmt.Print("(Y/n): ")
			} else {
				fmt.Print("(y/N): ")
			}
		} else {
			fmt.Print("(y/n): ")
		}

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		c := scanner.Text()

		switch strings.ToLower(c) {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			if defaultValue != nil {
				return *defaultValue
			}
		}
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomSecurePassword() string {
	b := make([]byte, 6)
	size := big.NewInt(int64(len(letterBytes)))
	for i := range b {
		c, _ := rand.Int(rand.Reader, size)
		b[i] = letterBytes[c.Int64()]
	}
	return string(b)
}
