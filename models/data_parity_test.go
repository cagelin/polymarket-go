package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/cagelam/polymarket-go/internal/testutil"
)

func TestPythonDataModels(t *testing.T) {
	raw, err := os.ReadFile("../testdata/data-models-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Model string
			Response    json.RawMessage
			Expected    struct {
				Value json.RawMessage
				Error string
			}
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(t *testing.T) {
			target := newDataParityTarget(test.Model)
			if target == nil {
				t.Fatalf("unknown model %s", test.Model)
			}
			decodeErr := json.Unmarshal(test.Response, target)
			if (decodeErr != nil) != (test.Expected.Error != "") {
				t.Fatalf("%s decode error = %v, Python = %s", test.Model, decodeErr, test.Expected.Error)
			}
			if decodeErr != nil {
				return
			}
			testutil.EqualJSON(t, target, test.Expected.Value)
		})
	}
}

func newDataParityTarget(name string) any {
	switch name {
	case "Position":
		return &Position{}
	case "ComboPositionMarketEvent":
		return &ComboPositionMarketEvent{}
	case "ComboPositionMarket":
		return &ComboPositionMarket{}
	case "ComboPositionLeg":
		return &ComboPositionLeg{}
	case "ComboPosition":
		return &ComboPosition{}
	case "PortfolioValue":
		return &PortfolioValue{}
	case "UserPnLPoint":
		return &UserPnLPoint{}
	case "UserStats":
		return &UserStats{}
	case "UserPnLSeries":
		return &UserPnLSeries{}
	case "UserVolume":
		return &UserVolume{}
	case "ComboCompressActivity":
		return &ComboCompressActivity{}
	case "ComboConvertActivity":
		return &ComboConvertActivity{}
	case "ComboMergeActivity":
		return &ComboMergeActivity{}
	case "ComboRedeemActivity":
		return &ComboRedeemActivity{}
	case "ComboSplitActivity":
		return &ComboSplitActivity{}
	case "ComboTradeActivity":
		return &ComboTradeActivity{}
	case "ComboUnwrapActivity":
		return &ComboUnwrapActivity{}
	case "ComboWrapActivity":
		return &ComboWrapActivity{}
	case "ConversionActivity":
		return &ConversionActivity{}
	case "DepositActivity":
		return &DepositActivity{}
	case "MakerRebateActivity":
		return &MakerRebateActivity{}
	case "MergeActivity":
		return &MergeActivity{}
	case "RedeemActivity":
		return &RedeemActivity{}
	case "ReferralRewardActivity":
		return &ReferralRewardActivity{}
	case "RewardActivity":
		return &RewardActivity{}
	case "SplitActivity":
		return &SplitActivity{}
	case "TakerRebateActivity":
		return &TakerRebateActivity{}
	case "Trade":
		return &Trade{}
	case "TradeActivity":
		return &TradeActivity{}
	case "UnknownActivity":
		return &UnknownActivity{}
	case "WithdrawalActivity":
		return &WithdrawalActivity{}
	case "YieldActivity":
		return &YieldActivity{}
	case "MigrationActivity":
		return &MigrationActivity{}
	case "TipActivity":
		return &TipActivity{}
	case "MarketLiveVolume":
		return &MarketLiveVolume{}
	case "LiveVolume":
		return &LiveVolume{}
	case "OpenInterest":
		return &OpenInterest{}
	case "Holder":
		return &Holder{}
	case "MetaHolder":
		return &MetaHolder{}
	case "TraderLeaderboardEntry":
		return &TraderLeaderboardEntry{}
	case "TraderLeaderboardStanding":
		return &TraderLeaderboardStanding{}
	case "MarketBiggestWinner":
		return &MarketBiggestWinner{}
	case "ComboBiggestWinner":
		return &ComboBiggestWinner{}
	case "BuilderStanding":
		return &BuilderStanding{}
	case "BuilderVolumePoint":
		return &BuilderVolumePoint{}
	case "TokenReference":
		return &TokenReference{}
	case "Resolution":
		return &Resolution{}
	case "CurrentRewardConfig":
		return &CurrentRewardConfig{}
	case "CurrentReward":
		return &CurrentReward{}
	case "MarketRewardConfig":
		return &MarketRewardConfig{}
	case "MarketRewardToken":
		return &MarketRewardToken{}
	case "MarketReward":
		return &MarketReward{}
	case "RewardMarket":
		return &RewardMarket{}
	case "RebatedFee":
		return &RebatedFee{}
	case "BuilderTrade":
		return &BuilderTrade{}
	case "ComboMarketOutcome":
		return &ComboMarketOutcome{}
	case "ComboMarketOutcomes":
		return &ComboMarketOutcomes{}
	case "ComboMarket":
		return &ComboMarket{}
	case "CategoryReference":
		return &CategoryReference{}
	case "CollectionReference":
		return &CollectionReference{}
	case "SeriesReference":
		return &SeriesReference{}
	case "TemplateReference":
		return &TemplateReference{}
	case "Chat":
		return &Chat{}
	case "SportsMarketTypes":
		return &SportsMarketTypes{}
	case "RelatedTag":
		return &RelatedTag{}
	case "Tag":
		return &Tag{}
	case "Series":
		return &Series{}
	case "Comment":
		return &Comment{}
	case "CommentMedia":
		return &CommentMedia{}
	case "CommentPosition":
		return &CommentPosition{}
	case "CommentProfile":
		return &CommentProfile{}
	case "Reaction":
		return &Reaction{}
	case "PublicProfile":
		return &PublicProfile{}
	case "PublicProfileUser":
		return &PublicProfileUser{}
	case "MarketBestBidAskEvent":
		return &MarketBestBidAskEvent{}
	case "MarketBestBidAskPayload":
		return &MarketBestBidAskPayload{}
	case "MarketBookEvent":
		return &MarketBookEvent{}
	case "MarketBookPayload":
		return &MarketBookPayload{}
	case "MarketEventMessage":
		return &MarketEventMessage{}
	case "MarketLastTradePriceEvent":
		return &MarketLastTradePriceEvent{}
	case "MarketLastTradePricePayload":
		return &MarketLastTradePricePayload{}
	case "MarketPriceChangeEvent":
		return &MarketPriceChangeEvent{}
	case "MarketPriceChangePayload":
		return &MarketPriceChangePayload{}
	case "MarketResolvedEvent":
		return &MarketResolvedEvent{}
	case "MarketResolvedPayload":
		return &MarketResolvedPayload{}
	case "MarketTickSizeChangeEvent":
		return &MarketTickSizeChangeEvent{}
	case "MarketTickSizeChangePayload":
		return &MarketTickSizeChangePayload{}
	case "NewMarketEvent":
		return &NewMarketEvent{}
	case "NewMarketPayload":
		return &NewMarketPayload{}
	case "PriceChange":
		return &PriceChange{}
	case "SportsResultEvent":
		return &SportsResultEvent{}
	case "SportsGameResult":
		return &SportsGameResult{}
	case "CommentCreatedEvent":
		return &CommentCreatedEvent{}
	case "CommentRemovedEvent":
		return &CommentRemovedEvent{}
	case "CommentRemovedPayload":
		return &CommentRemovedPayload{}
	case "CryptoPricesBinanceEvent":
		return &CryptoPricesBinanceEvent{}
	case "CryptoPricesChainlinkEvent":
		return &CryptoPricesChainlinkEvent{}
	case "CryptoPricesChainlinkTwapEvent":
		return &CryptoPricesChainlinkTwapEvent{}
	case "CryptoPricesChainlinkTwapPayload":
		return &CryptoPricesChainlinkTwapPayload{}
	case "EquityPriceSnapshotEntry":
		return &EquityPriceSnapshotEntry{}
	case "EquityPriceSubscribePayload":
		return &EquityPriceSubscribePayload{}
	case "EquityPriceUpdatePayload":
		return &EquityPriceUpdatePayload{}
	case "EquityPricesSubscribeEvent":
		return &EquityPricesSubscribeEvent{}
	case "EquityPricesUpdateEvent":
		return &EquityPricesUpdateEvent{}
	case "PriceUpdatePayload":
		return &PriceUpdatePayload{}
	case "ReactionCreatedEvent":
		return &ReactionCreatedEvent{}
	case "ReactionRemovedEvent":
		return &ReactionRemovedEvent{}
	case "UserOrderEvent":
		return &UserOrderEvent{}
	case "UserOrderPayload":
		return &UserOrderPayload{}
	case "UserTradeEvent":
		return &UserTradeEvent{}
	case "UserTradeMakerOrder":
		return &UserTradeMakerOrder{}
	case "UserTradePayload":
		return &UserTradePayload{}
	case "RealtimeEquityTwapSnapshot":
		return &RealtimeEquityTwapSnapshot{}
	case "RealtimeEquityTwapUpdate":
		return &RealtimeEquityTwapUpdate{}
	case "RealtimeEquityPriceSnapshot":
		return &RealtimeEquityPriceSnapshot{}
	case "RealtimeEquityPriceUpdate":
		return &RealtimeEquityPriceUpdate{}
	case "CryptoPriceSnapshotEvent":
		return &CryptoPriceSnapshotEvent{}
	case "CryptoPriceUpdateEvent":
		return &CryptoPriceUpdateEvent{}
	case "CryptoTwapPriceSnapshotEvent":
		return &CryptoTwapPriceSnapshotEvent{}
	case "CryptoTwapPriceUpdateEvent":
		return &CryptoTwapPriceUpdateEvent{}
	case "EquityPriceSnapshotEvent":
		return &EquityPriceSnapshotEvent{}
	case "EquityPriceUpdateEvent":
		return &EquityPriceUpdateEvent{}
	case "EquityTwapPriceSnapshotEvent":
		return &EquityTwapPriceSnapshotEvent{}
	case "EquityTwapPriceUpdateEvent":
		return &EquityTwapPriceUpdateEvent{}
	case "RealtimePricePoint":
		return &RealtimePricePoint{}
	case "RealtimePriceSnapshot":
		return &RealtimePriceSnapshot{}
	case "RealtimePriceUpdate":
		return &RealtimePriceUpdate{}
	case "RealtimeTwapSnapshot":
		return &RealtimeTwapSnapshot{}
	case "RealtimeTwapUpdate":
		return &RealtimeTwapUpdate{}
	case "OpenOrder":
		return &OpenOrder{}
	case "MakerOrder":
		return &MakerOrder{}
	case "ClobTrade":
		return &ClobTrade{}
	case "BuilderAPIKeyInfo":
		return &BuilderAPIKeyInfo{}
	case "BuilderFeeRates":
		return &BuilderFeeRates{}
	case "CancelOrdersResponse":
		return &CancelOrdersResponse{}
	case "OrderHeartbeat":
		return &OrderHeartbeat{}
	case "OrderHeartbeatMismatchResponse":
		return &OrderHeartbeatMismatchResponse{}
	case "RawOrderResponse":
		return &RawOrderResponse{}
	case "AcceptedOrder":
		return &AcceptedOrder{}
	case "RejectedOrder":
		return &RejectedOrder{}
	case "UserEarning":
		return &UserEarning{}
	case "TotalUserEarning":
		return &TotalUserEarning{}
	case "UserRewardsEarning":
		return &UserRewardsEarning{}
	case "AutoRedeemedNotification":
		return &AutoRedeemedNotification{}
	case "AutoRedeemedNotificationPayload":
		return &AutoRedeemedNotificationPayload{}
	case "ChildCommentCreatedNotification":
		return &ChildCommentCreatedNotification{}
	case "ChildCommentNotificationPayload":
		return &ChildCommentNotificationPayload{}
	case "ComboAutoRedeemedNotification":
		return &ComboAutoRedeemedNotification{}
	case "ComboAutoRedeemedNotificationPayload":
		return &ComboAutoRedeemedNotificationPayload{}
	case "MarketNotificationPayload":
		return &MarketNotificationPayload{}
	case "MarketNotificationRewards":
		return &MarketNotificationRewards{}
	case "MarketNotificationRewardsRate":
		return &MarketNotificationRewardsRate{}
	case "MarketNotificationToken":
		return &MarketNotificationToken{}
	case "MarketRegisteredNotification":
		return &MarketRegisteredNotification{}
	case "MarketResolvedNotification":
		return &MarketResolvedNotification{}
	case "OrderCancellationNotification":
		return &OrderCancellationNotification{}
	case "OrderFillFailedNotification":
		return &OrderFillFailedNotification{}
	case "OrderFillNotification":
		return &OrderFillNotification{}
	case "OrderNotificationPayload":
		return &OrderNotificationPayload{}
	case "RewardPayoutNotification":
		return &RewardPayoutNotification{}
	case "RewardPayoutNotificationPayload":
		return &RewardPayoutNotificationPayload{}
	case "TakerTierUpgradedNotification":
		return &TakerTierUpgradedNotification{}
	case "TakerTierUpgradedNotificationPayload":
		return &TakerTierUpgradedNotificationPayload{}
	case "YieldPayoutNotification":
		return &YieldPayoutNotification{}
	case "YieldPayoutNotificationPayload":
		return &YieldPayoutNotificationPayload{}
	case "EarningBreakdown":
		return &EarningBreakdown{}
	case "UserRewardsConfig":
		return &UserRewardsConfig{}
	default:
		return nil
	}
}
