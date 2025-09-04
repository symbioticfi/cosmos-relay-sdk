package simulation

import (
	"context"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	"github.com/cosmos/cosmos-sdk/x/symslashing/types"
)

// MsgUpdateParamsFactory creates a gov proposal for param updates
func MsgUpdateParamsFactory() simsx.SimMsgFactoryFn[*types.MsgUpdateParams] {
	return func(_ context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *types.MsgUpdateParams) {
		r := testData.Rand()
		params := types.DefaultParams()
		params.SignedBlocksWindow = int64(r.IntInRange(1, 1000))
		params.MinSignedPerWindow = sdkmath.LegacyNewDecWithPrec(int64(r.IntInRange(1, 100)), 2)
		params.SlashFractionDoubleSign = sdkmath.LegacyNewDecWithPrec(int64(r.IntInRange(1, 100)), 2)
		params.SlashFractionDowntime = sdkmath.LegacyNewDecWithPrec(int64(r.IntInRange(1, 100)), 2)

		return nil, &types.MsgUpdateParams{
			Authority: testData.ModuleAccountAddress(reporter, "gov"),
			Params:    params,
		}
	}
}
