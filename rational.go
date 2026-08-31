package decimal

import (
	"errors"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// maxPrecision mirrors Dart's maxPrecision constant: fractional digits
// beyond this are truncated (not rounded) when parsing.
const maxPrecision = 18

var pattern = regexp.MustCompile(`^([+-]?\d*)(\.\d*)?([eE][+-]?\d+)?$`)

// RationalNumber is a numerator/denominator pair backed by arbitrary
// precision integers, ported from lib/src/rational.dart.
type RationalNumber struct {
	Numerator   *big.Int
	Denominator *big.Int
}

// NewRationalNumber mirrors the Dart RationalNumber constructor, which
// asserts the denominator is non-zero.
func NewRationalNumber(numerator, denominator *big.Int) RationalNumber {
	if denominator.Sign() == 0 {
		panic("Division by zero is not allowed")
	}
	return RationalNumber{Numerator: numerator, Denominator: denominator}
}

// logBase10 mirrors utils.dart's logBase10: it is digitCount-1, correct
// only for powers of ten (which is all that this package feeds it).
func logBase10(n *big.Int) int {
	if n.Sign() <= 0 {
		panic("Logarithm undefined for non-positive values")
	}
	m := new(big.Int).Set(n)
	ten := big.NewInt(10)
	one := big.NewInt(1)
	count := 0
	for m.Cmp(one) > 0 {
		m.Quo(m, ten)
		count++
	}
	return count
}

// exponentialString mirrors Dart's num.toStringAsExponential() (no
// argument): the shortest round-trip exponential representation.
func exponentialString(v float64) string {
	return strconv.FormatFloat(v, 'e', -1, 64)
}

// Parse mirrors RationalNumber.parse(value) with no precision argument.
func Parse(value string) (RationalNumber, error) {
	return ParseWithPrecision(value, 0)
}

// ParseWithPrecision mirrors RationalNumber.parse(value, precision: precision).
func ParseWithPrecision(value string, precision int) (RationalNumber, error) {
	m := pattern.FindStringSubmatch(value)
	if m == nil {
		return RationalNumber{}, errors.New("invalid number format")
	}

	integerDigits := m[1]
	floatingPointDigits := m[2]
	exponentDigits := m[3]

	var numerator *big.Int
	denominator := big.NewInt(1)

	if floatingPointDigits != "" {
		ignoredDigits := 0
		if len(floatingPointDigits)-1 > maxPrecision {
			ignoredDigits = len(floatingPointDigits) - 1 - maxPrecision
		}

		limit := len(floatingPointDigits)
		if maxPrecision+1 < limit {
			limit = maxPrecision + 1
		}
		for i := 1; i < limit; i++ {
			denominator.Mul(denominator, big.NewInt(10))
		}

		numeratorString := integerDigits + floatingPointDigits[1:len(floatingPointDigits)-ignoredDigits]
		var ok bool
		numerator, ok = new(big.Int).SetString(numeratorString, 10)
		if !ok {
			return RationalNumber{}, errors.New("invalid number format")
		}
	} else {
		var ok bool
		numerator, ok = new(big.Int).SetString(integerDigits, 10)
		if !ok {
			return RationalNumber{}, errors.New("invalid number format")
		}
	}

	if exponentDigits != "" {
		expValue, err := strconv.Atoi(exponentDigits[1:])
		if err != nil {
			return RationalNumber{}, errors.New("invalid number format")
		}
		exponent := expValue - precision
		if exponent > 0 {
			numeratorString := numerator.String() + strings.Repeat("0", exponent)
			var ok bool
			numerator, ok = new(big.Int).SetString(numeratorString, 10)
			if !ok {
				return RationalNumber{}, errors.New("invalid number format")
			}
		} else {
			denominator.Mul(denominator, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-exponent)), nil))
		}
	}

	return NewRationalNumber(numerator, denominator).removeTrailingZeros().limitToMaxPrecision(maxPrecision), nil
}

// NewRationalNumberFromDecimal mirrors RationalNumber.fromDecimal(number, precision).
func NewRationalNumberFromDecimal(number float64, precision int) (RationalNumber, error) {
	return ParseWithPrecision(exponentialString(number), precision)
}

// ToValidDouble mirrors RationalNumber.toValidDouble(). Dart's BigInt
// division operator (used here) is numerator.toDouble() / denominator.toDouble():
// each operand is independently rounded to the nearest float64 first, then
// divided with native float64 division -- NOT a correctly-rounded
// conversion of the exact rational value. Replicated here rather than
// "fixed", since tests depend on this exact rounding behavior.
func (r RationalNumber) ToValidDouble() float64 {
	s := r.simplify()
	numeratorFloat, _ := new(big.Float).SetInt(s.Numerator).Float64()
	denominatorFloat, _ := new(big.Float).SetInt(s.Denominator).Float64()
	return numeratorFloat / denominatorFloat
}

func (r RationalNumber) simplify() RationalNumber {
	gcd := new(big.Int).GCD(nil, nil, r.Numerator, r.Denominator)
	if gcd.Sign() == 0 {
		return r
	}
	return RationalNumber{
		Numerator:   new(big.Int).Quo(r.Numerator, gcd),
		Denominator: new(big.Int).Quo(r.Denominator, gcd),
	}
}

// RoundTo mirrors RationalNumber.roundTo(precision).
func (r RationalNumber) RoundTo(precision int) RationalNumber {
	return r.limitToMaxPrecision(precision)
}

// Add mirrors RationalNumber operator+.
func (r RationalNumber) Add(other RationalNumber) RationalNumber {
	newNumerator := new(big.Int).Add(
		new(big.Int).Mul(r.Numerator, other.Denominator),
		new(big.Int).Mul(other.Numerator, r.Denominator),
	)
	newDenominator := new(big.Int).Mul(r.Denominator, other.Denominator)
	return NewRationalNumber(newNumerator, newDenominator).removeTrailingZeros().limitToMaxPrecision(maxPrecision)
}

// Sub mirrors RationalNumber operator-.
func (r RationalNumber) Sub(other RationalNumber) RationalNumber {
	newNumerator := new(big.Int).Sub(
		new(big.Int).Mul(r.Numerator, other.Denominator),
		new(big.Int).Mul(other.Numerator, r.Denominator),
	)
	newDenominator := new(big.Int).Mul(r.Denominator, other.Denominator)
	return NewRationalNumber(newNumerator, newDenominator).removeTrailingZeros().limitToMaxPrecision(maxPrecision)
}

// Mul mirrors RationalNumber operator*.
func (r RationalNumber) Mul(other RationalNumber) RationalNumber {
	if r.Numerator.Sign() == 0 || other.Numerator.Sign() == 0 {
		return NewRationalNumber(big.NewInt(0), big.NewInt(1))
	}

	newNumerator := new(big.Int).Mul(r.Numerator, other.Numerator)
	newDenominator := new(big.Int).Mul(r.Denominator, other.Denominator)
	return NewRationalNumber(newNumerator, newDenominator).removeTrailingZeros().limitToMaxPrecision(maxPrecision)
}

// Div mirrors RationalNumber operator/.
func (r RationalNumber) Div(other RationalNumber) RationalNumber {
	if other.Numerator.Sign() == 0 {
		panic("Division by zero is not allowed")
	}

	if r.Numerator.Sign() == 0 {
		return NewRationalNumber(big.NewInt(0), big.NewInt(1))
	}

	if other.Numerator.Cmp(other.Denominator) == 0 {
		return r
	}

	newNumerator := new(big.Int).Mul(r.Numerator, other.Denominator)
	newDenominator := new(big.Int).Mul(r.Denominator, other.Numerator)
	return NewRationalNumber(newNumerator, newDenominator).removeTrailingZeros().limitToMaxPrecision(maxPrecision)
}

func (r RationalNumber) removeTrailingZeros() RationalNumber {
	ten := big.NewInt(10)

	countTrailingZeros := func(n *big.Int) int {
		if n.Sign() == 0 {
			return 0
		}
		count := 0
		m := new(big.Int).Set(n)
		mod := new(big.Int)
		for {
			mod.Mod(m, ten)
			if mod.Sign() != 0 {
				break
			}
			count++
			m.Quo(m, ten)
		}
		return count
	}

	numeratorZeros := countTrailingZeros(r.Numerator)
	denominatorZeros := countTrailingZeros(r.Denominator)

	zeros := numeratorZeros
	if denominatorZeros < zeros {
		zeros = denominatorZeros
	}

	newNumerator := new(big.Int).Set(r.Numerator)
	newDenominator := new(big.Int).Set(r.Denominator)

	if zeros > 0 {
		divisor := new(big.Int).Exp(ten, big.NewInt(int64(zeros)), nil)
		newNumerator.Quo(newNumerator, divisor)
		newDenominator.Quo(newDenominator, divisor)
	}

	return NewRationalNumber(newNumerator, newDenominator)
}

func (r RationalNumber) limitToMaxPrecision(precision int) RationalNumber {
	newNumerator := new(big.Int).Set(r.Numerator)
	newDenominator := new(big.Int).Set(r.Denominator)

	powerOfTen := len(newDenominator.String()) - 1
	diffPrecision := powerOfTen - precision

	ten := big.NewInt(10)
	five := big.NewInt(5)
	remainder := new(big.Int)

	for i := 0; i < diffPrecision; i++ {
		remainder.Mod(newNumerator, ten)

		if remainder.Cmp(five) >= 0 {
			if newNumerator.Sign() < 0 {
				newNumerator.Sub(newNumerator, new(big.Int).Sub(ten, remainder))
			} else {
				newNumerator.Add(newNumerator, new(big.Int).Sub(ten, remainder))
			}
		} else {
			newNumerator.Sub(newNumerator, remainder)
		}

		newNumerator.Quo(newNumerator, ten)
		newDenominator.Quo(newDenominator, ten)
	}

	if newNumerator.Sign() == 0 {
		newDenominator = big.NewInt(1)
	}

	return NewRationalNumber(newNumerator, newDenominator)
}
