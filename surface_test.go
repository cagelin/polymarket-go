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

func TestSecureClientSurface(t *testing.T) {
	raw, err := os.ReadFile("testdata/secure-surface.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Members []struct {
			Python, Go string
			Stage      int
			Excluded   bool
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	client := reflect.TypeFor[*SecureClient]()
	for _, member := range fixture.Members {
		if member.Excluded {
			continue
		}
		if member.Python == "create" {
			continue
		}
		if _, ok := client.MethodByName(member.Go); !ok {
			t.Errorf("missing Python %s equivalent %s", member.Python, member.Go)
		}
	}
}
