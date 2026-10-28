// Canonical public names; implementations live in domain packages.
// Regenerate with scripts/generate_exports.py. This file is linted.

package polymarket

import (
	"github.com/cagelam/polymarket-go/errors"
	"github.com/cagelam/polymarket-go/internal/clob"
	"github.com/cagelam/polymarket-go/internal/data"
	"github.com/cagelam/polymarket-go/internal/gamma"
	"github.com/cagelam/polymarket-go/internal/rfq"
	"github.com/cagelam/polymarket-go/models"
	"github.com/cagelam/polymarket-go/pagination"
	"github.com/cagelam/polymarket-go/streams"
)

// Activity is the canonical SDK type.
type Activity = models.Activity

// ActivityType is the canonical SDK type.
type ActivityType = models.ActivityType

// ActivityTypeConversion is the canonical SDK value.
const ActivityTypeConversion = models.ActivityTypeConversion

// ActivityTypeDeposit is the canonical SDK value.
const ActivityTypeDeposit = models.ActivityTypeDeposit

// ActivityTypeMakerRebate is the canonical SDK value.
const ActivityTypeMakerRebate = models.ActivityTypeMakerRebate

// ActivityTypeMerge is the canonical SDK value.
const ActivityTypeMerge = models.ActivityTypeMerge

// ActivityTypeMigration is the canonical SDK value.
const ActivityTypeMigration = models.ActivityTypeMigration

// ActivityTypeRedeem is the canonical SDK value.
const ActivityTypeRedeem = models.ActivityTypeRedeem

// ActivityTypeReferralReward is the canonical SDK value.
const ActivityTypeReferralReward = models.ActivityTypeReferralReward

// ActivityTypeReward is the canonical SDK value.
const ActivityTypeReward = models.ActivityTypeReward

// ActivityTypeSplit is the canonical SDK value.
const ActivityTypeSplit = models.ActivityTypeSplit

// ActivityTypeTakerRebate is the canonical SDK value.
const ActivityTypeTakerRebate = models.ActivityTypeTakerRebate

// ActivityTypeTip is the canonical SDK value.
const ActivityTypeTip = models.ActivityTypeTip

// ActivityTypeTrade is the canonical SDK value.
const ActivityTypeTrade = models.ActivityTypeTrade

// ActivityTypeWithdrawal is the canonical SDK value.
const ActivityTypeWithdrawal = models.ActivityTypeWithdrawal

// ActivityTypeYield is the canonical SDK value.
const ActivityTypeYield = models.ActivityTypeYield

// AssetQuery is the canonical SDK type.
type AssetQuery = clob.AssetQuery

// AssetsQuery is the canonical SDK type.
type AssetsQuery = clob.AssetsQuery

// BestLine is the canonical SDK type.
type BestLine = models.BestLine

// BestLineID is the canonical SDK type.
type BestLineID = models.BestLineID

// BiggestWinner is the canonical SDK type.
type BiggestWinner = models.BiggestWinner

// BiggestWinnerKind is the canonical SDK type.
type BiggestWinnerKind = models.BiggestWinnerKind

// BiggestWinnerKindCombo is the canonical SDK value.
const BiggestWinnerKindCombo = models.BiggestWinnerKindCombo

// BiggestWinnerKindMarket is the canonical SDK value.
const BiggestWinnerKindMarket = models.BiggestWinnerKindMarket

// BuilderStanding is the canonical SDK type.
type BuilderStanding = models.BuilderStanding

// BuilderTrade is the canonical SDK type.
type BuilderTrade = models.BuilderTrade

// BuilderVolumePoint is the canonical SDK type.
type BuilderVolumePoint = models.BuilderVolumePoint

// Buy is the canonical SDK value.
const Buy = models.Buy

// CategoryID is the canonical SDK type.
type CategoryID = models.CategoryID

// CategoryReference is the canonical SDK type.
type CategoryReference = models.CategoryReference

// Chat is the canonical SDK type.
type Chat = models.Chat

// ChatID is the canonical SDK type.
type ChatID = models.ChatID

// ClobAssetID is the canonical SDK type.
type ClobAssetID = models.ClobAssetID

// ClobReward is the canonical SDK type.
type ClobReward = models.ClobReward

// ClobRewardID is the canonical SDK type.
type ClobRewardID = models.ClobRewardID

// CollectionID is the canonical SDK type.
type CollectionID = models.CollectionID

// CollectionReference is the canonical SDK type.
type CollectionReference = models.CollectionReference

// ComboActivity is the canonical SDK type.
type ComboActivity = models.ComboActivity

// ComboActivityID is the canonical SDK type.
type ComboActivityID = models.ComboActivityID

// ComboActivityType is the canonical SDK type.
type ComboActivityType = models.ComboActivityType

// ComboActivityTypeCompress is the canonical SDK value.
const ComboActivityTypeCompress = models.ComboActivityTypeCompress

// ComboActivityTypeConvert is the canonical SDK value.
const ComboActivityTypeConvert = models.ComboActivityTypeConvert

// ComboActivityTypeMerge is the canonical SDK value.
const ComboActivityTypeMerge = models.ComboActivityTypeMerge

// ComboActivityTypeRedeem is the canonical SDK value.
const ComboActivityTypeRedeem = models.ComboActivityTypeRedeem

// ComboActivityTypeSplit is the canonical SDK value.
const ComboActivityTypeSplit = models.ComboActivityTypeSplit

// ComboActivityTypeUnwrap is the canonical SDK value.
const ComboActivityTypeUnwrap = models.ComboActivityTypeUnwrap

// ComboActivityTypeWrap is the canonical SDK value.
const ComboActivityTypeWrap = models.ComboActivityTypeWrap

// ComboBiggestWinner is the canonical SDK type.
type ComboBiggestWinner = models.ComboBiggestWinner

// ComboCompressActivity is the canonical SDK type.
type ComboCompressActivity = models.ComboCompressActivity

// ComboConditionID is the canonical SDK type.
type ComboConditionID = models.ComboConditionID

// ComboConvertActivity is the canonical SDK type.
type ComboConvertActivity = models.ComboConvertActivity

// ComboMarket is the canonical SDK type.
type ComboMarket = models.ComboMarket

// ComboMarketOutcome is the canonical SDK type.
type ComboMarketOutcome = models.ComboMarketOutcome

// ComboMarketOutcomes is the canonical SDK type.
type ComboMarketOutcomes = models.ComboMarketOutcomes

// ComboMergeActivity is the canonical SDK type.
type ComboMergeActivity = models.ComboMergeActivity

// ComboPosition is the canonical SDK type.
type ComboPosition = models.ComboPosition

// ComboPositionLeg is the canonical SDK type.
type ComboPositionLeg = models.ComboPositionLeg

// ComboPositionMarket is the canonical SDK type.
type ComboPositionMarket = models.ComboPositionMarket

// ComboPositionMarketEvent is the canonical SDK type.
type ComboPositionMarketEvent = models.ComboPositionMarketEvent

// ComboPositionStatus is the canonical SDK type.
type ComboPositionStatus = models.ComboPositionStatus

// ComboPositionStatusOpen is the canonical SDK value.
const ComboPositionStatusOpen = models.ComboPositionStatusOpen

// ComboPositionStatusPartial is the canonical SDK value.
const ComboPositionStatusPartial = models.ComboPositionStatusPartial

// ComboPositionStatusRedeemable is the canonical SDK value.
const ComboPositionStatusRedeemable = models.ComboPositionStatusRedeemable

// ComboPositionStatusResolvedLoss is the canonical SDK value.
const ComboPositionStatusResolvedLoss = models.ComboPositionStatusResolvedLoss

// ComboPositionStatusResolvedPartial is the canonical SDK value.
const ComboPositionStatusResolvedPartial = models.ComboPositionStatusResolvedPartial

// ComboPositionStatusResolvedWin is the canonical SDK value.
const ComboPositionStatusResolvedWin = models.ComboPositionStatusResolvedWin

// ComboRedeemActivity is the canonical SDK type.
type ComboRedeemActivity = models.ComboRedeemActivity

// ComboSplitActivity is the canonical SDK type.
type ComboSplitActivity = models.ComboSplitActivity

// ComboTradeActivity is the canonical SDK type.
type ComboTradeActivity = models.ComboTradeActivity

// ComboUnwrapActivity is the canonical SDK type.
type ComboUnwrapActivity = models.ComboUnwrapActivity

// ComboWrapActivity is the canonical SDK type.
type ComboWrapActivity = models.ComboWrapActivity

// Comment is the canonical SDK type.
type Comment = models.Comment

// CommentCreatedEvent is the canonical SDK type.
type CommentCreatedEvent = models.CommentCreatedEvent

// CommentID is the canonical SDK type.
type CommentID = models.CommentID

// CommentMedia is the canonical SDK type.
type CommentMedia = models.CommentMedia

// CommentParentEntityType is the canonical SDK type.
type CommentParentEntityType = gamma.CommentParentEntityType

// CommentParentEvent is the canonical SDK value.
const CommentParentEvent = gamma.CommentParentEvent

// CommentParentSeries is the canonical SDK value.
const CommentParentSeries = gamma.CommentParentSeries

// CommentPosition is the canonical SDK type.
type CommentPosition = models.CommentPosition

// CommentProfile is the canonical SDK type.
type CommentProfile = models.CommentProfile

// CommentRemovedEvent is the canonical SDK type.
type CommentRemovedEvent = models.CommentRemovedEvent

// CommentRemovedPayload is the canonical SDK type.
type CommentRemovedPayload = models.CommentRemovedPayload

// CommentsSpec is the canonical SDK type.
type CommentsSpec = streams.CommentsSpec

// ConditionID is the canonical SDK type.
type ConditionID = models.ConditionID

// ConversionActivity is the canonical SDK type.
type ConversionActivity = models.ConversionActivity

// CryptoPricesBinanceEvent is the canonical SDK type.
type CryptoPricesBinanceEvent = models.CryptoPricesBinanceEvent

// CryptoPricesChainlinkEvent is the canonical SDK type.
type CryptoPricesChainlinkEvent = models.CryptoPricesChainlinkEvent

// CryptoPricesChainlinkTwapEvent is the canonical SDK type.
type CryptoPricesChainlinkTwapEvent = models.CryptoPricesChainlinkTwapEvent

// CryptoPricesChainlinkTwapPayload is the canonical SDK type.
type CryptoPricesChainlinkTwapPayload = models.CryptoPricesChainlinkTwapPayload

// CryptoPricesChainlinkTwapSpec is the canonical SDK type.
type CryptoPricesChainlinkTwapSpec = streams.CryptoPricesChainlinkTwapSpec

// CryptoPricesSpec is the canonical SDK type.
type CryptoPricesSpec = streams.CryptoPricesSpec

// CurrentReward is the canonical SDK type.
type CurrentReward = models.CurrentReward

// CurrentRewardConfig is the canonical SDK type.
type CurrentRewardConfig = models.CurrentRewardConfig

// Date is the canonical SDK type.
type Date = models.Date

// DepositActivity is the canonical SDK type.
type DepositActivity = models.DepositActivity

// DownloadAccountingSnapshotParams is the canonical SDK type.
type DownloadAccountingSnapshotParams = data.DownloadAccountingSnapshotParams

// ERC1155TradingApproval is the canonical SDK type.
type ERC1155TradingApproval = models.ERC1155TradingApproval

// ERC20TradingApproval is the canonical SDK type.
type ERC20TradingApproval = models.ERC20TradingApproval

// EVMAddress is the canonical SDK type.
type EVMAddress = models.EVMAddress

// Environment is the canonical SDK type.
type Environment = models.Environment

// EquityPriceSnapshotEntry is the canonical SDK type.
type EquityPriceSnapshotEntry = models.EquityPriceSnapshotEntry

// EquityPriceSubscribePayload is the canonical SDK type.
type EquityPriceSubscribePayload = models.EquityPriceSubscribePayload

// EquityPriceUpdatePayload is the canonical SDK type.
type EquityPriceUpdatePayload = models.EquityPriceUpdatePayload

// EquityPricesSpec is the canonical SDK type.
type EquityPricesSpec = streams.EquityPricesSpec

// EquityPricesSubscribeEvent is the canonical SDK type.
type EquityPricesSubscribeEvent = models.EquityPricesSubscribeEvent

// EquityPricesUpdateEvent is the canonical SDK type.
type EquityPricesUpdateEvent = models.EquityPricesUpdateEvent

// EstimateMarketPriceParams is the canonical SDK type.
type EstimateMarketPriceParams = clob.EstimateMarketPriceParams

// Event is the canonical SDK type.
type Event = models.Event

// EventCreator is the canonical SDK type.
type EventCreator = models.EventCreator

// EventCreatorID is the canonical SDK type.
type EventCreatorID = models.EventCreatorID

// EventDisplay is the canonical SDK type.
type EventDisplay = models.EventDisplay

// EventEstimation is the canonical SDK type.
type EventEstimation = models.EventEstimation

// EventExternalPartnerMappingID is the canonical SDK type.
type EventExternalPartnerMappingID = models.EventExternalPartnerMappingID

// EventID is the canonical SDK type.
type EventID = models.EventID

// EventMetrics is the canonical SDK type.
type EventMetrics = models.EventMetrics

// EventPartner is the canonical SDK type.
type EventPartner = models.EventPartner

// EventQuery is the canonical SDK type.
type EventQuery = gamma.EventQuery

// EventResolution is the canonical SDK type.
type EventResolution = models.EventResolution

// EventSchedule is the canonical SDK type.
type EventSchedule = models.EventSchedule

// EventSeries is the canonical SDK type.
type EventSeries = models.EventSeries

// EventSportsMetadata is the canonical SDK type.
type EventSportsMetadata = models.EventSportsMetadata

// EventState is the canonical SDK type.
type EventState = models.EventState

// EventTag is the canonical SDK type.
type EventTag = models.EventTag

// EventTrading is the canonical SDK type.
type EventTrading = models.EventTrading

// FAK is the canonical SDK value.
const FAK = clob.FAK

// FOK is the canonical SDK value.
const FOK = clob.FOK

// FeeSchedule is the canonical SDK type.
type FeeSchedule = models.FeeSchedule

// GetBuilderVolumesParams is the canonical SDK type.
type GetBuilderVolumesParams = data.GetBuilderVolumesParams

// GetCurrentRebatesParams is the canonical SDK type.
type GetCurrentRebatesParams = clob.GetCurrentRebatesParams

// GetEventLiveVolumeParams is the canonical SDK type.
type GetEventLiveVolumeParams = data.GetEventLiveVolumeParams

// GetOpenInterestsParams is the canonical SDK type.
type GetOpenInterestsParams = data.GetOpenInterestsParams

// GetPortfolioValueParams is the canonical SDK type.
type GetPortfolioValueParams = data.GetPortfolioValueParams

// GetResolutionsParams is the canonical SDK type.
type GetResolutionsParams = data.GetResolutionsParams

// GetTokenReferencesParams is the canonical SDK type.
type GetTokenReferencesParams = data.GetTokenReferencesParams

// GetTraderLeaderboardStandingParams is the canonical SDK type.
type GetTraderLeaderboardStandingParams = data.GetTraderLeaderboardStandingParams

// GetUserPnLParams is the canonical SDK type.
type GetUserPnLParams = data.GetUserPnLParams

// GetUserStatsParams is the canonical SDK type.
type GetUserStatsParams = data.GetUserStatsParams

// GetUserVolumeParams is the canonical SDK type.
type GetUserVolumeParams = data.GetUserVolumeParams

// Handle is the canonical SDK type.
type Handle = streams.Handle

// HexString is the canonical SDK type.
type HexString = models.HexString

// HistoryAll is the canonical SDK value.
const HistoryAll = models.HistoryAll

// HistoryDay is the canonical SDK value.
const HistoryDay = models.HistoryDay

// HistoryHour is the canonical SDK value.
const HistoryHour = models.HistoryHour

// HistoryMax is the canonical SDK value.
const HistoryMax = models.HistoryMax

// HistoryMonth is the canonical SDK value.
const HistoryMonth = models.HistoryMonth

// HistorySixHours is the canonical SDK value.
const HistorySixHours = models.HistorySixHours

// HistoryWeek is the canonical SDK value.
const HistoryWeek = models.HistoryWeek

// Holder is the canonical SDK type.
type Holder = models.Holder

// ImageOptimization is the canonical SDK type.
type ImageOptimization = models.ImageOptimization

// ImageOptimizationID is the canonical SDK type.
type ImageOptimizationID = models.ImageOptimizationID

// InsufficientLiquidityError is the canonical SDK type.
type InsufficientLiquidityError = errors.InsufficientLiquidityError

// LastTradePrice is the canonical SDK type.
type LastTradePrice = models.LastTradePrice

// LastTradePriceForAsset is the canonical SDK type.
type LastTradePriceForAsset = models.LastTradePriceForAsset

// ListActivityParams is the canonical SDK type.
type ListActivityParams = data.ListActivityParams

// ListBiggestWinnersParams is the canonical SDK type.
type ListBiggestWinnersParams = data.ListBiggestWinnersParams

// ListBuilderLeaderboardParams is the canonical SDK type.
type ListBuilderLeaderboardParams = data.ListBuilderLeaderboardParams

// ListBuilderTradesParams is the canonical SDK type.
type ListBuilderTradesParams = clob.ListBuilderTradesParams

// ListComboActivityParams is the canonical SDK type.
type ListComboActivityParams = data.ListComboActivityParams

// ListComboMarketsParams is the canonical SDK type.
type ListComboMarketsParams = rfq.ListComboMarketsParams

// ListComboPositionsParams is the canonical SDK type.
type ListComboPositionsParams = data.ListComboPositionsParams

// ListCommentsByUserAddressParams is the canonical SDK type.
type ListCommentsByUserAddressParams = gamma.ListCommentsByUserAddressParams

// ListCommentsParams is the canonical SDK type.
type ListCommentsParams = gamma.ListCommentsParams

// ListCurrentRewardsParams is the canonical SDK type.
type ListCurrentRewardsParams = clob.ListCurrentRewardsParams

// ListEventsParams is the canonical SDK type.
type ListEventsParams = gamma.ListEventsParams

// ListMarketHoldersParams is the canonical SDK type.
type ListMarketHoldersParams = data.ListMarketHoldersParams

// ListMarketRewardsParams is the canonical SDK type.
type ListMarketRewardsParams = clob.ListMarketRewardsParams

// ListMarketsParams is the canonical SDK type.
type ListMarketsParams = gamma.ListMarketsParams

// ListPositionsParams is the canonical SDK type.
type ListPositionsParams = data.ListPositionsParams

// ListRewardMarketsParams is the canonical SDK type.
type ListRewardMarketsParams = clob.ListRewardMarketsParams

// ListSeriesParams is the canonical SDK type.
type ListSeriesParams = gamma.ListSeriesParams

// ListTagsParams is the canonical SDK type.
type ListTagsParams = gamma.ListTagsParams

// ListTeamsParams is the canonical SDK type.
type ListTeamsParams = gamma.ListTeamsParams

// ListTraderLeaderboardParams is the canonical SDK type.
type ListTraderLeaderboardParams = data.ListTraderLeaderboardParams

// ListTradesParams is the canonical SDK type.
type ListTradesParams = data.ListTradesParams

// LiveVolume is the canonical SDK type.
type LiveVolume = models.LiveVolume

// MakerRebateActivity is the canonical SDK type.
type MakerRebateActivity = models.MakerRebateActivity

// Market is the canonical SDK type.
type Market = models.Market

// MarketBestBidAskEvent is the canonical SDK type.
type MarketBestBidAskEvent = models.MarketBestBidAskEvent

// MarketBestBidAskPayload is the canonical SDK type.
type MarketBestBidAskPayload = models.MarketBestBidAskPayload

// MarketBiggestWinner is the canonical SDK type.
type MarketBiggestWinner = models.MarketBiggestWinner

// MarketBookEvent is the canonical SDK type.
type MarketBookEvent = models.MarketBookEvent

// MarketBookPayload is the canonical SDK type.
type MarketBookPayload = models.MarketBookPayload

// MarketEvent is the canonical SDK type.
type MarketEvent = models.MarketEvent

// MarketEventMessage is the canonical SDK type.
type MarketEventMessage = models.MarketEventMessage

// MarketID is the canonical SDK type.
type MarketID = models.MarketID

// MarketLastTradePriceEvent is the canonical SDK type.
type MarketLastTradePriceEvent = models.MarketLastTradePriceEvent

// MarketLastTradePricePayload is the canonical SDK type.
type MarketLastTradePricePayload = models.MarketLastTradePricePayload

// MarketLiveVolume is the canonical SDK type.
type MarketLiveVolume = models.MarketLiveVolume

// MarketMetrics is the canonical SDK type.
type MarketMetrics = models.MarketMetrics

// MarketOrderType is the canonical SDK type.
type MarketOrderType = clob.MarketOrderType

// MarketOutcome is the canonical SDK type.
type MarketOutcome = models.MarketOutcome

// MarketOutcomes is the canonical SDK type.
type MarketOutcomes = models.MarketOutcomes

// MarketPriceChangeEvent is the canonical SDK type.
type MarketPriceChangeEvent = models.MarketPriceChangeEvent

// MarketPriceChangePayload is the canonical SDK type.
type MarketPriceChangePayload = models.MarketPriceChangePayload

// MarketPrices is the canonical SDK type.
type MarketPrices = models.MarketPrices

// MarketQuery is the canonical SDK type.
type MarketQuery = gamma.MarketQuery

// MarketResolution is the canonical SDK type.
type MarketResolution = models.MarketResolution

// MarketResolutionStatus is the canonical SDK type.
type MarketResolutionStatus = models.MarketResolutionStatus

// MarketResolvedEvent is the canonical SDK type.
type MarketResolvedEvent = models.MarketResolvedEvent

// MarketResolvedPayload is the canonical SDK type.
type MarketResolvedPayload = models.MarketResolvedPayload

// MarketReward is the canonical SDK type.
type MarketReward = models.MarketReward

// MarketRewardConfig is the canonical SDK type.
type MarketRewardConfig = models.MarketRewardConfig

// MarketRewardToken is the canonical SDK type.
type MarketRewardToken = models.MarketRewardToken

// MarketRewards is the canonical SDK type.
type MarketRewards = models.MarketRewards

// MarketSpec is the canonical SDK type.
type MarketSpec = streams.MarketSpec

// MarketSportsMetadata is the canonical SDK type.
type MarketSportsMetadata = models.MarketSportsMetadata

// MarketState is the canonical SDK type.
type MarketState = models.MarketState

// MarketTag is the canonical SDK type.
type MarketTag = models.MarketTag

// MarketTickSizeChangeEvent is the canonical SDK type.
type MarketTickSizeChangeEvent = models.MarketTickSizeChangeEvent

// MarketTickSizeChangePayload is the canonical SDK type.
type MarketTickSizeChangePayload = models.MarketTickSizeChangePayload

// MarketTrading is the canonical SDK type.
type MarketTrading = models.MarketTrading

// MergeActivity is the canonical SDK type.
type MergeActivity = models.MergeActivity

// MetaHolder is the canonical SDK type.
type MetaHolder = models.MetaHolder

// MigrationActivity is the canonical SDK type.
type MigrationActivity = models.MigrationActivity

// MissingTradingApprovals is the canonical SDK type.
type MissingTradingApprovals = models.MissingTradingApprovals

// NewMarketEvent is the canonical SDK type.
type NewMarketEvent = models.NewMarketEvent

// NewMarketPayload is the canonical SDK type.
type NewMarketPayload = models.NewMarketPayload

// OpenInterest is the canonical SDK type.
type OpenInterest = models.OpenInterest

// OrderBook is the canonical SDK type.
type OrderBook = models.OrderBook

// OrderBookLevel is the canonical SDK type.
type OrderBookLevel = models.OrderBookLevel

// OrderSide is the canonical SDK type.
type OrderSide = models.OrderSide

// PaginationLimitError is the canonical SDK type.
type PaginationLimitError = errors.PaginationLimitError

// Partner is the canonical SDK type.
type Partner = models.Partner

// PartnerID is the canonical SDK type.
type PartnerID = models.PartnerID

// PortfolioValue is the canonical SDK type.
type PortfolioValue = models.PortfolioValue

// Position is the canonical SDK type.
type Position = models.Position

// PositionID is the canonical SDK type.
type PositionID = models.PositionID

// PositionStatus is the canonical SDK type.
type PositionStatus = models.PositionStatus

// PositionStatusClosed is the canonical SDK value.
const PositionStatusClosed = models.PositionStatusClosed

// PositionStatusMergeable is the canonical SDK value.
const PositionStatusMergeable = models.PositionStatusMergeable

// PositionStatusOpen is the canonical SDK value.
const PositionStatusOpen = models.PositionStatusOpen

// PositionStatusRedeemable is the canonical SDK value.
const PositionStatusRedeemable = models.PositionStatusRedeemable

// PositionStatusRedeemableLost is the canonical SDK value.
const PositionStatusRedeemableLost = models.PositionStatusRedeemableLost

// PriceChange is the canonical SDK type.
type PriceChange = models.PriceChange

// PriceHistoryInterval is the canonical SDK type.
type PriceHistoryInterval = models.PriceHistoryInterval

// PriceHistoryParams is the canonical SDK type.
type PriceHistoryParams = data.PriceHistoryParams

// PriceHistoryPoint is the canonical SDK type.
type PriceHistoryPoint = models.PriceHistoryPoint

// PriceRequest is the canonical SDK type.
type PriceRequest = clob.PriceRequest

// PriceUpdatePayload is the canonical SDK type.
type PriceUpdatePayload = models.PriceUpdatePayload

// Profile is the canonical SDK type.
type Profile = models.Profile

// ProtocolV1 is the canonical SDK value.
const ProtocolV1 = models.ProtocolV1

// ProtocolV2 is the canonical SDK value.
const ProtocolV2 = models.ProtocolV2

// ProtocolVersion is the canonical SDK type.
type ProtocolVersion = models.ProtocolVersion

// PublicProfile is the canonical SDK type.
type PublicProfile = models.PublicProfile

// PublicProfileUser is the canonical SDK type.
type PublicProfileUser = models.PublicProfileUser

// QuestionID is the canonical SDK type.
type QuestionID = models.QuestionID

// RateLimitError is the canonical SDK type.
type RateLimitError = errors.RateLimitError

// RateLimitUpdate is the canonical SDK type.
type RateLimitUpdate = errors.RateLimitUpdate

// Reaction is the canonical SDK type.
type Reaction = models.Reaction

// ReactionCreatedEvent is the canonical SDK type.
type ReactionCreatedEvent = models.ReactionCreatedEvent

// ReactionRemovedEvent is the canonical SDK type.
type ReactionRemovedEvent = models.ReactionRemovedEvent

// RebatedFee is the canonical SDK type.
type RebatedFee = models.RebatedFee

// Recurrence is the canonical SDK type.
type Recurrence = gamma.Recurrence

// RecurrenceDaily is the canonical SDK value.
const RecurrenceDaily = gamma.RecurrenceDaily

// RecurrenceMonthly is the canonical SDK value.
const RecurrenceMonthly = gamma.RecurrenceMonthly

// RecurrenceWeekly is the canonical SDK value.
const RecurrenceWeekly = gamma.RecurrenceWeekly

// RedeemActivity is the canonical SDK type.
type RedeemActivity = models.RedeemActivity

// ReferralRewardActivity is the canonical SDK type.
type ReferralRewardActivity = models.ReferralRewardActivity

// RelatedTag is the canonical SDK type.
type RelatedTag = models.RelatedTag

// RelatedTagResourcesQuery is the canonical SDK type.
type RelatedTagResourcesQuery = gamma.RelatedTagResourcesQuery

// RelatedTagsQuery is the canonical SDK type.
type RelatedTagsQuery = gamma.RelatedTagsQuery

// RequestRejectedError is the canonical SDK type.
type RequestRejectedError = errors.RequestRejectedError

// Resolution is the canonical SDK type.
type Resolution = models.Resolution

// ResolutionActive is the canonical SDK value.
const ResolutionActive = models.ResolutionActive

// ResolutionInactive is the canonical SDK value.
const ResolutionInactive = models.ResolutionInactive

// ResolutionMarketType is the canonical SDK type.
type ResolutionMarketType = models.ResolutionMarketType

// ResolutionMarketTypeAtomicNegrisk is the canonical SDK value.
const ResolutionMarketTypeAtomicNegrisk = models.ResolutionMarketTypeAtomicNegrisk

// ResolutionMarketTypeBinary is the canonical SDK value.
const ResolutionMarketTypeBinary = models.ResolutionMarketTypeBinary

// ResolutionMarketTypeIncrementalNegrisk is the canonical SDK value.
const ResolutionMarketTypeIncrementalNegrisk = models.ResolutionMarketTypeIncrementalNegrisk

// ResolutionReporter is the canonical SDK type.
type ResolutionReporter = models.ResolutionReporter

// ResolutionReporterChainlink is the canonical SDK value.
const ResolutionReporterChainlink = models.ResolutionReporterChainlink

// ResolutionReporterEoa is the canonical SDK value.
const ResolutionReporterEoa = models.ResolutionReporterEoa

// ResolutionReporterUMAOptimisticOracle is the canonical SDK value.
const ResolutionReporterUMAOptimisticOracle = models.ResolutionReporterUMAOptimisticOracle

// ResolutionRequestID is the canonical SDK type.
type ResolutionRequestID = models.ResolutionRequestID

// ResolutionResolved is the canonical SDK value.
const ResolutionResolved = models.ResolutionResolved

// ResolutionSettlementTimeBasis is the canonical SDK type.
type ResolutionSettlementTimeBasis = models.ResolutionSettlementTimeBasis

// ResolutionSettlementTimeBasisDvmRoundEstimate is the canonical SDK value.
const ResolutionSettlementTimeBasisDvmRoundEstimate = models.ResolutionSettlementTimeBasisDvmRoundEstimate

// ResolutionSettlementTimeBasisLiveness is the canonical SDK value.
const ResolutionSettlementTimeBasisLiveness = models.ResolutionSettlementTimeBasisLiveness

// ResolutionSettlementTimeBasisManagedProposalExpiration is the canonical SDK value.
const ResolutionSettlementTimeBasisManagedProposalExpiration = models.ResolutionSettlementTimeBasisManagedProposalExpiration

// ResolutionSettlementTimeBasisProposalExpiration is the canonical SDK value.
const ResolutionSettlementTimeBasisProposalExpiration = models.ResolutionSettlementTimeBasisProposalExpiration

// ResolutionSource is the canonical SDK type.
type ResolutionSource = models.ResolutionSource

// ResolutionSourceDerived is the canonical SDK value.
const ResolutionSourceDerived = models.ResolutionSourceDerived

// ResolutionSourceReported is the canonical SDK value.
const ResolutionSourceReported = models.ResolutionSourceReported

// ResolutionStatus is the canonical SDK type.
type ResolutionStatus = models.ResolutionStatus

// ResolutionStatusActive is the canonical SDK value.
const ResolutionStatusActive = models.ResolutionStatusActive

// ResolutionStatusArbitration is the canonical SDK value.
const ResolutionStatusArbitration = models.ResolutionStatusArbitration

// ResolutionStatusChallenged is the canonical SDK value.
const ResolutionStatusChallenged = models.ResolutionStatusChallenged

// ResolutionStatusDisputed is the canonical SDK value.
const ResolutionStatusDisputed = models.ResolutionStatusDisputed

// ResolutionStatusInitialized is the canonical SDK value.
const ResolutionStatusInitialized = models.ResolutionStatusInitialized

// ResolutionStatusPosed is the canonical SDK value.
const ResolutionStatusPosed = models.ResolutionStatusPosed

// ResolutionStatusProposed is the canonical SDK value.
const ResolutionStatusProposed = models.ResolutionStatusProposed

// ResolutionStatusReproposed is the canonical SDK value.
const ResolutionStatusReproposed = models.ResolutionStatusReproposed

// ResolutionStatusResolved is the canonical SDK value.
const ResolutionStatusResolved = models.ResolutionStatusResolved

// RewardActivity is the canonical SDK type.
type RewardActivity = models.RewardActivity

// RewardMarket is the canonical SDK type.
type RewardMarket = models.RewardMarket

// SearchByClosedTime is the canonical SDK value.
const SearchByClosedTime = gamma.SearchByClosedTime

// SearchByCompetitive is the canonical SDK value.
const SearchByCompetitive = gamma.SearchByCompetitive

// SearchByEndDate is the canonical SDK value.
const SearchByEndDate = gamma.SearchByEndDate

// SearchByLiquidity is the canonical SDK value.
const SearchByLiquidity = gamma.SearchByLiquidity

// SearchByStartDate is the canonical SDK value.
const SearchByStartDate = gamma.SearchByStartDate

// SearchByVolume is the canonical SDK value.
const SearchByVolume = gamma.SearchByVolume

// SearchByVolume24hr is the canonical SDK value.
const SearchByVolume24hr = gamma.SearchByVolume24hr

// SearchParams is the canonical SDK type.
type SearchParams = gamma.SearchParams

// SearchResults is the canonical SDK type.
type SearchResults = models.SearchResults

// SearchSort is the canonical SDK type.
type SearchSort = gamma.SearchSort

// SearchTag is the canonical SDK type.
type SearchTag = models.SearchTag

// Sell is the canonical SDK value.
const Sell = models.Sell

// Series is the canonical SDK type.
type Series = models.Series

// SeriesID is the canonical SDK type.
type SeriesID = models.SeriesID

// SeriesReference is the canonical SDK type.
type SeriesReference = models.SeriesReference

// Spec is the canonical SDK type.
type Spec = streams.Spec

// SplitActivity is the canonical SDK type.
type SplitActivity = models.SplitActivity

// SportID is the canonical SDK type.
type SportID = models.SportID

// SportsGameResult is the canonical SDK type.
type SportsGameResult = models.SportsGameResult

// SportsMarketTypes is the canonical SDK type.
type SportsMarketTypes = models.SportsMarketTypes

// SportsMetadata is the canonical SDK type.
type SportsMetadata = models.SportsMetadata

// SportsResultEvent is the canonical SDK type.
type SportsResultEvent = models.SportsResultEvent

// SportsSpec is the canonical SDK type.
type SportsSpec = streams.SportsSpec

// Stats is the canonical SDK type.
type Stats = streams.Stats

// StreamEvent is the canonical SDK type.
type StreamEvent = models.StreamEvent

// Tag is the canonical SDK type.
type Tag = models.Tag

// TagID is the canonical SDK type.
type TagID = models.TagID

// TagMatch is the canonical SDK type.
type TagMatch = gamma.TagMatch

// TagMatchAll is the canonical SDK value.
const TagMatchAll = gamma.TagMatchAll

// TagMatchAny is the canonical SDK value.
const TagMatchAny = gamma.TagMatchAny

// TagQuery is the canonical SDK type.
type TagQuery = gamma.TagQuery

// TagReference is the canonical SDK type.
type TagReference = models.TagReference

// TakerRebateActivity is the canonical SDK type.
type TakerRebateActivity = models.TakerRebateActivity

// Team is the canonical SDK type.
type Team = models.Team

// TeamAway is the canonical SDK value.
const TeamAway = models.TeamAway

// TeamHome is the canonical SDK value.
const TeamHome = models.TeamHome

// TeamID is the canonical SDK type.
type TeamID = models.TeamID

// TeamOrdering is the canonical SDK type.
type TeamOrdering = models.TeamOrdering

// TemplateID is the canonical SDK type.
type TemplateID = models.TemplateID

// TemplateReference is the canonical SDK type.
type TemplateReference = models.TemplateReference

// TipActivity is the canonical SDK type.
type TipActivity = models.TipActivity

// TipSide is the canonical SDK type.
type TipSide = models.TipSide

// TipSideIn is the canonical SDK value.
const TipSideIn = models.TipSideIn

// TipSideOut is the canonical SDK value.
const TipSideOut = models.TipSideOut

// TokenID is the canonical SDK type.
type TokenID = models.TokenID

// TokenModule is the canonical SDK type.
type TokenModule = models.TokenModule

// TokenModuleBinary is the canonical SDK value.
const TokenModuleBinary = models.TokenModuleBinary

// TokenModuleCombo is the canonical SDK value.
const TokenModuleCombo = models.TokenModuleCombo

// TokenModuleNegRisk is the canonical SDK value.
const TokenModuleNegRisk = models.TokenModuleNegRisk

// TokenModuleV1Ctf is the canonical SDK value.
const TokenModuleV1Ctf = models.TokenModuleV1Ctf

// TokenReference is the canonical SDK type.
type TokenReference = models.TokenReference

// Trade is the canonical SDK type.
type Trade = models.Trade

// TradeActivity is the canonical SDK type.
type TradeActivity = models.TradeActivity

// TraderLeaderboardEntry is the canonical SDK type.
type TraderLeaderboardEntry = models.TraderLeaderboardEntry

// TraderLeaderboardStanding is the canonical SDK type.
type TraderLeaderboardStanding = models.TraderLeaderboardStanding

// TradingApprovalsState is the canonical SDK type.
type TradingApprovalsState = models.TradingApprovalsState

// TradingCancelOnly is the canonical SDK value.
const TradingCancelOnly = errors.TradingCancelOnly

// TradingPostOnly is the canonical SDK value.
const TradingPostOnly = errors.TradingPostOnly

// TradingRestarting is the canonical SDK value.
const TradingRestarting = errors.TradingRestarting

// TradingRestriction is the canonical SDK type.
type TradingRestriction = errors.TradingRestriction

// TransactionHash is the canonical SDK type.
type TransactionHash = models.TransactionHash

// TransportError is the canonical SDK type.
type TransportError = errors.TransportError

// UMADisputed is the canonical SDK value.
const UMADisputed = models.UMADisputed

// UMAProposed is the canonical SDK value.
const UMAProposed = models.UMAProposed

// UMARequested is the canonical SDK value.
const UMARequested = models.UMARequested

// UMAResolutionStatus is the canonical SDK type.
type UMAResolutionStatus = models.UMAResolutionStatus

// UMAResolved is the canonical SDK value.
const UMAResolved = models.UMAResolved

// UMASettled is the canonical SDK value.
const UMASettled = models.UMASettled

// UnexpectedResponseError is the canonical SDK type.
type UnexpectedResponseError = errors.UnexpectedResponseError

// UnknownActivity is the canonical SDK type.
type UnknownActivity = models.UnknownActivity

// UserInputError is the canonical SDK type.
type UserInputError = errors.UserInputError

// UserPnLFidelity is the canonical SDK type.
type UserPnLFidelity = models.UserPnLFidelity

// UserPnLFidelityEighteenHours is the canonical SDK value.
const UserPnLFidelityEighteenHours = models.UserPnLFidelityEighteenHours

// UserPnLFidelityOneDay is the canonical SDK value.
const UserPnLFidelityOneDay = models.UserPnLFidelityOneDay

// UserPnLFidelityOneHour is the canonical SDK value.
const UserPnLFidelityOneHour = models.UserPnLFidelityOneHour

// UserPnLFidelityThreeHours is the canonical SDK value.
const UserPnLFidelityThreeHours = models.UserPnLFidelityThreeHours

// UserPnLFidelityTwelveHours is the canonical SDK value.
const UserPnLFidelityTwelveHours = models.UserPnLFidelityTwelveHours

// UserPnLInterval is the canonical SDK type.
type UserPnLInterval = models.UserPnLInterval

// UserPnLIntervalAll is the canonical SDK value.
const UserPnLIntervalAll = models.UserPnLIntervalAll

// UserPnLIntervalMax is the canonical SDK value.
const UserPnLIntervalMax = models.UserPnLIntervalMax

// UserPnLIntervalOneDay is the canonical SDK value.
const UserPnLIntervalOneDay = models.UserPnLIntervalOneDay

// UserPnLIntervalOneMonth is the canonical SDK value.
const UserPnLIntervalOneMonth = models.UserPnLIntervalOneMonth

// UserPnLIntervalOneWeek is the canonical SDK value.
const UserPnLIntervalOneWeek = models.UserPnLIntervalOneWeek

// UserPnLIntervalSixHours is the canonical SDK value.
const UserPnLIntervalSixHours = models.UserPnLIntervalSixHours

// UserPnLIntervalTwelveHours is the canonical SDK value.
const UserPnLIntervalTwelveHours = models.UserPnLIntervalTwelveHours

// UserPnLPoint is the canonical SDK type.
type UserPnLPoint = models.UserPnLPoint

// UserPnLSeries is the canonical SDK type.
type UserPnLSeries = models.UserPnLSeries

// UserStats is the canonical SDK type.
type UserStats = models.UserStats

// UserVolume is the canonical SDK type.
type UserVolume = models.UserVolume

// WithdrawalActivity is the canonical SDK type.
type WithdrawalActivity = models.WithdrawalActivity

// YieldActivity is the canonical SDK type.
type YieldActivity = models.YieldActivity

// Page holds a batch of results and continuation metadata.
type Page[T any] = pagination.Page[T]

// Paginator lazily retrieves results.
type Paginator[T any] = pagination.Paginator[T]

// Asset selects one CLOB asset.
func Asset(id ClobAssetID) AssetQuery { return clob.Asset(id) }
