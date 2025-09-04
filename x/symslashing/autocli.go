package slashing

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/symslashing/types"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current symslashing parameters",
				},
				{
					RpcMethod: "SigningInfo",
					Use:       "signing-info [validator-conspub/address]",
					Short:     "Query a validator's signing information",
					Long:      "Query a validator's signing information, with a pubkey ('<appd> comet show-validator') or a validator consensus address",
					Example:   fmt.Sprintf(`%s query symslashing signing-info '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"OauFcTKbN5Lx3fJL689cikXBqe+hcp6Y+x0rYUdR9Jk="}'`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "cons_address"},
					},
				},
				{
					RpcMethod: "SigningInfos",
					Use:       "signing-infos",
					Short:     "Query signing information of all validators",
				},
			},
		},
	}
}
