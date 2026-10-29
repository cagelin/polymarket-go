package streams

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/cagelin/polymarket-go/internal/testutil"

	"github.com/cagelin/polymarket-go/models"
)

type specFixture struct {
	Kind  string
	Input json.RawMessage
}

func (t specFixture) decode(test *testing.T) Spec {
	test.Helper()
	var raw struct {
		AssetIDs   []string `json:"asset_ids"`
		Custom     bool     `json:"custom_feature_enabled"`
		Types      []string
		ParentID   *int64  `json:"parent_entity_id"`
		ParentType *string `json:"parent_entity_type"`
		Topic      string
		Symbols    []string
		Window     int `json:"window_seconds"`
		Symbol     string
	}
	if err := json.Unmarshal(t.Input, &raw); err != nil {
		test.Fatal(err)
	}
	switch t.Kind {
	case "market":
		return MarketSpec{AssetIDs: raw.AssetIDs, CustomFeatureEnabled: raw.Custom}
	case "sports":
		return SportsSpec{}
	case "comments":
		return CommentsSpec{Types: raw.Types, ParentEntityID: raw.ParentID, ParentEntityType: raw.ParentType}
	case "crypto":
		return CryptoPricesSpec{Topic: raw.Topic, Symbols: raw.Symbols}
	case "twap":
		return CryptoPricesChainlinkTwapSpec{WindowSeconds: raw.Window, Symbols: raw.Symbols}
	case "equity":
		return EquityPricesSpec{Symbol: raw.Symbol, Types: raw.Types}
	default:
		test.Fatalf("unknown spec %s", t.Kind)
		return nil
	}
}
func TestPythonStreamProtocol(t *testing.T) {
	raw, err := os.ReadFile("../testdata/streams-protocol-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Service string
			Spec          specFixture
			Expected      struct {
				Initial json.RawMessage
				Matches []string
				Error   bool
			}
		}
		Transitions []struct {
			Service       string
			Before, After []specFixture
			Expected      json.RawMessage
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	eventRaw, err := os.ReadFile("../testdata/streams-events-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var eventFixture struct {
		Cases []struct {
			Name, Service string
			Raw           json.RawMessage
			Expected      struct{ Error bool }
		}
	}
	if err = json.Unmarshal(eventRaw, &eventFixture); err != nil {
		t.Fatal(err)
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(subtest *testing.T) {
			spec, parseErr := test.Spec.decode(subtest).normalize()
			if (parseErr != nil) != test.Expected.Error {
				subtest.Fatalf("error=%v; Python=%v", parseErr, test.Expected.Error)
			}
			if parseErr != nil {
				return
			}
			frames := initialFrames(test.Service, deriveState([]selection{spec}))
			var initial any
			if len(frames) > 0 {
				initial = frames[0]
			}
			testutil.EqualJSON(subtest, initial, test.Expected.Initial)
			matches := make([]string, 0) // Match the fixture's empty JSON array.
			for _, event := range eventFixture.Cases {
				if event.Service != test.Service || event.Expected.Error {
					continue
				}
				parsed, decodeErr := models.ParseStreamEvent(event.Service, event.Raw)
				if decodeErr != nil {
					subtest.Fatal(decodeErr)
				}
				if spec.matches(parsed) {
					matches = append(matches, event.Name)
				}
			}
			if !reflect.DeepEqual(matches, test.Expected.Matches) {
				subtest.Errorf("matched %v; Python=%v", matches, test.Expected.Matches)
			}
		})
	}
	for _, test := range fixture.Transitions {
		var before, after []selection
		for _, value := range test.Before {
			spec, parseErr := value.decode(t).normalize()
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			before = append(before, spec)
		}
		for _, value := range test.After {
			spec, parseErr := value.decode(t).normalize()
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			after = append(after, spec)
		}
		testutil.EqualJSON(t, diffFrames(test.Service, deriveState(before), deriveState(after)), test.Expected)
	}
}
