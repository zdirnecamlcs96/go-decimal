# Go Decimal

## Features

Go library that solves the floating-point issue by using `big.Int`.

**Note:** `Decimal` supports precision up to 18 digits because it uses `int64` instead of `big.Int`. The purpose is to avoid unnecessarily large numbers or high precision. Additionally, `int64` uses a fixed amount of memory and may be more efficient than `big.Int`.

## Getting started

```bash
go get github.com/zdirnecamlcs96/go-decimal
```

## Usage

```go
import "github.com/zdirnecamlcs96/go-decimal"

// 1.005 * 1000 = 1004.9999999999999 (floating-point issue)
a := decimal.New(1005, 3) // 1005 / 10^3 = 1.005
b := decimal.ParseFloat(1000)
result := a.Sub(b) // 1.005
fmt.Println(result) // "1.005"
```

## API

### Constructors

| Function | Description |
|---|---|
| `New(amount int64, precision int)` | Create from integer amount and decimal precision. `New(1005, 3)` = `1.005` |
| `ParseFloat(value float64)` | Parse from a float64 value |
| `FromDecimal(amount float64, precision int)` | Parse a float value scaled by `10^(-precision)` |

### Methods

| Method | Description |
|---|---|
| `Add(other GoDecimal) GoDecimal` | Addition |
| `Sub(other GoDecimal) GoDecimal` | Subtraction |
| `Mul(other GoDecimal) GoDecimal` | Multiplication |
| `Div(other GoDecimal) GoDecimal` | Division (panics on divide-by-zero) |
| `RoundToPrecision(precision int) GoDecimal` | Round to N decimal places |
| `ToFloat64() float64` | Convert to float64 |
| `CompareTo(other GoDecimal) int` | Compare: returns -1, 0, or 1 |

### RationalNumber (low-level)

`RationalNumber` holds an exact rational value as `Numerator / Denominator` using `big.Int`. It is used internally by `Decimal` but is also exported for direct use.

| Function / Method | Description |
|---|---|
| `NewRationalNumber(num, den *big.Int)` | Construct directly |
| `Parse(value string)` | Parse a string |
| `ParseWithPrecision(value string, precision int)` | Parse a string; precision shifts the exponent by `-precision` |
| `NewRationalNumberFromDecimal(number float64, precision int)` | Create from a scaled float64 |
| `Add / Sub / Mul / Div` | Arithmetic, returning a new `RationalNumber` |
| `RoundTo(precision int)` | Round to N decimal places |
| `ToValidDouble() float64` | Convert to float64 |

## JSON

`GoDecimal` implements `json.Marshaler` and `json.Unmarshaler`. It marshals as a plain JSON number (e.g. `12.34`) and unmarshals either a JSON number or a quoted decimal string (`"12.34"`); `null` leaves the value unchanged. Both directions go through `float64` (`ToFloat64` / `ParseFloat`), so precision beyond ~15 significant digits is not preserved on the wire.

## Edge Cases

### Intermediate rounding with limited precision

When performing calculations with limited precision (e.g., rounding to 10 decimal places), discrepancies can arise if intermediate results are rounded too early. Rounding early may cause small errors when those rounded values are used in subsequent calculations, leading to minor discrepancies in the final result.

#### Example 1: Continuous Calculation (No Intermediate Rounding)

```go
// No rounding until the final result (a / b * c)
a := decimal.ParseFloat(114.4)
b := decimal.ParseFloat(15.2)
c := a.Div(b)          // 7.526315789473684 (full precision)
result := c.Mul(b)     // 114.4 (exact result)
```

#### Example 2: Intermediate Rounding (Round After Division)

```go
// Intermediate rounding before multiplying back
a := decimal.ParseFloat(114.4)
b := decimal.ParseFloat(15.2)
c := a.Div(b).RoundToPrecision(10) // 7.5263157895 (rounded to 10 decimal places)
result := c.Mul(b)                 // 114.4000000004 (slight discrepancy due to rounding)
```

### How to handle it

1. **Avoid Intermediate Rounding** to prevent such discrepancies.
2. For the use case above, **check if `b` is the same as `c`** and return `a` directly without further manipulation to ensure the expected outcome.

## Disclaimer

This project is intended for personal use only. Use at your own risk. No warranty or support is provided. However, you are welcome to raise issues or submit pull requests.

## License

This project is licensed under the [MIT License](./LICENSE).

## Acknowledgements

- Also available as a [Dart package](https://github.com/zdirnecamlcs96/dart-decimal).
