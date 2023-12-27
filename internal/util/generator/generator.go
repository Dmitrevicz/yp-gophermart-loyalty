package generator

import (
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/ShiraazMoollatjie/goluhn"
)

var (
	ErrOrderGeneratorLimitReached = errors.New("generator limit reached, values started from the very beginning")
)

type OrderNumberGenerator struct {
	counter atomic.Uint64
}

// NewOrderNumberGenerator - seed may be provided as a start point
// (e.g to continue sequense after app restart).
// This generator has a limit of uint64 max value numbers.
// Parameter seed is a value for counter to start counting from;
// seed must be valid numeric luhn sequence.
func NewOrderNumberGenerator(seed ...string) (numgen *OrderNumberGenerator, err error) {
	numgen = &OrderNumberGenerator{}

	if len(seed) > 0 {
		err = numgen.SetSeed(seed[0])
	}

	return numgen, err
}

// New generates next order number.
// Order number value is just incremented counter value
// with a luhn check digit appended to the end.
func (g *OrderNumberGenerator) New() (number string, err error) {
	count := g.counter.Add(1)
	if count == 0 {
		return "0", ErrOrderGeneratorLimitReached
	}

	s := strconv.FormatUint(count, 10)

	_, number, err = goluhn.Calculate(s)

	return
}

// SetSeed seed must be valid numeric luhn sequence.
func (g *OrderNumberGenerator) SetSeed(seed string) (err error) {
	// value "0" is expected to appear on the very first app start
	// when there is not a single order exists yet
	if seed != "0" {
		if err = goluhn.Validate(seed); err != nil {
			return err
		}
	}

	if len(seed) > 1 {
		seed = seed[:len(seed)-1]
	}

	num, err := strconv.ParseUint(seed, 10, 64)
	if err != nil {
		return err
	}

	g.counter.Store(num)

	return nil
}
