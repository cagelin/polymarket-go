package models

import (
	"encoding/json"
	"fmt"

	"github.com/cagelin/polymarket-go/internal/codec"
)

// Activity is one of the concrete public response variants. Use a type switch to inspect it.
type Activity interface{ activityVariant() }

func (*TradeActivity) activityVariant()          {}
func (*SplitActivity) activityVariant()          {}
func (*MergeActivity) activityVariant()          {}
func (*RedeemActivity) activityVariant()         {}
func (*ConversionActivity) activityVariant()     {}
func (*RewardActivity) activityVariant()         {}
func (*DepositActivity) activityVariant()        {}
func (*WithdrawalActivity) activityVariant()     {}
func (*MakerRebateActivity) activityVariant()    {}
func (*TakerRebateActivity) activityVariant()    {}
func (*ReferralRewardActivity) activityVariant() {}
func (*YieldActivity) activityVariant()          {}
func (*MigrationActivity) activityVariant()      {}
func (*TipActivity) activityVariant()            {}
func (*ComboTradeActivity) activityVariant()     {}
func (*UnknownActivity) activityVariant()        {}

// ParseActivities validates and decodes the tagged response variants.
func ParseActivities(raw []byte) ([]Activity, error) {
	items, err := codec.ParseList[json.RawMessage](raw)
	if err != nil {
		return nil, err
	}
	result := make([]Activity, 0, len(items))
	for _, item := range items {
		value, parseErr := ParseActivity(item)
		if parseErr != nil {
			return nil, parseErr
		}
		result = append(result, value)
	}
	return result, nil
}

// ParseActivity preserves the concrete response variant.
func ParseActivity(raw []byte) (Activity, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	var kind string
	_ = json.Unmarshal(fields["type"], &kind)
	var result Activity
	switch kind {
	case "TRADE":
		result = &TradeActivity{}
	case "SPLIT":
		result = &SplitActivity{}
	case "MERGE":
		result = &MergeActivity{}
	case "REDEEM":
		result = &RedeemActivity{}
	case "CONVERSION":
		result = &ConversionActivity{}
	case "REWARD":
		result = &RewardActivity{}
	case "DEPOSIT":
		result = &DepositActivity{}
	case "WITHDRAWAL":
		result = &WithdrawalActivity{}
	case "MAKER_REBATE":
		result = &MakerRebateActivity{}
	case "TAKER_REBATE":
		result = &TakerRebateActivity{}
	case "REFERRAL_REWARD":
		result = &ReferralRewardActivity{}
	case "YIELD":
		result = &YieldActivity{}
	case "MIGRATION":
		result = &MigrationActivity{}
	case "TIP":
		result = &TipActivity{}
	default:
		original, encodeErr := json.Marshal(fields)
		if encodeErr != nil {
			return nil, encodeErr
		}
		fields["raw"] = original
		fields["type"], err = json.Marshal(kind)
		if err != nil {
			return nil, err
		}
		raw, err = json.Marshal(fields)
		if err != nil {
			return nil, err
		}
		result = &UnknownActivity{}
	}
	if kind == "TRADE" && string(fields["is_combo"]) == "true" {
		result = &ComboTradeActivity{}
	}
	if err = json.Unmarshal(raw, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ComboActivity is one of the concrete public response variants. Use a type switch to inspect it.
type ComboActivity interface{ comboActivityVariant() }

func (*ComboSplitActivity) comboActivityVariant()    {}
func (*ComboMergeActivity) comboActivityVariant()    {}
func (*ComboConvertActivity) comboActivityVariant()  {}
func (*ComboCompressActivity) comboActivityVariant() {}
func (*ComboWrapActivity) comboActivityVariant()     {}
func (*ComboUnwrapActivity) comboActivityVariant()   {}
func (*ComboRedeemActivity) comboActivityVariant()   {}

// ParseComboActivities validates and decodes the tagged response variants.
func ParseComboActivities(raw []byte) ([]ComboActivity, error) {
	items, err := codec.ParseList[json.RawMessage](raw)
	if err != nil {
		return nil, err
	}
	result := make([]ComboActivity, 0, len(items))
	for _, item := range items {
		value, parseErr := ParseComboActivity(item)
		if parseErr != nil {
			return nil, parseErr
		}
		result = append(result, value)
	}
	return result, nil
}

// ParseComboActivity preserves the concrete response variant.
func ParseComboActivity(raw []byte) (ComboActivity, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	var kind string
	_ = json.Unmarshal(fields["type"], &kind)
	var result ComboActivity
	switch kind {
	case "SPLIT":
		result = &ComboSplitActivity{}
	case "MERGE":
		result = &ComboMergeActivity{}
	case "CONVERT":
		result = &ComboConvertActivity{}
	case "COMPRESS":
		result = &ComboCompressActivity{}
	case "WRAP":
		result = &ComboWrapActivity{}
	case "UNWRAP":
		result = &ComboUnwrapActivity{}
	case "REDEEM":
		result = &ComboRedeemActivity{}
	default:
		return nil, fmt.Errorf("unknown ComboActivity variant %q", kind)
	}
	if err = json.Unmarshal(raw, result); err != nil {
		return nil, err
	}
	return result, nil
}

// BiggestWinner is one of the concrete public response variants. Use a type switch to inspect it.
type BiggestWinner interface{ biggestWinnerVariant() }

func (*MarketBiggestWinner) biggestWinnerVariant() {}
func (*ComboBiggestWinner) biggestWinnerVariant()  {}

// ParseBiggestWinners validates and decodes the tagged response variants.
func ParseBiggestWinners(raw []byte) ([]BiggestWinner, error) {
	items, err := codec.ParseList[json.RawMessage](raw)
	if err != nil {
		return nil, err
	}
	result := make([]BiggestWinner, 0, len(items))
	for _, item := range items {
		value, parseErr := ParseBiggestWinner(item)
		if parseErr != nil {
			return nil, parseErr
		}
		result = append(result, value)
	}
	return result, nil
}

// ParseBiggestWinner preserves the concrete response variant.
func ParseBiggestWinner(raw []byte) (BiggestWinner, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	var kind string
	_ = json.Unmarshal(fields["kind"], &kind)
	var result BiggestWinner
	switch kind {
	case "market":
		result = &MarketBiggestWinner{}
	case "combo":
		result = &ComboBiggestWinner{}
	default:
		return nil, fmt.Errorf("unknown BiggestWinner variant %q", kind)
	}
	if err = json.Unmarshal(raw, result); err != nil {
		return nil, err
	}
	return result, nil
}
