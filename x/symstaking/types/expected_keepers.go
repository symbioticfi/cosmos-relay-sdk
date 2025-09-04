package types

import (
	"context"

	"cosmossdk.io/core/address"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuthKeeper defines the expected interface for the Auth module.
type AuthKeeper interface {
	AddressCodec() address.Codec
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}

// Event Hooks
// These can be utilized to communicate between a staking keeper and another
// keeper which must take particular actions when validators/delegators change
// state. The second keeper must implement this interface, which then the
// staking keeper can call.

// SymStakingHooks event hooks for staking validator object (noalias)
type SymStakingHooks interface {
	AfterValidatorCreated(ctx context.Context, consPubKey cryptotypes.PubKey) error  // Must be called when a validator is created
	AfterValidatorModified(ctx context.Context, consPubKey cryptotypes.PubKey) error // Must be called when a validator's state changes
	AfterValidatorRemoved(ctx context.Context, consPubKey cryptotypes.PubKey) error  // Must be called when a validator is deleted
}

// SymStakingHooksWrapper is a wrapper for modules to inject StakingHooks using depinject.
type SymStakingHooksWrapper struct{ SymStakingHooks }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (SymStakingHooksWrapper) IsOnePerModuleType() {}
