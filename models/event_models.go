// Event and search models ported from polymarket-client 0.14.0.
// Regenerate with tests/parity/generate_event_types.py. These files are linted.

package models

import (
	"encoding/json"
	"time"

	"github.com/ericlagergren/decimal"
)

// BestLineID identifies a BestLine resource.
type BestLineID string

// EventCreatorID identifies a EventCreator resource.
type EventCreatorID string

// EventExternalPartnerMappingID identifies a EventExternalPartnerMapping resource.
type EventExternalPartnerMappingID int64

// ImageOptimizationID identifies a ImageOptimization resource.
type ImageOptimizationID string

// PartnerID identifies a Partner resource.
type PartnerID int64

// SeriesID identifies a Series resource.
type SeriesID string

// SportID identifies a Sport resource.
type SportID int64

// TeamID identifies a Team resource.
type TeamID int64

// TeamOrdering identifies a team's side in a sports event.
type TeamOrdering string

// Supported team positions.
const (
	TeamHome TeamOrdering = "home"
	TeamAway TeamOrdering = "away"
)

// ImageOptimization represents a ImageOptimization response.
type ImageOptimization struct {
	ID                        ImageOptimizationID `json:"id"`
	ImageURLSource            *string             `json:"image_url_source"`
	ImageURLOptimized         *string             `json:"image_url_optimized"`
	ImageSizeKBSource         *float64            `json:"image_size_kb_source"`
	ImageSizeKBOptimized      *float64            `json:"image_size_kb_optimized"`
	ImageOptimizedComplete    *bool               `json:"image_optimized_complete"`
	ImageOptimizedLastUpdated *time.Time          `json:"image_optimized_last_updated"`
	RelID                     *int64              `json:"rel_id"`
	Field                     *string             `json:"field"`
	Relname                   *string             `json:"relname"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *ImageOptimization) UnmarshalJSON(data []byte) error {
	type plain ImageOptimization
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "image_url_source", aliases: []string{"imageUrlSource"}},
		{name: "image_url_optimized", aliases: []string{"imageUrlOptimized"}},
		{name: "image_size_kb_source", aliases: []string{"imageSizeKbSource"}, kind: "float"},
		{name: "image_size_kb_optimized", aliases: []string{"imageSizeKbOptimized"}, kind: "float"},
		{name: "image_optimized_complete", aliases: []string{"imageOptimizedComplete"}, kind: "bool"},
		{name: "image_optimized_last_updated", aliases: []string{"imageOptimizedLastUpdated"}, kind: "epoch_datetime"},
		{name: "rel_id", aliases: []string{"relID"}, kind: "int"},
		{name: "field"},
		{name: "relname"},
	}); err != nil {
		return err
	}
	*m = ImageOptimization(result)
	return nil
}

// Partner represents a Partner response.
type Partner struct {
	ID   PartnerID `json:"id"`
	Slug string    `json:"slug"`
	Name string    `json:"name"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Partner) UnmarshalJSON(data []byte) error {
	type plain Partner
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "int", required: true},
		{name: "slug", required: true},
		{name: "name", required: true},
	}); err != nil {
		return err
	}
	*m = Partner(result)
	return nil
}

// BestLine represents a BestLine response.
type BestLine struct {
	ID       BestLineID `json:"id"`
	LineType *string    `json:"line_type"`
	Line     *float64   `json:"line"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *BestLine) UnmarshalJSON(data []byte) error {
	type plain BestLine
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "line_type", aliases: []string{"lineType"}},
		{name: "line", kind: "float"},
	}); err != nil {
		return err
	}
	*m = BestLine(result)
	return nil
}

// Team represents a Team response.
type Team struct {
	ID           TeamID        `json:"id"`
	Name         *string       `json:"name"`
	League       *string       `json:"league"`
	Record       *string       `json:"record"`
	Logo         *string       `json:"logo"`
	Abbreviation *string       `json:"abbreviation"`
	Alias        *string       `json:"alias"`
	CreatedAt    *time.Time    `json:"created_at"`
	UpdatedAt    *time.Time    `json:"updated_at"`
	ProviderID   *int64        `json:"provider_id"`
	Color        *string       `json:"color"`
	Ordering     *TeamOrdering `json:"ordering"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Team) UnmarshalJSON(data []byte) error {
	type plain Team
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "int", required: true},
		{name: "name"},
		{name: "league"},
		{name: "record"},
		{name: "logo"},
		{name: "abbreviation"},
		{name: "alias"},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
		{name: "provider_id", aliases: []string{"providerId"}, kind: "int"},
		{name: "color"},
		{name: "ordering", kind: "team_ordering"},
	}); err != nil {
		return err
	}
	*m = Team(result)
	return nil
}

// SportsMetadata represents a SportsMetadata response.
type SportsMetadata struct {
	ID         SportID    `json:"id"`
	Sport      string     `json:"sport"`
	Name       *string    `json:"name"`
	Image      string     `json:"image"`
	Resolution string     `json:"resolution"`
	Ordering   string     `json:"ordering"`
	Tags       string     `json:"tags"`
	Series     string     `json:"series"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *SportsMetadata) UnmarshalJSON(data []byte) error {
	type plain SportsMetadata
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "int", required: true},
		{name: "sport", required: true},
		{name: "name"},
		{name: "image", required: true},
		{name: "resolution", required: true},
		{name: "ordering", required: true},
		{name: "tags", required: true},
		{name: "series", required: true},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = SportsMetadata(result)
	return nil
}

// TagReference represents a TagReference response.
type TagReference struct {
	ID                  TagID      `json:"id"`
	Label               *string    `json:"label"`
	Slug                *string    `json:"slug"`
	ForceShow           *bool      `json:"force_show"`
	PublishedAt         *time.Time `json:"published_at"`
	CreatedBy           *int64     `json:"created_by"`
	UpdatedBy           *int64     `json:"updated_by"`
	CreatedAt           *time.Time `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at"`
	ForceHide           *bool      `json:"force_hide"`
	IsCarousel          *bool      `json:"is_carousel"`
	RequiresTranslation *bool      `json:"requires_translation"`
	ActiveEventsCount   *int64     `json:"active_events_count"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *TagReference) UnmarshalJSON(data []byte) error {
	type plain TagReference
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
	}); err != nil {
		return err
	}
	*m = TagReference(result)
	return nil
}

// Event represents a Polymarket event.
type Event struct {
	ID            EventID                    `json:"id"`
	Version       *ProtocolVersion           `json:"version"`
	ParentEventID *EventID                   `json:"parent_event_id"`
	Ticker        *string                    `json:"ticker"`
	Slug          *string                    `json:"slug"`
	Title         *string                    `json:"title"`
	Subtitle      *string                    `json:"subtitle"`
	Description   *string                    `json:"description"`
	Category      *string                    `json:"category"`
	Subcategory   *string                    `json:"subcategory"`
	Image         *string                    `json:"image"`
	Icon          *string                    `json:"icon"`
	FeaturedImage *string                    `json:"featured_image"`
	CreatedAt     *time.Time                 `json:"created_at"`
	UpdatedAt     *time.Time                 `json:"updated_at"`
	PublishedAt   *time.Time                 `json:"published_at"`
	State         EventState                 `json:"state"`
	Schedule      EventSchedule              `json:"schedule"`
	Metrics       EventMetrics               `json:"metrics"`
	Display       EventDisplay               `json:"display"`
	Trading       EventTrading               `json:"trading"`
	Resolution    EventResolution            `json:"resolution"`
	Estimation    EventEstimation            `json:"estimation"`
	Sports        EventSportsMetadata        `json:"sports"`
	Partners      []EventPartner             `json:"partners"`
	Metadata      map[string]json.RawMessage `json:"metadata"`
	Markets       []Market                   `json:"markets"`
	Series        []EventSeries              `json:"series"`
	Tags          []EventTag                 `json:"tags"`
	Creators      []EventCreator             `json:"creators"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Event) UnmarshalJSON(data []byte) error {
	var err error
	data, err = normalizeEvent(data)
	if err != nil {
		return err
	}
	type plain Event
	var result plain
	if err = decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "id", required: true},
		{name: "version", kind: "version"},
		{name: "parent_event_id", aliases: []string{"parentEventId"}, kind: "id"},
		{name: "ticker"},
		{name: "slug"},
		{name: "title"},
		{name: "subtitle"},
		{name: "description"},
		{name: "category"},
		{name: "subcategory"},
		{name: "image"},
		{name: "icon"},
		{name: "featured_image", aliases: []string{"featuredImage"}},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
		{name: "published_at", aliases: []string{"publishedAt"}, kind: "datetime"},
		{name: "state", required: true},
		{name: "schedule", required: true},
		{name: "metrics", required: true},
		{name: "display", required: true},
		{name: "trading", required: true},
		{name: "resolution", required: true},
		{name: "estimation", required: true},
		{name: "sports", required: true},
		{name: "partners", kind: "sequence", required: true},
		{name: "metadata"},
		{name: "markets", kind: "sequence", required: true},
		{name: "series", kind: "sequence", required: true},
		{name: "tags", kind: "sequence", required: true},
		{name: "creators", kind: "sequence", required: true},
	}); err != nil {
		return err
	}
	*m = Event(result)
	return nil
}

// EventCreator represents a EventCreator response.
type EventCreator struct {
	ID        EventCreatorID `json:"id"`
	Name      *string        `json:"name"`
	Handle    *string        `json:"handle"`
	URL       *string        `json:"url"`
	Image     *string        `json:"image"`
	CreatedAt *time.Time     `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventCreator) UnmarshalJSON(data []byte) error {
	type plain EventCreator
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "name"},
		{name: "handle"},
		{name: "url"},
		{name: "image"},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = EventCreator(result)
	return nil
}

// EventDisplay represents a EventDisplay response.
type EventDisplay struct {
	SortBy                 *string            `json:"sort_by"`
	ShowAllOutcomes        *bool              `json:"show_all_outcomes"`
	ShowMarketImages       *bool              `json:"show_market_images"`
	GMPChartMode           *string            `json:"gmp_chart_mode"`
	Color                  *string            `json:"color"`
	FeaturedOrder          *int64             `json:"featured_order"`
	CountryName            *string            `json:"country_name"`
	ElectionType           *string            `json:"election_type"`
	ImageOptimized         *ImageOptimization `json:"image_optimized"`
	IconOptimized          *ImageOptimization `json:"icon_optimized"`
	FeaturedImageOptimized *ImageOptimization `json:"featured_image_optimized"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventDisplay) UnmarshalJSON(data []byte) error {
	type plain EventDisplay
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "sort_by", aliases: []string{"sortBy"}},
		{name: "show_all_outcomes", aliases: []string{"showAllOutcomes"}, kind: "bool"},
		{name: "show_market_images", aliases: []string{"showMarketImages"}, kind: "bool"},
		{name: "gmp_chart_mode", aliases: []string{"gmpChartMode"}},
		{name: "color"},
		{name: "featured_order", aliases: []string{"featuredOrder"}, kind: "int"},
		{name: "country_name", aliases: []string{"countryName"}},
		{name: "election_type", aliases: []string{"electionType"}},
		{name: "image_optimized", aliases: []string{"imageOptimized"}},
		{name: "icon_optimized", aliases: []string{"iconOptimized"}},
		{name: "featured_image_optimized", aliases: []string{"featuredImageOptimized"}},
	}); err != nil {
		return err
	}
	*m = EventDisplay(result)
	return nil
}

// EventEstimation represents a EventEstimation response.
type EventEstimation struct {
	EstimateValue  *bool        `json:"estimate_value"`
	CantEstimate   *bool        `json:"cant_estimate"`
	EstimatedValue *decimal.Big `json:"estimated_value"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventEstimation) UnmarshalJSON(data []byte) error {
	type plain EventEstimation
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "estimate_value", aliases: []string{"estimateValue"}, kind: "bool"},
		{name: "cant_estimate", aliases: []string{"cantEstimate"}, kind: "bool"},
		{name: "estimated_value", aliases: []string{"estimatedValue"}, kind: "decimal"},
	}); err != nil {
		return err
	}
	*m = EventEstimation(result)
	return nil
}

// EventMetrics represents a EventMetrics response.
type EventMetrics struct {
	Liquidity     *decimal.Big `json:"liquidity"`
	LiquidityAMM  *decimal.Big `json:"liquidity_amm"`
	LiquidityClob *decimal.Big `json:"liquidity_clob"`
	Volume        *decimal.Big `json:"volume"`
	Volume24hr    *decimal.Big `json:"volume_24hr"`
	Volume1wk     *decimal.Big `json:"volume_1wk"`
	Volume1mo     *decimal.Big `json:"volume_1mo"`
	Volume1yr     *decimal.Big `json:"volume_1yr"`
	OpenInterest  *decimal.Big `json:"open_interest"`
	Competitive   *float64     `json:"competitive"`
	CommentCount  *int64       `json:"comment_count"`
	TweetCount    *int64       `json:"tweet_count"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventMetrics) UnmarshalJSON(data []byte) error {
	type plain EventMetrics
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "liquidity", kind: "decimal"},
		{name: "liquidity_amm", aliases: []string{"liquidityAmm"}, kind: "decimal"},
		{name: "liquidity_clob", aliases: []string{"liquidityClob"}, kind: "decimal"},
		{name: "volume", kind: "decimal"},
		{name: "volume_24hr", aliases: []string{"volume24hr"}, kind: "decimal"},
		{name: "volume_1wk", aliases: []string{"volume1wk"}, kind: "decimal"},
		{name: "volume_1mo", aliases: []string{"volume1mo"}, kind: "decimal"},
		{name: "volume_1yr", aliases: []string{"volume1yr"}, kind: "decimal"},
		{name: "open_interest", aliases: []string{"openInterest"}, kind: "decimal"},
		{name: "competitive", kind: "float"},
		{name: "comment_count", aliases: []string{"commentCount"}, kind: "int"},
		{name: "tweet_count", aliases: []string{"tweetCount"}, kind: "int"},
	}); err != nil {
		return err
	}
	*m = EventMetrics(result)
	return nil
}

// EventPartner represents a EventPartner response.
type EventPartner struct {
	ID         EventExternalPartnerMappingID `json:"id"`
	ExternalID string                        `json:"external_id"`
	Partner    *Partner                      `json:"partner"`
	CreatedAt  *time.Time                    `json:"created_at"`
	UpdatedAt  *time.Time                    `json:"updated_at"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventPartner) UnmarshalJSON(data []byte) error {
	type plain EventPartner
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "int", required: true},
		{name: "external_id", aliases: []string{"externalId"}, required: true},
		{name: "partner"},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = EventPartner(result)
	return nil
}

// EventResolution represents a EventResolution response.
type EventResolution struct {
	Source                *string `json:"source"`
	AutomaticallyResolved *bool   `json:"automatically_resolved"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventResolution) UnmarshalJSON(data []byte) error {
	type plain EventResolution
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "source"},
		{name: "automatically_resolved", aliases: []string{"automaticallyResolved"}, kind: "bool"},
	}); err != nil {
		return err
	}
	*m = EventResolution(result)
	return nil
}

// EventSchedule represents a EventSchedule response.
type EventSchedule struct {
	StartDate    *time.Time `json:"start_date"`
	CreationDate *time.Time `json:"creation_date"`
	EndDate      *time.Time `json:"end_date"`
	ClosedTime   *time.Time `json:"closed_time"`
	StartTime    *time.Time `json:"start_time"`
	EventDate    *Date      `json:"event_date"`
	EventWeek    *int64     `json:"event_week"`
	FinishedAt   *time.Time `json:"finished_at"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventSchedule) UnmarshalJSON(data []byte) error {
	type plain EventSchedule
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "start_date", aliases: []string{"startDate"}, kind: "datetime"},
		{name: "creation_date", aliases: []string{"creationDate"}, kind: "datetime"},
		{name: "end_date", aliases: []string{"endDate"}, kind: "datetime"},
		{name: "closed_time", aliases: []string{"closedTime"}, kind: "datetime"},
		{name: "start_time", aliases: []string{"startTime"}, kind: "datetime"},
		{name: "event_date", aliases: []string{"eventDate"}, kind: "optional_date"},
		{name: "event_week", aliases: []string{"eventWeek"}, kind: "int"},
		{name: "finished_at", aliases: []string{"finishedAt"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = EventSchedule(result)
	return nil
}

// EventSeries represents a EventSeries response.
type EventSeries struct {
	ID          SeriesID     `json:"id"`
	Slug        *string      `json:"slug"`
	Title       *string      `json:"title"`
	Subtitle    *string      `json:"subtitle"`
	Description *string      `json:"description"`
	Image       *string      `json:"image"`
	Icon        *string      `json:"icon"`
	Active      *bool        `json:"active"`
	Closed      *bool        `json:"closed"`
	Archived    *bool        `json:"archived"`
	Volume      *decimal.Big `json:"volume"`
	Liquidity   *decimal.Big `json:"liquidity"`
	StartDate   *time.Time   `json:"start_date"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventSeries) UnmarshalJSON(data []byte) error {
	type plain EventSeries
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", kind: "id", required: true},
		{name: "slug"},
		{name: "title"},
		{name: "subtitle"},
		{name: "description"},
		{name: "image"},
		{name: "icon"},
		{name: "active", kind: "bool"},
		{name: "closed", kind: "bool"},
		{name: "archived", kind: "bool"},
		{name: "volume", kind: "decimal"},
		{name: "liquidity", kind: "decimal"},
		{name: "start_date", aliases: []string{"startDate"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = EventSeries(result)
	return nil
}

// EventSportsMetadata represents a EventSportsMetadata response.
type EventSportsMetadata struct {
	SeriesSlug            *string         `json:"series_slug"`
	Score                 *string         `json:"score"`
	Elapsed               *string         `json:"elapsed"`
	Period                *string         `json:"period"`
	GameStatus            *string         `json:"game_status"`
	GameID                *int64          `json:"game_id"`
	RescheduledFromGameID *int64          `json:"rescheduled_from_game_id"`
	SportsradarMatchID    *string         `json:"sportsradar_match_id"`
	HomeTeamName          *string         `json:"home_team_name"`
	AwayTeamName          *string         `json:"away_team_name"`
	SpreadsMainLine       *float64        `json:"spreads_main_line"`
	TotalsMainLine        *float64        `json:"totals_main_line"`
	BestLines             []BestLine      `json:"best_lines"`
	Teams                 []Team          `json:"teams"`
	Sport                 *SportsMetadata `json:"sport"`
	LastHighlight         *string         `json:"last_highlight"`
	LastHighlightType     *string         `json:"last_highlight_type"`
	LastHighlightAt       *time.Time      `json:"last_highlight_at"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventSportsMetadata) UnmarshalJSON(data []byte) error {
	type plain EventSportsMetadata
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "series_slug", aliases: []string{"seriesSlug"}},
		{name: "score"},
		{name: "elapsed"},
		{name: "period"},
		{name: "game_status", aliases: []string{"gameStatus"}},
		{name: "game_id", aliases: []string{"gameId"}, kind: "int"},
		{name: "rescheduled_from_game_id", aliases: []string{"rescheduledFromGameId"}, kind: "int"},
		{name: "sportsradar_match_id", aliases: []string{"sportsradarMatchId"}},
		{name: "home_team_name", aliases: []string{"homeTeamName"}},
		{name: "away_team_name", aliases: []string{"awayTeamName"}},
		{name: "spreads_main_line", aliases: []string{"spreadsMainLine"}, kind: "float"},
		{name: "totals_main_line", aliases: []string{"totalsMainLine"}, kind: "float"},
		{name: "best_lines", kind: "default_sequence"},
		{name: "teams", kind: "default_sequence"},
		{name: "sport"},
		{name: "last_highlight", aliases: []string{"lastHighlight"}},
		{name: "last_highlight_type", aliases: []string{"lastHighlightType"}},
		{name: "last_highlight_at", aliases: []string{"lastHighlightAt"}, kind: "datetime"},
	}); err != nil {
		return err
	}
	*m = EventSportsMetadata(result)
	return nil
}

// EventState represents a EventState response.
type EventState struct {
	Active              *bool `json:"active"`
	Closed              *bool `json:"closed"`
	Archived            *bool `json:"archived"`
	New                 *bool `json:"new"`
	Featured            *bool `json:"featured"`
	Restricted          *bool `json:"restricted"`
	CYOM                *bool `json:"cyom"`
	Live                *bool `json:"live"`
	Ended               *bool `json:"ended"`
	AutomaticallyActive *bool `json:"automatically_active"`
	CommentsEnabled     *bool `json:"comments_enabled"`
	RequiresTranslation *bool `json:"requires_translation"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventState) UnmarshalJSON(data []byte) error {
	type plain EventState
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "active", kind: "bool"},
		{name: "closed", kind: "bool"},
		{name: "archived", kind: "bool"},
		{name: "new", kind: "bool"},
		{name: "featured", kind: "bool"},
		{name: "restricted", kind: "bool"},
		{name: "cyom", kind: "bool"},
		{name: "live", kind: "bool"},
		{name: "ended", kind: "bool"},
		{name: "automatically_active", aliases: []string{"automaticallyActive"}, kind: "bool"},
		{name: "comments_enabled", aliases: []string{"commentsEnabled"}, kind: "bool"},
		{name: "requires_translation", aliases: []string{"requiresTranslation"}, kind: "bool"},
	}); err != nil {
		return err
	}
	*m = EventState(result)
	return nil
}

// EventTag represents a EventTag response.
type EventTag struct {
	ID    TagID   `json:"id"`
	Slug  *string `json:"slug"`
	Label *string `json:"label"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventTag) UnmarshalJSON(data []byte) error {
	type plain EventTag
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "slug"},
		{name: "label"},
	}); err != nil {
		return err
	}
	*m = EventTag(result)
	return nil
}

// EventTrading represents a EventTrading response.
type EventTrading struct {
	EnableOrderBook   *bool    `json:"enable_order_book"`
	NegRisk           *bool    `json:"neg_risk"`
	NegRiskMarketID   *string  `json:"neg_risk_market_id"`
	NegRiskFeeBips    *float64 `json:"neg_risk_fee_bips"`
	EnableNegRisk     *bool    `json:"enable_neg_risk"`
	NegRiskAugmented  *bool    `json:"neg_risk_augmented"`
	CumulativeMarkets *bool    `json:"cumulative_markets"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *EventTrading) UnmarshalJSON(data []byte) error {
	type plain EventTrading
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "enable_order_book", aliases: []string{"enableOrderBook"}, kind: "bool"},
		{name: "neg_risk", aliases: []string{"negRisk"}, kind: "bool"},
		{name: "neg_risk_market_id", aliases: []string{"negRiskMarketId"}},
		{name: "neg_risk_fee_bips", aliases: []string{"negRiskFeeBips"}, kind: "float"},
		{name: "enable_neg_risk", aliases: []string{"enableNegRisk"}, kind: "bool"},
		{name: "neg_risk_augmented", aliases: []string{"negRiskAugmented"}, kind: "bool"},
		{name: "cumulative_markets", aliases: []string{"cumulativeMarkets"}, kind: "bool"},
	}); err != nil {
		return err
	}
	*m = EventTrading(result)
	return nil
}

// Profile represents a Profile response.
type Profile struct {
	ID                     *string            `json:"id"`
	Name                   *string            `json:"name"`
	User                   *int64             `json:"user"`
	Referral               *string            `json:"referral"`
	CreatedBy              *int64             `json:"created_by"`
	UpdatedBy              *int64             `json:"updated_by"`
	CreatedAt              *time.Time         `json:"created_at"`
	UpdatedAt              *time.Time         `json:"updated_at"`
	UTMSource              *string            `json:"utm_source"`
	UTMMedium              *string            `json:"utm_medium"`
	UTMCampaign            *string            `json:"utm_campaign"`
	UTMContent             *string            `json:"utm_content"`
	UTMTerm                *string            `json:"utm_term"`
	WalletActivated        *bool              `json:"wallet_activated"`
	Pseudonym              *string            `json:"pseudonym"`
	DisplayUsernamePublic  *bool              `json:"display_username_public"`
	ProfileImage           *string            `json:"profile_image"`
	Bio                    *string            `json:"bio"`
	Wallet                 *EVMAddress        `json:"wallet"`
	ProfileImageOptimized  *ImageOptimization `json:"profile_image_optimized"`
	IsCloseOnly            *bool              `json:"is_close_only"`
	IsCertReq              *bool              `json:"is_cert_req"`
	CertReqDate            *time.Time         `json:"cert_req_date"`
	DiscordUsername        *string            `json:"discord_username"`
	XUsername              *string            `json:"x_username"`
	VerifiedBadge          *bool              `json:"verified_badge"`
	DubPartnerID           *string            `json:"dub_partner_id"`
	TermsAcceptedAt        *time.Time         `json:"terms_accepted_at"`
	ViewOnlyAcknowledgedAt *time.Time         `json:"view_only_acknowledged_at"`
	IsReferralRestricted   *bool              `json:"is_referral_restricted"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *Profile) UnmarshalJSON(data []byte) error {
	var err error
	data, err = normalizeProfile(data)
	if err != nil {
		return err
	}
	type plain Profile
	var result plain
	if err = decodeGamma(data, &result, []gammaField{
		{name: "id"},
		{name: "name"},
		{name: "user", kind: "int"},
		{name: "referral"},
		{name: "created_by", aliases: []string{"createdBy"}, kind: "int"},
		{name: "updated_by", aliases: []string{"updatedBy"}, kind: "int"},
		{name: "created_at", aliases: []string{"createdAt"}, kind: "datetime"},
		{name: "updated_at", aliases: []string{"updatedAt"}, kind: "datetime"},
		{name: "utm_source", aliases: []string{"utmSource"}},
		{name: "utm_medium", aliases: []string{"utmMedium"}},
		{name: "utm_campaign", aliases: []string{"utmCampaign"}},
		{name: "utm_content", aliases: []string{"utmContent"}},
		{name: "utm_term", aliases: []string{"utmTerm"}},
		{name: "wallet_activated", aliases: []string{"walletActivated"}, kind: "bool"},
		{name: "pseudonym"},
		{name: "display_username_public", aliases: []string{"displayUsernamePublic"}, kind: "bool"},
		{name: "profile_image", aliases: []string{"profileImage"}},
		{name: "bio"},
		{name: "wallet"},
		{name: "profile_image_optimized", aliases: []string{"profileImageOptimized"}},
		{name: "is_close_only", aliases: []string{"isCloseOnly"}, kind: "bool"},
		{name: "is_cert_req", aliases: []string{"isCertReq"}, kind: "bool"},
		{name: "cert_req_date", aliases: []string{"certReqDate"}, kind: "datetime"},
		{name: "discord_username", aliases: []string{"discordUsername"}},
		{name: "x_username", aliases: []string{"xUsername"}},
		{name: "verified_badge", aliases: []string{"verifiedBadge"}, kind: "bool"},
		{name: "dub_partner_id", aliases: []string{"dubPartnerId"}},
		{name: "terms_accepted_at", aliases: []string{"termsAcceptedAt"}, kind: "datetime"},
		{name: "view_only_acknowledged_at", aliases: []string{"viewOnlyAcknowledgedAt"}, kind: "datetime"},
		{name: "is_referral_restricted", aliases: []string{"isReferralRestricted"}, kind: "bool"},
	}); err != nil {
		return err
	}
	*m = Profile(result)
	return nil
}

// SearchTag represents a SearchTag response.
type SearchTag struct {
	ID         TagID   `json:"id"`
	Label      *string `json:"label"`
	Slug       *string `json:"slug"`
	EventCount *int64  `json:"event_count"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *SearchTag) UnmarshalJSON(data []byte) error {
	type plain SearchTag
	var result plain
	if err := decodeGamma(data, &result, []gammaField{
		{name: "id", required: true},
		{name: "label"},
		{name: "slug"},
		{name: "event_count", aliases: []string{"event_count"}, kind: "int"},
	}); err != nil {
		return err
	}
	*m = SearchTag(result)
	return nil
}

// SearchResults represents a SearchResults response.
type SearchResults struct {
	Events   []Event     `json:"events"`
	Tags     []SearchTag `json:"tags"`
	Profiles []Profile   `json:"profiles"`
}

// UnmarshalJSON validates response fields and normalizes wire names.
func (m *SearchResults) UnmarshalJSON(data []byte) error {
	var err error
	data, err = normalizeSearchResults(data)
	if err != nil {
		return err
	}
	type plain SearchResults
	var result plain
	if err = decodeGamma(data, &result, []gammaField{
		{name: "events", kind: "default_sequence"},
		{name: "tags", kind: "default_sequence"},
		{name: "profiles", kind: "default_sequence"},
	}); err != nil {
		return err
	}
	*m = SearchResults(result)
	return nil
}
