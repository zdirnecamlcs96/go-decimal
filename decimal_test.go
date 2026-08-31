package decimal

import (
	"math"
	"testing"
)

func TestGoDecimalAdditionSubtraction(t *testing.T) {
	t.Run("0 - 0.00039806763285024154", func(t *testing.T) {
		left := New(0, 0)
		right := FromDecimal(-3.9806763285024154, 4)
		result := left.Add(right)
		if result.Amount != -398067632850242 {
			t.Errorf("result.Amount = %d, want -398067632850242", result.Amount)
		}
		if result.Precision != 18 {
			t.Errorf("result.Precision = %d, want 18", result.Precision)
		}
	})

	t.Run("0.3824 - 0.00039806763285024154 = 0.3820019324", func(t *testing.T) {
		left := FromDecimal(3824, 4)                 // 3824 / 10000
		right := FromDecimal(-3.9806763285024154, 4) // -39806763285024154 / 100000000000000000000
		result := left.Add(right)
		if result.Amount != 382001932367149758 { // 0.3824 - 0.0004 = 0.3820
			t.Errorf("result.Amount = %d, want 382001932367149758", result.Amount)
		}
		if result.Precision != 18 {
			t.Errorf("result.Precision = %d, want 18", result.Precision)
		}
	})

	t.Run("1 + 0.5 == 1.5", func(t *testing.T) {
		a := New(1, 0)
		b := New(5, 1)
		result := a.Add(b)
		if result.Amount != 15 {
			t.Errorf("result.Amount = %d, want 15", result.Amount)
		}
		if result.Precision != 1 {
			t.Errorf("result.Precision = %d, want 1", result.Precision)
		}
	})

	t.Run("1 - 0.5 == 0.5", func(t *testing.T) {
		a := New(1, 0)
		b := New(5, 1)
		result := a.Sub(b)
		if result.Amount != 5 {
			t.Errorf("result.Amount = %d, want 5", result.Amount)
		}
		if result.Precision != 1 {
			t.Errorf("result.Precision = %d, want 1", result.Precision)
		}
	})
}

func TestGoDecimalMultiplication(t *testing.T) {
	t.Run("10 * 0.3 == 3", func(t *testing.T) {
		left := New(10, 0)
		right := New(3, 1)
		result := left.Mul(right)
		if result.Amount != 3 {
			t.Errorf("result.Amount = %d, want 3", result.Amount)
		}
		if result.Precision != 0 {
			t.Errorf("result.Precision = %d, want 0", result.Precision)
		}
	})

	t.Run("10 * 1 == 1", func(t *testing.T) {
		left := New(10, 0)
		right := New(1, 0)
		result := left.Mul(right)
		if result.Amount != 10 {
			t.Errorf("result.Amount = %d, want 10", result.Amount)
		}
		if result.Precision != 0 {
			t.Errorf("result.Precision = %d, want 0", result.Precision)
		}
	})

	t.Run("0.3 * 0.3 == 0.09", func(t *testing.T) {
		left := New(3, 1)
		right := New(3, 1)
		result := left.Mul(right)
		if result.Amount != 9 {
			t.Errorf("result.Amount = %d, want 9", result.Amount)
		}
		if result.Precision != 2 {
			t.Errorf("result.Precision = %d, want 2", result.Precision)
		}
	})

	t.Run("3 * 3 == 9", func(t *testing.T) {
		left := New(3, 0)
		right := New(3, 0)
		result := left.Mul(right)
		if result.Amount != 9 {
			t.Errorf("result.Amount = %d, want 9", result.Amount)
		}
		if result.Precision != 0 {
			t.Errorf("result.Precision = %d, want 0", result.Precision)
		}
	})

	t.Run("1.005 * 1000 = 1005", func(t *testing.T) {
		a := ParseFloat(1.005)
		b := ParseFloat(1000)
		result := a.Mul(b)
		if result.Amount != 1005 {
			t.Errorf("result.Amount = %d, want 1005", result.Amount)
		}
		if result.Precision != 0 {
			t.Errorf("result.Precision = %d, want 0", result.Precision)
		}
	})
}

func TestGoDecimalDivision(t *testing.T) {
	t.Run("0.002 / 3 == 0", func(t *testing.T) {
		left := New(2, 3)
		right := New(3, 0)
		result := left.Div(right)
		if math.Trunc(result.ToFloat64()) != 0 {
			t.Errorf("truncated result = %v, want 0", math.Trunc(result.ToFloat64()))
		}
	})

	t.Run("575 / 50 == 11.5", func(t *testing.T) {
		left := New(575, 0)
		right := New(50, 0)
		result := left.Div(right)
		if result.ToFloat64() != 11.5 {
			t.Errorf("result.ToFloat64() = %v, want 11.5", result.ToFloat64())
		}
	})

	t.Run("114.4 / 15.2 = 7.526315789473684 * 15.2 = 114.4000000004", func(t *testing.T) {
		a := FromDecimal(11440, 2)
		b := FromDecimal(152, 1)
		r1 := a.Div(b).Mul(b)

		// 7.526315789473684 * 15.2 = 114.4 (rounded to 1 precision)
		if r1.Amount != 1144 { // 114.4
			t.Errorf("r1.Amount = %d, want 1144", r1.Amount)
		}
		if r1.Precision != 1 {
			t.Errorf("r1.Precision = %d, want 1", r1.Precision)
		}
	})
}

func TestGoDecimalRoundToPrecision(t *testing.T) {
	t.Run("0.4 to 3 precision == 0.4 instead of 0.400", func(t *testing.T) {
		d := New(4, 1)
		result := d.RoundToPrecision(3)
		if result.Amount != 4 {
			t.Errorf("result.Amount = %d, want 4", result.Amount)
		}
		if result.Precision != 1 {
			t.Errorf("result.Precision = %d, want 1", result.Precision)
		}
	})

	t.Run("0.47 to 1 precision == 0.5", func(t *testing.T) {
		d := New(47, 2)
		result := d.RoundToPrecision(1)
		if result.Amount != 5 {
			t.Errorf("result.Amount = %d, want 5", result.Amount)
		}
		if result.Precision != 1 {
			t.Errorf("result.Precision = %d, want 1", result.Precision)
		}
	})

	t.Run("0.22 to 1 precision == 0.2", func(t *testing.T) {
		d := New(22, 2)
		result := d.RoundToPrecision(1)
		if result.Amount != 2 {
			t.Errorf("result.Amount = %d, want 2", result.Amount)
		}
		if result.Precision != 1 {
			t.Errorf("result.Precision = %d, want 1", result.Precision)
		}
	})

	t.Run("890.9913293236 - 1.446468288", func(t *testing.T) {
		left := New(8909913293236, 10)
		right := FromDecimal(1446468288, 9)
		result := left.Sub(right)
		roundToTen := result.RoundToPrecision(10)
		if roundToTen.Amount != 8895448610356 {
			t.Errorf("roundToTen.Amount = %d, want 8895448610356", roundToTen.Amount)
		}
		if roundToTen.Precision != 10 {
			t.Errorf("roundToTen.Precision = %d, want 10", roundToTen.Precision)
		}
	})
}

func TestGoDecimalDivideAndMultiply(t *testing.T) {
	t.Run("890.9913293236 / 37160", func(t *testing.T) {
		left := New(8909913293236, 10)
		right := New(37160, 0)
		result := left.Div(right)
		if result.Amount != 23977161714843917 {
			t.Errorf("result.Amount = %d, want 23977161714843917", result.Amount)
		}
		if result.Precision != 18 {
			t.Errorf("result.Precision = %d, want 18", result.Precision)
		}

		multiplied := result.Mul(New(60, 0))
		if multiplied.Amount != 1438629702890635 {
			t.Errorf("multiplied.Amount = %d, want 1438629702890635", multiplied.Amount)
		}
		if multiplied.Precision != 15 {
			t.Errorf("multiplied.Precision = %d, want 15", multiplied.Precision)
		}
	})
}
