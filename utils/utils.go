package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
	"os"
	"os/signal"
	"syscall"
)

// TrapSignal traps SIGINT and SIGTERM and terminates the server correctly.
func TrapSignal(cleanupFunc func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		if cleanupFunc != nil {
			cleanupFunc()
		}
		exitCode := 128
		switch sig {
		case syscall.SIGINT:
			exitCode += int(syscall.SIGINT)
		case syscall.SIGTERM:
			exitCode += int(syscall.SIGTERM)
		}
		os.Exit(exitCode)
	}()
}

// Character sets.
const (
	Digits       = "0123456789"
	Alphabets    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AlphaNumeric = Digits + Alphabets
)

// GenerateRandomCode generates a random string of given length using the provided charset.
// If charset is empty, Digits will be used by default.
// Returns an error if cryptographic randomness fails.
func GenerateRandomCode(length uint8, charset string) (string, error) {
	if length == 0 {
		return "", errors.New("length must be greater than zero")
	}

	if charset == "" {
		charset = Digits
	}

	max := big.NewInt(int64(len(charset)))
	code := make([]byte, length)

	for i := range code {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err // propagate error to caller
		}
		code[i] = charset[num.Int64()]
	}

	return string(code), nil
}
