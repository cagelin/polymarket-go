package signing

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/cagelin/polymarket-go/internal/chain"

	"github.com/cagelin/polymarket-go/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	proxyFactory          = chain.ProxyFactory
	proxyImplementation   = "44e999d5c2F66Ef0861317f9A4805AC2e90aEB4f"
	safeFactory           = chain.SafeFactory
	safeCodeHash          = "2bce2127ff07fb632d16c8347c4ebf501f4841168bed00d9e6ef715ddb6fcecf"
	depositFactory        = chain.DepositFactory
	depositImplementation = "58CA52ebe0DadfdF531Cde7062e76746de4Db1eB"
	depositBeacon         = "7A18EDfe055488A3128f01F563e5B479D92ffc3a"
)

func fixedHex(value string) []byte {
	raw, err := hex.DecodeString(strings.TrimPrefix(value, "0x"))
	if err != nil {
		panic(fmt.Sprintf("invalid embedded wallet bytecode: %v", err))
	}
	return raw
}
func create2(factory string, salt, codeHash []byte) string {
	factoryBytes := fixedHex(factory)
	if len(factoryBytes) != 20 || len(salt) != 32 || len(codeHash) != 32 {
		panic("invalid embedded CREATE2 input length")
	}
	raw := append([]byte{0xff}, factoryBytes...)
	raw = append(raw, salt...)
	raw = append(raw, codeHash...)
	return common.BytesToAddress(crypto.Keccak256(raw)[12:]).Hex()
}

// Wallets derives all owner wallet forms using the pinned production contracts.
func (s *Signer) Wallets() map[string]string {
	address := common.HexToAddress(s.Address()).Bytes()
	padded := common.LeftPadBytes(address, 32)
	proxyCode := fixedHex("3d3d606380380380913d393d73" + strings.TrimPrefix(proxyFactory, "0x") + "5af4602a57600080fd5b602d8060366000396000f3363d3d373d3d3d363d73" + proxyImplementation + "5af43d82803e903d91602b57fd5bf352e831dd" + strings.Repeat("0", 62) + "20" + strings.Repeat("0", 64))
	args := append(common.LeftPadBytes(common.HexToAddress(depositFactory).Bytes(), 32), padded...)
	salt := crypto.Keccak256(args)
	uupsPrefix, _ := new(big.Int).SetString("61003D3D8160233D3973", 16)
	uupsPrefix.Add(uupsPrefix, new(big.Int).Lsh(big.NewInt(int64(len(args))), 56))
	uups := append(common.LeftPadBytes(uupsPrefix.Bytes(), 10), fixedHex(depositImplementation+"6009"+"5155f3363d3d373d3d363d7f360894a13ba1a3210667c828492db98dca3e2076"+"cc3735a920a3ca505d382bbc545af43d6000803e6038573d6000fd5b3d6000f3")...)
	uups = append(uups, args...)
	beaconPrefix, _ := new(big.Int).SetString("6100523D8160233D3973", 16)
	beaconPrefix.Add(beaconPrefix, new(big.Int).Lsh(big.NewInt(int64(len(args))), 56))
	beacon := append(common.LeftPadBytes(beaconPrefix.Bytes(), 10), fixedHex(depositBeacon+"60195155f3363d3d373d3d363d602036600436635c60da"+"1b60e01b36527fa3f0ad74e5423aebfd80d3ef4346578335a9a72aeaee59ff6c"+"b3582b35133d50545afa5036515af43d6000803e604d573d6000fd5b3d6000f3")...)
	beacon = append(beacon, args...)
	return map[string]string{"EOA": s.Address(), "POLY_PROXY": create2(proxyFactory, crypto.Keccak256(address), crypto.Keccak256(proxyCode)), "GNOSIS_SAFE": create2(safeFactory, crypto.Keccak256(padded), fixedHex(safeCodeHash)), "UUPS": create2(depositFactory, salt, crypto.Keccak256(uups)), "BEACON": create2(depositFactory, salt, crypto.Keccak256(beacon))}
}

// Classify distinguishes owner-derived wallets from session-key deposit accounts.
func (s *Signer) Classify(wallet string) (models.WalletType, models.SignerType) {
	for kind, address := range s.Wallets() {
		if !strings.EqualFold(wallet, address) {
			continue
		}
		if kind == "UUPS" || kind == "BEACON" {
			return models.WalletDeposit, models.SignerOwner
		}
		return models.WalletType(kind), models.SignerOwner
	}
	return models.WalletDeposit, models.SignerSessionKey
}

// SignatureType maps wallet ownership to the exchange wire code.
func SignatureType(wallet models.WalletType) int {
	switch wallet {
	case models.WalletProxy:
		return 1
	case models.WalletSafe:
		return 2
	case models.WalletDeposit:
		return 3
	default:
		return 0
	}
}
