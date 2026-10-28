package wallet

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/cagelam/polymarket-go/internal/data"
	"github.com/cagelam/polymarket-go/internal/evm"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/transactions"
)

// ApproveERC20 sets a token allowance in base units. Nil amount selects maximum allowance.
func (c *Client) ApproveERC20(ctx context.Context, token, spender models.EVMAddress, amount *big.Int, metadata ...string) (*transactions.Handle, error) {
	label := "max"
	if amount != nil {
		label = amount.String()
	}
	if amount == nil {
		amount = evm.MaxUint256()
	}
	call, err := evm.ERC20Call(string(token), string(spender), amount, true)
	if err != nil {
		return nil, err
	}
	return c.dispatchNamed(ctx, call, fmt.Sprintf("Approve %s of %s to %s", label, call.To, common.HexToAddress(string(spender)).Hex()), metadata)
}

// ApproveERC1155ForAll grants or revokes an operator.
func (c *Client) ApproveERC1155ForAll(ctx context.Context, token, operator models.EVMAddress, approved bool, metadata ...string) (*transactions.Handle, error) {
	call, err := evm.OperatorCall(string(token), string(operator), approved)
	if err != nil {
		return nil, err
	}
	verb := "Approve"
	if !approved {
		verb = "Revoke"
	}
	return c.dispatchNamed(ctx, call, fmt.Sprintf("%s %s on %s", verb, common.HexToAddress(string(operator)).Hex(), call.To), metadata)
}

// TransferERC20 transfers a base-unit amount to an explicit address.
func (c *Client) TransferERC20(ctx context.Context, token, to models.EVMAddress, amount *big.Int, metadata ...string) (*transactions.Handle, error) {
	call, err := evm.ERC20Call(string(token), string(to), amount, false)
	if err != nil {
		return nil, err
	}
	return c.dispatchNamed(ctx, call, fmt.Sprintf("Transfer %s of %s to %s", amount, call.To, common.HexToAddress(string(to)).Hex()), metadata)
}

// SetupTradingApprovals reads live allowances and waits for all missing grants.
func (c *Client) SetupTradingApprovals(ctx context.Context) error {
	required := data.RequiredApprovals()
	queries := make([]models.TransactionCall, 0, len(required.ERC20)+len(required.ERC1155))
	for _, approval := range required.ERC20 {
		call, err := evm.ReadApproval(string(approval.TokenAddress), string(c.config.Wallet), string(approval.Spender), true)
		if err != nil {
			return err
		}
		queries = append(queries, call)
	}
	for _, approval := range required.ERC1155 {
		call, err := evm.ReadApproval(string(approval.TokenAddress), string(c.config.Wallet), string(approval.Operator), false)
		if err != nil {
			return err
		}
		queries = append(queries, call)
	}
	values, err := c.rpc.CallBatch(ctx, queries)
	if err != nil {
		return err
	}
	var calls []models.TransactionCall
	for i, approval := range required.ERC20 {
		decoded, decodeErr := evm.Unpack(values[i], "uint256")
		if decodeErr != nil {
			return decodeErr
		}
		amount, ok := decoded.(*big.Int)
		if !ok {
			return response("invalid allowance return type")
		}
		if amount.Cmp(approval.Amount) < 0 {
			call, callErr := evm.ERC20Call(string(approval.TokenAddress), string(approval.Spender), approval.Amount, true)
			if callErr != nil {
				return callErr
			}
			calls = append(calls, call)
		}
	}
	for i, approval := range required.ERC1155 {
		decoded, decodeErr := evm.Unpack(values[len(required.ERC20)+i], "bool")
		if decodeErr != nil {
			return decodeErr
		}
		approved, ok := decoded.(bool)
		if !ok {
			return response("invalid approval return type")
		}
		if !approved {
			call, callErr := evm.OperatorCall(string(approval.TokenAddress), string(approval.Operator), true)
			if callErr != nil {
				return callErr
			}
			calls = append(calls, call)
		}
	}
	if len(calls) == 0 {
		return nil
	}
	metadata := "Setup trading approvals"
	handle, err := c.ExecuteTransaction(ctx, calls, &metadata)
	if err != nil {
		return err
	}
	_, err = handle.Wait(ctx)
	return err
}

func (c *Client) dispatchNamed(ctx context.Context, call models.TransactionCall, label string, metadata []string) (*transactions.Handle, error) {
	if len(metadata) > 1 {
		return nil, input("at most one metadata string is accepted")
	}
	if len(metadata) == 1 {
		label = metadata[0]
	}
	return c.ExecuteTransaction(ctx, []models.TransactionCall{call}, &label)
}
