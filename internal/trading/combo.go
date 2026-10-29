package trading

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

const comboRequestsPath = "/v1/builder/rfq/requests"

// RequestComboQuoteParams sizes BUY in collateral Amount and SELL in position Size. Side defaults to YES.
type RequestComboQuoteParams struct {
	LegPositionIDs []string
	Direction      models.OrderSide
	Amount, Size   *decimal.Big
	Side           *string
}

func unsignedPosition(value string) (string, error) {
	digits := codec.ASCIIDigits(value)
	if digits == "" || strings.Trim(digits, "0123456789") != "" {
		return "", input("position ID must contain only decimal digits")
	}
	number, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return "", input("invalid position ID")
	}
	return number.String(), nil
}
func comboUnits(value *decimal.Big) (*big.Int, error) {
	if !positive(value) {
		return nil, input("amount must be positive and finite")
	}
	scaled := new(decimal.Big)
	if err := codec.DecimalArithmetic("mul", scaled, value, decimal.New(1000000, 0)); err != nil {
		return nil, input(err.Error())
	}
	if !scaled.Rat(nil).IsInt() {
		return nil, input("amount must have at most six decimal places")
	}
	return scaled.Int(nil), nil
}
func (c *Client) comboRequest(p RequestComboQuoteParams) (map[string]any, error) {
	if p.Direction != models.Buy && p.Direction != models.Sell {
		return nil, input("direction must be BUY or SELL")
	}
	side := "YES"
	if p.Side != nil {
		side = *p.Side
	}
	if side != "YES" {
		return nil, input("combo side must be YES")
	}
	if len(p.LegPositionIDs) < 2 || len(p.LegPositionIDs) > 50 {
		return nil, input("combo requires between 2 and 50 legs")
	}
	legs := make([]string, len(p.LegPositionIDs))
	seen := map[string]bool{}
	for i, id := range p.LegPositionIDs {
		value, err := unsignedPosition(id)
		if err != nil {
			return nil, err
		}
		if seen[value] {
			return nil, input("combo legs must be unique")
		}
		seen[value] = true
		legs[i] = value
	}
	slices.SortFunc(legs, func(a, b string) int {
		if len(a) != len(b) {
			return len(a) - len(b)
		}
		return strings.Compare(a, b)
	})
	quantity, unit := p.Amount, "notional"
	if p.Direction == models.Buy {
		if p.Size != nil {
			return nil, input("BUY uses amount, not size")
		}
	} else {
		if p.Amount != nil {
			return nil, input("SELL uses size, not amount")
		}
		quantity, unit = p.Size, "shares"
	}
	amount, err := comboUnits(quantity)
	if err != nil {
		return nil, err
	}
	signer := c.Signer()
	kind := signing.SignatureType(c.walletType)
	if kind == 3 {
		signer = c.wallet
	}
	return map[string]any{"signer_address": signer, "maker_address": c.wallet, "signature_type": kind, "leg_position_ids": legs, "direction": p.Direction, "side": side, "requested_size": map[string]any{"unit": unit, "value_e6": amount.String()}}, nil
}
func (c *Client) comboAllowed() error {
	if c.signerType == models.SignerSessionKey {
		return input("combo quotes do not support session keys")
	}
	return nil
}
func (c *Client) builderHeaders(method, path string, body []byte) (http.Header, error) {
	headers, err := c.headers(method, path, body)
	if err != nil {
		return nil, err
	}
	if method != "GET" && c.builderKey != nil {
		timestamp := c.now().Unix()
		signature, signErr := signing.HMAC(c.builderKey.Secret, timestamp, method, path, body)
		if signErr != nil {
			return nil, signErr
		}
		headers.Set("POLY_BUILDER_API_KEY", c.builderKey.Key)
		headers.Set("POLY_BUILDER_PASSPHRASE", c.builderKey.Passphrase)
		headers.Set("POLY_BUILDER_TIMESTAMP", strconv.FormatInt(timestamp, 10))
		headers.Set("POLY_BUILDER_SIGNATURE", signature)
	}
	return headers, nil
}
func (c *Client) comboHTTP(ctx context.Context, method, path string, body any) ([]byte, error) {
	var raw []byte
	var err error
	if method == "GET" {
		raw, err = c.transport.AuthenticatedRequest(ctx, c.transport.BuilderGatewayURL, method, path, nil, body, c.builderHeaders)
	} else {
		raw, err = c.transport.HeldRequest(ctx, c.transport.BuilderGatewayURL, method, path, body, c.builderHeaders)
	}
	if rejected, ok := errors.AsType[*sdkerrors.RequestRejectedError](err); ok {
		return nil, &sdkerrors.RFQRequestRejectedError{Rejection: rejected}
	}
	return raw, err
}

// RequestComboQuote waits through the competition window and returns a quote or a normal no-quote outcome.
func (c *Client) RequestComboQuote(ctx context.Context, p RequestComboQuoteParams) (*models.ComboQuoteResult, error) {
	if err := c.comboAllowed(); err != nil {
		return nil, err
	}
	body, err := c.comboRequest(p)
	if err != nil {
		return nil, err
	}
	if c.builderKey == nil {
		return nil, input("combo quote requests require a builder API key")
	}
	raw, err := c.comboHTTP(ctx, "POST", comboRequestsPath, body)
	if err != nil {
		return nil, err
	}
	return parseComboQuote(raw, body)
}

// FetchRFQStatus retrieves an accepted RFQ's lifecycle and execution state.
func (c *Client) FetchRFQStatus(ctx context.Context, id string) (*models.RFQStatusResult, error) {
	if err := c.comboAllowed(); err != nil {
		return nil, err
	}
	if err := nonempty(id); err != nil {
		return nil, err
	}
	raw, err := c.comboHTTP(ctx, "GET", comboRequestsPath+"/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	return parseRFQStatus(raw, id)
}
func (c *Client) comboAcceptanceBody(quote models.ComboQuote) (map[string]any, error) {
	if err := nonempty(quote.RFQID, quote.QuoteID); err != nil {
		return nil, err
	}
	if quote.Direction != models.Buy && quote.Direction != models.Sell {
		return nil, input("invalid combo quote direction")
	}
	if quote.Direction == models.Sell && quote.NetReceive == nil {
		return nil, input("SELL quote requires net receive")
	}
	position, err := unsignedPosition(quote.PositionID)
	if err != nil {
		return nil, err
	}
	if err = validateBuilder(quote.BuilderCode); err != nil {
		return nil, err
	}
	if quote.ExpiresAt < 0 {
		return nil, input("quote expiration must be nonnegative")
	}
	maker, err := comboUnits(quote.MakerAmount)
	if err != nil {
		return nil, err
	}
	taker, err := comboUnits(quote.TakerAmount)
	if err != nil {
		return nil, err
	}
	salt, err := c.newSalt(64)
	if err != nil {
		return nil, err
	}
	kind := signing.SignatureType(c.walletType)
	signer := c.Signer()
	if kind == 3 {
		signer = c.wallet
	}
	order := models.SignedOrder{Builder: quote.BuilderCode, Maker: c.wallet, Signer: signer, MakerAmount: maker, TakerAmount: taker, Salt: salt, Side: quote.Direction, SignatureType: kind, Timestamp: c.now().Unix(), AssetID: models.ClobAssetID(position), Metadata: signing.ZeroBytes32}
	signature, err := c.signer.Order(order, "0xe3333700cA9d93003F00f0F71f8515005F6c00Aa", models.SignerOwner)
	if err != nil {
		return nil, err
	}
	side := 0
	if quote.Direction == models.Sell {
		side = 1
	}
	return map[string]any{"quote_id": quote.QuoteID, "signed_order": map[string]any{"salt": salt.String(), "maker": order.Maker, "signer": order.Signer, "tokenId": position, "makerAmount": maker.String(), "takerAmount": taker.String(), "side": side, "signatureType": kind, "timestamp": strconv.FormatInt(order.Timestamp, 10), "builder": order.Builder, "metadata": order.Metadata, "signature": signature}}, nil
}

// AcceptComboQuote signs and accepts a persisted quote. Only interrupted or malformed acceptance responses retry, using the same signed body.
func (c *Client) AcceptComboQuote(ctx context.Context, quote models.ComboQuote) (*models.ComboQuoteAcceptance, error) {
	encoded, encodeErr := json.Marshal(quote)
	if encodeErr != nil {
		return nil, input("invalid persisted quote")
	}
	var validated models.ComboQuote
	if decodeErr := json.Unmarshal(encoded, &validated); decodeErr != nil {
		return nil, input("invalid persisted quote")
	}
	quote = validated
	if err := c.comboAllowed(); err != nil {
		return nil, err
	}
	if c.builderKey == nil {
		return nil, input("combo acceptance requires a builder API key")
	}
	body, err := c.comboAcceptanceBody(quote)
	if err != nil {
		return nil, err
	}
	path := comboRequestsPath + "/" + url.PathEscape(quote.RFQID) + "/accept"
	var status *models.RFQStatusResult
	for range 2 {
		var raw []byte
		raw, err = c.comboHTTP(ctx, "POST", path, body)
		if err == nil {
			status, err = parseRFQStatus(raw, quote.RFQID)
		}
		var transportError *sdkerrors.TransportError
		var malformed *sdkerrors.UnexpectedResponseError
		if err == nil || ctx.Err() != nil || !errors.As(err, &transportError) && !errors.As(err, &malformed) {
			break
		}
	}
	if err != nil {
		var rejected *sdkerrors.RequestRejectedError
		if errors.As(err, &rejected) && rejected.Code == "EXPIRED_RFQ" {
			reason := "ACCEPTANCE_WINDOW_EXPIRED"
			return &models.ComboQuoteAcceptance{RFQID: quote.RFQID, Status: "failed", Reason: &reason, Error: &models.RFQErrorDetail{Code: rejected.Code, Message: rejected.Error()}}, nil
		}
		return nil, err
	}
	if status == nil {
		return nil, responseMessage("combo acceptance response is missing")
	}
	hash := status.TakerOrderHash
	deadline := time.Now().Add(30 * time.Second)
	for status.Status == "AWAITING_MAKER_CONFIRMATION" {
		if time.Now().After(deadline) {
			return nil, &sdkerrors.RFQTimeoutError{RFQID: quote.RFQID}
		}
		if err = waitContext(ctx, 500*time.Millisecond); err != nil {
			return nil, err
		}
		status, err = c.FetchRFQStatus(ctx, quote.RFQID)
		if err != nil {
			return nil, err
		}
	}
	result := &models.ComboQuoteAcceptance{RFQID: quote.RFQID, Status: "executing", TakerOrderHash: hash}
	if slices.Contains([]string{"FAILED", "EXPIRED", "CANCELED"}, status.Status) {
		reason := "EXECUTION_FAILED"
		if status.Status == "EXPIRED" || status.Error != nil && status.Error.Code == "EXPIRED_RFQ" {
			reason = "ACCEPTANCE_WINDOW_EXPIRED"
		} else if status.Error != nil && status.Error.Code == "MAKER_DECLINED" {
			reason = "MAKER_DECLINED"
		}
		result.Status = "failed"
		result.Reason = &reason
		result.Error = status.Error
		result.TakerOrderHash = nil
	}
	return result, nil
}

// WaitForComboFill polls until filled, failed, expired or canceled. Zero durations use thirty seconds and one second.
func (c *Client) WaitForComboFill(ctx context.Context, id string, timeout, interval time.Duration) (*models.ComboFillResult, error) {
	if err := c.comboAllowed(); err != nil {
		return nil, err
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if interval == 0 {
		interval = time.Second
	}
	if timeout < 0 || interval < 0 {
		return nil, input("timeout and polling interval must be positive")
	}
	deadline := time.Now().Add(timeout)
	for {
		status, err := c.FetchRFQStatus(ctx, id)
		if err != nil {
			return nil, err
		}
		switch status.Status {
		case "FILLED", "CONFIRMED":
			if status.TxHash == nil {
				return nil, responseMessage("filled RFQ has no transaction hash")
			}
			return &models.ComboFillResult{RFQID: id, Status: "FILLED", TxHash: status.TxHash}, nil
		case "FAILED", "EXPIRED", "CANCELED":
			return &models.ComboFillResult{RFQID: id, Status: status.Status, Error: status.Error}, nil
		}
		if time.Now().After(deadline) {
			return nil, &sdkerrors.RFQTimeoutError{RFQID: id}
		}
		if err = waitContext(ctx, min(interval, time.Until(deadline))); err != nil {
			return nil, err
		}
	}
}
func waitContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
