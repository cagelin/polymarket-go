// Package evm implements ABI encoding and JSON-RPC for wallet operations.
package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/models"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Uint256 validates and copies a nonnegative EVM integer.
func Uint256(value *big.Int) (*big.Int, error) {
	if value == nil || value.Sign() < 0 || value.BitLen() > 256 {
		return nil, &sdkerrors.UserInputError{Message: "amount must be a uint256"}
	}
	return new(big.Int).Set(value), nil
}

// MaxUint256 returns a fresh maximum allowance.
func MaxUint256() *big.Int {
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
}

// Address validates an address without silently padding or truncating it.
func Address(value string) (common.Address, error) {
	if !common.IsHexAddress(value) {
		return common.Address{}, &sdkerrors.UserInputError{Message: "invalid EVM address"}
	}
	return common.HexToAddress(value), nil
}

// Hex decodes an even-length hex string with an optional prefix.
func Hex(value string) ([]byte, error) {
	raw, err := hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X"))
	if err != nil {
		return nil, &sdkerrors.UserInputError{Message: "invalid hex data"}
	}
	return raw, nil
}

// Pack encodes a method selector and its ABI arguments.
func Pack(signature string, types []string, values ...any) (string, error) {
	args, err := arguments(types)
	if err != nil {
		return "", err
	}
	raw, err := args.Pack(values...)
	if err != nil {
		return "", &sdkerrors.UserInputError{Message: fmt.Sprintf("invalid %s arguments: %v", signature, err)}
	}
	return "0x" + hex.EncodeToString(append(crypto.Keccak256([]byte(signature))[:4], raw...)), nil
}
func arguments(types []string) (abi.Arguments, error) {
	args := make(abi.Arguments, len(types))
	for i, name := range types {
		value, err := abi.NewType(name, "", nil)
		if err != nil {
			return nil, err
		}
		args[i] = abi.Argument{Type: value}
	}
	return args, nil
}

// Unpack validates one ABI return value.
func Unpack(data, kind string) (any, error) {
	raw, err := Hex(data)
	if err != nil {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid RPC hex data", Cause: err}
	}
	args, err := arguments([]string{kind})
	if err != nil {
		return nil, err
	}
	values, err := args.Unpack(raw)
	if err != nil || len(values) != 1 {
		return nil, &sdkerrors.UnexpectedResponseError{Message: "invalid ABI return data", Cause: err}
	}
	return values[0], nil
}

// Call builds a validated call to an explicit address.
func Call(to, signature string, types []string, values ...any) (models.TransactionCall, error) {
	address, err := Address(to)
	if err != nil {
		return models.TransactionCall{}, err
	}
	data, err := Pack(signature, types, values...)
	return models.TransactionCall{To: models.EVMAddress(address.Hex()), Data: data}, err
}

// ERC20Call builds approval or transfer calldata.
func ERC20Call(token, target string, amount *big.Int, approve bool) (models.TransactionCall, error) {
	address, err := Address(target)
	if err != nil {
		return models.TransactionCall{}, err
	}
	value, err := Uint256(amount)
	if err != nil {
		return models.TransactionCall{}, err
	}
	signature := "transfer(address,uint256)"
	if approve {
		signature = "approve(address,uint256)"
	}
	return Call(token, signature, []string{"address", "uint256"}, address, value)
}

// OperatorCall grants or revokes an ERC-1155 operator.
func OperatorCall(token, operator string, approved bool) (models.TransactionCall, error) {
	address, err := Address(operator)
	if err != nil {
		return models.TransactionCall{}, err
	}
	return Call(token, "setApprovalForAll(address,bool)", []string{"address", "bool"}, address, approved)
}

// ReadApproval builds the allowance or operator query.
func ReadApproval(token, owner, spender string, erc20 bool) (models.TransactionCall, error) {
	from, err := Address(owner)
	if err != nil {
		return models.TransactionCall{}, err
	}
	to, err := Address(spender)
	if err != nil {
		return models.TransactionCall{}, err
	}
	signature := "isApprovedForAll(address,address)"
	if erc20 {
		signature = "allowance(address,address)"
	}
	return Call(token, signature, []string{"address", "address"}, from, to)
}

// SafeMultisend encodes ordered calls as a Safe delegate-call payload.
func SafeMultisend(calls []models.TransactionCall) (string, error) {
	if len(calls) == 0 {
		return "", &sdkerrors.UserInputError{Message: "at least one call is required"}
	}
	var inner []byte
	for _, call := range calls {
		address, err := Address(string(call.To))
		if err != nil {
			return "", err
		}
		data, err := Hex(call.Data)
		if err != nil {
			return "", err
		}
		value := call.Value
		if value == nil {
			value = new(big.Int)
		}
		value, err = Uint256(value)
		if err != nil {
			return "", err
		}
		inner = append(inner, 0)
		inner = append(inner, address.Bytes()...)
		inner = append(inner, value.FillBytes(make([]byte, 32))...)
		inner = append(inner, new(big.Int).SetInt64(int64(len(data))).FillBytes(make([]byte, 32))...)
		inner = append(inner, data...)
	}
	return Pack("multiSend(bytes)", []string{"bytes"}, inner)
}

// ProxyCalls encodes ordered calls for the legacy proxy factory.
func ProxyCalls(calls []models.TransactionCall) (string, error) {
	if len(calls) == 0 {
		return "", &sdkerrors.UserInputError{Message: "at least one call is required"}
	}
	tuple, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{{Name: "kind", Type: "uint8"}, {Name: "to", Type: "address"}, {Name: "value", Type: "uint256"}, {Name: "data", Type: "bytes"}})
	if err != nil {
		return "", err
	}
	type item struct {
		Kind  uint8
		To    common.Address
		Value *big.Int
		Data  []byte
	}
	values := make([]item, len(calls))
	for i, call := range calls {
		to, addressErr := Address(string(call.To))
		if addressErr != nil {
			return "", addressErr
		}
		data, dataErr := Hex(call.Data)
		if dataErr != nil {
			return "", dataErr
		}
		value := call.Value
		if value == nil {
			value = new(big.Int)
		}
		amount, amountErr := Uint256(value)
		if amountErr != nil {
			return "", amountErr
		}
		values[i] = item{Kind: 1, To: to, Value: amount, Data: data}
	}
	raw, err := (abi.Arguments{{Type: tuple}}).Pack(values)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(append(crypto.Keccak256([]byte("proxy((uint8,address,uint256,bytes)[])"))[:4], raw...)), nil
}
