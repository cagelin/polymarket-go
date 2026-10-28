// Data API enumerations from the pinned SDK. This file is linted.

package models

// ActivityType is a supported Data API value.
type ActivityType string

// Supported ActivityType values.
const (
	ActivityTypeTrade          ActivityType = "TRADE"
	ActivityTypeSplit          ActivityType = "SPLIT"
	ActivityTypeMerge          ActivityType = "MERGE"
	ActivityTypeRedeem         ActivityType = "REDEEM"
	ActivityTypeReward         ActivityType = "REWARD"
	ActivityTypeConversion     ActivityType = "CONVERSION"
	ActivityTypeMigration      ActivityType = "MIGRATION"
	ActivityTypeDeposit        ActivityType = "DEPOSIT"
	ActivityTypeWithdrawal     ActivityType = "WITHDRAWAL"
	ActivityTypeYield          ActivityType = "YIELD"
	ActivityTypeMakerRebate    ActivityType = "MAKER_REBATE"
	ActivityTypeTakerRebate    ActivityType = "TAKER_REBATE"
	ActivityTypeReferralReward ActivityType = "REFERRAL_REWARD"
	ActivityTypeTip            ActivityType = "TIP"
)

// BiggestWinnerKind is a supported Data API value.
type BiggestWinnerKind string

// Supported BiggestWinnerKind values.
const (
	BiggestWinnerKindMarket BiggestWinnerKind = "market"
	BiggestWinnerKindCombo  BiggestWinnerKind = "combo"
)

// ComboActivityType is a supported Data API value.
type ComboActivityType string

// Supported ComboActivityType values.
const (
	ComboActivityTypeSplit    ComboActivityType = "SPLIT"
	ComboActivityTypeMerge    ComboActivityType = "MERGE"
	ComboActivityTypeConvert  ComboActivityType = "CONVERT"
	ComboActivityTypeCompress ComboActivityType = "COMPRESS"
	ComboActivityTypeWrap     ComboActivityType = "WRAP"
	ComboActivityTypeUnwrap   ComboActivityType = "UNWRAP"
	ComboActivityTypeRedeem   ComboActivityType = "REDEEM"
)

// ComboPositionStatus is a supported Data API value.
type ComboPositionStatus string

// Supported ComboPositionStatus values.
const (
	ComboPositionStatusOpen            ComboPositionStatus = "OPEN"
	ComboPositionStatusRedeemable      ComboPositionStatus = "REDEEMABLE"
	ComboPositionStatusPartial         ComboPositionStatus = "PARTIAL"
	ComboPositionStatusResolvedPartial ComboPositionStatus = "RESOLVED_PARTIAL"
	ComboPositionStatusResolvedWin     ComboPositionStatus = "RESOLVED_WIN"
	ComboPositionStatusResolvedLoss    ComboPositionStatus = "RESOLVED_LOSS"
)

// PositionStatus is a supported Data API value.
type PositionStatus string

// Supported PositionStatus values.
const (
	PositionStatusOpen           PositionStatus = "OPEN"
	PositionStatusRedeemable     PositionStatus = "REDEEMABLE"
	PositionStatusRedeemableLost PositionStatus = "REDEEMABLE_LOST"
	PositionStatusMergeable      PositionStatus = "MERGEABLE"
	PositionStatusClosed         PositionStatus = "CLOSED"
)

// ResolutionMarketType is a supported Data API value.
type ResolutionMarketType string

// Supported ResolutionMarketType values.
const (
	ResolutionMarketTypeBinary             ResolutionMarketType = "BINARY"
	ResolutionMarketTypeIncrementalNegrisk ResolutionMarketType = "INCREMENTAL_NEGRISK"
	ResolutionMarketTypeAtomicNegrisk      ResolutionMarketType = "ATOMIC_NEGRISK"
)

// ResolutionReporter is a supported Data API value.
type ResolutionReporter string

// Supported ResolutionReporter values.
const (
	ResolutionReporterUMAOptimisticOracle ResolutionReporter = "UMA_OO"
	ResolutionReporterChainlink           ResolutionReporter = "CHAINLINK"
	ResolutionReporterEoa                 ResolutionReporter = "EOA"
)

// ResolutionSettlementTimeBasis is a supported Data API value.
type ResolutionSettlementTimeBasis string

// Supported ResolutionSettlementTimeBasis values.
const (
	ResolutionSettlementTimeBasisManagedProposalExpiration ResolutionSettlementTimeBasis = "managed_proposal_expiration"
	ResolutionSettlementTimeBasisProposalExpiration        ResolutionSettlementTimeBasis = "proposal_expiration"
	ResolutionSettlementTimeBasisLiveness                  ResolutionSettlementTimeBasis = "liveness"
	ResolutionSettlementTimeBasisDvmRoundEstimate          ResolutionSettlementTimeBasis = "dvm_round_estimate"
)

// ResolutionSource is a supported Data API value.
type ResolutionSource string

// Supported ResolutionSource values.
const (
	ResolutionSourceReported ResolutionSource = "reported"
	ResolutionSourceDerived  ResolutionSource = "derived"
)

// ResolutionStatus is a supported Data API value.
type ResolutionStatus string

// Supported ResolutionStatus values.
const (
	ResolutionStatusInitialized ResolutionStatus = "initialized"
	ResolutionStatusPosed       ResolutionStatus = "posed"
	ResolutionStatusProposed    ResolutionStatus = "proposed"
	ResolutionStatusChallenged  ResolutionStatus = "challenged"
	ResolutionStatusReproposed  ResolutionStatus = "reproposed"
	ResolutionStatusDisputed    ResolutionStatus = "disputed"
	ResolutionStatusActive      ResolutionStatus = "active"
	ResolutionStatusArbitration ResolutionStatus = "arbitration"
	ResolutionStatusResolved    ResolutionStatus = "resolved"
)

// TipSide is a supported Data API value.
type TipSide string

// Supported TipSide values.
const (
	TipSideIn  TipSide = "IN"
	TipSideOut TipSide = "OUT"
)

// TokenModule is a supported Data API value.
type TokenModule string

// Supported TokenModule values.
const (
	TokenModuleV1Ctf   TokenModule = "v1_ctf"
	TokenModuleBinary  TokenModule = "binary"
	TokenModuleNegRisk TokenModule = "neg_risk"
	TokenModuleCombo   TokenModule = "combo"
)

// UserPnLFidelity is a supported Data API value.
type UserPnLFidelity string

// Supported UserPnLFidelity values.
const (
	UserPnLFidelityOneDay        UserPnLFidelity = "1d"
	UserPnLFidelityEighteenHours UserPnLFidelity = "18h"
	UserPnLFidelityTwelveHours   UserPnLFidelity = "12h"
	UserPnLFidelityThreeHours    UserPnLFidelity = "3h"
	UserPnLFidelityOneHour       UserPnLFidelity = "1h"
)

// UserPnLInterval is a supported Data API value.
type UserPnLInterval string

// Supported UserPnLInterval values.
const (
	UserPnLIntervalMax         UserPnLInterval = "max"
	UserPnLIntervalAll         UserPnLInterval = "all"
	UserPnLIntervalOneMonth    UserPnLInterval = "1m"
	UserPnLIntervalOneWeek     UserPnLInterval = "1w"
	UserPnLIntervalOneDay      UserPnLInterval = "1d"
	UserPnLIntervalTwelveHours UserPnLInterval = "12h"
	UserPnLIntervalSixHours    UserPnLInterval = "6h"
)
