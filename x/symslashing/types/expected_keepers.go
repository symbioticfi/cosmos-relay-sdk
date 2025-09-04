package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

// AccountKeeper expected account keeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	IterateAccounts(ctx context.Context, process func(sdk.AccountI) (stop bool))
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	HasKeyTable() bool
	WithKeyTable(table paramtypes.KeyTable) paramtypes.Subspace
	Get(ctx sdk.Context, key []byte, ptr any)
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	SetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
}

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	SlashWithInfractionReason(context.Context, []byte, int64, int64, math.LegacyDec, stakingtypes.Infraction) (string, error)
	ConsensusAddressCodec() address.Codec
	IterateValidators(ctx context.Context, fn func(index int64, validator abci.ValidatorUpdate) (stop bool)) error
}

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	AfterValidatorCreated(ctx context.Context, consPubKey cryptotypes.PubKey) error  // Must be called when a validator is created
	AfterValidatorModified(ctx context.Context, consPubKey cryptotypes.PubKey) error // Must be called when a validator's state changes
	AfterValidatorRemoved(ctx context.Context, consPubKey cryptotypes.PubKey) error  // Must be called when a validator is deleted
}
