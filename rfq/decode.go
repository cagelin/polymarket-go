package rfq

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

func parseEvent(raw []byte) (Event, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return Event{}, err
	}
	kind, err := codec.Required[string](fields, "type")
	if err != nil {
		return Event{}, err
	}
	target := map[string]string{"RFQ_REQUEST": "quote_request", "RFQ_CONFIRMATION_REQUEST": "confirmation_request", "RFQ_EXECUTION_UPDATE": "execution_update", "RFQ_TRADE": "trade"}[kind]
	if target == "" {
		return Event{}, fmt.Errorf("unknown RFQ event")
	}
	var event Event
	if err = json.Unmarshal(raw, &event); err != nil {
		return Event{}, err
	}
	event.Type = target
	required := []string{"rfq_id"}
	if kind == "RFQ_EXECUTION_UPDATE" {
		if !slices.Contains([]string{"MATCHED", "MINED", "CONFIRMED", "RETRYING", "FAILED"}, event.Status) {
			return Event{}, fmt.Errorf("invalid execution status")
		}
	} else {
		if event.Direction != models.Buy && event.Direction != models.Sell || event.Side != "YES" {
			return Event{}, fmt.Errorf("invalid RFQ direction or side")
		}
		condition, conditionErr := codec.Condition(string(event.ConditionID), "combo")
		if conditionErr != nil {
			return Event{}, conditionErr
		}
		event.ConditionID = models.ConditionID(condition)
		if _, err = codec.Required[[]string](fields, "leg_position_ids"); err != nil {
			return Event{}, err
		}
		if string(fields["leg_position_ids"]) == "null" {
			return Event{}, fmt.Errorf("legs must be an array")
		}
		switch kind {
		case "RFQ_REQUEST":
			required = append(required, "requestor_public_id", "yes_position_id", "no_position_id")
			event.SubmissionDeadline, err = codec.Required[int64](fields, "submission_deadline")
			if err != nil {
				return Event{}, err
			}
			size, sizeErr := codec.Object(fields["requested_size"])
			if sizeErr != nil {
				return Event{}, sizeErr
			}
			unit, sizeErr := codec.Required[string](size, "unit")
			if sizeErr != nil {
				return Event{}, sizeErr
			}
			if unit != "shares" && unit != "notional" {
				return Event{}, fmt.Errorf("invalid RFQ size unit")
			}
			amount, sizeErr := e6(size, "value_e6")
			if sizeErr != nil {
				return Event{}, sizeErr
			}
			event.RequestedSize = RequestedSize{Unit: unit, Value: amount}
		case "RFQ_CONFIRMATION_REQUEST":
			required = append(required, "quote_id", "signer_address", "maker_address", "yes_position_id", "no_position_id")
			event.ConfirmBy, err = codec.Required[int64](fields, "confirm_by")
			if err != nil {
				return Event{}, err
			}
			event.SignatureType, err = codec.Required[int](fields, "signature_type")
			if err != nil {
				return Event{}, err
			}
			event.FillSize, err = e6(fields, "fill_size_e6")
			if err != nil {
				return Event{}, err
			}
			event.Price, err = e6(fields, "price_e6")
			if err != nil {
				return Event{}, err
			}
		case "RFQ_TRADE":
			required = append(required, "requester_id")
			event.ExecutedAt, err = codec.Required[int64](fields, "executed_at")
			if err != nil {
				return Event{}, err
			}
			event.Price, err = e6(fields, "price_e6")
			if err != nil {
				return Event{}, err
			}
			event.Size, err = e6(fields, "size_e6")
			if err != nil {
				return Event{}, err
			}
		}
	}
	for _, name := range required {
		value, fieldErr := codec.Required[string](fields, name)
		if fieldErr != nil {
			return Event{}, fieldErr
		}
		if value == "" {
			return Event{}, fmt.Errorf("empty RFQ %s", name)
		}
	}
	return event, nil
}
func e6(fields map[string]json.RawMessage, name string) (*decimal.Big, error) {
	text, err := codec.Required[string](fields, name)
	if err != nil {
		return nil, err
	}
	text = codec.ASCIIDigits(text)
	if text == "" || strings.Trim(text, "0123456789") != "" {
		return nil, fmt.Errorf("invalid RFQ unsigned amount")
	}
	amount, ok := new(decimal.Big).SetString(text)
	if !ok {
		return nil, fmt.Errorf("invalid decimal")
	}
	result := new(decimal.Big)
	if err = codec.DecimalArithmetic("mul", result, amount, decimal.New(1, 6)); err != nil {
		return nil, err
	}
	return result, nil
}
