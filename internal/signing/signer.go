// Package signing implements the pinned exchange's authentication and EIP-712 signatures.
package signing

import (
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Signer owns an EVM key. Its formatting never includes key material.
type Signer struct{ key *ecdsa.PrivateKey }

func (s *Signer) String() string { return "Signer(<redacted>)" }

// GoString redacts sensitive fields during Go-syntax formatting.
func (s *Signer) GoString() string { return s.String() }

// New validates and copies a 32-byte private key.
func New(privateKey string) (*Signer, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return nil, &sdkerrors.UserInputError{Message: "invalid private key"}
	}
	return &Signer{key: key}, nil
}

// Address returns the signer's checksummed EVM address.
func (s *Signer) Address() string { return crypto.PubkeyToAddress(s.key.PublicKey).Hex() }

// Address validates a complete EVM address and returns its checksum form.
func Address(value string) (string, error) {
	if !common.IsHexAddress(value) {
		return "", &sdkerrors.UserInputError{Message: "invalid EVM address"}
	}
	return common.HexToAddress(value).Hex(), nil
}

// TypedData signs a validated EIP-712 message with Ethereum's recovery byte.
func (s *Signer) TypedData(data apitypes.TypedData) (string, error) {
	digest, _, err := apitypes.TypedDataAndHash(data)
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	raw, err := crypto.Sign(digest, s.key)
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	raw[64] += 27
	return "0x" + hex.EncodeToString(raw), nil
}

// L1 authenticates ownership of the signer for API-key creation or derivation.
func (s *Signer) L1(timestamp, nonce int64) (http.Header, error) {
	if nonce < 0 {
		return nil, &sdkerrors.UserInputError{Message: "nonce must be nonnegative"}
	}
	data := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}},
			"ClobAuth":     {{Name: "address", Type: "address"}, {Name: "timestamp", Type: "string"}, {Name: "nonce", Type: "uint256"}, {Name: "message", Type: "string"}},
		},
		PrimaryType: "ClobAuth",
		Domain:      apitypes.TypedDataDomain{Name: "ClobAuthDomain", Version: "1", ChainId: (*math.HexOrDecimal256)(big.NewInt(137))},
		Message:     apitypes.TypedDataMessage{"address": s.Address(), "timestamp": strconv.FormatInt(timestamp, 10), "nonce": big.NewInt(nonce), "message": "This message attests that I control the given wallet"},
	}
	signature, err := s.TypedData(data)
	if err != nil {
		return nil, err
	}
	return http.Header{"POLY_ADDRESS": {s.Address()}, "POLY_NONCE": {strconv.FormatInt(nonce, 10)}, "POLY_TIMESTAMP": {strconv.FormatInt(timestamp, 10)}, "POLY_SIGNATURE": {signature}}, nil
}

// HMAC authenticates the exact UTF-8 request body, excluding the query string.
func HMAC(secret string, timestamp int64, method, path string, body []byte) (string, error) {
	raw, err := decodeSecret(secret)
	if err != nil {
		return "", &sdkerrors.SigningError{Cause: err}
	}
	digest := hmac.New(sha256.New, raw)
	_, _ = digest.Write([]byte(strconv.FormatInt(timestamp, 10) + method + path)) // Hash writes cannot fail.
	_, _ = digest.Write(body)
	return base64.URLEncoding.EncodeToString(digest.Sum(nil)), nil
}

// decodeSecret follows Python's non-strict base64 decoding, including padding.
func decodeSecret(secret string) ([]byte, error) {
	secret += strings.Repeat("=", (4-len(secret)%4)%4)
	var data strings.Builder
	pads, position := 0, 0
	complete := false
	for _, ch := range secret {
		if ch > 127 {
			return nil, fmt.Errorf("non-ASCII API secret")
		}
		if ch == '=' {
			pads++
			if position >= 2 && position+pads >= 4 {
				complete = true
				break
			}
			continue
		}
		if ch == '-' {
			ch = '+'
		}
		if ch == '_' {
			ch = '/'
		}
		valid := ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '+' || ch == '/'
		if !valid {
			continue
		}
		pads = 0
		position = (position + 1) % 4
		data.WriteRune(ch)
	}
	if !complete && position != 0 {
		return nil, fmt.Errorf("invalid API secret encoding")
	}
	raw, err := base64.RawStdEncoding.DecodeString(data.String())
	if err != nil {
		return nil, fmt.Errorf("invalid API secret encoding")
	}
	return raw, nil
}
