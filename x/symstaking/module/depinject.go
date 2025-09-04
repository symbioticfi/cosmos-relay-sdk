package symstaking

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"sort"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	v1 "github.com/symbioticfi/relay/api/client/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/symstaking/keeper"
	"github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

const (
	MockRelayRPCAddress = "mock-rpc"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.Register(
		&types.Module{},
		appconfig.Provide(ProvideModule),
		appconfig.Invoke(InvokeSetStakingHooks),
	)
}

type ModuleInputs struct {
	depinject.In

	Config                *types.Module
	StoreService          store.KVStoreService
	Cdc                   codec.Codec
	AddressCodec          runtime.ConsensusAddressCodec
	ConsensusAddressCodec runtime.ConsensusAddressCodec

	AuthKeeper types.AuthKeeper
	BankKeeper types.BankKeeper

	Logger log.Logger
}

type ModuleOutputs struct {
	depinject.Out

	SymstakingKeeper *keeper.Keeper
	Module           appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	// create relay client
	if in.Config.RelayClientRpc == "" {
		in.Config.RelayClientRpc = MockRelayRPCAddress
		in.Logger.Info("no relay client rpc address configured, defaulting to mock relay client (SYMBIOTIC_KEY_FILE based)", "address", MockRelayRPCAddress)
	}

	var client types.RelayClient
	if in.Config.RelayClientRpc == MockRelayRPCAddress {
		client = types.NewMockRelayClient(types.ValidatorFromFileGetter(os.Getenv("SYMBIOTIC_KEY_FILE")))
	} else {
		conn, err := GetGRPCConnection(in.Config.RelayClientRpc)
		if err != nil {
			panic(err)
		}
		client = v1.NewSymbioticClient(conn)
	}

	k := keeper.NewKeeper(
		in.Logger,
		in.StoreService,
		in.Cdc,
		in.AddressCodec,
		in.ConsensusAddressCodec,
		authority,
		client,
	)
	m := NewAppModule(in.Cdc, k, in.AuthKeeper, in.BankKeeper)

	return ModuleOutputs{SymstakingKeeper: k, Module: m}
}

func InvokeSetStakingHooks(
	config *types.Module,
	keeper *keeper.Keeper,
	stakingHooks map[string]types.SymStakingHooksWrapper,
) error {
	// all arguments to invokers are optional
	if keeper == nil || config == nil {
		return nil
	}

	modNames := slices.Collect(maps.Keys(stakingHooks))
	order := config.HooksOrder
	if len(order) == 0 {
		order = modNames
		sort.Strings(order)
	}

	if len(order) != len(modNames) {
		return fmt.Errorf("len(hooks_order: %v) != len(hooks modules: %v)", order, modNames)
	}

	if len(modNames) == 0 {
		return nil
	}

	var multiHooks types.MultiSymStakingHooks
	for _, modName := range order {
		hook, ok := stakingHooks[modName]
		if !ok {
			return fmt.Errorf("can't find staking hooks for module %s", modName)
		}

		multiHooks = append(multiHooks, hook)
	}
	keeper.SetHooks(multiHooks)
	return nil
}

func GetGRPCConnection(address string) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
	}
	unaryInterceptors := []grpc.UnaryClientInterceptor{grpc_retry.UnaryClientInterceptor(retryOpts...)}
	opts := []grpc.DialOption{
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryInterceptors...)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024), grpc.MaxCallSendMsgSize(100*1024*1024)),
	}

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	return grpc.NewClient(address, opts...)
}
