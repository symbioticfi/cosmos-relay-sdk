package types

import (
	errorsmod "cosmossdk.io/errors"
)

// NewParams creates a new Params instance.
func NewParams() Params {
	return Params{
		ValidatorKeyTag:    43, // type 2 (Ed25519) with id 11 (suggested for validator keys)
		EpochCheckInterval: 10, // every 10 cosmos blocks
		SigningKeyTag:      15, // Default symbiotic signing key
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.ValidatorKeyTag>>4 != 2 {
		return errorsmod.Wrapf(ErrInvalidKeyTag, "expected key tag to be of type 2 (indicating a ed25519 key), got %d", p.ValidatorKeyTag>>4)
	}
	return nil
}
