package wallet

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/cagelin/polymarket-go/internal/chain"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/evm"
	"github.com/cagelin/polymarket-go/internal/gamma"
	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/transactions"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const collateralToken = chain.Collateral
const collateralAdapter = chain.CollateralAdapter
const negRiskCollateralAdapter = chain.NegRiskCollateralAdapter

const conditionalTokens = chain.ConditionalTokens
const negRiskAdapter = chain.NegRiskAdapter
const positionManager = chain.PositionManager
const router = chain.Router
const combinatorialModule = chain.CombinatorialModule

type positionContext struct {
	condition, token, adapter string
	v2                        bool
	ids                       []*big.Int
	prepare                   *models.TransactionCall
}

func input(message string) error    { return &sdkerrors.UserInputError{Message: message} }
func response(message string) error { return &sdkerrors.UnexpectedResponseError{Message: message} }
func positionInteger(id string) (*big.Int, error) {
	value, ok := new(big.Int).SetString(strings.TrimSpace(id), 10)
	if !ok {
		return nil, input("position ID must be a uint256")
	}
	return evm.Uint256(value)
}
func v2Position(id string) (positionContext, int, error) {
	value, err := chain.ParsePositionID(id)
	if err != nil {
		return positionContext{}, 0, err
	}
	raw := value.FillBytes(make([]byte, 32))
	if raw[0] < 1 || raw[0] > 3 || raw[31] > 1 {
		return positionContext{}, 0, input("position ID must identify a V2 binary outcome")
	}
	base := new(big.Int).And(value, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(256)))
	return positionContext{condition: "0x" + hex.EncodeToString(raw[:31]), v2: true, token: positionManager, ids: []*big.Int{base, new(big.Int).Add(base, big.NewInt(1))}}, int(raw[31]), nil
}
func comboContext(legs []string) (positionContext, error) {
	if len(legs) < 1 || len(legs) > 50 {
		return positionContext{}, input("combo requires 1 to 50 legs")
	}
	ids := make([]*big.Int, len(legs))
	for i, leg := range legs {
		value, err := positionInteger(leg)
		if err != nil {
			return positionContext{}, err
		}
		raw := value.FillBytes(make([]byte, 32))
		if raw[0] < 1 || raw[0] > 2 || raw[31] > 1 {
			return positionContext{}, input("combo legs must be binary or neg-risk outcome IDs")
		}
		ids[i] = value
	}
	slices.SortFunc(ids, func(a, b *big.Int) int { return a.Cmp(b) })
	for i := 1; i < len(ids); i++ {
		if new(big.Int).Rsh(new(big.Int).Set(ids[i-1]), 8).Cmp(new(big.Int).Rsh(new(big.Int).Set(ids[i]), 8)) == 0 {
			return positionContext{}, input("combo legs must have distinct conditions")
		}
	}
	encoded, err := (abi.Arguments{{Type: positionArrayType}}).Pack(ids)
	if err != nil {
		return positionContext{}, err
	}
	raw, err := (abi.Arguments{{Type: positionUintType}, {Type: positionBytesType}}).Pack(big.NewInt(3), encoded)
	if err != nil {
		return positionContext{}, err
	}
	hash := crypto.Keccak256(raw)
	condition := "0x03" + hex.EncodeToString(hash[16:]) + strings.Repeat("0", 28)
	base, ok := new(big.Int).SetString(condition[2:]+"00", 16)
	if !ok {
		return positionContext{}, input("invalid combo condition")
	}
	call, err := evm.Call(combinatorialModule, "prepareCondition(uint256[])", []string{"uint256[]"}, ids)
	if err != nil {
		return positionContext{}, err
	}
	return positionContext{condition: condition, token: positionManager, v2: true, ids: []*big.Int{base, new(big.Int).Add(base, big.NewInt(1))}, prepare: &call}, nil
}
func (c *Client) marketContext(ctx context.Context, condition, market string, closed *bool) (positionContext, error) {
	if (condition == "") == (market == "") {
		return positionContext{}, input("provide exactly one condition ID or market ID")
	}
	size := 1
	// Gamma defaults to V1 when no version filter is supplied.
	p := gamma.ListMarketsParams{PageSize: &size, Closed: closed, Versions: []models.ProtocolVersion{models.ProtocolV1, models.ProtocolV2}}
	if condition != "" {
		p.ConditionIDs = []string{condition}
	} else {
		value, err := strconv.ParseInt(strings.TrimSpace(market), 10, 64)
		if err != nil || value <= 0 {
			return positionContext{}, input("market ID must be positive")
		}
		p.IDs = []int64{value}
	}
	pages, err := gamma.New(c.transport).ListMarkets(p)
	if err != nil {
		return positionContext{}, err
	}
	page, err := pages.FirstPage(ctx)
	if err != nil {
		return positionContext{}, err
	}
	if len(page.Items) != 1 {
		return positionContext{}, input("expected exactly one matching market")
	}
	m := page.Items[0]
	if m.ConditionID == nil || m.Version == nil {
		return positionContext{}, response("market is missing condition or protocol version")
	}
	result := positionContext{condition: string(*m.ConditionID), v2: *m.Version == "v2", token: conditionalTokens, adapter: collateralAdapter}
	outcomes := []models.MarketOutcome{m.Outcomes.Yes, m.Outcomes.No}
	if result.v2 {
		result.token = positionManager
	} else {
		if m.State.NegRisk == nil {
			return positionContext{}, response("CTF market is missing negative-risk flag")
		}
		if *m.State.NegRisk {
			result.token, result.adapter = negRiskAdapter, negRiskCollateralAdapter
		}
	}
	for _, outcome := range outcomes {
		var text string
		if result.v2 {
			if outcome.PositionID == nil {
				return positionContext{}, response("market is missing a position ID")
			}
			text = string(*outcome.PositionID)
		} else {
			if outcome.TokenID == nil {
				return positionContext{}, response("market is missing a token ID")
			}
			text = string(*outcome.TokenID)
		}
		value, parseErr := positionInteger(text)
		if parseErr != nil {
			return positionContext{}, parseErr
		}
		result.ids = append(result.ids, value)
	}
	return result, nil
}
func (p positionContext) call(action string, amount *big.Int, outcome int) (models.TransactionCall, error) {
	raw, err := evm.Hex(p.condition)
	if err != nil {
		return models.TransactionCall{}, err
	}
	if p.v2 {
		if len(raw) == 32 && (raw[31] == 0 || raw[31] == 1) {
			raw = raw[:31]
		}
		if len(raw) != 31 {
			return models.TransactionCall{}, input("V2 condition must be bytes31")
		}
		var condition [31]byte
		copy(condition[:], raw)
		value, amountErr := evm.Uint256(amount)
		if amountErr != nil {
			return models.TransactionCall{}, amountErr
		}
		if action == "redeem" {
			return evm.Call(router, "redeem(bytes31,uint256,uint256)", []string{"bytes31", "uint256", "uint256"}, condition, big.NewInt(int64(outcome)), value)
		}
		return evm.Call(router, action+"(bytes31,uint256)", []string{"bytes31", "uint256"}, condition, value)
	}
	if len(raw) != 32 {
		return models.TransactionCall{}, input("CTF condition must be bytes32")
	}
	var condition [32]byte
	copy(condition[:], raw)
	args := []any{common.HexToAddress(collateralToken), [32]byte{}, condition, []*big.Int{big.NewInt(1), big.NewInt(2)}}
	types := []string{"address", "bytes32", "bytes32", "uint256[]"}
	if action == "redeem" {
		return evm.Call(p.adapter, "redeemPositions(address,bytes32,bytes32,uint256[])", types, args...)
	}
	value, err := evm.Uint256(amount)
	if err != nil {
		return models.TransactionCall{}, err
	}
	args = append(args, value)
	types = append(types, "uint256")
	name := "splitPosition"
	if action == "merge" {
		name = "mergePositions"
	}
	return evm.Call(p.adapter, name+"(address,bytes32,bytes32,uint256[],uint256)", types, args...)
}
func (c *Client) balances(ctx context.Context, p positionContext) ([]*big.Int, error) {
	owner, err := evm.Address(string(c.config.Wallet))
	if err != nil {
		return nil, err
	}
	owners := make([]common.Address, len(p.ids))
	for i := range owners {
		owners[i] = owner
	}
	call, err := evm.Call(p.token, "balanceOfBatch(address[],uint256[])", []string{"address[]", "uint256[]"}, owners, p.ids)
	if err != nil {
		return nil, err
	}
	raw, err := c.rpc.Call(ctx, call)
	if err != nil {
		return nil, err
	}
	decoded, err := evm.Unpack(raw, "uint256[]")
	if err != nil {
		return nil, err
	}
	values, ok := decoded.([]*big.Int)
	if !ok {
		return nil, response("invalid balance array return type")
	}
	if len(values) != len(p.ids) {
		return nil, response("position balance count mismatch")
	}
	return values, nil
}
func (c *Client) mergeCall(ctx context.Context, p positionContext, amount *big.Int) (models.TransactionCall, error) {
	balances, err := c.balances(ctx, p)
	if err != nil {
		return models.TransactionCall{}, err
	}
	if len(balances) != 2 {
		return models.TransactionCall{}, response("expected two position balances")
	}
	limit := balances[0]
	if balances[1].Cmp(limit) < 0 {
		limit = balances[1]
	}
	if amount == nil {
		amount = limit
	}
	if amount.Sign() <= 0 || amount.Cmp(limit) > 0 {
		return models.TransactionCall{}, input("merge amount must be positive and within complementary balances")
	}
	return p.call("merge", amount, 0)
}

// SplitPosition splits collateral into a market's two outcomes or a prepared combo.
func (c *Client) SplitPosition(ctx context.Context, p models.SplitPositionParams) (*transactions.Handle, error) {
	if (p.ConditionID == "") == (p.Legs == nil) {
		return nil, input("provide exactly one condition ID or legs")
	}
	var position positionContext
	var err error
	if p.Legs != nil {
		if p.Amount == nil || p.Amount.Sign() <= 0 {
			return nil, input("combo split amount must be positive")
		}
		position, err = comboContext(p.Legs)
	} else {
		position, err = c.marketContext(ctx, string(p.ConditionID), "", nil)
	}
	if err != nil {
		return nil, err
	}
	call, err := position.call("split", p.Amount, 0)
	if err != nil {
		return nil, err
	}
	var calls []models.TransactionCall
	if position.prepare != nil {
		calls = append(calls, *position.prepare)
	}
	calls = append(calls, call)
	metadata := p.Metadata
	if metadata == nil {
		label := fmt.Sprintf("Split %s positions for condition %s", p.Amount, position.condition)
		if position.prepare != nil {
			label = fmt.Sprintf("Split %s combo positions for condition %s", p.Amount, position.condition)
		}
		metadata = &label
	}
	return c.ExecuteTransaction(ctx, calls, metadata)
}

// MergePositions merges complementary outcomes; nil Amount selects the full balanced amount.
func (c *Client) MergePositions(ctx context.Context, p models.MergePositionsParams) (*transactions.Handle, error) {
	if (p.ConditionID == "") == (p.Legs == nil) {
		return nil, input("provide exactly one condition ID or legs")
	}
	var position positionContext
	var err error
	if p.Legs != nil {
		position, err = comboContext(p.Legs)
	} else {
		position, err = c.marketContext(ctx, string(p.ConditionID), "", nil)
	}
	if err != nil {
		return nil, err
	}
	call, err := c.mergeCall(ctx, position, p.Amount)
	if err != nil {
		return nil, err
	}
	var calls []models.TransactionCall
	if position.prepare != nil {
		calls = append(calls, *position.prepare)
	}
	calls = append(calls, call)
	metadata := p.Metadata
	if metadata == nil {
		label := fmt.Sprintf("Merge positions for condition %s", position.condition)
		metadata = &label
	}
	return c.ExecuteTransaction(ctx, calls, metadata)
}

// MergeMultiplePositions validates the full batch before submitting any transaction.
func (c *Client) MergeMultiplePositions(ctx context.Context, requests []models.MergePositionRequest, metadata *string) (*transactions.Handle, error) {
	if len(requests) == 0 {
		return nil, input("at least one merge request is required")
	}
	combo := requests[0].PositionID != ""
	seen := map[string]bool{}
	calls := make([]models.TransactionCall, 0, len(requests))
	for _, request := range requests {
		if (request.PositionID != "") != combo {
			return nil, input("cannot mix market and V2 position IDs")
		}
		var p positionContext
		var err error
		if combo {
			if request.ConditionID != "" || request.MarketID != "" {
				return nil, input("provide exactly one identifier")
			}
			p, _, err = v2Position(string(request.PositionID))
		} else {
			p, err = c.marketContext(ctx, string(request.ConditionID), string(request.MarketID), nil)
		}
		if err != nil {
			return nil, err
		}
		condition := strings.ToLower(p.condition)
		if seen[condition] {
			return nil, input("merge requests must have distinct conditions")
		}
		seen[condition] = true
		call, callErr := c.mergeCall(ctx, p, request.Amount)
		if callErr != nil {
			return nil, callErr
		}
		calls = append(calls, call)
	}
	if metadata == nil {
		label := fmt.Sprintf("Merge %d positions", len(calls))
		if combo {
			label = fmt.Sprintf("Merge %d Polymarket V2 positions", len(calls))
		}
		metadata = &label
	}
	return c.ExecuteTransaction(ctx, calls, metadata)
}

// RedeemPositions redeems a resolved market or one V2 outcome's current balance.
func (c *Client) RedeemPositions(ctx context.Context, p models.RedeemPositionsParams) (*transactions.Handle, error) {
	var position positionContext
	var err error
	outcome := -1
	if p.PositionID != "" {
		if p.ConditionID != "" || p.MarketID != "" {
			return nil, input("provide exactly one identifier")
		}
		position, outcome, err = v2Position(string(p.PositionID))
	} else {
		closed := true
		position, err = c.marketContext(ctx, string(p.ConditionID), string(p.MarketID), &closed)
	}
	if err != nil {
		return nil, err
	}
	var calls []models.TransactionCall
	if !position.v2 {
		call, callErr := position.call("redeem", nil, 0)
		if callErr != nil {
			return nil, callErr
		}
		calls = append(calls, call)
	} else {
		var balances []*big.Int
		var balanceErr error
		if outcome >= 0 {
			owner := common.HexToAddress(string(c.config.Wallet))
			call, callErr := evm.Call(position.token, "balanceOf(address,uint256)", []string{"address", "uint256"}, owner, position.ids[outcome])
			if callErr != nil {
				return nil, callErr
			}
			raw, readErr := c.rpc.Call(ctx, call)
			if readErr != nil {
				return nil, readErr
			}
			decoded, decodeErr := evm.Unpack(raw, "uint256")
			if decodeErr != nil {
				return nil, decodeErr
			}
			value, ok := decoded.(*big.Int)
			if !ok {
				return nil, response("invalid balance return type")
			}
			switch outcome {
			case 0:
				balances = []*big.Int{value, new(big.Int)}
			case 1:
				balances = []*big.Int{new(big.Int), value}
			default:
				return nil, input("invalid outcome index")
			}
		} else {
			balances, balanceErr = c.balances(ctx, position)
		}
		if balanceErr != nil {
			return nil, balanceErr
		}
		for i, amount := range balances {
			if (outcome < 0 || outcome == i) && amount.Sign() > 0 {
				call, callErr := position.call("redeem", amount, i)
				if callErr != nil {
					return nil, callErr
				}
				calls = append(calls, call)
			}
		}
		if len(calls) == 0 {
			return nil, input("positions have no balance to redeem")
		}
	}
	metadata := p.Metadata
	if metadata == nil {
		label := "Redeem positions for condition " + position.condition
		if p.PositionID != "" {
			label = "Redeem position " + string(p.PositionID)
		}
		metadata = &label
	}
	return c.ExecuteTransaction(ctx, calls, metadata)
}

var (
	positionArrayType = positionABIType("uint256[]")
	positionUintType  = positionABIType("uint256")
	positionBytesType = positionABIType("bytes")
)

func positionABIType(name string) abi.Type {
	parsed, err := abi.NewType(name, "", nil)
	if err != nil {
		panic(fmt.Sprintf("invalid embedded position ABI type: %v", err))
	}
	return parsed
}
