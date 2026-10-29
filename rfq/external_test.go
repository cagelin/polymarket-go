package rfq_test

import (
	"testing"

	"github.com/cagelin/polymarket-go/models"
	"github.com/cagelin/polymarket-go/rfq"
)

type externalSigner struct{}

func (externalSigner) Address() string { return "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf" }
func (externalSigner) Order(models.SignedOrder, string, models.SignerType) (string, error) {
	return "signature", nil
}
func TestConfigAcceptsExternalSigner(t *testing.T) {
	config := rfq.Config{Signer: externalSigner{}}
	if config.Signer.Address() == "" {
		t.Error("Config(external signer) lost signer")
	}
}

func TestTypedNilSignerRejectedBeforeDial(t *testing.T) {
	var signer *externalSigner
	if _, err := rfq.Open(t.Context(), rfq.Config{Signer: signer, URL: "ws://invalid.invalid"}); err == nil {
		t.Error("Open(typed nil signer) succeeded, want input error")
	}
}
