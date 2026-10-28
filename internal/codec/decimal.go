package codec

import (
	"fmt"

	"github.com/ericlagergren/decimal"
)

// DecimalContext returns Python's default arithmetic context without changing
// process-wide settings. Parsing retains all digits independently of this context.
func DecimalContext() decimal.Context {
	return decimal.Context{
		Precision: 28, RoundingMode: decimal.ToNearestEven, OperatingMode: decimal.GDA,
		MaxScale: 999999, MinScale: -999999,
		Traps: decimal.InvalidOperation | decimal.Overflow | decimal.DivisionByZero,
	}
}

// DecimalArithmetic performs the operations used by price estimation. Rounding
// is applied once, including subnormals, with Quantize: the dependency's Round
// incorrectly rounds some underflows up instead of using half-even.
func DecimalArithmetic(operation string, result, x, y *decimal.Big) error {
	ctx := decimal.Context{Precision: 28, RoundingMode: decimal.ToNearestEven, OperatingMode: decimal.Go}
	switch operation {
	case "add":
		ctx.Add(result, x, y)
	case "sub":
		ctx.Sub(result, x, y)
	case "mul":
		ctx.Mul(result, x, y)
	default:
		return fmt.Errorf("unsupported decimal operation %q", operation)
	}
	return RoundDecimal(result)
}

// RoundDecimal applies Python's 28-digit half-even context to an exact result.
func RoundDecimal(value *decimal.Big) error {
	if !value.IsFinite() {
		return fmt.Errorf("decimal overflow or invalid operation")
	}
	const tinyScale = 1000026
	adjusted := value.Precision() - value.Scale() - 1
	if adjusted < -tinyScale-1 {
		negative := value.Signbit()
		value.SetMantScale(0, tinyScale).SetSignbit(negative)
		return nil
	}
	ctx := DecimalContext()
	if value.Precision() > 28 || value.Scale() > tinyScale {
		scale := min(value.Scale()-max(value.Precision()-28, 0), tinyScale)
		ctx.Quantize(value, scale)
		if value.IsFinite() && value.Scale() < scale && value.Precision()+scale-value.Scale() <= 28 {
			ctx.Quantize(value, scale)
		}
	}
	if !value.IsFinite() || value.Precision()-value.Scale()-1 > 999999 && value.Sign() != 0 {
		return fmt.Errorf("decimal overflow or invalid operation")
	}
	if value.Sign() == 0 && value.Scale() < -999999 {
		value.SetScale(-999999)
	}
	return nil
}
