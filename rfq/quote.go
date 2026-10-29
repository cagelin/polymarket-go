package rfq

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
	"github.com/cagelin/polymarket-go/internal/codec"
	"github.com/cagelin/polymarket-go/internal/signing"
	"github.com/cagelin/polymarket-go/models"
	"github.com/ericlagergren/decimal"
)

func units(value *decimal.Big) (*big.Int, error) {
	if value == nil || !value.IsFinite() || value.Sign() <= 0 {
		return nil, &sdkerrors.UserInputError{Message: "quote amount must be positive and finite"}
	}
	scaled := new(decimal.Big)
	if err := codec.DecimalArithmetic("mul", scaled, value, decimal.New(1000000, 0)); err != nil {
		return nil, err
	}
	if !scaled.Rat(nil).IsInt() {
		return nil, &sdkerrors.UserInputError{Message: "quote amount must have at most six decimal places"}
	}
	return scaled.Int(nil), nil
}
func (s *Session) signedQuote(request Event, p QuoteParams, salt *big.Int, timestamp int64) (map[string]any, error) {
	price, err := units(p.Price)
	if err != nil {
		return nil, err
	}
	million := big.NewInt(1000000)
	if price.Cmp(million) >= 0 {
		return nil, &sdkerrors.UserInputError{Message: "quote price must be less than one"}
	}
	source := p.Source
	if source == "" {
		source = "collateral"
	}
	if source != "collateral" && source != "inventory" {
		return nil, &sdkerrors.UserInputError{Message: "quote source must be collateral or inventory"}
	}
	if request.Direction != models.Buy && request.Direction != models.Sell {
		return nil, &sdkerrors.UserInputError{Message: "invalid RFQ direction"}
	}
	var size *big.Int
	if p.Size != nil {
		size, err = units(p.Size)
	} else {
		size, err = units(request.RequestedSize.Value)
		if err == nil {
			switch request.RequestedSize.Unit {
			case "notional":
				size.Quo(new(big.Int).Mul(size, million), price)
			case "shares":
			default:
				return nil, &sdkerrors.UserInputError{Message: "invalid requested size unit"}
			}
		}
	}
	if err != nil {
		return nil, err
	}
	complement := request.Direction == models.Buy && source == "collateral" || request.Direction == models.Sell && source == "inventory"
	orderPrice := new(big.Int).Set(price)
	asset := request.YesPositionID
	if complement {
		orderPrice.Sub(million, price)
		asset = request.NoPositionID
	}
	side := models.Buy
	maker, taker := new(big.Int), new(big.Int).Set(size)
	product := new(big.Int).Mul(orderPrice, size)
	if source == "inventory" {
		side = models.Sell
		maker.Set(size)
		taker.Quo(product, million)
	} else {
		maker.Quo(new(big.Int).Add(product, big.NewInt(999999)), million)
	}
	order := models.SignedOrder{Builder: signing.ZeroBytes32, Metadata: signing.ZeroBytes32, Maker: s.config.Wallet, Signer: s.orderSigner(), AssetID: models.ClobAssetID(asset), MakerAmount: maker, TakerAmount: taker, Side: side, Salt: salt, Timestamp: timestamp, SignatureType: signing.SignatureType(s.config.WalletType), OrderType: models.GTC}
	signature, err := s.config.Signer.Order(order, "0xe3333700cA9d93003F00f0F71f8515005F6c00Aa", models.SignerOwner)
	if err != nil {
		return nil, err
	}
	sideNumber := 0
	if side == models.Sell {
		sideNumber = 1
	}
	signed := map[string]any{"salt": salt.String(), "maker": order.Maker, "signer": order.Signer, "tokenId": asset, "makerAmount": maker.String(), "takerAmount": taker.String(), "side": sideNumber, "signatureType": order.SignatureType, "timestamp": big.NewInt(timestamp).String(), "builder": order.Builder, "metadata": order.Metadata, "signature": signature}
	return map[string]any{"type": "RFQ_QUOTE", "rfq_id": request.RFQID, "price_e6": price.String(), "size_e6": size.String(), "signed_order": signed}, nil
}

// Quote signs one quote locally and waits for its acknowledgement. Nil Size uses the request size.
func (s *Session) Quote(ctx context.Context, request Event, p QuoteParams) (QuoteReference, error) {
	salt, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 64))
	if err != nil {
		return QuoteReference{}, err
	}
	body, err := s.signedQuote(request, p, salt, time.Now().Unix())
	if err != nil {
		return QuoteReference{}, err
	}
	return s.request(ctx, ackKey{kind: "ACK_RFQ_QUOTE", rfq: request.RFQID}, body)
}
