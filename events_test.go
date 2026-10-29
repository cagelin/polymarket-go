package polymarket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/cagelin/polymarket-go/internal/codec"
)

func TestSearchIteration(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		query := r.URL.Query()
		if query.Get("q") != "test" || query.Get("limit_per_type") != "2" || query.Get("events_tag") != "tag" || query.Get("sort") != "volume" {
			t.Errorf("search query changed after construction: %v", query)
		}
		page, err := strconv.Atoi(query.Get("page"))
		if err != nil || page < 1 || page > 3 {
			t.Errorf("search page = %s, want 1..3", query.Get("page"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// An empty first page may still have a continuation. Category sizes do not
		// control iteration; the server's pagination metadata does.
		responses := []string{
			`{"events":[],"pagination":{"hasMore":true,"totalResults":2}}`,
			`{"events":[{"id":"2"}],"pagination":{"hasMore":true,"totalResults":2}}`,
			`{"events":[{"id":"3"}],"pagination":{"hasMore":false,"totalResults":2}}`,
		}
		if _, err = w.Write([]byte(responses[page-1])); err != nil {
			t.Errorf("write search fixture: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := NewPublicClient(ClientConfig{GammaURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Errorf("close client: %v", closeErr)
		}
	})
	size, sort, tags := 2, SearchByVolume, []string{"tag"}
	pager, err := client.Search(SearchParams{Query: "test", PageSize: &size, Sort: &sort, EventsTag: tags})
	if err != nil {
		t.Fatal(err)
	}
	size, sort, tags[0] = 99, SearchByLiquidity, "mutated"
	if calls.Load() != 0 {
		t.Fatal("Search eagerly requested a page")
	}
	pages := 0
	for page, pageErr := range pager.Pages(t.Context()) {
		if pageErr != nil {
			t.Fatal(pageErr)
		}
		pages++
		if len(page.Items) != 1 || page.TotalCount == nil || *page.TotalCount != 2 {
			t.Fatalf("search page %d = %#v, want one result group and total 2", pages, page)
		}
		if pages == 1 && len(page.Items[0].Events) != 0 {
			t.Error("first search page should be empty")
		}
		if pages > 1 && (len(page.Items[0].Events) != 1 || string(page.Items[0].Events[0].ID) != strconv.Itoa(pages)) {
			t.Errorf("search page %d = %#v, want matching event ID", pages, page.Items[0].Events)
		}
	}
	if pages != 3 || calls.Load() != 3 {
		t.Errorf("search iteration pages/calls = %d/%d, want 3/3", pages, calls.Load())
	}
	before := calls.Load()
	for _, pageErr := range pager.Items(t.Context()) {
		if pageErr != nil {
			t.Fatal(pageErr)
		}
		break
	}
	if calls.Load() != before+1 {
		t.Error("search break prefetched an extra page")
	}
	empty, err := pager.FromCursor(nil).FirstPage(t.Context())
	if err != nil || len(empty.Items) != 0 || calls.Load() != before+1 {
		t.Errorf("empty cursor made a request: page=%#v err=%v", empty, err)
	}
}

func BenchmarkDecodeEvent(b *testing.B) {
	data, err := os.ReadFile("testdata/events-parity.json")
	if err != nil {
		b.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name     string
			Response json.RawMessage
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		b.Fatal(err)
	}
	if len(fixture.Cases) == 0 || fixture.Cases[0].Name != "event/full" {
		b.Fatal("missing full event fixture")
	}
	response := fixture.Cases[0].Response
	b.ReportAllocs()
	b.SetBytes(int64(len(response)))
	b.ResetTimer()
	for b.Loop() {
		if _, err = codec.DecodeResponse[Event](response); err != nil {
			b.Fatal(err)
		}
	}
}
