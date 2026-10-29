package signing

import (
	"testing"
)

func TestWalletAddressesPinned(t *testing.T) {
	signer, err := New("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	// These independently retained expectations must survive fixture regeneration.
	expected := map[string]string{
		"EOA":         "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf",
		"POLY_PROXY":  "0x909dEdA326ED2aE677F5F8c2527E32977c9838c3",
		"GNOSIS_SAFE": "0x51b7C68A71dCcBc0b7FA4400934a293D8f4d3Ba8",
		"UUPS":        "0x73310C2E333fB7fb177Dcb43e164958a9D443d6f",
		"BEACON":      "0x6F908B3B67b9F9C40413775fFf48b286aeF9a081",
	}
	actual := signer.Wallets()
	for kind, want := range expected {
		if actual[kind] != want {
			t.Errorf("Wallets()[%s] = %s, want %s", kind, actual[kind], want)
		}
	}
}
func TestInvalidEmbeddedHexFails(t *testing.T) {
	for _, text := range []string{"0", "xz", "0xabc"} {
		t.Run(text, func(subtest *testing.T) {
			defer func() {
				if recover() == nil {
					subtest.Errorf("fixedHex(%q) did not panic", text)
				}
			}()
			fixedHex(text)
		})
	}
}
