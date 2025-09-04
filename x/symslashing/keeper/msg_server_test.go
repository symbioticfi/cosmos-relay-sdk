package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/symslashing/types"
)

func (s *KeeperTestSuite) TestUpdateParams() {
	require := s.Require()

	minSignedPerWindow, err := sdkmath.LegacyNewDecFromStr("0.60")
	require.NoError(err)

	slashFractionDoubleSign, err := sdkmath.LegacyNewDecFromStr("0.022")
	require.NoError(err)

	slashFractionDowntime, err := sdkmath.LegacyNewDecFromStr("0.0089")
	require.NoError(err)

	invalidVal, err := sdkmath.LegacyNewDecFromStr("-1")
	require.NoError(err)

	testCases := []struct {
		name      string
		request   *slashingtypes.MsgUpdateParams
		expectErr bool
		expErrMsg string
	}{
		{
			name: "set invalid authority",
			request: &slashingtypes.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
			expErrMsg: "invalid authority",
		},
		{
			name: "set invalid signed blocks window",
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
					SignedBlocksWindow:      0,
					MinSignedPerWindow:      minSignedPerWindow,
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: true,
			expErrMsg: "signed blocks window must be positive",
		},
		{
			name: "set invalid min signed per window",
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      invalidVal,
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: true,
			expErrMsg: "min signed per window cannot be negative",
		},
		{
			name: "set invalid downtime jail duration",
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      minSignedPerWindow,
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: true,
			expErrMsg: "downtime jail duration must be positive",
		},
		{
			name: "set invalid slash fraction double sign",
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      minSignedPerWindow,
					SlashFractionDoubleSign: invalidVal,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: true,
			expErrMsg: "double sign slash fraction cannot be negative",
		},
		{
			name: "set invalid slash fraction downtime",
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      minSignedPerWindow,
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   invalidVal,
				},
			},
			expectErr: true,
			expErrMsg: "downtime slash fraction cannot be negative",
		},
		{
			name: "set full valid params",
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      minSignedPerWindow,
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.request)
			if tc.expectErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}
