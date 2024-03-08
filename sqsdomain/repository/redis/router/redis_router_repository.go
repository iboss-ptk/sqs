package routerredisrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/sqs/sqsdomain"
	"github.com/osmosis-labs/sqs/sqsdomain/repository"
)

// RouterRepository represent the router's repository contract
type RouterRepository interface {
	GetTakerFee(ctx context.Context, denom0, denom1 string) (osmomath.Dec, error)
	GetAllTakerFees(ctx context.Context) (sqsdomain.TakerFeeMap, error)
	SetTakerFee(ctx context.Context, tx repository.Tx, denom0, denom1 string, takerFee osmomath.Dec) error
}

type redisRouterRepo struct {
	repositoryManager repository.TxManager
}

const (
	keySeparator = "~"

	routerPrefix   = "r" + keySeparator
	takerFeePrefix = routerPrefix + "tf" + keySeparator
	routesPrefix   = routerPrefix + "r" + keySeparator
)

var (
	_ RouterRepository = &redisRouterRepo{}
)

// New will create an implementation of pools.Repository
func New(repositoryManager repository.TxManager, routesCacheExpirySeconds uint64) RouterRepository {
	return &redisRouterRepo{
		repositoryManager: repositoryManager,
	}
}

// GetAllTakerFees implements mvc.RouterRepository.
func (r *redisRouterRepo) GetAllTakerFees(ctx context.Context) (sqsdomain.TakerFeeMap, error) {
	tx := r.repositoryManager.StartTx()

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return nil, err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return nil, err
	}

	result := pipeliner.HGetAll(ctx, takerFeePrefix)

	_, err = pipeliner.Exec(ctx)
	if err != nil {
		return nil, err
	}

	resultMap, err := result.Result()
	if err != nil {
		return nil, err
	}

	// Parse taker fee map
	takerFeeMap := make(sqsdomain.TakerFeeMap, len(resultMap))
	for denomPairStr, takerFeeStr := range resultMap {
		takerFee, err := osmomath.NewDecFromStr(takerFeeStr)
		if err != nil {
			return nil, err
		}

		denoms := strings.Split(denomPairStr, keySeparator)

		if len(denoms) != 2 {
			return nil, fmt.Errorf("invalid denom pair string key %s. must have 2 denoms, had (%d)", denomPairStr, len(denoms))
		}

		if denoms[0] > denoms[1] {
			return nil, fmt.Errorf("invalid denom pair string key %s. must be in increasing lexicographic order", denomPairStr)
		}

		takerFeeMap[sqsdomain.DenomPair{
			Denom0: denoms[0],
			Denom1: denoms[1],
		}] = takerFee
	}

	return takerFeeMap, nil
}

// GetTakerFee implements mvc.RouterRepository.
func (r *redisRouterRepo) GetTakerFee(ctx context.Context, denom0 string, denom1 string) (osmomath.Dec, error) {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	tx := r.repositoryManager.StartTx()

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return osmomath.Dec{}, err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return osmomath.Dec{}, err
	}

	result := pipeliner.HGet(ctx, takerFeePrefix, denom0+keySeparator+denom1)

	_, err = pipeliner.Exec(ctx)
	if err != nil {
		return osmomath.Dec{}, err
	}

	resultStr, err := result.Result()
	if err != nil {
		return osmomath.Dec{}, err
	}

	return osmomath.NewDecFromStr(resultStr)
}

// SetTakerFee sets taker fee for a denom pair.
func (r *redisRouterRepo) SetTakerFee(ctx context.Context, tx repository.Tx, denom0, denom1 string, takerFee osmomath.Dec) error {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return err
	}
	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return err
	}

	cmd := pipeliner.HSet(ctx, takerFeePrefix, denom0+keySeparator+denom1, takerFee.String())
	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

func getRoutesPrefixByDenoms(denom0, denom1 string) string {
	return routesPrefix + denom0 + keySeparator + denom1
}
