package signing

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/models"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// ZeroBytes32 is the exchange's empty metadata and builder value.
const ZeroBytes32 = "0x0000000000000000000000000000000000000000000000000000000000000000"
const orderTypeString = "Order(uint256 salt,address maker,address signer,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,uint256 timestamp,bytes32 metadata,bytes32 builder)"

// V2Position reports whether an asset occupies the protocol-v2 namespace.
func V2Position(asset string) bool {
	raw := strings.TrimSpace(asset)
	base := 10
	if len(raw) > 2 {
		switch strings.ToLower(raw[:2]) {
		case "0x":
			base = 16
			raw = raw[2:]
		case "0b":
			base = 2
			raw = raw[2:]
		case "0o":
			base = 8
			raw = raw[2:]
		}
	}
	if raw == "" || strings.ContainsAny(raw, "_+- ") {
		return false
	}
	number, ok := new(big.Int).SetString(raw, base)
	if !ok || number.Sign() < 0 || number.BitLen() > 256 {
		return false
	}
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1))
	mask.Lsh(mask, 40)
	return new(big.Int).And(number, mask).Sign() == 0
}

// Exchange chooses the contract whose domain signs this order.
func Exchange(asset string, negativeRisk bool) string {
	if V2Position(asset) {
		return "0xe3333700cA9d93003F00f0F71f8515005F6c00Aa"
	}
	if negativeRisk {
		return "0xe2222d279d744050d28e00520010520000310F59"
	}
	return "0xE111180000d2663C0091e4f400237545B87B996B"
}

// Order signs either the standard order or the deposit-wallet EIP-1271 wrapper.
func (s *Signer) Order(order models.SignedOrder, exchange string, authority models.SignerType) (string, error) {
	data, err := orderData(order, exchange)
	if err != nil {
		return "", err
	}
	standard := data
	if order.SignatureType == 3 {
		data.Types["TypedDataSign"] = []apitypes.Type{{Name: "contents", Type: "Order"}, {Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}, {Name: "salt", Type: "bytes32"}}
		data.PrimaryType = "TypedDataSign"
		data.Message = apitypes.TypedDataMessage{"contents": standard.Message, "name": "DepositWallet", "version": "1", "chainId": big.NewInt(137), "verifyingContract": string(order.Signer), "salt": ZeroBytes32}
	}
	signature, err := s.TypedData(data)
	if err != nil {
		return "", err
	}
	if order.SignatureType != 3 {
		return signature, nil
	}
	domain, err := standard.HashStruct("EIP712Domain", standard.Domain.Map())
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	contents, err := standard.HashStruct("Order", standard.Message)
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	signature += hex.EncodeToString(domain) + hex.EncodeToString(contents) + hex.EncodeToString([]byte(orderTypeString)) + fmt.Sprintf("%04x", len(orderTypeString))
	return wrapDepositSignature(signature, s.Address(), authority)
}
func orderData(order models.SignedOrder, exchange string) (apitypes.TypedData, error) {
	asset, ok := new(big.Int).SetString(string(order.AssetID), 10)
	if !ok || asset.Sign() < 0 || asset.BitLen() > 256 {
		return apitypes.TypedData{}, &sdkerrors.SigningError{Cause: fmt.Errorf("invalid order asset")}
	}
	version := "2"
	if strings.EqualFold(exchange, "0xe3333700cA9d93003F00f0F71f8515005F6c00Aa") {
		version = "3"
	}
	side := int64(0)
	if order.Side == models.Sell {
		side = 1
	}
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}},
			"Order":        {{Name: "salt", Type: "uint256"}, {Name: "maker", Type: "address"}, {Name: "signer", Type: "address"}, {Name: "tokenId", Type: "uint256"}, {Name: "makerAmount", Type: "uint256"}, {Name: "takerAmount", Type: "uint256"}, {Name: "side", Type: "uint8"}, {Name: "signatureType", Type: "uint8"}, {Name: "timestamp", Type: "uint256"}, {Name: "metadata", Type: "bytes32"}, {Name: "builder", Type: "bytes32"}},
		},
		PrimaryType: "Order",
		Domain:      apitypes.TypedDataDomain{Name: "Polymarket CTF Exchange", Version: version, ChainId: (*math.HexOrDecimal256)(big.NewInt(137)), VerifyingContract: exchange},
		Message:     apitypes.TypedDataMessage{"salt": order.Salt, "maker": string(order.Maker), "signer": string(order.Signer), "tokenId": asset, "makerAmount": order.MakerAmount, "takerAmount": order.TakerAmount, "side": big.NewInt(side), "signatureType": big.NewInt(int64(order.SignatureType)), "timestamp": big.NewInt(order.Timestamp), "metadata": order.Metadata, "builder": order.Builder},
	}, nil
}
