package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/ethereum/go-ethereum/common/hexutil"
	v1 "github.com/symbioticfi/relay/api/client/v1"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type MockRelayValidatorGetter func(epoch uint64) []*v1.Validator
type MockRelayClient struct {
	validatorDataGetter MockRelayValidatorGetter
	currentEpoch        uint64
}

func NewMockRelayClient(getter MockRelayValidatorGetter) *MockRelayClient {
	return &MockRelayClient{
		validatorDataGetter: getter,
	}
}

func (m *MockRelayClient) GetCurrentEpoch(ctx context.Context, in *v1.GetCurrentEpochRequest, opts ...grpc.CallOption) (*v1.GetCurrentEpochResponse, error) {
	return &v1.GetCurrentEpochResponse{
		Epoch: m.currentEpoch,
	}, nil
}

func (m *MockRelayClient) GetSuggestedEpoch(ctx context.Context, in *v1.GetSuggestedEpochRequest, opts ...grpc.CallOption) (*v1.GetSuggestedEpochResponse, error) {
	return &v1.GetSuggestedEpochResponse{
		Epoch: m.currentEpoch,
	}, nil
}

func (m *MockRelayClient) GetValidatorSet(ctx context.Context, in *v1.GetValidatorSetRequest, opts ...grpc.CallOption) (*v1.GetValidatorSetResponse, error) {
	if m.currentEpoch <= *in.Epoch {
		m.currentEpoch += 5
	}
	vals := m.validatorDataGetter(*in.Epoch)
	return &v1.GetValidatorSetResponse{
		Epoch:      *in.Epoch,
		Validators: vals,
	}, nil
}

func (m *MockRelayClient) SignMessage(ctx context.Context, in *v1.SignMessageRequest, opts ...grpc.CallOption) (*v1.SignMessageResponse, error) {
	hasher := sha256.New()
	_, err := hasher.Write(in.Message)
	if err != nil {
		return nil, err
	}
	return &v1.SignMessageResponse{
		RequestHash: hexutil.Encode(hasher.Sum(nil)),
		Epoch:       m.currentEpoch,
	}, nil
}

func ValidatorFromFileGetter(filePath string) MockRelayValidatorGetter {
	return func(epoch uint64) []*v1.Validator {
		data, _, err := readKeysFromFile(filePath)
		if err != nil {
			panic(err)
		}

		var targetEpoch uint64 = 0
		for e := range data {
			if e <= epoch && e > targetEpoch {
				targetEpoch = e
			}
		}

		keys := data[targetEpoch]
		vals := make([]*v1.Validator, len(keys))
		for i := range keys {
			vals[i] = &v1.Validator{
				Operator:    fmt.Sprintf("0xValidator%v", i),
				VotingPower: "10000",
				IsActive:    true,
				Keys: []*v1.Key{
					{Tag: 43, Payload: keys[i].PubKey().Bytes()},
				},
			}
		}
		return vals
	}
}

// read file where each like is a hex encoded key, copnverrt to aray of byte arrays
func readKeysFromFile(filePath string) (map[uint64][]p2p.NodeKey, uint64, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, 0, err
	}
	typedData := map[uint64][]string{}
	if err := json.Unmarshal(data, &typedData); err != nil {
		return nil, 0, err
	}
	largestEpoch := uint64(0)
	keys := map[uint64][]p2p.NodeKey{}
	for epoch, k := range typedData {
		if epoch > largestEpoch {
			largestEpoch = epoch
		}
		finalKeys := []p2p.NodeKey{}
		for _, key := range k {
			if key == "" {
				continue
			}
			decoded, err := hex.DecodeString(key)
			if err != nil {
				return nil, 0, err
			}
			finalKeys = append(finalKeys, p2p.NodeKey{PrivKey: ed25519.PrivKey(decoded)})
		}
		keys[epoch] = finalKeys
	}
	return keys, largestEpoch, nil
}
