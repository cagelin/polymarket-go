package data

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/models"
	"golang.org/x/crypto/sha3"
)

// GetTradingApprovalsState reads indexed allowances without requesting wallet access or submitting transactions.
func (c *Client) GetTradingApprovalsState(ctx context.Context, wallet models.EVMAddress) (*models.TradingApprovalsState, error) {
	owner, err := checksumAddress(string(wallet))
	if err != nil {
		return nil, input(err.Error())
	}
	return getData(ctx, c, "/v2/approvals", map[string]any{"user": owner}, func(raw []byte) (*models.TradingApprovalsState, error) { return parseApprovals(raw, owner) })
}
func checksumAddress(address string) (string, error) {
	if strings.HasPrefix(address, "0x") || strings.HasPrefix(address, "0X") {
		address = address[2:]
	}
	if len(address) != 40 {
		return "", fmt.Errorf("wallet must be a 20-byte hex address")
	}
	if _, err := hex.DecodeString(address); err != nil {
		return "", fmt.Errorf("invalid wallet: %w", err)
	}
	address = strings.ToLower(address)
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write([]byte(address))
	digest := hex.EncodeToString(hash.Sum(nil))
	result := []byte(address)
	for i, ch := range result {
		if ch >= 'a' && ch <= 'f' && digest[i] >= '8' {
			result[i] = ch - 'a' + 'A'
		}
	}
	return "0x" + string(result), nil
}
func maxUint256() *big.Int {
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
}
func parseApprovals(raw []byte, owner string) (*models.TradingApprovalsState, error) {
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	address, err := codec.Required[string](fields, "address")
	if err != nil || !strings.EqualFold(address, owner) {
		return nil, fmt.Errorf("approval owner differs from requested wallet")
	}
	chain, err := codec.Required[int64](fields, "chain_id")
	if err != nil || chain != 137 {
		return nil, fmt.Errorf("approval chain differs from production")
	}
	rows, err := codec.Required[[]json.RawMessage](fields, "contracts")
	if err != nil {
		return nil, err
	}
	required := RequiredApprovals()
	pairs := map[string]bool{}
	for _, a := range required.ERC20 {
		pairs[strings.ToLower(string(a.TokenAddress)+"/"+string(a.Spender))] = true
	}
	for _, a := range required.ERC1155 {
		pairs[strings.ToLower(string(a.TokenAddress)+"/"+string(a.Operator))] = true
	}
	type indexedApproval struct {
		standard string
		approved bool
		amount   *big.Int
	}
	indexed := map[string]indexedApproval{}
	for _, row := range rows {
		item, parseErr := codec.Object(row)
		if parseErr != nil {
			return nil, parseErr
		}
		token, tokenErr := codec.Required[string](item, "token")
		spender, spenderErr := codec.Required[string](item, "spender")
		if tokenErr != nil || spenderErr != nil {
			continue
		}
		pair := strings.ToLower(token + "/" + spender)
		if !pairs[pair] {
			continue
		}
		standard, standardErr := codec.Required[string](item, "standard")
		approved, approvedErr := codec.Required[bool](item, "approved")
		if standardErr != nil || approvedErr != nil || (standard != "ERC20" && standard != "ERC1155") {
			return nil, fmt.Errorf("invalid approval standard or flag")
		}
		if _, exists := indexed[pair]; exists {
			return nil, fmt.Errorf("duplicate approval pair")
		}
		var amount *big.Int
		if standard == "ERC20" {
			text, amountErr := codec.Required[string](item, "amount")
			if amountErr != nil {
				return nil, amountErr
			}
			if text == "max" {
				amount = maxUint256()
			} else {
				if len(text) < 1 || len(text) > 78 || strings.Trim(text, "0123456789") != "" || (len(text) > 1 && text[0] == '0') {
					return nil, fmt.Errorf("invalid uint256 allowance")
				}
				var valid bool
				amount, valid = new(big.Int).SetString(text, 10)
				if !valid || amount.BitLen() > 256 {
					return nil, fmt.Errorf("allowance exceeds uint256")
				}
			}
		}
		indexed[pair] = indexedApproval{standard: standard, approved: approved, amount: amount}
	}
	missing := models.MissingTradingApprovals{ERC20: []models.ERC20TradingApproval{}, ERC1155: []models.ERC1155TradingApproval{}}
	for _, a := range required.ERC20 {
		row, ok := indexed[strings.ToLower(string(a.TokenAddress)+"/"+string(a.Spender))]
		if !ok || row.standard != "ERC20" || row.amount == nil {
			return nil, fmt.Errorf("missing required ERC20 pair")
		}
		if !row.approved || row.amount.Cmp(a.Amount) < 0 {
			missing.ERC20 = append(missing.ERC20, a)
		}
	}
	for _, a := range required.ERC1155 {
		row, ok := indexed[strings.ToLower(string(a.TokenAddress)+"/"+string(a.Operator))]
		if !ok || row.standard != "ERC1155" {
			return nil, fmt.Errorf("missing required ERC1155 pair")
		}
		if !row.approved {
			missing.ERC1155 = append(missing.ERC1155, a)
		}
	}
	return &models.TradingApprovalsState{Missing: missing, IsFullyApproved: len(missing.ERC20) == 0 && len(missing.ERC1155) == 0}, nil
}
