package types

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// combine multiple staking hooks, all hook functions are run in array sequence
var _ SymStakingHooks = &MultiSymStakingHooks{}

type MultiSymStakingHooks []SymStakingHooks

func NewMultiStakingHooks(hooks ...SymStakingHooks) MultiSymStakingHooks {
	return hooks
}

func (h MultiSymStakingHooks) AfterValidatorCreated(ctx context.Context, consPubKey cryptotypes.PubKey) error {
	for i := range h {
		if err := h[i].AfterValidatorCreated(ctx, consPubKey); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiSymStakingHooks) AfterValidatorModified(ctx context.Context, consPubKey cryptotypes.PubKey) error {
	for i := range h {
		if err := h[i].AfterValidatorModified(ctx, consPubKey); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiSymStakingHooks) AfterValidatorRemoved(ctx context.Context, consPubKey cryptotypes.PubKey) error {
	for i := range h {
		if err := h[i].AfterValidatorRemoved(ctx, consPubKey); err != nil {
			return err
		}
	}
	return nil
}
