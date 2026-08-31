package decimal

import (
	"math/big"
	"strconv"
	"testing"
)

func bigFromString(t *testing.T, s string) *big.Int {
	t.Helper()
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		t.Fatalf("failed to parse big int %q", s)
	}
	return n
}

func pow10(exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
}

func mustParse(t *testing.T, value string) RationalNumber {
	t.Helper()
	r, err := Parse(value)
	if err != nil {
		t.Fatalf("Parse(%q) error: %v", value, err)
	}
	return r
}

func mustParseWithPrecision(t *testing.T, value string, precision int) RationalNumber {
	t.Helper()
	r, err := ParseWithPrecision(value, precision)
	if err != nil {
		t.Fatalf("ParseWithPrecision(%q, %d) error: %v", value, precision, err)
	}
	return r
}

func assertBigIntEqual(t *testing.T, got, want *big.Int, label string) {
	t.Helper()
	if got.Cmp(want) != 0 {
		t.Errorf("%s = %s, want %s", label, got.String(), want.String())
	}
}

func TestRationalScientificNotation(t *testing.T) {
	t.Run("scientific notation", func(t *testing.T) {
		a := mustParse(t, "9.223372036854776e+24")
		assertBigIntEqual(t, a.Numerator, bigFromString(t, "9223372036854776000000000"), "a.Numerator")

		b := mustParse(t, "1.2345e+5")
		assertBigIntEqual(t, b.Numerator, big.NewInt(123450), "b.Numerator")
		assertBigIntEqual(t, b.Denominator, big.NewInt(1), "b.Denominator")

		c := mustParse(t, "1.000000000000000000")
		assertBigIntEqual(t, c.Numerator, big.NewInt(1), "c.Numerator")
		assertBigIntEqual(t, c.Denominator, big.NewInt(1), "c.Denominator")

		d := mustParse(t, "1.000050000000000000")
		assertBigIntEqual(t, d.Numerator, big.NewInt(100005), "d.Numerator")
		assertBigIntEqual(t, d.Denominator, pow10(5), "d.Denominator")

		// Max precision is 18, so the last 2 digits are ignored
		e := mustParse(t, "0.1234567890123456789")
		assertBigIntEqual(t, e.Numerator, big.NewInt(123456789012345678), "e.Numerator")
		assertBigIntEqual(t, e.Denominator, pow10(18), "e.Denominator")

		f := mustParse(t, "0.1999999999999999999")
		assertBigIntEqual(t, f.Numerator, big.NewInt(199999999999999999), "f.Numerator")
		assertBigIntEqual(t, f.Denominator, pow10(18), "f.Denominator")
	})

	t.Run("scientific notation with precision", func(t *testing.T) {
		amount := 9.223372036854776e+24
		precision := 50

		a, err := NewRationalNumberFromDecimal(amount, precision)
		if err != nil {
			t.Fatalf("NewRationalNumberFromDecimal error: %v", err)
		}
		assertBigIntEqual(t, a.Numerator, big.NewInt(0), "a.Numerator")
		assertBigIntEqual(t, a.Denominator, big.NewInt(1), "a.Denominator")

		b := mustParseWithPrecision(t, exponentialString(11440), 2)
		assertBigIntEqual(t, b.Numerator, big.NewInt(1144), "b.Numerator")
		assertBigIntEqual(t, b.Denominator, big.NewInt(10), "b.Denominator")
	})

	t.Run("Subtraction", func(t *testing.T) {
		right, err := NewRationalNumberFromDecimal(3.9806763285024154, 4)
		if err != nil {
			t.Fatalf("NewRationalNumberFromDecimal error: %v", err)
		}
		result := mustParse(t, "0").Sub(right)

		assertBigIntEqual(t, result.Numerator, big.NewInt(-398067632850242), "result.Numerator")
		assertBigIntEqual(t, result.Denominator, bigFromString(t, "1000000000000000000"), "result.Denominator")

		doubleResult := result.ToValidDouble()
		if doubleResult != -0.000398067632850242 {
			t.Errorf("doubleResult = %v, want -0.000398067632850242", doubleResult)
		}
	})

	t.Run("Division", func(t *testing.T) {
		result := mustParse(t, "120").Div(mustParse(t, "228"))
		assertBigIntEqual(t, result.Numerator, big.NewInt(120), "result.Numerator")
		assertBigIntEqual(t, result.Denominator, big.NewInt(228), "result.Denominator")

		doubleResult := result.ToValidDouble()
		if doubleResult != 0.5263157894736842 {
			t.Errorf("doubleResult = %v, want 0.5263157894736842", doubleResult)
		}

		doubleToRational := mustParse(t, strconv.FormatFloat(doubleResult, 'g', -1, 64))
		assertBigIntEqual(t, doubleToRational.Numerator, big.NewInt(5263157894736842), "doubleToRational.Numerator")
		assertBigIntEqual(t, doubleToRational.Denominator, pow10(16), "doubleToRational.Denominator")

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic dividing by zero")
				}
			}()
			mustParse(t, "1").Div(mustParse(t, "0"))
		}()

		zero1 := mustParse(t, "0").Div(mustParse(t, "228"))
		assertBigIntEqual(t, zero1.Numerator, big.NewInt(0), "zero1.Numerator")
	})

	t.Run("Multiply", func(t *testing.T) {
		zero1 := mustParse(t, "1234").Mul(mustParse(t, "0"))
		assertBigIntEqual(t, zero1.Numerator, big.NewInt(0), "zero1.Numerator")
		assertBigIntEqual(t, zero1.Denominator, big.NewInt(1), "zero1.Denominator")

		zero2 := mustParse(t, "0").Mul(mustParse(t, "77834"))
		assertBigIntEqual(t, zero2.Numerator, big.NewInt(0), "zero2.Numerator")
		assertBigIntEqual(t, zero1.Denominator, big.NewInt(1), "zero1.Denominator")
	})

	t.Run("Round Up", func(t *testing.T) {
		doubleToRational := mustParse(t, "0.5263157894736842")
		assertBigIntEqual(t, doubleToRational.Numerator, big.NewInt(5263157894736842), "doubleToRational.Numerator")
		assertBigIntEqual(t, doubleToRational.Denominator, pow10(16), "doubleToRational.Denominator")

		rounded := doubleToRational.RoundTo(10)
		assertBigIntEqual(t, rounded.Numerator, big.NewInt(5263157895), "rounded.Numerator")
		assertBigIntEqual(t, rounded.Denominator, pow10(10), "rounded.Denominator")
	})
}
