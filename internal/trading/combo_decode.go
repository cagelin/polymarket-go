package trading

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

type rfqFields struct {
	fields map[string]json.RawMessage
	err    *error
}

func readRFQ(raw []byte) rfqFields {
	fields, err := codec.Object(raw)
	return rfqFields{fields: fields, err: &err}
}
func (r rfqFields) fail(err error) {
	if *r.err == nil {
		*r.err = err
	}
}
func (r rfqFields) text(key string) string {
	value, err := codec.Required[string](r.fields, key)
	r.fail(err)
	if value == "" {
		r.fail(fmt.Errorf("missing RFQ %s", key))
	}
	return value
}
func (r rfqFields) optional(key string) *string {
	if r.fields[key] == nil {
		return nil
	}
	value := r.text(key)
	return &value
}
func (r rfqFields) object(key string) rfqFields {
	fields, err := codec.Object(r.fields[key])
	r.fail(err)
	return rfqFields{fields: fields, err: r.err}
}
func (r rfqFields) position(key string) string {
	value, err := unsignedPosition(r.text(key))
	r.fail(err)
	return value
}
func (r rfqFields) amount(key string) *decimal.Big {
	value := r.text(key)
	normalized, err := unsignedPosition(value)
	r.fail(err)
	if err != nil {
		return nil
	}
	amount, ok := new(decimal.Big).SetString(normalized)
	if !ok {
		r.fail(fmt.Errorf("invalid RFQ amount"))
		return nil
	}
	amount.SetScale(6)
	r.fail(codec.RoundDecimal(amount))
	amount.Context.Precision = decimal.UnlimitedPrecision
	amount.Reduce()
	scale := min(0, 28-amount.Precision()+amount.Scale())
	if amount.Scale() < scale {
		ctx := codec.DecimalContext()
		ctx.Quantize(amount, scale)
	}
	return amount
}
func (r rfqFields) status() string {
	status := r.text("status")
	if !slices.Contains([]string{"AWAITING_REQUESTER_ACCEPTANCE", "AWAITING_MAKER_CONFIRMATION", "EXECUTING", "FILLED", "FAILED", "EXPIRED", "CANCELED", "MATCHED", "MINED", "CONFIRMED", "RETRYING"}, status) {
		r.fail(fmt.Errorf("unknown RFQ status"))
	}
	return status
}
func (r rfqFields) detail() *models.RFQErrorDetail {
	if r.fields["error"] == nil || string(r.fields["error"]) == "null" {
		return nil
	}
	value := r.object("error")
	return &models.RFQErrorDetail{Code: value.text("code"), Message: value.text("message")}
}
func parseRFQStatus(raw []byte, id string) (*models.RFQStatusResult, error) {
	fields := readRFQ(raw)
	actual := fields.text("rfq_id")
	if actual != id {
		fields.fail(fmt.Errorf("RFQ response ID differs from request"))
	}
	result := &models.RFQStatusResult{RFQID: actual, Status: fields.status(), TakerOrderHash: fields.optional("taker_order_hash"), Error: fields.detail()}
	if hash := fields.optional("tx_hash"); hash != nil {
		value := models.TransactionHash(*hash)
		result.TxHash = &value
	}
	if *fields.err != nil {
		return nil, response(*fields.err)
	}
	return result, nil
}
func parseComboQuote(raw []byte, submitted map[string]any) (*models.ComboQuoteResult, error) {
	fields := readRFQ(raw)
	id := fields.text("rfq_id")
	status := fields.status()
	if fields.fields["quote"] == nil || string(fields.fields["quote"]) == "null" {
		if !slices.Contains([]string{"FAILED", "EXPIRED", "CANCELED"}, status) {
			fields.fail(fmt.Errorf("RFQ response has no quote before terminal failure"))
		}
		detail := fields.detail()
		if *fields.err != nil {
			return nil, response(*fields.err)
		}
		if detail != nil && (detail.Code == "NO_QUOTES" || detail.Code == "SIZE_TOO_LARGE") {
			return &models.ComboQuoteResult{RFQID: id, Reason: &detail.Code}, nil
		}
		code, message := "", "The combo quote request failed."
		if detail != nil {
			code, message = detail.Code, detail.Message
		}
		return nil, &sdkerrors.RFQRequestRejectedError{Rejection: &sdkerrors.RequestRejectedError{Status: 200, Code: code, Message: message}}
	}
	if status != "AWAITING_REQUESTER_ACCEPTANCE" {
		fields.fail(fmt.Errorf("quote arrived in an invalid RFQ state"))
	}
	request := fields.object("request")
	direction := request.text("direction")
	side := request.text("side")
	if request.text("rfq_id") != id || direction != fmt.Sprint(submitted["direction"]) || side != fmt.Sprint(submitted["side"]) {
		fields.fail(fmt.Errorf("RFQ identity, direction or side differs from request"))
	}
	legs, err := codec.Required[[]string](request.fields, "leg_position_ids")
	fields.fail(err)
	for _, leg := range legs {
		_, err = unsignedPosition(leg)
		fields.fail(err)
	}
	if !reflect.DeepEqual(legs, submitted["leg_position_ids"]) {
		fields.fail(fmt.Errorf("RFQ legs differ from request"))
	}
	size := request.object("requested_size")
	unit := size.text("unit")
	value := size.position("value_e6")
	expected, ok := submitted["requested_size"].(map[string]any)
	if !ok {
		fields.fail(fmt.Errorf("invalid submitted size"))
	} else if unit != expected["unit"] || value != expected["value_e6"] {
		fields.fail(fmt.Errorf("RFQ size differs from request"))
	}
	if _, err = codec.Condition(request.text("condition_id"), "combo"); err != nil {
		fields.fail(err)
	}
	_ = request.position("no_position_id")
	quote := fields.object("quote")
	builder := fields.text("builder_code")
	fields.fail(validateBuilder(builder))
	expires, err := codec.Required[int64](fields.fields, "expires_at")
	fields.fail(err)
	result := &models.ComboQuote{RFQID: id, QuoteID: quote.text("quote_id"), BuilderCode: builder, Direction: models.OrderSide(direction), PositionID: request.position("yes_position_id"), BlendedPrice: quote.amount("blended_price_e6"), MakerAmount: quote.amount("maker_amount_e6"), TakerAmount: quote.amount("taker_amount_e6"), TotalRequired: quote.amount("total_required_e6"), ExpiresAt: expires}
	if quote.fields["net_receive_e6"] != nil {
		net := quote.amount("net_receive_e6")
		if result.Direction == models.Sell {
			result.NetReceive = net
		}
	} else if result.Direction == models.Sell {
		fields.fail(fmt.Errorf("SELL quote omitted net proceeds"))
	}
	if *fields.err != nil {
		return nil, response(*fields.err)
	}
	return &models.ComboQuoteResult{RFQID: id, Quote: result}, nil
}
