package testsuite

import (
	"encoding/hex"
	"math/rand"
	"slices"
	"testing"
	"time"
	"unsafe"
)

// TestSuite provides a set of helper functions for testing purposes.
// All the random values are generated based on a logged seed.
// By using a pre-generated seed, it is possible to reproduce failed tests
// by re-evaluating all the random values. This helps in identifying and debugging
// failures in testing conditions.
type TestSuite struct {
	Seed int64
	Rand *rand.Rand
}

func GenerateSeed() int64 {
	return time.Now().UTC().UnixNano()
}

// NewTestSuiteFromSeed creates a new TestSuite with the given seed.
func NewTestSuiteFromSeed(t *testing.T, seed int64) *TestSuite {
	t.Helper()

	return &TestSuite{
		Seed: seed,
		//nolint:gosec // to reproduce the failed tests
		Rand: rand.New(rand.NewSource(seed)),
	}
}

// NewTestSuite creates a new TestSuite by generating new seed.
func NewTestSuite(t *testing.T) *TestSuite {
	t.Helper()

	seed := GenerateSeed()
	t.Logf("%v seed is %v", t.Name(), seed)

	return NewTestSuiteFromSeed(t, seed)
}

// Integer is a constraint that matches any integer type.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// RandOptions holds configuration for random number generation.
type RandOptions[T Integer] struct {
	Min T // minimum value (inclusive)
	Max T // maximum value (exclusive)
}

// defaultRandOptions returns default options for any integer type.
func defaultRandOptions[T Integer]() RandOptions[T] {
	var zero T
	var max T

	// Detect signed vs unsigned by comparison
	if zero-1 < zero {
		// signed
		bits := uint(unsafe.Sizeof(zero)) * 8
		max = T(1<<(bits-1) - 1)
	} else {
		// unsigned
		max = ^T(0)
	}

	return RandOptions[T]{
		Min: zero, // 0 for all types
		Max: max,  // safe default max value for type
	}
}

// RandOption is a functional option for configuring random generation.
type RandOption[T Integer] func(*RandOptions[T])

// WithMin sets the minimum value.
func WithMin[T Integer](min T) RandOption[T] {
	return func(opts *RandOptions[T]) {
		opts.Min = min
	}
}

// WithMax sets the maximum value.
func WithMax[T Integer](max T) RandOption[T] {
	return func(opts *RandOptions[T]) {
		opts.Max = max
	}
}

// randInt generates a random integer of type T with the given options.
func randInt[T Integer](suite *TestSuite,
	defaultRandOptions func() RandOptions[T], opts ...RandOption[T],
) T {
	cfg := defaultRandOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	min := int64(cfg.Min)
	max := int64(cfg.Max)

	return T(suite.Rand.Int63n(max-min) + min)
}

// RandBool returns a random boolean value.
func (ts *TestSuite) RandBool() bool {
	return ts.RandInt(WithMax(2)) == 0
}

// RandInt8 returns a random int8 with optional configuration.
func (ts *TestSuite) RandInt8(opts ...RandOption[int8]) int8 {
	return randInt(ts, defaultRandOptions[int8], opts...)
}

// RandUint8 returns a random uint8 with optional configuration.
func (ts *TestSuite) RandUint8(opts ...RandOption[uint8]) uint8 {
	return randInt(ts, defaultRandOptions[uint8], opts...)
}

// RandInt16 returns a random int16 with optional configuration.
func (ts *TestSuite) RandInt16(opts ...RandOption[int16]) int16 {
	return randInt(ts, defaultRandOptions[int16], opts...)
}

// RandUint16 returns a random uint16 with optional configuration.
func (ts *TestSuite) RandUint16(opts ...RandOption[uint16]) uint16 {
	return randInt(ts, defaultRandOptions[uint16], opts...)
}

// RandInt32 returns a random int32 with optional configuration.
func (ts *TestSuite) RandInt32(opts ...RandOption[int32]) int32 {
	return randInt(ts, defaultRandOptions[int32], opts...)
}

// RandUint32 returns a random uint32 with optional configuration.
func (ts *TestSuite) RandUint32(opts ...RandOption[uint32]) uint32 {
	return randInt(ts, defaultRandOptions[uint32], opts...)
}

// RandInt64 returns a random int64 with optional configuration.
func (ts *TestSuite) RandInt64(opts ...RandOption[int64]) int64 {
	return randInt(ts, defaultRandOptions[int64], opts...)
}

// RandUint64 returns a random uint64 with optional configuration.
func (ts *TestSuite) RandUint64(opts ...RandOption[uint64]) uint64 {
	return randInt(ts, defaultRandOptions[uint64], opts...)
}

// RandInt returns a random int with optional configuration.
func (ts *TestSuite) RandInt(opts ...RandOption[int]) int {
	return randInt(ts, defaultRandOptions[int], opts...)
}

// RandUint returns a random uint with optional configuration.
func (ts *TestSuite) RandUint(opts ...RandOption[uint]) uint {
	return randInt(ts, defaultRandOptions[uint], opts...)
}

// RandBytes returns a slice of random bytes of the given length.
func (ts *TestSuite) RandBytes(length int) []byte {
	buf := make([]byte, length)
	_, err := ts.Rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return buf
}

// RandSlice generates a random non-repeating slice of int32 elements with the specified length.
func (ts *TestSuite) RandSlice(length int) []int32 {
	slice := []int32{}
	for {
		randInt := ts.RandInt32()
		if !slices.Contains(slice, randInt) {
			slice = append(slice, randInt)
		}

		if len(slice) == length {
			return slice
		}
	}
}

// Predefined charsets.
const (
	CharsetAlphabet     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetAlphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CharsetHex          = "0123456789abcdef"
)

type randStringConfig struct {
	charset string
}

type RandStringOption func(*randStringConfig)

func WithCharset(charset string) RandStringOption {
	return func(c *randStringConfig) {
		if charset != "" {
			c.charset = charset
		}
	}
}

// RandString generates a random string of the given length.
func (ts *TestSuite) RandString(length int, opts ...RandStringOption) string {
	cfg := randStringConfig{
		charset: CharsetAlphabet, // default
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	max := len(cfg.charset)
	b := make([]byte, length)
	for i := range b {
		b[i] = cfg.charset[ts.RandInt(WithMax(max))]
	}

	return string(b)
}

// RandHash32 generates a random 32-character hexadecimal string.
func (ts *TestSuite) RandHash32() string {
	return ts.RandString(32, WithCharset(CharsetHex))
}

// DecodingHex decodes the input string from hexadecimal format and returns the resulting byte slice.
func (*TestSuite) DecodingHex(in string) []byte {
	d, err := hex.DecodeString(in)
	if err != nil {
		panic(err)
	}

	return d
}
