package abci

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/symstaking/keeper"
	symstakingTypes "github.com/cosmos/cosmos-sdk/x/symstaking/types"
)

type ProposalHandler struct {
	logger log.Logger
	keeper *keeper.Keeper
}

func NewProposalHandler(logger log.Logger, keeper *keeper.Keeper) *ProposalHandler {
	return &ProposalHandler{
		logger: logger,
		keeper: keeper,
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		proposalTxs := req.Txs

		params, err := h.keeper.Params.Get(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get params")
		}
		if req.Height%params.EpochCheckInterval != 0 {
			return &abci.ResponsePrepareProposal{
				Txs: proposalTxs,
			}, nil
		}

		epoch, err := h.keeper.GetCurrentEpoch(ctx)
		if err != nil {
			epoch = &symstakingTypes.StoreEpoch{Epoch: 0}
		}

		latestEpoch, err := h.keeper.GetLatestEpoch(ctx)
		if err != nil {
			h.logger.Error("PrepareProposal: failed to get latest epoch from relay", "err", err)
			return &abci.ResponsePrepareProposal{
				Txs: proposalTxs,
			}, nil
		}

		if latestEpoch <= epoch.Epoch {
			// if no new epoch found push existing one
			latestEpoch = epoch.Epoch
		}

		data := symstakingTypes.StoreEpoch{
			Epoch: latestEpoch,
		}
		bz, err := proto.Marshal(&data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode injected vote extension tx")
		}

		// Inject a "fake" tx into the proposal s.t. validators can decode, verify,
		// and store the canonical stake-weighted average prices.
		proposalTxs = append([][]byte{bz}, proposalTxs...)

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		params, err := h.keeper.Params.Get(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get params")
		}
		if req.Height%params.EpochCheckInterval != 0 || len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		var epoch symstakingTypes.StoreEpoch
		if err := proto.Unmarshal(req.Txs[0], &epoch); err != nil {
			h.logger.Error("ProcessProposal: failed to decode injected epoch tx", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		currentEpoch, err := h.keeper.GetCurrentEpoch(ctx)
		if err != nil {
			currentEpoch = &symstakingTypes.StoreEpoch{Epoch: 0}
		}

		if epoch.Epoch < currentEpoch.Epoch {
			h.logger.Error("ProcessProposal: invalid epoch number", "expected >=", currentEpoch.Epoch, "got", epoch.Epoch)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func (h *ProposalHandler) PreBlocker() sdk.PreBlocker {
	return func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		params, err := h.keeper.Params.Get(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get params")
		}
		if req.Height%params.EpochCheckInterval != 0 || len(req.Txs) == 0 {
			return &sdk.ResponsePreBlock{
				ConsensusParamsChanged: false,
			}, nil
		}

		var epoch symstakingTypes.StoreEpoch
		if err := proto.Unmarshal(req.Txs[0], &epoch); err != nil {
			return &sdk.ResponsePreBlock{
				ConsensusParamsChanged: false,
			}, errors.Wrap(err, "failed to decode injected epoch tx")
		}

		if err := h.keeper.SetCurrentEpoch(ctx, &epoch); err != nil {
			return &sdk.ResponsePreBlock{
				ConsensusParamsChanged: false,
			}, errors.Wrap(err, "failed to set current epoch")
		}
		return &sdk.ResponsePreBlock{
			ConsensusParamsChanged: false,
		}, nil
	}
}
