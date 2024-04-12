package pools

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/sqs/domain"
	"github.com/osmosis-labs/sqs/sqsdomain"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/cosmwasm/msg"
	cwpoolmodel "github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v24/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v24/x/poolmanager/types"
)

const (
	// placeholder for the code id of the pool that is not a cosm wasm pool
	notCosmWasmPoolCodeID = 0
)

var _ sqsdomain.RoutablePool = &routableCosmWasmPoolImpl{}

// routableCosmWasmPool is an implemenation of the cosm wasm pool
// that interacts with the chain for quotes and spot price.
type routableCosmWasmPoolImpl struct {
	ChainPool     *cwpoolmodel.CosmWasmPool "json:\"pool\""
	Balances      sdk.Coins                 "json:\"balances\""
	TokenOutDenom string                    "json:\"token_out_denom\""
	TakerFee      osmomath.Dec              "json:\"taker_fee\""
	SpreadFactor  osmomath.Dec              "json:\"spread_factor\""
	wasmClient    wasmtypes.QueryClient     "json:\"-\""
}

// GetId implements sqsdomain.RoutablePool.
func (r *routableCosmWasmPoolImpl) GetId() uint64 {
	return r.ChainPool.PoolId
}

// GetPoolDenoms implements sqsdomain.RoutablePool.
func (r *routableCosmWasmPoolImpl) GetPoolDenoms() []string {
	return r.Balances.Denoms()
}

// GetType implements sqsdomain.RoutablePool.
func (*routableCosmWasmPoolImpl) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.CosmWasm
}

// GetSpreadFactor implements sqsdomain.RoutablePool.
func (r *routableCosmWasmPoolImpl) GetSpreadFactor() math.LegacyDec {
	return r.SpreadFactor
}

// CalculateTokenOutByTokenIn implements sqsdomain.RoutablePool.
// It calculates the amount of token out given the amount of token in for a transmuter pool.
// Transmuter pool allows no slippage swaps. It just returns the same amount of token out as token in
// Returns error if:
// - the underlying chain pool set on the routable pool is not of transmuter type
// - the token in amount is greater than the balance of the token in
// - the token in amount is greater than the balance of the token out
func (r *routableCosmWasmPoolImpl) CalculateTokenOutByTokenIn(ctx context.Context, tokenIn sdk.Coin) (sdk.Coin, error) {
	poolType := r.GetType()

	// Ensure that the pool is cosmwasm
	if poolType != poolmanagertypes.CosmWasm {
		return sdk.Coin{}, domain.InvalidPoolTypeError{PoolType: int32(poolType)}
	}

	// Configure the calc query message
	calcMessage := msg.NewCalcOutAmtGivenInRequest(tokenIn, r.TokenOutDenom, r.SpreadFactor)

	calcOutAmtGivenInResponse := msg.CalcOutAmtGivenInResponse{}
	if err := queryCosmwasmContract(ctx, r.wasmClient, r.ChainPool.ContractAddress, &calcMessage, &calcOutAmtGivenInResponse); err != nil {
		return sdk.Coin{}, err
	}

	// No slippage swaps - just return the same amount of token out as token in
	// as long as there is enough liquidity in the pool.
	return calcOutAmtGivenInResponse.TokenOut, nil
}

// GetTokenOutDenom implements RoutablePool.
func (r *routableCosmWasmPoolImpl) GetTokenOutDenom() string {
	return r.TokenOutDenom
}

// String implements sqsdomain.RoutablePool.
func (r *routableCosmWasmPoolImpl) String() string {
	return fmt.Sprintf("pool (%d), pool type (%d) Generalized CosmWasm, pool denoms (%v), token out (%s)", r.ChainPool.PoolId, poolmanagertypes.CosmWasm, r.GetPoolDenoms(), r.TokenOutDenom)
}

// ChargeTakerFeeExactIn implements sqsdomain.RoutablePool.
// Returns tokenInAmount and does not charge any fee for transmuter pools.
func (r *routableCosmWasmPoolImpl) ChargeTakerFeeExactIn(tokenIn sdk.Coin) (inAmountAfterFee sdk.Coin) {
	tokenInAfterTakerFee, _ := poolmanager.CalcTakerFeeExactIn(tokenIn, r.GetTakerFee())
	return tokenInAfterTakerFee
}

// GetTakerFee implements sqsdomain.RoutablePool.
func (r *routableCosmWasmPoolImpl) GetTakerFee() math.LegacyDec {
	return r.TakerFee
}

// SetTokenOutDenom implements sqsdomain.RoutablePool.
func (r *routableCosmWasmPoolImpl) SetTokenOutDenom(tokenOutDenom string) {
	r.TokenOutDenom = tokenOutDenom
}

// CalcSpotPrice implements sqsdomain.RoutablePool.
func (r *routableCosmWasmPoolImpl) CalcSpotPrice(ctx context.Context, baseDenom string, quoteDenom string) (osmomath.BigDec, error) {
	request := msg.SpotPriceQueryMsg{
		SpotPrice: msg.SpotPrice{
			QuoteAssetDenom: quoteDenom,
			BaseAssetDenom:  baseDenom,
		},
	}

	response := &msg.SpotPriceQueryMsgResponse{}
	if err := queryCosmwasmContract(ctx, r.wasmClient, r.ChainPool.ContractAddress, &request, response); err != nil {
		return osmomath.BigDec{}, err
	}

	return osmomath.MustNewBigDecFromStr(response.SpotPrice), nil
}

// IsGeneralizedCosmWasmPool implements sqsdomain.RoutablePool.
func (*routableCosmWasmPoolImpl) IsGeneralizedCosmWasmPool() bool {
	return true
}

// GetCodeID implements sqsdomain.RoutablePool.
func (r *routableCosmWasmPoolImpl) GetCodeID() uint64 {
	return r.ChainPool.CodeId
}
