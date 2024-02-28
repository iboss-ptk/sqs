package usecase

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/domain"
	"github.com/osmosis-labs/sqs/router/usecase/route"
	"github.com/osmosis-labs/sqs/sqsdomain"
)

type (
	RouterUseCaseImpl = routerUseCaseImpl

	QuoteImpl = quoteImpl

	CandidatePoolWrapper = candidatePoolWrapper
)

const (
	OsmoPrecisionMultiplier = osmoPrecisionMultiplier
	NoTotalValueLockedError = noTotalValueLockedError
)

func (r *Router) ValidateAndFilterRoutes(candidateRoutes [][]candidatePoolWrapper, tokenInDenom string) (sqsdomain.CandidateRoutes, error) {
	return r.validateAndFilterRoutes(candidateRoutes, tokenInDenom)
}

func (r *routerUseCaseImpl) InitializeDefaultRouter() *Router {
	return r.initializeDefaultRouter()
}

func (r *routerUseCaseImpl) HandleRoutes(ctx context.Context, router *Router, tokenInDenom, tokenOutDenom string) (candidateRoutes sqsdomain.CandidateRoutes, err error) {
	return r.handleCandidateRoutes(ctx, router, tokenInDenom, tokenOutDenom)
}

func (r *Router) EstimateAndRankSingleRouteQuote(ctx context.Context, routes []route.RouteImpl, tokenIn sdk.Coin) (domain.Quote, []RouteWithOutAmount, error) {
	return r.estimateAndRankSingleRouteQuote(ctx, routes, tokenIn)
}

// GetSortedPoolIDs returns the sorted pool IDs.
// The sorting is initialized in NewRouter() by preferredPoolIDs and TVL.
// Only used for tests.
func (r Router) GetSortedPoolIDs() []uint64 {
	sortedPoolIDs := make([]uint64, len(r.sortedPools))
	for i, pool := range r.sortedPools {
		sortedPoolIDs[i] = pool.GetId()
	}
	return sortedPoolIDs
}

func FilterDuplicatePoolIDRoutes(rankedRoutes []route.RouteImpl) []route.RouteImpl {
	return filterDuplicatePoolIDRoutes(rankedRoutes)
}

func ConvertRankedToCandidateRoutes(rankedRoutes []route.RouteImpl) sqsdomain.CandidateRoutes {
	return convertRankedToCandidateRoutes(rankedRoutes)
}

func FormatRankedRouteCacheKey(tokenInDenom string, tokenOutDenom string, tokenIOrderOfMagnitude int) string {
	return formatRankedRouteCacheKey(tokenInDenom, tokenOutDenom, tokenIOrderOfMagnitude)
}

func FormatRouteCacheKey(tokenInDenom string, tokenOutDenom string) string {
	return formatRouteCacheKey(tokenInDenom, tokenOutDenom)
}
