package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	v1 "github.com/symbioticfi/relay/api/client/v1"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

type Keeper struct {
	logger log.Logger

	storeService          corestore.KVStoreService
	cdc                   codec.Codec
	addressCodec          address.Codec
	consensusAddressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema collections.Schema
	Params collections.Item[types.Params]

	// Relay Client
	relayClient types.RelayClient

	hooks types.SymStakingHooks
}

const (
	RelayCurrentEpochKey              = "relay_current_epoch"
	RelayCurrentEpochValidatorInfoKey = "relay_epoch_validator_info"
)

func NewKeeper(
	logger log.Logger,
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	consensusAddressCodec address.Codec,
	authority []byte,
	relayClient types.RelayClient,
) *Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		logger:                logger,
		storeService:          storeService,
		cdc:                   cdc,
		addressCodec:          addressCodec,
		consensusAddressCodec: consensusAddressCodec,
		authority:             authority,
		relayClient:           relayClient,
		Params:                collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		hooks:                 nil,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return &k
}

func (k *Keeper) SetHooks(sh types.SymStakingHooks) {
	if k.hooks != nil {
		panic("cannot set symstaking hooks twice")
	}
	k.hooks = sh
}

func (k *Keeper) ConsensusAddressCodec() address.Codec {
	return k.consensusAddressCodec
}

func (k *Keeper) Hooks() types.SymStakingHooks {
	if k.hooks == nil {
		// return a no-op implementation if no hooks are set
		return types.MultiSymStakingHooks{}
	}

	return k.hooks
}

func (k *Keeper) GetCurrentEpoch(ctx context.Context) (*types.StoreEpoch, error) {
	data, err := k.storeService.OpenKVStore(ctx).Get([]byte(RelayCurrentEpochKey))
	if err != nil {
		return nil, errors.Wrap(err, "could not get current epoch")
	}

	if data == nil {
		return &types.StoreEpoch{Epoch: 0}, nil
	}
	var storeValue types.StoreEpoch
	if err := k.cdc.Unmarshal(data, &storeValue); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal current epoch")
	}
	return &storeValue, nil
}

func (k *Keeper) SetCurrentEpoch(ctx context.Context, storeValue *types.StoreEpoch) error {
	storeBytes, err := k.cdc.Marshal(storeValue)
	if err != nil {
		return errors.Wrap(err, "failed to marshal epoch")
	}

	if err := k.storeService.OpenKVStore(ctx).Set([]byte(RelayCurrentEpochKey), storeBytes); err != nil {
		return errors.Wrap(err, "failed to set epoch in store")
	}
	return nil
}

func (k *Keeper) SetLastValidatorSet(ctx context.Context, storeValue *types.LastValidatorSet) error {
	storeBytes, err := k.cdc.Marshal(storeValue)
	if err != nil {
		return errors.Wrap(err, "failed to marshal epoch")
	}

	if err := k.storeService.OpenKVStore(ctx).Set([]byte(RelayCurrentEpochValidatorInfoKey), storeBytes); err != nil {
		return errors.Wrap(err, "failed to set epoch in store")
	}
	return nil
}

func (k *Keeper) GetLastValidatorSet(ctx context.Context) (*types.LastValidatorSet, error) {
	data, err := k.storeService.OpenKVStore(ctx).Get([]byte(RelayCurrentEpochValidatorInfoKey))
	if err != nil {
		return nil, errors.Wrap(err, "could not get last validator set")
	}

	if data == nil {
		return &types.LastValidatorSet{Epoch: 0, Updates: nil}, nil
	}
	var storeValue types.LastValidatorSet
	if err := k.cdc.Unmarshal(data, &storeValue); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal last validator set")
	}
	return &storeValue, nil
}

// GetAuthority returns the module's authority.
func (k *Keeper) GetAuthority() []byte {
	return k.authority
}

func (k *Keeper) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	current, err := k.GetLastValidatorSet(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get last validator set")
	}
	currentEpoch, err := k.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get current epoch")
	}
	if current.Epoch == currentEpoch.Epoch {
		return nil, nil
	}
	newValset, err := k.GetValidatorSet(ctx, currentEpoch.Epoch)
	if err != nil {
		return nil, errors.Wrap(err, "could not get new validator set")
	}

	removed, added, updated := k.diffValidatorSets(current.Updates, newValset)
	if err := k.SetLastValidatorSet(ctx, &types.LastValidatorSet{
		Epoch:   currentEpoch.Epoch,
		Updates: newValset,
	}); err != nil {
		return nil, errors.Wrap(err, "could not set last validator set")
	}

	merged := append(updated, added...)
	merged = append(merged, removed...)

	for _, item := range removed {
		if err := k.Hooks().AfterValidatorRemoved(ctx, &ed25519.PubKey{Key: item.PubKey.GetEd25519()}); err != nil {
			return nil, err
		}
	}
	for _, item := range added {
		if err := k.Hooks().AfterValidatorCreated(ctx, &ed25519.PubKey{Key: item.PubKey.GetEd25519()}); err != nil {
			return nil, err
		}
	}
	for _, item := range updated {
		if err := k.Hooks().AfterValidatorModified(ctx, &ed25519.PubKey{Key: item.PubKey.GetEd25519()}); err != nil {
			return nil, err
		}
	}
	newSet := "New validator set:\n"
	for _, v := range merged {
		newSet += fmt.Sprintf("PubKey: %X, Power: %d\n", v.PubKey.GetEd25519(), v.Power)
	}
	k.logger.Info(newSet)
	// only return updates/new validators
	return merged, nil
}

func (k *Keeper) diffValidatorSets(old, new []abci.ValidatorUpdate) (removed, added, updated []abci.ValidatorUpdate) {
	oldMap := make(map[string]abci.ValidatorUpdate)
	newMap := make(map[string]abci.ValidatorUpdate)

	for _, val := range old {
		oldMap[val.PubKey.String()] = val
	}
	for _, val := range new {
		newMap[val.PubKey.String()] = val
	}

	for key, oldVal := range oldMap {
		if newVal, exists := newMap[key]; !exists {
			oldVal.Power = 0
			removed = append(removed, oldVal)
		} else if oldVal.Power != newVal.Power {
			updated = append(updated, newVal)
		}
	}

	for key, newVal := range newMap {
		if _, exists := oldMap[key]; !exists {
			added = append(added, newVal)
		}
	}

	return removed, added, updated
}

type slashMessage struct {
	ValidatorPk     string         `json:"validatorPk"`
	InfractionType  string         `json:"infractionType"`
	InfractionHeigh int64          `json:"infractionHeigh"`
	Power           int64          `json:"power"`
	SlashFactor     math.LegacyDec `json:"slashFactor"`
}

// SlashWithInfractionReason implementation doesn't require the infraction (types.Infraction) to work but is required by Interchain Security.
func (k *Keeper) SlashWithInfractionReason(ctx context.Context, validatorPubKey []byte, infractionHeight, power int64, slashFactor math.LegacyDec, typ types.Infraction) (string, error) {
	msg := slashMessage{
		ValidatorPk:     hexutil.Encode(validatorPubKey),
		InfractionType:  typ.String(),
		InfractionHeigh: infractionHeight,
		Power:           power,
		SlashFactor:     slashFactor,
	}
	dataBytes, err := json.Marshal(msg)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal slash message")
	}
	params, err := k.Params.Get(ctx)
	if err != nil {
		return "", errors.Wrap(err, "could not get params")
	}
	resp, err := k.relayClient.SignMessage(ctx, &v1.SignMessageRequest{
		KeyTag:  params.SigningKeyTag,
		Message: dataBytes,
	})
	if err != nil {
		return "", errors.Wrap(err, "could not sign message")
	}
	return resp.RequestHash, nil
}

// IterateValidators iterates through the validator set and perform the provided function
func (k *Keeper) IterateValidators(ctx context.Context, fn func(index int64, validator abci.ValidatorUpdate) (stop bool)) error {
	valset, err := k.GetLastValidatorSet(ctx)
	if err != nil {
		return err
	}
	for i, validator := range valset.Updates {
		stop := fn(int64(i), validator)

		if stop {
			break
		}
	}

	return nil
}
