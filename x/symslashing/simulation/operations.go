package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/symslashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/symslashing/types"
)

// Simulation operation weights constants
// will be removed in the future
const (
	OpWeightMsgUnjail = "op_weight_msg_unjail"

	DefaultWeightMsgUnjail = 100
)

// WeightedOperations returns all the operations from the module with their respective weights
// migrate to the msg factories instead, this method will be removed in the future
func WeightedOperations(
	registry codectypes.InterfaceRegistry,
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
	sk types.StakingKeeper,
) simulation.WeightedOperations {
	var weightMsgUnjail int
	appParams.GetOrGenerate(OpWeightMsgUnjail, &weightMsgUnjail, nil, func(_ *rand.Rand) {
		weightMsgUnjail = DefaultWeightMsgUnjail
	})

	return simulation.WeightedOperations{}
}
