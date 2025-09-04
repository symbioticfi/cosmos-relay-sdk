package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/symslashing/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/symslashing/types"
)

func (s *KeeperTestSuite) TestValidatorSigningInfo() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		ctx.BlockHeight(),
		int64(3),
		int64(10),
	)

	// set the validator signing information
	require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signingInfo))

	require.True(keeper.HasValidatorSigningInfo(ctx, consAddr))
	info, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.NoError(err)
	require.Equal(info.StartHeight, ctx.BlockHeight())
	require.Equal(info.IndexOffset, int64(3))
	require.Equal(info.MissedBlocksCounter, int64(10))

	var signingInfos []slashingtypes.ValidatorSigningInfo

	require.NoError(keeper.IterateValidatorSigningInfos(ctx, func(consAddr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	}))

	require.Equal(signingInfos[0].Address, signingInfo.Address)
}

func (s *KeeperTestSuite) TestValidatorMissedBlockBitmap_SmallWindow() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	for _, window := range []int64{100, 32_000} {
		params := testutil.TestParams()
		params.SignedBlocksWindow = window
		require.NoError(keeper.SetParams(ctx, params))

		// validator misses all blocks in the window
		var valIdxOffset int64
		for valIdxOffset < params.SignedBlocksWindow {
			idx := valIdxOffset % params.SignedBlocksWindow
			err := keeper.SetMissedBlockBitmapValue(ctx, consAddr, idx, true)
			require.NoError(err)

			missed, err := keeper.GetMissedBlockBitmapValue(ctx, consAddr, idx)
			require.NoError(err)
			require.True(missed)

			valIdxOffset++
		}

		// validator should have missed all blocks
		missedBlocks, err := keeper.GetValidatorMissedBlocks(ctx, consAddr)
		require.NoError(err)
		require.Len(missedBlocks, int(params.SignedBlocksWindow))

		// sign next block, which rolls the missed block bitmap
		idx := valIdxOffset % params.SignedBlocksWindow
		err = keeper.SetMissedBlockBitmapValue(ctx, consAddr, idx, false)
		require.NoError(err)

		missed, err := keeper.GetMissedBlockBitmapValue(ctx, consAddr, idx)
		require.NoError(err)
		require.False(missed)

		// validator should have missed all blocks except the last one
		missedBlocks, err = keeper.GetValidatorMissedBlocks(ctx, consAddr)
		require.NoError(err)
		require.Len(missedBlocks, int(params.SignedBlocksWindow)-1)
	}
}
