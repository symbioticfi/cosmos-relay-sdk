package slashing

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)

	// Dev: disabled to prevent original slashing module from interfering with symSlashing
	// sdkCtx := sdk.UnwrapSDKContext(ctx)
	// for _, voteInfo := range sdkCtx.VoteInfos() {
	//	err := k.HandleValidatorSignature(ctx, voteInfo.Validator.Address, voteInfo.Validator.Power, comet.BlockIDFlag(voteInfo.BlockIdFlag))
	//	if err != nil {
	//		return err
	//	}
	// }
	return nil
}
