// Catalog models ported from the pinned Python SDK. This file is linted.

package models

import (
	"time"

	"github.com/ericlagergren/decimal"
)

// CategoryID identifies a catalog resource.
type CategoryID string

// ChatID identifies a catalog resource.
type ChatID string

// CollectionID identifies a catalog resource.
type CollectionID string

// CommentID identifies a catalog resource.
type CommentID string

// TemplateID identifies a catalog resource.
type TemplateID string

// CategoryReference represents a CategoryReference response.
type CategoryReference struct {
	ID             CategoryID `json:"id"`
	Label          *string    `json:"label"`
	ParentCategory *string    `json:"parent_category"`
	Slug           *string    `json:"slug"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedBy      *string    `json:"created_by"`
	UpdatedBy      *string    `json:"updated_by"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *CategoryReference) UnmarshalJSON(data []byte) error {
	type plain CategoryReference
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "label"},
		{name: "parent_category", aliases: []string{"parentCategory"}},
		{name: "slug"},
		{name: "published_at", aliases: []string{"publishedAt"}, kind: "datetime"},
		{name: "created_by", aliases: []string{"createdBy"}},
		{name: "updated_by", aliases: []string{"updatedBy"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = CategoryReference(result)
	return nil
}

// CollectionReference represents a CollectionReference response.
type CollectionReference struct {
	ID                CollectionID `json:"id"`
	Ticker            *string      `json:"ticker"`
	Slug              *string      `json:"slug"`
	Title             *string      `json:"title"`
	Subtitle          *string      `json:"subtitle"`
	CollectionType    *string      `json:"collection_type"`
	Description       *string      `json:"description"`
	Tags              *string      `json:"tags"`
	Image             *string      `json:"image"`
	Icon              *string      `json:"icon"`
	HeaderImage       *string      `json:"header_image"`
	Layout            *string      `json:"layout"`
	Active            *bool        `json:"active"`
	Closed            *bool        `json:"closed"`
	Archived          *bool        `json:"archived"`
	New               *bool        `json:"new"`
	Featured          *bool        `json:"featured"`
	Restricted        *bool        `json:"restricted"`
	IsTemplate        *bool        `json:"is_template"`
	TemplateVariables *string      `json:"template_variables"`
	PublishedAt       *time.Time   `json:"published_at"`
	CreatedBy         *string      `json:"created_by"`
	UpdatedBy         *string      `json:"updated_by"`
	CreatedAt         *time.Time   `json:"created_at"`
	UpdatedAt         *time.Time   `json:"updated_at"`
	DisqusThread      *string      `json:"disqus_thread"`
	CommentsEnabled   *bool        `json:"comments_enabled"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *CollectionReference) UnmarshalJSON(data []byte) error {
	type plain CollectionReference
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "ticker"},
		{name: "slug"},
		{name: "title"},
		{name: "subtitle"},
		{name: "collection_type", aliases: []string{"collectionType"}},
		{name: "description"},
		{name: "tags"},
		{name: "image"},
		{name: "icon"},
		{name: "header_image", aliases: []string{"headerImage"}},
		{name: "layout"},
		{name: "active", kind: "bool"},
		{name: "closed", kind: "bool"},
		{name: "archived", kind: "bool"},
		{name: "new", kind: "bool"},
		{name: "featured", kind: "bool"},
		{name: "restricted", kind: "bool"},
		{name: "is_template", aliases: []string{"isTemplate"}, kind: "bool"},
		{name: "template_variables", aliases: []string{"templateVariables"}},
		{name: "published_at", aliases: []string{"publishedAt"}, kind: "datetime"},
		{name: "created_by", aliases: []string{"createdBy"}},
		{name: "updated_by", aliases: []string{"updatedBy"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
		{name: "disqus_thread", aliases: []string{"disqusThread"}},
		{name: "comments_enabled", aliases: []string{"commentsEnabled"}, kind: "bool"},
	}); err != nil {
		return err
	}
	*m = CollectionReference(result)
	return nil
}

// SeriesReference represents a SeriesReference response.
type SeriesReference struct {
	ID                  SeriesID     `json:"id"`
	Ticker              *string      `json:"ticker"`
	Slug                *string      `json:"slug"`
	Title               *string      `json:"title"`
	Subtitle            *string      `json:"subtitle"`
	SeriesType          *string      `json:"series_type"`
	Recurrence          *string      `json:"recurrence"`
	Description         *string      `json:"description"`
	Image               *string      `json:"image"`
	Icon                *string      `json:"icon"`
	Layout              *string      `json:"layout"`
	Active              *bool        `json:"active"`
	Closed              *bool        `json:"closed"`
	Archived            *bool        `json:"archived"`
	New                 *bool        `json:"new"`
	Featured            *bool        `json:"featured"`
	Restricted          *bool        `json:"restricted"`
	IsTemplate          *bool        `json:"is_template"`
	TemplateVariables   *bool        `json:"template_variables"`
	PublishedAt         *time.Time   `json:"published_at"`
	CreatedBy           *string      `json:"created_by"`
	UpdatedBy           *string      `json:"updated_by"`
	CreatedAt           *time.Time   `json:"created_at"`
	UpdatedAt           *time.Time   `json:"updated_at"`
	CommentsEnabled     *bool        `json:"comments_enabled"`
	Competitive         *string      `json:"competitive"`
	Volume24hr          *decimal.Big `json:"volume_24hr"`
	Volume              *decimal.Big `json:"volume"`
	Liquidity           *decimal.Big `json:"liquidity"`
	StartDate           *time.Time   `json:"start_date"`
	PythTokenID         *string      `json:"pyth_token_id"`
	CgAssetName         *string      `json:"cg_asset_name"`
	Score               *int64       `json:"score"`
	CommentCount        *int64       `json:"comment_count"`
	RequiresTranslation *bool        `json:"requires_translation"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *SeriesReference) UnmarshalJSON(data []byte) error {
	type plain SeriesReference
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "id", required: true},
		{name: "ticker"},
		{name: "slug"},
		{name: "title"},
		{name: "subtitle"},
		{name: "series_type", aliases: []string{"seriesType"}},
		{name: "recurrence"},
		{name: "description"},
		{name: "image"},
		{name: "icon"},
		{name: "layout"},
		{name: "active", kind: "bool"},
		{name: "closed", kind: "bool"},
		{name: "archived", kind: "bool"},
		{name: "new", kind: "bool"},
		{name: "featured", kind: "bool"},
		{name: "restricted", kind: "bool"},
		{name: "is_template", aliases: []string{"isTemplate"}, kind: "bool"},
		{name: "template_variables", aliases: []string{"templateVariables"}, kind: "bool"},
		{name: "published_at", aliases: []string{"publishedAt"}, kind: "datetime"},
		{name: "created_by", aliases: []string{"createdBy"}},
		{name: "updated_by", aliases: []string{"updatedBy"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
		{name: "comments_enabled", aliases: []string{"commentsEnabled"}, kind: "bool"},
		{name: "competitive"},
		{name: "volume_24hr", aliases: []string{"volume24hr"}, kind: "decimal"},
		{name: "volume", kind: "decimal"},
		{name: "liquidity", kind: "decimal"},
		{name: "start_date", aliases: []string{"startDate"}, kind: "datetime"},
		{name: "pyth_token_id", aliases: []string{"pythTokenID"}},
		{name: "cg_asset_name", aliases: []string{"cgAssetName"}},
		{name: "score", kind: "int"},
		{name: "comment_count", aliases: []string{"commentCount"}, kind: "int"},
		{name: "requires_translation", aliases: []string{"requiresTranslation"}, kind: "bool"},
	}); err != nil {
		return err
	}
	*m = SeriesReference(result)
	return nil
}

// TemplateReference represents a TemplateReference response.
type TemplateReference struct {
	ID                      TemplateID `json:"id"`
	DisplayName             *string    `json:"display_name"`
	EventTitle              *string    `json:"event_title"`
	EventSlug               *string    `json:"event_slug"`
	Description             *string    `json:"description"`
	ResolutionSource        *string    `json:"resolution_source"`
	MarketsOrder            *string    `json:"markets_order"`
	MarketsNegRisk          *bool      `json:"markets_neg_risk"`
	MarketsAugmentedNegRisk *bool      `json:"markets_augmented_neg_risk"`
	MarketsShowImages       *bool      `json:"markets_show_images"`
	Markets                 *string    `json:"markets"`
	UserVariables           *string    `json:"user_variables"`
	CreatorUserID           *string    `json:"creator_user_id"`
	CreatedAt               *time.Time `json:"created_at"`
	UpdatedAt               *time.Time `json:"updated_at"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *TemplateReference) UnmarshalJSON(data []byte) error {
	type plain TemplateReference
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "display_name", aliases: []string{"displayName"}},
		{name: "event_title", aliases: []string{"eventTitle"}},
		{name: "event_slug", aliases: []string{"eventSlug"}},
		{name: "description"},
		{name: "resolution_source", aliases: []string{"resolutionSource"}},
		{name: "markets_order", aliases: []string{"marketsOrder"}},
		{name: "markets_neg_risk", aliases: []string{"marketsNegRisk"}, kind: "bool"},
		{name: "markets_augmented_neg_risk", aliases: []string{"marketsAugmentedNegRisk"}, kind: "bool"},
		{name: "markets_show_images", aliases: []string{"marketsShowImages"}, kind: "bool"},
		{name: "markets"},
		{name: "user_variables", aliases: []string{"userVariables"}},
		{name: "creator_user_id", aliases: []string{"creatorUserId"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = TemplateReference(result)
	return nil
}

// Chat represents a Chat response.
type Chat struct {
	ID           ChatID     `json:"id"`
	ChannelID    *string    `json:"channel_id"`
	ChannelName  *string    `json:"channel_name"`
	ChannelImage *string    `json:"channel_image"`
	Live         *bool      `json:"live"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Chat) UnmarshalJSON(data []byte) error {
	type plain Chat
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "channel_id", aliases: []string{"channelId"}},
		{name: "channel_name", aliases: []string{"channelName"}},
		{name: "channel_image", aliases: []string{"channelImage"}},
		{name: "live", kind: "bool"},
		{name: "start_time", aliases: []string{"startTime"}, kind: "datetime"},
		{name: "end_time", aliases: []string{"endTime"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = Chat(result)
	return nil
}

// SportsMarketTypes represents a SportsMarketTypes response.
type SportsMarketTypes struct {
	MarketTypes []string `json:"market_types"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *SportsMarketTypes) UnmarshalJSON(data []byte) error {
	type plain SportsMarketTypes
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "market_types", aliases: []string{"marketTypes"}},
	}); err != nil {
		return err
	}
	*m = SportsMarketTypes(result)
	return nil
}

// RelatedTag represents a RelatedTag response.
type RelatedTag struct {
	ID           string `json:"id"`
	TagID        *int64 `json:"tag_id"`
	RelatedTagID *int64 `json:"related_tag_id"`
	Rank         *int64 `json:"rank"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *RelatedTag) UnmarshalJSON(data []byte) error {
	type plain RelatedTag
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "tag_id", aliases: []string{"tagID"}, kind: "int"},
		{name: "related_tag_id", aliases: []string{"relatedTagID"}, kind: "int"},
		{name: "rank", kind: "int"},
	}); err != nil {
		return err
	}
	*m = RelatedTag(result)
	return nil
}

// Tag represents a Tag response.
type Tag struct {
	ID                  TagID               `json:"id"`
	Label               *string             `json:"label"`
	Slug                *string             `json:"slug"`
	ForceShow           *bool               `json:"force_show"`
	PublishedAt         *time.Time          `json:"published_at"`
	CreatedBy           *int64              `json:"created_by"`
	UpdatedBy           *int64              `json:"updated_by"`
	CreatedAt           *time.Time          `json:"created_at"`
	UpdatedAt           *time.Time          `json:"updated_at"`
	ForceHide           *bool               `json:"force_hide"`
	IsCarousel          *bool               `json:"is_carousel"`
	RequiresTranslation *bool               `json:"requires_translation"`
	ActiveEventsCount   *int64              `json:"active_events_count"`
	Templates           []TemplateReference `json:"templates"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Tag) UnmarshalJSON(data []byte) error {
	type plain Tag
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "label"},
		{name: "slug"},
		{name: "force_show", aliases: []string{"forceShow"}, kind: "bool"},
		{name: "published_at", aliases: []string{"publishedAt"}, kind: "datetime"},
		{name: "created_by", aliases: []string{"createdBy"}, kind: "int"},
		{name: "updated_by", aliases: []string{"updatedBy"}, kind: "int"},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
		{name: "force_hide", aliases: []string{"forceHide"}, kind: "bool"},
		{name: "is_carousel", aliases: []string{"isCarousel"}, kind: "bool"},
		{name: "requires_translation", aliases: []string{"requiresTranslation"}, kind: "bool"},
		{name: "active_events_count", aliases: []string{"activeEventsCount"}, kind: "int"},
		{name: "templates"},
	}); err != nil {
		return err
	}
	*m = Tag(result)
	return nil
}

// Series represents a Series response.
type Series struct {
	ID                  SeriesID              `json:"id"`
	Ticker              *string               `json:"ticker"`
	Slug                *string               `json:"slug"`
	Title               *string               `json:"title"`
	Subtitle            *string               `json:"subtitle"`
	SeriesType          *string               `json:"series_type"`
	Recurrence          *string               `json:"recurrence"`
	Description         *string               `json:"description"`
	Image               *string               `json:"image"`
	Icon                *string               `json:"icon"`
	Layout              *string               `json:"layout"`
	Active              *bool                 `json:"active"`
	Closed              *bool                 `json:"closed"`
	Archived            *bool                 `json:"archived"`
	New                 *bool                 `json:"new"`
	Featured            *bool                 `json:"featured"`
	Restricted          *bool                 `json:"restricted"`
	IsTemplate          *bool                 `json:"is_template"`
	TemplateVariables   *bool                 `json:"template_variables"`
	PublishedAt         *time.Time            `json:"published_at"`
	CreatedBy           *string               `json:"created_by"`
	UpdatedBy           *string               `json:"updated_by"`
	CreatedAt           *time.Time            `json:"created_at"`
	UpdatedAt           *time.Time            `json:"updated_at"`
	CommentsEnabled     *bool                 `json:"comments_enabled"`
	Competitive         *string               `json:"competitive"`
	Volume24hr          *decimal.Big          `json:"volume_24hr"`
	Volume              *decimal.Big          `json:"volume"`
	Liquidity           *decimal.Big          `json:"liquidity"`
	StartDate           *time.Time            `json:"start_date"`
	PythTokenID         *string               `json:"pyth_token_id"`
	CgAssetName         *string               `json:"cg_asset_name"`
	Score               *int64                `json:"score"`
	CommentCount        *int64                `json:"comment_count"`
	RequiresTranslation *bool                 `json:"requires_translation"`
	Events              []Event               `json:"events"`
	Collections         []CollectionReference `json:"collections"`
	Tags                []Tag                 `json:"tags"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Series) UnmarshalJSON(data []byte) error {
	type plain Series
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "id", required: true},
		{name: "ticker"},
		{name: "slug"},
		{name: "title"},
		{name: "subtitle"},
		{name: "series_type", aliases: []string{"seriesType"}},
		{name: "recurrence"},
		{name: "description"},
		{name: "image"},
		{name: "icon"},
		{name: "layout"},
		{name: "active", kind: "bool"},
		{name: "closed", kind: "bool"},
		{name: "archived", kind: "bool"},
		{name: "new", kind: "bool"},
		{name: "featured", kind: "bool"},
		{name: "restricted", kind: "bool"},
		{name: "is_template", aliases: []string{"isTemplate"}, kind: "bool"},
		{name: "template_variables", aliases: []string{"templateVariables"}, kind: "bool"},
		{name: "published_at", aliases: []string{"publishedAt"}, kind: "datetime"},
		{name: "created_by", aliases: []string{"createdBy"}},
		{name: "updated_by", aliases: []string{"updatedBy"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
		{name: "comments_enabled", aliases: []string{"commentsEnabled"}, kind: "bool"},
		{name: "competitive"},
		{name: "volume_24hr", aliases: []string{"volume24hr"}, kind: "decimal"},
		{name: "volume", kind: "decimal"},
		{name: "liquidity", kind: "decimal"},
		{name: "start_date", aliases: []string{"startDate"}, kind: "datetime"},
		{name: "pyth_token_id", aliases: []string{"pythTokenID"}},
		{name: "cg_asset_name", aliases: []string{"cgAssetName"}},
		{name: "score", kind: "int"},
		{name: "comment_count", aliases: []string{"commentCount"}, kind: "int"},
		{name: "requires_translation", aliases: []string{"requiresTranslation"}, kind: "bool"},
		{name: "events"},
		{name: "collections"},
		{name: "tags"},
	}); err != nil {
		return err
	}
	*m = Series(result)
	return nil
}

// Comment represents a Comment response.
type Comment struct {
	ID               CommentID       `json:"id"`
	Body             *string         `json:"body"`
	ParentEntityType *string         `json:"parent_entity_type"`
	ParentEntityID   *string         `json:"parent_entity_id"`
	ParentCommentID  *CommentID      `json:"parent_comment_id"`
	UserAddress      *EVMAddress     `json:"user_address"`
	ReplyAddress     *EVMAddress     `json:"reply_address"`
	CreatedAt        *time.Time      `json:"created_at"`
	UpdatedAt        *time.Time      `json:"updated_at"`
	Media            []CommentMedia  `json:"media"`
	Profile          *CommentProfile `json:"profile"`
	Reactions        []Reaction      `json:"reactions"`
	ReportCount      *int64          `json:"report_count"`
	ReactionCount    *int64          `json:"reaction_count"`
	TradeAsset       *string         `json:"trade_asset"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Comment) UnmarshalJSON(data []byte) error {
	var err error
	data, err = normalizeComment(data)
	if err != nil {
		return err
	}
	type plain Comment
	var result plain
	if err = decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "body"},
		{name: "parent_entity_type", aliases: []string{"parentEntityType"}},
		{name: "parent_entity_id", aliases: []string{"parentEntityID"}},
		{name: "parent_comment_id", aliases: []string{"parentCommentID"}},
		{name: "user_address", aliases: []string{"userAddress"}},
		{name: "reply_address", aliases: []string{"replyAddress"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
		{name: "media"},
		{name: "profile"},
		{name: "reactions"},
		{name: "report_count", aliases: []string{"reportCount"}, kind: "int"},
		{name: "reaction_count", aliases: []string{"reactionCount"}, kind: "int"},
		{name: "trade_asset", aliases: []string{"tradeAsset"}},
	}); err != nil {
		return err
	}
	*m = Comment(result)
	return nil
}

// CommentMedia represents a CommentMedia response.
type CommentMedia struct {
	ID              string     `json:"id"`
	CommentID       *int64     `json:"comment_id"`
	Provider        *string    `json:"provider"`
	ProviderMediaID *string    `json:"provider_media_id"`
	URL             *string    `json:"url"`
	MediaType       *string    `json:"media_type"`
	AltText         *string    `json:"alt_text"`
	CreatedAt       *time.Time `json:"created_at"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *CommentMedia) UnmarshalJSON(data []byte) error {
	type plain CommentMedia
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "comment_id", aliases: []string{"commentID"}, kind: "int"},
		{name: "provider"},
		{name: "provider_media_id", aliases: []string{"providerMediaId"}},
		{name: "url"},
		{name: "media_type", aliases: []string{"mediaType"}},
		{name: "alt_text", aliases: []string{"altText"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = CommentMedia(result)
	return nil
}

// CommentPosition represents a CommentPosition response.
type CommentPosition struct {
	TokenID      *TokenID `json:"token_id"`
	PositionSize *int64   `json:"position_size"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *CommentPosition) UnmarshalJSON(data []byte) error {
	type plain CommentPosition
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "token_id", aliases: []string{"tokenId"}},
		{name: "position_size", aliases: []string{"positionSize"}, kind: "int"},
	}); err != nil {
		return err
	}
	*m = CommentPosition(result)
	return nil
}

// CommentProfile represents a CommentProfile response.
type CommentProfile struct {
	Name                  *string            `json:"name"`
	Pseudonym             *string            `json:"pseudonym"`
	DisplayUsernamePublic *bool              `json:"display_username_public"`
	Bio                   *string            `json:"bio"`
	IsMod                 *bool              `json:"is_mod"`
	IsCreator             *bool              `json:"is_creator"`
	Wallet                *EVMAddress        `json:"wallet"`
	BaseAddress           *EVMAddress        `json:"base_address"`
	ProfileImage          *string            `json:"profile_image"`
	ProfileImageOptimized *ImageOptimization `json:"profile_image_optimized"`
	Positions             []CommentPosition  `json:"positions"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *CommentProfile) UnmarshalJSON(data []byte) error {
	var err error
	data, err = normalizeCommentProfile(data)
	if err != nil {
		return err
	}
	type plain CommentProfile
	var result plain
	if err = decodeGamma(data, &result, []gammaField{
		{name: "name"},
		{name: "pseudonym"},
		{name: "display_username_public", aliases: []string{"displayUsernamePublic"}, kind: "bool"},
		{name: "bio"},
		{name: "is_mod", aliases: []string{"isMod"}, kind: "bool"},
		{name: "is_creator", aliases: []string{"isCreator"}, kind: "bool"},
		{name: "wallet", aliases: []string{"wallet"}},
		{name: "base_address", aliases: []string{"baseAddress"}},
		{name: "profile_image", aliases: []string{"profileImage"}},
		{name: "profile_image_optimized", aliases: []string{"profileImageOptimized"}},
		{name: "positions"},
	}); err != nil {
		return err
	}
	*m = CommentProfile(result)
	return nil
}

// Reaction represents a Reaction response.
type Reaction struct {
	ID           string          `json:"id"`
	CommentID    *int64          `json:"comment_id"`
	ReactionType *string         `json:"reaction_type"`
	Icon         *string         `json:"icon"`
	UserAddress  *EVMAddress     `json:"user_address"`
	CreatedAt    *time.Time      `json:"created_at"`
	Profile      *CommentProfile `json:"profile"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Reaction) UnmarshalJSON(data []byte) error {
	type plain Reaction
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "comment_id", aliases: []string{"commentID"}, kind: "int"},
		{name: "reaction_type", aliases: []string{"reactionType"}},
		{name: "icon"},
		{name: "user_address", aliases: []string{"userAddress"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "profile"},
	}); err != nil {
		return err
	}
	*m = Reaction(result)
	return nil
}

// PublicProfile represents a PublicProfile response.
type PublicProfile struct {
	CreatedAt             *time.Time          `json:"created_at"`
	Wallet                *EVMAddress         `json:"wallet"`
	ProfileImage          *string             `json:"profile_image"`
	DisplayUsernamePublic *bool               `json:"display_username_public"`
	Bio                   *string             `json:"bio"`
	Pseudonym             *string             `json:"pseudonym"`
	Name                  *string             `json:"name"`
	Users                 []PublicProfileUser `json:"users"`
	XUsername             *string             `json:"x_username"`
	VerifiedBadge         *bool               `json:"verified_badge"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *PublicProfile) UnmarshalJSON(data []byte) error {
	var err error
	data, err = normalizePublicProfile(data)
	if err != nil {
		return err
	}
	type plain PublicProfile
	var result plain
	if err = decodeGamma(data, &result, []gammaField{
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "wallet", aliases: []string{"wallet"}},
		{name: "profile_image", aliases: []string{"profileImage"}},
		{name: "display_username_public", aliases: []string{"displayUsernamePublic"}, kind: "bool"},
		{name: "bio"},
		{name: "pseudonym"},
		{name: "name"},
		{name: "users"},
		{name: "x_username", aliases: []string{"xUsername"}},
		{name: "verified_badge", aliases: []string{"verifiedBadge"}, kind: "bool"},
	}); err != nil {
		return err
	}
	*m = PublicProfile(result)
	return nil
}

// PublicProfileUser represents a PublicProfileUser response.
type PublicProfileUser struct {
	ID           string `json:"id"`
	CommunityMod *bool  `json:"community_mod"`
	Creator      *bool  `json:"creator"`
	Mod          *bool  `json:"mod"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *PublicProfileUser) UnmarshalJSON(data []byte) error {
	type plain PublicProfileUser
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "community_mod", aliases: []string{"communityMod"}, kind: "bool"},
		{name: "creator", kind: "bool"},
		{name: "mod", kind: "bool"},
	}); err != nil {
		return err
	}
	*m = PublicProfileUser(result)
	return nil
}
