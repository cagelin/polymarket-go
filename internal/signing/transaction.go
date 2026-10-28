package signing

import (
	"encoding/hex"
	"math/big"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/evm"
	"github.com/cagelam/polymarket-go/models"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// SignLegacyTransaction serializes the pinned SDK's Polygon EIP-155 transaction.
func (s *Signer) SignLegacyTransaction(call models.TransactionCall, nonce, gas uint64, gasPrice *big.Int) (string, error) {
	to, err := evm.Address(string(call.To))
	if err != nil {
		return "", err
	}
	data, err := evm.Hex(call.Data)
	if err != nil {
		return "", err
	}
	value := call.Value
	if value == nil {
		value = new(big.Int)
	}
	value, err = evm.Uint256(value)
	if err != nil {
		return "", err
	}
	price, err := evm.Uint256(gasPrice)
	if err != nil {
		return "", err
	}
	tx := types.NewTx(&types.LegacyTx{Nonce: nonce, GasPrice: price, Gas: gas, To: &to, Value: value, Data: data})
	signed, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(137)), s.key)
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	raw, err := signed.MarshalBinary()
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	return "0x" + hex.EncodeToString(raw), nil
}

// SignDepositBatch signs ordered calls and wraps the signer authority byte.
func (s *Signer) SignDepositBatch(wallet string, calls []models.TransactionCall, nonce, deadline string, authority models.SignerType) (string, error) {
	messageCalls := make([]any, len(calls))
	for i, call := range calls {
		data, err := evm.Hex(call.Data)
		if err != nil {
			return "", err
		}
		value := call.Value
		if value == nil {
			value = new(big.Int)
		}
		messageCalls[i] = map[string]any{"target": string(call.To), "value": value, "data": data}
	}
	typed := apitypes.TypedData{Types: apitypes.Types{
		"EIP712Domain": {{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}},
		"Batch":        {{Name: "wallet", Type: "address"}, {Name: "nonce", Type: "uint256"}, {Name: "deadline", Type: "uint256"}, {Name: "calls", Type: "Call[]"}},
		"Call":         {{Name: "target", Type: "address"}, {Name: "value", Type: "uint256"}, {Name: "data", Type: "bytes"}},
	}, PrimaryType: "Batch", Domain: apitypes.TypedDataDomain{Name: "DepositWallet", Version: "1", ChainId: (*math.HexOrDecimal256)(big.NewInt(137)), VerifyingContract: wallet}, Message: apitypes.TypedDataMessage{"wallet": wallet, "nonce": nonce, "deadline": deadline, "calls": messageCalls}}
	signature, err := s.TypedData(typed)
	if err != nil {
		return "", err
	}
	return wrapDepositSignature(signature, s.Address(), authority)
}

// SignSafe signs the typed digest as an Ethereum personal message, using Safe's recovery byte.
func (s *Signer) SignSafe(wallet, to, data string, value *big.Int, operation int, nonce string) (string, error) {
	raw, err := evm.Hex(data)
	if err != nil {
		return "", err
	}
	if value == nil {
		value = new(big.Int)
	}
	zero := common.Address{}.Hex()
	typed := apitypes.TypedData{Types: apitypes.Types{
		"EIP712Domain": {{Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}},
		"SafeTx":       {{Name: "to", Type: "address"}, {Name: "value", Type: "uint256"}, {Name: "data", Type: "bytes"}, {Name: "operation", Type: "uint8"}, {Name: "safeTxGas", Type: "uint256"}, {Name: "baseGas", Type: "uint256"}, {Name: "gasPrice", Type: "uint256"}, {Name: "gasToken", Type: "address"}, {Name: "refundReceiver", Type: "address"}, {Name: "nonce", Type: "uint256"}},
	}, PrimaryType: "SafeTx", Domain: apitypes.TypedDataDomain{ChainId: (*math.HexOrDecimal256)(big.NewInt(137)), VerifyingContract: wallet}, Message: apitypes.TypedDataMessage{"to": to, "value": value, "data": raw, "operation": big.NewInt(int64(operation)), "safeTxGas": "0", "baseGas": "0", "gasPrice": "0", "gasToken": zero, "refundReceiver": zero, "nonce": nonce}}
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	signature, err := crypto.Sign(accounts.TextHash(digest), s.key)
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	signature[64] += 31
	return "0x" + hex.EncodeToString(signature), nil
}

// SignProxy authenticates the legacy relay envelope's packed hash.
func (s *Signer) SignProxy(to, data, gasLimit, nonce, relayHub, relay string) (string, error) {
	raw, err := evm.Hex(data)
	if err != nil {
		return "", err
	}
	target, err := evm.Address(to)
	if err != nil {
		return "", err
	}
	hub, err := evm.Address(relayHub)
	if err != nil {
		return "", err
	}
	relayer, err := evm.Address(relay)
	if err != nil {
		return "", err
	}
	payload := append([]byte("rlx:"), common.HexToAddress(s.Address()).Bytes()...)
	payload = append(payload, target.Bytes()...)
	payload = append(payload, raw...)
	for _, text := range []string{"0", "0", gasLimit, nonce} {
		value, ok := new(big.Int).SetString(text, 10)
		if !ok {
			return "", &sdkerrors.UserInputError{Message: "invalid proxy integer"}
		}
		value, err = evm.Uint256(value)
		if err != nil {
			return "", err
		}
		payload = append(payload, value.FillBytes(make([]byte, 32))...)
	}
	payload = append(payload, hub.Bytes()...)
	payload = append(payload, relayer.Bytes()...)
	signature, err := crypto.Sign(accounts.TextHash(crypto.Keccak256(payload)), s.key)
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	signature[64] += 27
	return "0x" + hex.EncodeToString(signature), nil
}

func wrapDepositSignature(signature, address string, authority models.SignerType) (string, error) {
	if authority != models.SignerSessionKey {
		return signature, nil
	}
	raw, err := evm.Hex(signature)
	if err != nil {
		return "", err
	}
	payload := common.LeftPadBytes(common.HexToAddress(address).Bytes(), 32)
	payload = append(payload, make([]byte, 32)...)
	payload = append(payload, big.NewInt(96).FillBytes(make([]byte, 32))...)
	payload = append(payload, big.NewInt(int64(len(raw))).FillBytes(make([]byte, 32))...)
	payload = append(payload, raw...)
	payload = append(payload, make([]byte, (32-len(raw)%32)%32)...)
	for range 16 {
		payload = append(payload, 0x64, 0x92)
	}
	return "0x" + hex.EncodeToString(payload), nil
}
