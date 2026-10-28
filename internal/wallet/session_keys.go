package wallet

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	sdkerrors "github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/codec"
	"github.com/cagelam/polymarket-go/internal/evm"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/transactions"
	"github.com/ethereum/go-ethereum/common"
)

func (c *Client) sessionOwner() error {
	if c.config.Kind != models.WalletDeposit || c.config.Authority != models.SignerOwner {
		return input("session keys require the Deposit Wallet owner")
	}
	return nil
}
func sessionAddress(address models.EVMAddress) (common.Address, error) {
	if !strings.HasPrefix(string(address), "0x") {
		return common.Address{}, input("session key address requires a 0x prefix")
	}
	value, err := evm.Address(string(address))
	if err != nil {
		return common.Address{}, err
	}
	if value == (common.Address{}) {
		return value, input("session key address cannot be zero")
	}
	return value, nil
}
func operationKey(value *string) (string, error) {
	if value != nil {
		key := strings.TrimSpace(*value)
		if key == "" {
			return "", input("idempotency key cannot be empty")
		}
		return key, nil
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	raw[6] = raw[6]&15 | 64
	raw[8] = raw[8]&63 | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[:4], raw[4:6], raw[6:8], raw[8:10], raw[10:]), nil
}

// FetchSessionKeys reads the active registry and verifies its wallet identity.
func (c *Client) FetchSessionKeys(ctx context.Context) ([]models.AuthorizedSessionKey, error) {
	if err := c.sessionOwner(); err != nil {
		return nil, err
	}
	raw, err := c.transport.AuthenticatedRequest(ctx, c.transport.CLOBURL, "GET", "/v1/user/session-signers", nil, nil, c.config.CLOBHeaders)
	if err != nil {
		return nil, err
	}
	fields, err := codec.Object(raw)
	if err != nil {
		return nil, err
	}
	wallet, err := codec.Required[string](fields, "wallet")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(wallet, string(c.config.Wallet)) {
		return nil, response("session registry wallet mismatch")
	}
	rows, err := codec.Required[[]json.RawMessage](fields, "signers")
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, response("session signers must be an array")
	}
	result := make([]models.AuthorizedSessionKey, len(rows))
	for i, row := range rows {
		item, decodeErr := codec.Object(row)
		if decodeErr != nil {
			return nil, decodeErr
		}
		address, decodeErr := codec.Required[string](item, "address")
		if decodeErr != nil {
			return nil, decodeErr
		}
		normalized, decodeErr := evm.Address(address)
		if decodeErr != nil {
			return nil, response("invalid session key address")
		}
		scopes, decodeErr := codec.Required[[]models.SessionKeyScope](item, "scopes")
		if decodeErr != nil {
			return nil, decodeErr
		}
		if len(scopes) == 0 || slices.Contains(scopes, "") {
			return nil, response("session scopes must be nonempty")
		}
		expiry, decodeErr := codec.Required[int64](item, "valid_until")
		if decodeErr != nil {
			return nil, decodeErr
		}
		if expiry < 0 || expiry > 253402300799 {
			return nil, response("invalid session key expiry")
		}
		result[i] = models.AuthorizedSessionKey{Address: models.EVMAddress(normalized.Hex()), Scopes: scopes, ValidUntil: time.Unix(expiry, 0).UTC()}
	}
	return result, nil
}
func transientSession(err error, registry bool) (bool, time.Duration) {
	if rate, ok := errors.AsType[*sdkerrors.RateLimitError](err); ok {
		if rate.RetryAfter != nil {
			return true, time.Duration(max(*rate.RetryAfter, 0) * float64(time.Second))
		}
		return true, 0
	}
	if rejected, ok := errors.AsType[*sdkerrors.RequestRejectedError](err); ok {
		retry := rejected.Status >= 500 && rejected.Status < 600 || registry && rejected.Status == 404
		if rejected.RetryAfter != nil {
			return retry, time.Duration(max(*rejected.RetryAfter, 0) * float64(time.Second))
		}
		return retry, 0
	}
	var transportError *sdkerrors.TransportError
	return errors.As(err, &transportError), 0
}
func (c *Client) sessionOperation(ctx context.Context, path string, body map[string]any, key string) (map[string]json.RawMessage, error) {
	headers := func(method, requestPath string, raw []byte) (http.Header, error) {
		result, err := c.headers(method, requestPath, raw)
		if err != nil {
			return nil, err
		}
		result.Set("Idempotency-Key", key)
		return result, nil
	}
	for attempt := 0; ; attempt++ {
		raw, err := c.transport.LongRequest(ctx, c.transport.RelayerURL, "POST", path, body, headers, 300*time.Second)
		if err == nil {
			return codec.Object(raw)
		}
		retry, delay := transientSession(err, false)
		if !retry || attempt == 2 {
			return nil, err
		}
		if delay == 0 {
			delay = c.config.PollInterval
		}
		if err = wait(ctx, delay); err != nil {
			return nil, err
		}
	}
}
func (c *Client) sessionBatch(ctx context.Context, address common.Address, expiry *big.Int) (map[string]any, error) {
	name, types, values := "revokeSessionSigner(address)", []string{"address"}, []any{address}
	if expiry != nil {
		name = "authorizeSessionSigner(address,uint256)"
		types = append(types, "uint256")
		values = append(values, expiry)
	}
	call, err := evm.Call(string(c.config.Wallet), name, types, values...)
	if err != nil {
		return nil, err
	}
	call.Value = new(big.Int)
	envelope, err := c.relayerPayload(ctx, []models.TransactionCall{call}, "", "")
	if err != nil {
		return nil, err
	}
	params, ok := envelope["depositWalletParams"].(map[string]any)
	if !ok {
		return nil, response("missing deposit wallet batch parameters")
	}
	return map[string]any{"deadline": params["deadline"], "nonce": envelope["nonce"], "sessionSignerAddress": address.Hex(), "signature": envelope["signature"], "walletAddress": c.config.Wallet}, nil
}
func (c *Client) waitSession(ctx context.Context, address models.EVMAddress, scopes []models.SessionKeyScope, expiry int64, revoke bool) (models.AuthorizedSessionKey, error) {
	expected := append([]models.SessionKeyScope(nil), scopes...)
	slices.Sort(expected)
	expected = slices.Compact(expected)
	for attempt := 0; attempt < c.config.MaxPolls; attempt++ {
		keys, err := c.FetchSessionKeys(ctx)
		if err != nil {
			retry, _ := transientSession(err, true)
			if !retry {
				return models.AuthorizedSessionKey{}, err
			}
		} else {
			found := false
			for _, key := range keys {
				if strings.EqualFold(string(key.Address), string(address)) {
					found = true
					actual := append([]models.SessionKeyScope(nil), key.Scopes...)
					slices.Sort(actual)
					actual = slices.Compact(actual)
					if !revoke && key.ValidUntil.Unix() == expiry && slices.Equal(actual, expected) {
						return key, nil
					}
				}
			}
			if revoke && !found {
				return models.AuthorizedSessionKey{}, nil
			}
		}
		if attempt+1 < c.config.MaxPolls {
			if err = wait(ctx, c.config.PollInterval); err != nil {
				return models.AuthorizedSessionKey{}, err
			}
		}
	}
	return models.AuthorizedSessionKey{}, &sdkerrors.TransactionTimeoutError{}
}

// AuthorizeSessionKey waits for chain confirmation and the exact registry grant.
func (c *Client) AuthorizeSessionKey(ctx context.Context, p models.AuthorizeSessionKeyParams) (*models.AuthorizeSessionKeyResult, error) {
	if err := c.sessionOwner(); err != nil {
		return nil, err
	}
	if c.config.BuilderKey == nil {
		return nil, input("session authorization requires BuilderKey")
	}
	address, err := sessionAddress(p.Address)
	if err != nil {
		return nil, err
	}
	scopes, err := sessionScopes(p.Scopes)
	if err != nil {
		return nil, err
	}
	key, err := operationKey(p.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	expiry := c.now().Unix() + 4315*60*60
	if err = c.lock(ctx); err != nil {
		return nil, err
	}
	defer c.unlock()
	body, err := c.sessionBatch(ctx, address, big.NewInt(expiry))
	if err != nil {
		return nil, err
	}
	body["scopes"], body["validUntil"] = scopes, strconv.FormatInt(expiry, 10)
	fields, err := c.sessionOperation(ctx, "/v1/session-signers/authorizations", body, key)
	if err != nil {
		return nil, err
	}
	status, err := codec.Required[string](fields, "status")
	if err != nil {
		return nil, err
	}
	switch status {
	case "FAILED", "SUPERSEDED", "REPAIR_REQUIRED":
		return nil, &sdkerrors.TransactionFailedError{Message: "session authorization failed: " + status}
	case "SUBMITTED", "REGISTRY_PENDING", "REGISTERED":
	default:
		return nil, response("invalid session authorization status")
	}
	id, err := codec.Required[string](fields, "transactionId")
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, response("missing session transaction ID")
	}
	var hash *string
	if value, ok := fields["transactionHash"]; ok {
		if err = json.Unmarshal(value, &hash); err != nil {
			return nil, err
		}
	}
	if hash != nil && *hash == "" {
		hash = nil
	}
	if hash != nil {
		decoded, parseErr := evm.Hex(*hash)
		if parseErr != nil || len(decoded) != 32 {
			return nil, response("invalid session transaction hash")
		}
	}
	handle := transactions.NewRelayed(c.transport, c.headers, id, hash, c.config.MaxPolls, c.config.PollInterval)
	outcome, err := handle.Wait(ctx)
	if err != nil {
		return nil, err
	}
	granted, err := c.waitSession(ctx, models.EVMAddress(address.Hex()), scopes, expiry, false)
	if err != nil {
		return nil, err
	}
	return &models.AuthorizeSessionKeyResult{SessionKey: granted, Transaction: *outcome}, nil
}

// RevokeSessionKey returns only after the key leaves the active registry.
func (c *Client) RevokeSessionKey(ctx context.Context, p models.RevokeSessionKeyParams) error {
	if err := c.sessionOwner(); err != nil {
		return err
	}
	if _, err := c.headers("POST", "", nil); err != nil {
		return err
	}
	address, err := sessionAddress(p.Address)
	if err != nil {
		return err
	}
	key, err := operationKey(p.IdempotencyKey)
	if err != nil {
		return err
	}
	if err = c.lock(ctx); err != nil {
		return err
	}
	defer c.unlock()
	body, err := c.sessionBatch(ctx, address, nil)
	if err != nil {
		return err
	}
	fields, err := c.sessionOperation(ctx, "/v1/session-signers/revocations", body, key)
	if err != nil {
		return err
	}
	status, err := codec.Required[string](fields, "status")
	if err != nil {
		return err
	}
	switch status {
	case "FAILED":
		return &sdkerrors.TransactionFailedError{Message: "session revocation failed"}
	case "PENDING", "FENCED", "SWEPT", "CHAIN_SUBMITTED", "CONFIRMED":
	default:
		return response("invalid session revocation status")
	}
	if _, err = codec.Required[bool](fields, "fenced"); err != nil {
		return err
	}
	id, err := codec.Required[string](fields, "transactionId")
	if err != nil {
		return err
	}
	if id == "" {
		return response("missing revocation transaction ID")
	}
	_, err = c.waitSession(ctx, models.EVMAddress(address.Hex()), nil, 0, true)
	return err
}

func sessionScopes(value []models.SessionKeyScope) ([]models.SessionKeyScope, error) {
	scopes := append([]models.SessionKeyScope(nil), value...)
	if value == nil {
		scopes = []models.SessionKeyScope{models.SessionScopeAll}
	}
	if len(scopes) == 0 || slices.Contains(scopes, "") {
		return nil, input("scopes must contain nonempty strings")
	}
	if slices.Contains(scopes, models.SessionScopeAll) {
		for _, scope := range scopes {
			if scope != models.SessionScopeAll {
				return nil, input("ALL cannot be combined with another scope")
			}
		}
	}
	return scopes, nil
}
