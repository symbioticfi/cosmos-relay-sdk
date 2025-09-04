package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/symstaking module sentinel errors
var (
	ErrInvalidSigner = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidKeyTag = errors.Register(ModuleName, 1101, "invalid key tag for validator")
)
