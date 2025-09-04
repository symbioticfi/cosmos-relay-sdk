package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

func (q queryServer) LastValidatorSet(ctx context.Context, req *types.QueryLastValidatorSetRequest) (*types.QueryLastValidatorSetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	storeValue, err := q.k.GetLastValidatorSet(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current epoch: %v", err)
	}

	return &types.QueryLastValidatorSetResponse{
		LastValidatorSet: storeValue,
	}, nil
}
