package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

func (q queryServer) CurrentEpoch(ctx context.Context, req *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	storeValue, err := q.k.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current epoch: %v", err)
	}

	return &types.QueryCurrentEpochResponse{
		Epoch: storeValue.Epoch,
	}, nil
}
