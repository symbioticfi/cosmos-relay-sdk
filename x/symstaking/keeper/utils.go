package keeper

import (
	"context"
	"strconv"

	"github.com/cometbft/cometbft/abci/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	v1 "github.com/symbioticfi/relay/api/client/v1"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	symStakingTypes "github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

func (k *Keeper) GetValidatorSet(ctx context.Context, epoch uint64) ([]types.ValidatorUpdate, error) {
	resp, err := k.relayClient.GetValidatorSet(ctx, &v1.GetValidatorSetRequest{
		Epoch: &epoch,
	})
	if err != nil {
		return nil, err
	}
	out := make([]types.ValidatorUpdate, len(resp.Validators))
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get params")
	}
	for i := range resp.Validators {
		pubKey, err := k.extractConsensusPubKey(resp.Validators[i].Keys, params.ValidatorKeyTag)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "failed to extract consensus pubkey for validator %s", resp.Validators[i].Operator)
		}
		votingPower, err := strconv.ParseInt(resp.Validators[i].VotingPower, 10, 64)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "failed to parse voting power for validator %s", resp.Validators[i].Operator)
		}
		out[i] = types.ValidatorUpdate{
			PubKey: *pubKey,
			Power:  votingPower,
		}
	}
	return out, nil
}

func (k *Keeper) GetLatestEpoch(ctx context.Context) (uint64, error) {
	// list through all settlement chains and find the lowest committed epoch
	resp, err := k.relayClient.GetLastAllCommitted(ctx, &v1.GetLastAllCommittedRequest{})
	if err != nil {
		return 0, err
	}

	// find the lowest epoch
	responseEpoch := uint64(0)
	for _, chainInfo := range resp.EpochInfos {
		if responseEpoch == 0 || chainInfo.LastCommittedEpoch < responseEpoch {
			responseEpoch = chainInfo.LastCommittedEpoch
		}
	}
	return responseEpoch, nil
}

// extractConsensusPubKey extracts the consensus public key from the key list
func (k *Keeper) extractConsensusPubKey(keys []*v1.Key, requiredKeyTag uint32) (*cmtprotocrypto.PublicKey, error) {
	for _, key := range keys {
		if key.Tag == requiredKeyTag {
			// Assuming the key payload is an Ed25519 public key (32 bytes)
			if len(key.Payload) != ed25519.PubKeySize {
				continue
			}

			pubKey := cmtprotocrypto.PublicKey{
				Sum: &cmtprotocrypto.PublicKey_Ed25519{
					Ed25519: key.Payload,
				},
			}
			return &pubKey, nil
		}
	}

	return nil, errorsmod.Wrapf(symStakingTypes.ErrInvalidKeyTag, "consensus key with tag %d not found", requiredKeyTag)
}
