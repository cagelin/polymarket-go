package polymarket

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPublicClientSurface(t *testing.T) {
	raw, err := os.ReadFile("testdata/public-surface.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Members []struct{ Python, Go string } }
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	client := reflect.TypeFor[*PublicClient]()
	for _, member := range fixture.Members {
		if _, ok := client.MethodByName(member.Go); !ok {
			t.Errorf("missing Python %s equivalent %s", member.Python, member.Go)
		}
	}
}
