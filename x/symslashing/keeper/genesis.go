package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/symslashing/types"
)

// InitGenesis initializes default parameters and the keeper's address to
// pubkey map.
func (k Keeper) InitGenesis(ctx sdk.Context, stakingKeeper types.StakingKeeper, data *types.GenesisState) {
	err := stakingKeeper.IterateValidators(ctx,
		func(index int64, validator abci.ValidatorUpdate) bool {
			pubkey := &ed25519.PubKey{Key: validator.PubKey.GetEd25519()}
			if err := k.AddPubkey(ctx, pubkey); err != nil {
				panic(err)
			}
			addr := sdk.ConsAddress(pubkey.Address())
			if err := k.SetValidatorSigningInfo(ctx, addr, types.NewValidatorSigningInfo(
				addr,
				ctx.BlockHeight(),
				0,
				0,
			)); err != nil {
				panic(err)
			}
			return false
		},
	)
	if err != nil {
		panic(err)
	}

	for _, info := range data.SigningInfos {
		address, err := k.sk.ConsensusAddressCodec().StringToBytes(info.Address)
		if err != nil {
			panic(err)
		}
		err = k.SetValidatorSigningInfo(ctx, address, info.ValidatorSigningInfo)
		if err != nil {
			panic(err)
		}
	}

	for _, array := range data.MissedBlocks {
		address, err := k.sk.ConsensusAddressCodec().StringToBytes(array.Address)
		if err != nil {
			panic(err)
		}

		for _, missed := range array.MissedBlocks {
			if err := k.SetMissedBlockBitmapValue(ctx, address, missed.Index, missed.Missed); err != nil {
				panic(err)
			}
		}
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	signingInfos := make([]types.SigningInfo, 0)
	missedBlocks := make([]types.ValidatorMissedBlocks, 0)
	err = k.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		bechAddr := address.String()
		signingInfos = append(signingInfos, types.SigningInfo{
			Address:              bechAddr,
			ValidatorSigningInfo: info,
		})

		localMissedBlocks, err := k.GetValidatorMissedBlocks(ctx, address)
		if err != nil {
			panic(err)
		}

		missedBlocks = append(missedBlocks, types.ValidatorMissedBlocks{
			Address:      bechAddr,
			MissedBlocks: localMissedBlocks,
		})

		return false
	})
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(params, signingInfos, missedBlocks)
}
