package models

// APIKeyCreds identifies a CLOB account. Formatting redacts all credential fields.
// JSON encoding is explicit data export and includes the values; do not log it.
type APIKeyCreds struct {
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
	Secret     string `json:"secret"`
}

// Value receivers redact values and pointers; decoding requires a pointer receiver.
// noinspection GoMixedReceiverTypes
func (c APIKeyCreds) String() string { return "APIKeyCreds(<redacted>)" }

// GoString redacts sensitive fields during Go-syntax formatting.
// Value receivers redact values and pointers; decoding requires a pointer receiver.
// noinspection GoMixedReceiverTypes
func (c APIKeyCreds) GoString() string { return c.String() }

// UnmarshalJSON accepts the exchange's apiKey field and the canonical key name.
// Value receivers redact values and pointers; decoding requires a pointer receiver.
// noinspection GoMixedReceiverTypes
func (c *APIKeyCreds) UnmarshalJSON(raw []byte) error {
	type plain APIKeyCreds
	var result plain
	if err := decodeData(raw, &result, []dataField{{name: "key", aliases: []string{"apiKey"}, required: true}, {name: "passphrase", required: true}, {name: "secret", required: true}}); err != nil {
		return err
	}
	*c = APIKeyCreds(result)
	return nil
}

// BuilderAPIKey authenticates builder-specific requests and redacts formatting.
type BuilderAPIKey struct {
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// Value receivers redact values and pointers; decoding requires a pointer receiver.
// noinspection GoMixedReceiverTypes
func (c BuilderAPIKey) String() string { return "BuilderAPIKey(<redacted>)" }

// GoString redacts sensitive fields during Go-syntax formatting.
// Value receivers redact values and pointers; decoding requires a pointer receiver.
// noinspection GoMixedReceiverTypes
func (c BuilderAPIKey) GoString() string { return c.String() }

// UnmarshalJSON accepts canonical credentials and exchange aliases.
// Value receivers redact values and pointers; decoding requires a pointer receiver.
// noinspection GoMixedReceiverTypes
func (c *BuilderAPIKey) UnmarshalJSON(raw []byte) error {
	type plain BuilderAPIKey
	var result plain
	if err := decodeData(raw, &result, []dataField{{name: "key", required: true}, {name: "secret", required: true}, {name: "passphrase", required: true}}); err != nil {
		return err
	}
	*c = BuilderAPIKey(result)
	return nil
}

// WalletType selects the exchange's four account signature modes.
type WalletType string

// Supported wallet ownership models.
const (
	WalletEOA     WalletType = "EOA"
	WalletProxy   WalletType = "POLY_PROXY"
	WalletSafe    WalletType = "GNOSIS_SAFE"
	WalletDeposit WalletType = "DEPOSIT_WALLET"
)

// SignerType distinguishes a wallet owner from an authorized session key.
type SignerType string

// Supported signing authorities.
const (
	SignerOwner      SignerType = "OWNER"
	SignerSessionKey SignerType = "SESSION_KEY"
)
