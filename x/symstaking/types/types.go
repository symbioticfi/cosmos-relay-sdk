package types

import (
	"context"

	v1 "github.com/symbioticfi/relay/api/client/v1"
	"google.golang.org/grpc"
)

type RelayClient interface {
	// Get current epoch
	GetCurrentEpoch(ctx context.Context, in *v1.GetCurrentEpochRequest, opts ...grpc.CallOption) (*v1.GetCurrentEpochResponse, error)
	// Get last committed epochs for all settlement chains
	GetLastAllCommitted(ctx context.Context, in *v1.GetLastAllCommittedRequest, opts ...grpc.CallOption) (*v1.GetLastAllCommittedResponse, error)
	// Get current validator set
	GetValidatorSet(ctx context.Context, in *v1.GetValidatorSetRequest, opts ...grpc.CallOption) (*v1.GetValidatorSetResponse, error)
	// Sign Message
	SignMessage(ctx context.Context, in *v1.SignMessageRequest, opts ...grpc.CallOption) (*v1.SignMessageResponse, error)
}
