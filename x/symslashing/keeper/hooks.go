package keeper

import (
	"context"

	"github.com/cometbft/cometbft/crypto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/symslashing/types"
	symStakingTypes "github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

var _ symStakingTypes.SymStakingHooks = Hooks{}

// Hooks wrapper struct for slashing keeper
type Hooks struct {
	k Keeper
}

// Hooks Return the slashing hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterValidatorBonded updates the signing info start height or create a new signing info
func (h Hooks) afterValidatorBonded(ctx context.Context, consPubKey cryptotypes.PubKey) error {
	consAddr := sdk.ConsAddress(consPubKey.Address())
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	signingInfo, err := h.k.GetValidatorSigningInfo(ctx, consAddr)
	if err == nil {
		signingInfo.StartHeight = sdkCtx.BlockHeight()
	} else {
		signingInfo = types.NewValidatorSigningInfo(
			consAddr,
			sdkCtx.BlockHeight(),
			0,
			0,
		)
	}

	return h.k.SetValidatorSigningInfo(ctx, consAddr, signingInfo)
}

// AfterValidatorRemoved deletes the address-pubkey relation when a validator is removed,
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consPubKey cryptotypes.PubKey) error {
	consAddr := sdk.ConsAddress(consPubKey.Address())
	return h.k.deleteAddrPubkeyRelation(ctx, crypto.Address(consAddr))
}

// AfterValidatorCreated adds the address-pubkey relation when a validator is created.
func (h Hooks) AfterValidatorCreated(ctx context.Context, consPubKey cryptotypes.PubKey) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := h.k.AddPubkey(sdkCtx, consPubKey); err != nil {
		return err
	}
	// we handle the creation of signing info here directly as symStaking doesn't have separate bonding events
	return h.afterValidatorBonded(ctx, consPubKey)
}

func (h Hooks) AfterValidatorModified(_ context.Context, _ cryptotypes.PubKey) error {
	return nil
}
