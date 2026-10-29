package signing

import (
	"math/big"
	"testing"

	"github.com/cagelin/polymarket-go/models"
)

func BenchmarkDepositBatchSignature(b *testing.B) {
	signer, err := New("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		b.Fatal(err)
	}
	wallet := signer.Wallets()["BEACON"]
	calls := []models.TransactionCall{{To: models.EVMAddress(signer.Address()), Data: "0x", Value: big.NewInt(1)}}
	b.ReportAllocs()
	for b.Loop() {
		if _, err = signer.SignDepositBatch(wallet, calls, "19", "1720000000", models.SignerOwner); err != nil {
			b.Fatal(err)
		}
	}
}
