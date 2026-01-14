package testsuite

import (
	"encoding/hex"
	"math/rand"
	"slices"
	"testing"
	"time"
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

// RandBool returns a random boolean value.
func (ts *TestSuite) RandBool() bool {
	return ts.RandInt64(2) == 0
}

// RandInt8 returns a random int8 between 0 and max: [0, max).
func (ts *TestSuite) RandInt8(max int8) int8 {
	return int8(ts.RandUint64(uint64(max)))
}

// RandUint8 returns a random uint8 between 0 and max: [0, max).
func (ts *TestSuite) RandUint8(max uint8) uint8 {
	return uint8(ts.RandUint64(uint64(max)))
}

// RandInt16 returns a random int16 between 0 and max: [0, max).
func (ts *TestSuite) RandInt16(max int16) int16 {
	return int16(ts.RandUint64(uint64(max)))
}

// RandUint16 returns a random uint16 between 0 and max: [0, max).
func (ts *TestSuite) RandUint16(max uint16) uint16 {
	return uint16(ts.RandUint64(uint64(max)))
}

// RandInt32 returns a random int32 between 0 and max: [0, max).
func (ts *TestSuite) RandInt32(max int32) int32 {
	return int32(ts.RandUint64(uint64(max)))
}

// RandUint32 returns a random uint32 between 0 and max: [0, max).
func (ts *TestSuite) RandUint32(max uint32) uint32 {
	return uint32(ts.RandUint64(uint64(max)))
}

// RandInt64 returns a random int64 between 0 and max: [0, max).
func (ts *TestSuite) RandInt64(max int64) int64 {
	return ts.Rand.Int63n(max)
}

// RandUint64 returns a random uint64 between 0 and max: [0, max).
func (ts *TestSuite) RandUint64(max uint64) uint64 {
	return uint64(ts.RandInt64(int64(max)))
}

// RandInt returns a random int between 0 and max: [0, max).
func (ts *TestSuite) RandInt(max int) int {
	return int(ts.RandInt64(int64(max)))
}

// RandInt16NonZero returns a random int16 between 1 and max+1: [1, max+1).
func (ts *TestSuite) RandInt16NonZero(max int16) int16 {
	return ts.RandInt16(max) + 1
}

// RandUint16NonZero returns a random uint16 between 1 and max+1: [1, max+1).
func (ts *TestSuite) RandUint16NonZero(max uint16) uint16 {
	return ts.RandUint16(max) + 1
}

// RandInt32NonZero returns a random int32 between 1 and max+1: [1, max+1).
func (ts *TestSuite) RandInt32NonZero(max int32) int32 {
	return ts.RandInt32(max) + 1
}

// RandUint32NonZero returns a random uint32 between 1 and max+1: [1, max+1).
func (ts *TestSuite) RandUint32NonZero(max uint32) uint32 {
	return ts.RandUint32(max) + 1
}

// RandInt64NonZero returns a random int64 between 1 and max+1: [1, max+1).
func (ts *TestSuite) RandInt64NonZero(max int64) int64 {
	return ts.RandInt64(max) + 1
}

// RandUint64NonZero returns a random uint64 between 1 and max+1: [1, max+1).
func (ts *TestSuite) RandUint64NonZero(max uint64) uint64 {
	return ts.RandUint64(max) + 1
}

// RandIntNonZero returns a random int between 1 and max+1: [1, max+1).
func (ts *TestSuite) RandIntNonZero(max int) int {
	return ts.RandInt(max) + 1
}

// RandIntRange returns a random int between [min, max).
func (ts *TestSuite) RandIntRange(min, max int) int {
	return ts.RandIntNonZero(max-min) + min
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
		randInt := ts.RandInt32(1000)
		if !slices.Contains(slice, randInt) {
			slice = append(slice, randInt)
		}

		if len(slice) == length {
			return slice
		}
	}
}

// RandString generates a random string of the given length.
func (ts *TestSuite) RandString(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[ts.RandInt(52)]
	}

	return string(b)
}

// DecodingHex decodes the input string from hexadecimal format and returns the resulting byte slice.
func (*TestSuite) DecodingHex(in string) []byte {
	d, err := hex.DecodeString(in)
	if err != nil {
		panic(err)
	}

	return d
}
