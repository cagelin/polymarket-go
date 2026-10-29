package trading

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	sdkerrors "github.com/cagelin/polymarket-go/errors"

	"github.com/cagelin/polymarket-go/internal/clob"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

// CreateLimitOrderParams describes a locally signed limit order.
type CreateLimitOrderParams struct {
	AssetID     models.ClobAssetID
	Price, Size *decimal.Big
	Side        models.OrderSide
	PostOnly    bool
	Expiration  *int64
	BuilderCode *string
}

// CreateLimitOrder resolves market metadata and signs without submitting a trade.
func (c *Client) CreateLimitOrder(ctx context.Context, p CreateLimitOrderParams) (*models.SignedOrder, error) {
	if err := validateOrder(string(p.AssetID), p.Side, p.BuilderCode); err != nil {
		return nil, err
	}
	if !positive(p.Price) || !positive(p.Size) {
		return nil, input("price and size must be positive and finite")
	}
	if p.Expiration != nil && *p.Expiration < c.now().Unix()+180 {
		return nil, input("expiration must be at least 180 seconds in the future")
	}
	metadata, err := c.market(ctx, string(p.AssetID), false)
	if err != nil {
		return nil, err
	}
	if err = validatePrice(p.Price, metadata.tick); err != nil {
		metadata, err = c.market(ctx, string(p.AssetID), true)
		if err != nil {
			return nil, err
		}
		if err = validatePrice(p.Price, metadata.tick); err != nil {
			return nil, err
		}
	}
	offered, requested, err := limitAmounts(p.Price, p.Size, metadata.tick, p.Side)
	if err != nil {
		return nil, input(err.Error())
	}
	order := models.SignedOrder{AssetID: p.AssetID, Side: p.Side, OrderType: models.GTC, MakerAmount: offered, TakerAmount: requested, PostOnly: p.PostOnly}
	if p.Expiration != nil {
		order.Expiration = *p.Expiration
		order.OrderType = models.GTD
	}
	return c.signOrder(order, p.BuilderCode, metadata.negativeRisk)
}

// CreateMarketOrderParams uses Amount when buying and Shares when selling. Empty OrderType selects FAK.
type CreateMarketOrderParams struct {
	AssetID                                      models.ClobAssetID
	Side                                         models.OrderSide
	Amount, Shares, MaxSpend, MaxPrice, MinPrice *decimal.Big
	OrderType                                    clob.MarketOrderType
	BuilderCode                                  *string
}

func validateOrder(asset string, side models.OrderSide, builder *string) error {
	if err := nonempty(asset); err != nil {
		return err
	}
	if side != models.Buy && side != models.Sell {
		return input("side must be BUY or SELL")
	}
	if builder != nil {
		return validateBuilder(*builder)
	}
	return nil
}
func validateMarket(p CreateMarketOrderParams) error {
	if err := validateOrder(string(p.AssetID), p.Side, p.BuilderCode); err != nil {
		return err
	}
	if p.OrderType != "" && p.OrderType != clob.FAK && p.OrderType != clob.FOK {
		return input("market order type must be FAK or FOK")
	}
	if p.Side == models.Buy {
		if !positive(p.Amount) || p.Shares != nil || p.MinPrice != nil {
			return input("BUY requires amount and forbids shares and min_price")
		}
	} else if !positive(p.Shares) || p.Amount != nil || p.MaxSpend != nil || p.MaxPrice != nil {
		return input("SELL requires shares and forbids amount, max_spend and max_price")
	}
	for _, value := range []*decimal.Big{p.MaxSpend, p.MaxPrice, p.MinPrice} {
		if value != nil && !positive(value) {
			return input("price bounds and max_spend must be positive and finite")
		}
	}
	return nil
}

// CreateMarketOrder signs a market order using the supplied price bound or current order book.
func (c *Client) CreateMarketOrder(ctx context.Context, p CreateMarketOrderParams) (*models.SignedOrder, error) {
	if err := validateMarket(p); err != nil {
		return nil, err
	}
	if p.OrderType == "" {
		p.OrderType = clob.FAK
	}
	amount, price := p.Amount, p.MaxPrice
	if p.Side == models.Sell {
		amount, price = p.Shares, p.MinPrice
	}
	protected := price != nil
	var metadata marketInfo
	if protected {
		for attempt := range 2 {
			current, err := c.market(ctx, string(p.AssetID), attempt == 1)
			if err != nil {
				return nil, err
			}
			metadata = current
			order, buildErr := c.marketDraft(ctx, p, amount, price, metadata, true)
			if buildErr == nil {
				return c.signOrder(*order, p.BuilderCode, metadata.negativeRisk)
			}
			if attempt == 1 || !isInput(buildErr) {
				return nil, buildErr
			}
		}
	}
	book, err := clob.New(c.transport).GetOrderBook(ctx, clob.Asset(p.AssetID))
	if err != nil {
		return nil, err
	}
	if book.AssetID != p.AssetID {
		return nil, responseMessage("order book asset differs from requested asset")
	}
	price, err = clob.EstimatePrice(book, p.Side, amount, p.OrderType)
	if err != nil {
		return nil, err
	}
	metadata = marketInfo{tick: book.TickSize, negativeRisk: book.NegRisk}
	if p.Side == models.Buy && p.MaxSpend != nil {
		feeMetadata, fetchErr := c.market(ctx, string(p.AssetID), false)
		if fetchErr != nil {
			return nil, fetchErr
		}
		metadata.rate, metadata.exponent = feeMetadata.rate, feeMetadata.exponent
	}
	order, err := c.marketDraft(ctx, p, amount, price, metadata, false)
	if err != nil {
		return nil, err
	}
	return c.signOrder(*order, p.BuilderCode, metadata.negativeRisk)
}
func (c *Client) marketDraft(ctx context.Context, p CreateMarketOrderParams, amount, price *decimal.Big, metadata marketInfo, protected bool) (*models.SignedOrder, error) {
	if protected {
		if err := validatePrice(price, metadata.tick); err != nil {
			return nil, err
		}
	}
	if p.Side == models.Buy && p.MaxSpend != nil {
		code := ""
		if p.BuilderCode != nil {
			code = *p.BuilderCode
		}
		builder, err := c.builderRate(ctx, code)
		if err != nil {
			return nil, err
		}
		amount, err = adjustBuyAmount(amount, price, p.MaxSpend, metadata.rate, metadata.exponent, builder)
		if err != nil {
			return nil, response(err)
		}
	}
	offered, requested, err := marketAmounts(price, amount, metadata.tick, p.Side)
	if err != nil {
		return nil, input(err.Error())
	}
	if protected && p.Side == models.Buy {
		if err = protectedBuy(offered, requested, price); err != nil {
			return nil, err
		}
	}
	return &models.SignedOrder{AssetID: p.AssetID, Side: p.Side, OrderType: models.OrderType(p.OrderType), MakerAmount: offered, TakerAmount: requested}, nil
}
func (c *Client) signOrder(order models.SignedOrder, builder *string, negativeRisk bool) (*models.SignedOrder, error) {
	salt, err := c.newSalt(53)
	if err != nil {
		return nil, err
	}
	order.Salt = salt
	order.Timestamp = c.now().UnixMilli()
	order.Maker = c.wallet
	order.Signer = c.Signer()
	order.SignatureType = signing.SignatureType(c.walletType)
	if order.SignatureType == 3 {
		order.Signer = c.wallet
	}
	order.Builder = signing.ZeroBytes32
	if builder != nil {
		order.Builder = *builder
	}
	order.Metadata = signing.ZeroBytes32
	order.Signature, err = c.signer.Order(order, signing.Exchange(string(order.AssetID), negativeRisk), c.signerType)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func isInput(err error) bool               { var target *sdkerrors.UserInputError; return errors.As(err, &target) }
func responseMessage(message string) error { return response(fmt.Errorf("%s", message)) }

// PlaceLimitOrder creates, signs and submits a limit order. Chain allowance recovery belongs to the wallet milestone.
func (c *Client) PlaceLimitOrder(ctx context.Context, p CreateLimitOrderParams) (models.OrderResponse, error) {
	order, err := c.CreateLimitOrder(ctx, p)
	if err != nil {
		return nil, err
	}
	return c.postWithAllowanceRecovery(ctx, *order)
}

// PlaceMarketOrder creates, signs and submits a market order. Chain allowance recovery belongs to the wallet milestone.
func (c *Client) PlaceMarketOrder(ctx context.Context, p CreateMarketOrderParams) (models.OrderResponse, error) {
	order, err := c.CreateMarketOrder(ctx, p)
	if err != nil {
		return nil, err
	}
	return c.postWithAllowanceRecovery(ctx, *order)
}

func (c *Client) newSalt(bits int) (*big.Int, error) {
	if c.salt != nil {
		return c.salt(bits)
	}
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), uint(bits)))
}
