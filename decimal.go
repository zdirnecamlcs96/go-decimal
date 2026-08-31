package decimal

import (
	"math"
	"math/big"
	"strconv"
)

// GoDecimal is ported from lib/src/decimal.dart.
type GoDecimal struct {
	Amount    int64
	Precision int

	rational RationalNumber
}

// New mirrors the GoDecimal({amount, precision}) constructor. Dart's
// BigInt.pow(precision) raises a RangeError for a negative exponent; Go's
// big.Int.Exp does not, so this panics explicitly to match.
func New(amount int64, precision int) GoDecimal {
	if precision < 0 {
		panic("precision must be non-negative")
	}
	denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil)
	rational := NewRationalNumber(big.NewInt(amount), denominator)
	return GoDecimal{Amount: amount, Precision: precision, rational: rational}
}

// FromDecimal mirrors GoDecimal.fromDecimal({amount, precision}).
func FromDecimal(amount float64, precision int) GoDecimal {
	rationalNum, err := NewRationalNumberFromDecimal(amount, precision)
	if err != nil {
		panic(err)
	}
	return New(rationalNum.Numerator.Int64(), logBase10(rationalNum.Denominator))
}

// ParseFloat mirrors GoDecimal.parse(value).
func ParseFloat(value float64) GoDecimal {
	rationalNum, err := Parse(exponentialString(value))
	if err != nil {
		panic(err)
	}
	return New(rationalNum.Numerator.Int64(), logBase10(rationalNum.Denominator))
}

// ToFloat64 mirrors GoDecimal.toDouble().
func (d GoDecimal) ToFloat64() float64 {
	precision := d.Precision
	if precision > 18 {
		precision = 18
	}
	v := float64(d.Amount) / math.Pow(10, float64(precision))
	s := strconv.FormatFloat(v, 'f', precision, 64)
	result, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return result
}

// RoundToPrecision mirrors GoDecimal.roundToPrecision(precision).
func (d GoDecimal) RoundToPrecision(precision int) GoDecimal {
	r := d.rational.RoundTo(precision)
	return New(r.Numerator.Int64(), logBase10(r.Denominator))
}

// Add mirrors GoDecimal operator+.
func (d GoDecimal) Add(other GoDecimal) GoDecimal {
	newAmount := d.rational.Add(other.rational)
	return New(newAmount.Numerator.Int64(), logBase10(newAmount.Denominator))
}

// Sub mirrors GoDecimal operator-.
func (d GoDecimal) Sub(other GoDecimal) GoDecimal {
	newAmount := d.rational.Sub(other.rational)
	return New(newAmount.Numerator.Int64(), logBase10(newAmount.Denominator))
}

// Mul mirrors GoDecimal operator*.
func (d GoDecimal) Mul(other GoDecimal) GoDecimal {
	newAmount := d.rational.Mul(other.rational)
	return ParseFloat(newAmount.ToValidDouble())
}

// Div mirrors GoDecimal operator/.
func (d GoDecimal) Div(other GoDecimal) GoDecimal {
	newAmount := d.rational.Div(other.rational)
	return ParseFloat(newAmount.ToValidDouble())
}

// CompareTo mirrors GoDecimal.compareTo(b).
func (d GoDecimal) CompareTo(other GoDecimal) int {
	a := d.ToFloat64()
	b := other.ToFloat64()
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}
