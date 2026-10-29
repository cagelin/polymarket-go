package trading

import (
	"fmt"
	"math/big"

	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

type rounding struct{ amount, price int }

var tickRoundings = [...]struct {
	tick   *decimal.Big
	config rounding
}{
	{decimal.New(1, 1), rounding{3, 1}}, {decimal.New(1, 2), rounding{4, 2}},
	{decimal.New(5, 3), rounding{5, 3}}, {decimal.New(25, 4), rounding{6, 4}},
	{decimal.New(1, 3), rounding{5, 3}}, {decimal.New(1, 4), rounding{6, 4}},
}

func tickRounding(tick *decimal.Big) (rounding, error) {
	if tick != nil && tick.IsFinite() {
		for _, entry := range tickRoundings {
			if tick.Cmp(entry.tick) == 0 {
				return entry.config, nil
			}
		}
	}
	return rounding{}, response(fmt.Errorf("unsupported tick size"))
}
func positive(value *decimal.Big) bool { return value != nil && value.IsFinite() && value.Sign() > 0 }
func places(value *decimal.Big) int {
	v := new(decimal.Big).Copy(value)
	if codec.RoundDecimal(v) != nil {
		return 0
	}
	ctx := codec.DecimalContext()
	ctx.Reduce(v)
	return max(0, v.Scale())
}
func validatePrice(price, tick *decimal.Big) error {
	if !positive(price) {
		return input("price must be positive and finite")
	}
	ceiling := new(decimal.Big)
	if err := codec.DecimalArithmetic("sub", ceiling, decimal.New(1, 0), tick); err != nil {
		return response(err)
	}
	if price.Cmp(tick) < 0 || price.Cmp(ceiling) > 0 || places(price) > places(tick) {
		return input("price is outside the tick grid")
	}
	ratio := new(big.Rat).Quo(price.Rat(nil), tick.Rat(nil))
	if !ratio.IsInt() {
		return input("price must be a multiple of tick size")
	}
	return nil
}

// orderMath retains the first arithmetic failure through a single amount calculation.
type orderMath struct{ err error }

func (m *orderMath) op(operation string, x, y *decimal.Big) *decimal.Big {
	result := new(decimal.Big)
	if m.err != nil {
		return result
	}
	if operation == "div" || operation == "pow" {
		ctx := codec.DecimalContext()
		ctx.OperatingMode = decimal.Go
		if operation == "div" {
			ctx.Quo(result, x, y)
		} else {
			ctx.Pow(result, x, y)
		}
		m.err = codec.RoundDecimal(result)
	} else {
		m.err = codec.DecimalArithmetic(operation, result, x, y)
	}
	return result
}
func (m *orderMath) round(value *decimal.Big, digits int, mode decimal.RoundingMode) *decimal.Big {
	result := new(decimal.Big).Copy(value)
	if m.err != nil || places(value) <= digits {
		return result
	}
	ctx := codec.DecimalContext()
	ctx.RoundingMode = mode
	ctx.Quantize(result, digits)
	if !result.IsFinite() {
		m.err = fmt.Errorf("order amount exceeds decimal precision")
	}
	return result
}
func (m *orderMath) amount(value *decimal.Big) *big.Int {
	scaled := m.op("mul", value, decimal.New(1000000, 0))
	scaled = m.round(scaled, 0, decimal.ToNearestEven)
	if m.err != nil {
		return nil
	}
	if !scaled.IsFinite() || scaled.Precision()-scaled.Scale() > 28 {
		m.err = fmt.Errorf("order amount exceeds decimal precision")
		return nil
	}
	return scaled.Int(nil)
}
func (m *orderMath) amountRound(value *decimal.Big, digits int) *decimal.Big {
	return m.round(m.round(value, digits+4, decimal.ToPositiveInf), digits, decimal.ToNegativeInf)
}
func limitAmounts(price, size, tick *decimal.Big, side models.OrderSide) (*big.Int, *big.Int, error) {
	config, err := tickRounding(tick)
	if err != nil {
		return nil, nil, err
	}
	m := orderMath{}
	shares := m.round(size, 2, decimal.ToNegativeInf)
	amount := m.amountRound(m.op("mul", shares, price), config.amount)
	offered, requested := m.amount(amount), m.amount(shares)
	if side == models.Sell {
		offered, requested = requested, offered
	}
	return offered, requested, m.err
}
func marketAmounts(price, amount, tick *decimal.Big, side models.OrderSide) (*big.Int, *big.Int, error) {
	config, err := tickRounding(tick)
	if err != nil {
		return nil, nil, err
	}
	m := orderMath{}
	price = m.round(price, config.price, decimal.ToNegativeInf)
	maker := m.round(amount, 2, decimal.ToNegativeInf)
	operation := "mul"
	if side == models.Buy {
		operation = "div"
	}
	taker := m.amountRound(m.op(operation, maker, price), config.amount)
	offered, requested := m.amount(maker), m.amount(taker)
	return offered, requested, m.err
}
func adjustBuyAmount(amount, price, maxSpend, rate, exponent, builderRate *decimal.Big) (*decimal.Big, error) {
	m := orderMath{}
	one := decimal.New(1, 0)
	effective := m.op("mul", rate, m.op("pow", m.op("mul", price, m.op("sub", one, price)), exponent))
	platform := m.op("mul", m.op("div", amount, price), effective)
	builder := m.op("mul", amount, builderRate)
	total := m.op("add", m.op("add", amount, platform), builder)
	if m.err != nil {
		return nil, m.err
	}
	if maxSpend.Cmp(total) > 0 {
		return new(decimal.Big).Copy(amount), nil
	}
	result := m.op("div", maxSpend, m.op("add", m.op("add", one, m.op("div", effective, price)), builderRate))
	return result, m.err
}
func protectedBuy(offered, requested *big.Int, price *decimal.Big) error {
	if offered.Sign() <= 0 || requested.Sign() <= 0 {
		return input("protected BUY amount rounds to zero")
	}
	next := new(big.Rat).Add(price.Rat(nil), big.NewRat(1, 10000))
	left := new(big.Int).Mul(offered, next.Denom())
	right := new(big.Int).Mul(requested, next.Num())
	if left.Cmp(right) >= 0 {
		return input("amount precision cannot preserve the maximum execution price")
	}
	return nil
}
