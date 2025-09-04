package types

import (
	"context"

	v1 "github.com/symbioticfi/relay/api/client/v1"
	"google.golang.org/grpc"
)

type RelayClient interface {
	// Get current epoch
	GetCurrentEpoch(ctx context.Context, in *v1.GetCurrentEpochRequest, opts ...grpc.CallOption) (*v1.GetCurrentEpochResponse, error)
	// Get suggested epoch to request sign
	GetSuggestedEpoch(ctx context.Context, in *v1.GetSuggestedEpochRequest, opts ...grpc.CallOption) (*v1.GetSuggestedEpochResponse, error)
	// Get current validator set
	GetValidatorSet(ctx context.Context, in *v1.GetValidatorSetRequest, opts ...grpc.CallOption) (*v1.GetValidatorSetResponse, error)
	// Sign Message
	SignMessage(ctx context.Context, in *v1.SignMessageRequest, opts ...grpc.CallOption) (*v1.SignMessageResponse, error)
}
