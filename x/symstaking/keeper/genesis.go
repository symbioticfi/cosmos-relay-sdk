package keeper

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) []abci.ValidatorUpdate {
	if err := k.Params.Set(ctx, genState.Params); err != nil {
		panic(err)
	}

	k.logger.Info("Initializing genesis state", "epoch", genState.GenesisEpoch)
	// set epoch
	if err := k.SetCurrentEpoch(ctx, &types.StoreEpoch{Epoch: genState.GenesisEpoch}); err != nil {
		panic(err)
	}
	// get validator set for current epoch
	valset, err := k.GetValidatorSet(ctx, genState.GenesisEpoch)
	if err != nil {
		panic(err)
	}
	// set last validator set
	if err = k.SetLastValidatorSet(ctx, &types.LastValidatorSet{
		Epoch:   genState.GenesisEpoch,
		Updates: valset,
	}); err != nil {
		panic(err)
	}

	return valset
}

// ExportGenesis returns the module's exported genesis.
func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params, err = k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	storeEpoch, err := k.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, err
	}
	genesis.GenesisEpoch = storeEpoch.Epoch

	return genesis, nil
}
