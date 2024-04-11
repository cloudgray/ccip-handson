package testhelpers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	helpers "touchstone.com/ccip/handson/common"
	"touchstone.com/ccip/handson/contracts/generated/link_token_interface"
	"touchstone.com/ccip/handson/contracts/generated/lock_release_token_pool"
	"touchstone.com/ccip/handson/contracts/generated/lock_release_token_pool_1_0_0"
	"touchstone.com/ccip/handson/contracts/generated/price_registry"
	"touchstone.com/ccip/handson/contracts/generated/router"
	"touchstone.com/ccip/handson/contracts/generated/weth9"
)

func updateCCIPSrcContracts(src *SourceChain, destChainID uint64) {
	destChainSelector := mustGetChainByEvmID(destChainID).Selector

	provideLiquidityToLinkPool(src.User, src.Chain, src.LinkToken, src.LinkTokenPool)
	provideLiquidityToWeth9Pool(src.User, src.Chain, src.WrappedNative, src.WrappedNativePool)

	applyPriceRegistryPriceUpdate(src.User, src.Chain, src.PriceRegistry, src.LinkToken.Address(), src.WrappedNative.Address(), destChainSelector)
	applyLockReleaseTokenPoolChainUpdates(src.User, src.Chain, destChainSelector, src.LinkTokenPool)
	applyLockReleaseTokenPool_1_0_0RampUpdates(src.User, src.Chain, src.WrappedNativePool, src.OnRamp.Address())
	applyRouterRampUpdates(src.User, src.Chain, src.Router, []router.RouterOnRamp{{DestChainSelector: destChainSelector, OnRamp: src.OnRamp.Address()}}, nil, nil)
}

func updateCCIPDestContracts(dest *DestinationChain, src *SourceChain) {
	provideLiquidityToLinkPool(dest.User, dest.Chain, dest.LinkToken, dest.LinkTokenPool)
	provideLiquidityToWeth9Pool(dest.User, dest.Chain, dest.WrappedNative, dest.WrappedNativePool)

	applyLockReleaseTokenPoolChainUpdates(dest.User, dest.Chain, src.ChainSelector, dest.LinkTokenPool)
	applyLockReleaseTokenPool_1_0_0RampUpdates(dest.User, dest.Chain, dest.WrappedNativePool, dest.OffRamp.Address())
	applyPriceRegistryUpdatersUpdate(dest.User, dest.Chain, dest.PriceRegistry, []common.Address{dest.CommitStore.Address()}, []common.Address{})
	applyRouterRampUpdates(dest.User, dest.Chain, dest.Router, nil, []router.RouterOffRamp{{SourceChainSelector: src.ChainSelector, OffRamp: dest.OffRamp.Address()}}, nil)
}

func provideLiquidityToLinkPool(
	transactor *bind.TransactOpts,
	chainClient *ethclient.Client,
	linkToken *link_token_interface.LinkToken,
	linkPool *lock_release_token_pool.LockReleaseTokenPool,
) {
	chainID := getChainID(chainClient)

	_, err := linkPool.Owner(nil)
	helpers.PrintAndPanicErr("error getting owner of dest pool: %v", err)

	tx, err := linkPool.SetRebalancer(transactor, transactor.From)
	helpers.PrintAndPanicErr("error setting rebalancer: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	
	tx, err = linkToken.Approve(transactor, linkPool.Address(), Link(200))
	helpers.PrintAndPanicErr("error approving link token: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)

	tx, err = linkPool.ProvideLiquidity(transactor, Link(200))
	helpers.PrintAndPanicErr("error providing liquidity: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	fmt.Printf("Provided liquidity to link pool!\n\n")
}

func provideLiquidityToWeth9Pool(
	transactor *bind.TransactOpts,
	chainClient *ethclient.Client,
	weth9 *weth9.WETH9,
	weth9Pool *lock_release_token_pool_1_0_0.LockReleaseTokenPool,
) {
	chainID := getChainID(chainClient)

	poolFloatValue := big.NewInt(1e12)
	transactor.Value = poolFloatValue
	tx, err := weth9.Deposit(transactor)
	helpers.PrintAndPanicErr("error depositing weth: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	fmt.Printf("Deposited wrapped native token!\n\n")

	transactor.Value = nil
	tx, err = weth9.Transfer(transactor, weth9Pool.Address(), poolFloatValue)
	helpers.PrintAndPanicErr("error transferring weth: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	fmt.Printf("Provided liquidity to wrapped native pool!\n\n")
}

func applyLockReleaseTokenPool_1_0_0RampUpdates(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client,
	pool *lock_release_token_pool_1_0_0.LockReleaseTokenPool,  
	onRamp common.Address,
) {
	chainID := getChainID(chainClient)
	tx, err := pool.ApplyRampUpdates(
		transactor,
		[]lock_release_token_pool_1_0_0.TokenPoolRampUpdate{
			{
				Ramp: onRamp, 
				Allowed: true,
				RateLimiterConfig: lock_release_token_pool_1_0_0.RateLimiterConfig{
					IsEnabled: true,
					Capacity:  Link(100),
					Rate:      big.NewInt(1e18),
				},
			},
		},
		[]lock_release_token_pool_1_0_0.TokenPoolRampUpdate{},
	)
	helpers.PrintAndPanicErr("error applying token pool chain update: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	fmt.Printf("Token pool chain updates applied!\n\n")
}


func applyLockReleaseTokenPoolChainUpdates(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client,
	destChainSelector uint64, 
	pool *lock_release_token_pool.LockReleaseTokenPool,
) {
	chainID := getChainID(chainClient)
	tx, err := pool.ApplyChainUpdates(
		transactor,
		[]lock_release_token_pool.TokenPoolChainUpdate{{
			RemoteChainSelector: destChainSelector,
			Allowed:             true,
			OutboundRateLimiterConfig: lock_release_token_pool.RateLimiterConfig{
				IsEnabled: true,
				Capacity:  Link(100),
				Rate:      big.NewInt(1e18),
			},
			InboundRateLimiterConfig: lock_release_token_pool.RateLimiterConfig{
				IsEnabled: true,
				Capacity:  Link(100),
				Rate:      big.NewInt(1e18),
			},
		}},
	)

	helpers.PrintAndPanicErr("error applying token pool chain update: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	fmt.Printf("Token pool chain updates applied!\n\n")
}

func applyPriceRegistryPriceUpdate(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client,
	srcPriceRegistry *price_registry.PriceRegistry, 
	srcLinkToken common.Address, 
	srcWeth9 common.Address, 
	destChainSelector uint64,
) {
	chainID := getChainID(chainClient)
	tx, err := srcPriceRegistry.UpdatePrices(transactor, price_registry.InternalPriceUpdates{
		TokenPriceUpdates: []price_registry.InternalTokenPriceUpdate{
			{
				SourceToken: srcLinkToken,
				UsdPerToken: new(big.Int).Mul(big.NewInt(1e18), big.NewInt(20)),
			},
			{
				SourceToken: srcWeth9,
				UsdPerToken: new(big.Int).Mul(big.NewInt(1e18), big.NewInt(2000)),
			},
		},
		GasPriceUpdates: []price_registry.InternalGasPriceUpdate{
			{
				DestChainSelector: destChainSelector,
				UsdPerUnitGas:     big.NewInt(20000e9),
			},
		},
	})
	helpers.PrintAndPanicErr("error updating prices: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	fmt.Printf("Price registry update applied!\n\n")
}

func applyPriceRegistryUpdatersUpdate(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client,
	destPriceRegistry *price_registry.PriceRegistry, 
	priceUpdatersToAdd []common.Address, 
	priceUpdatersToRemove []common.Address, 
) {
	chainID := getChainID(chainClient)

	tx, err := destPriceRegistry.ApplyPriceUpdatersUpdates(transactor, priceUpdatersToRemove, priceUpdatersToRemove)
	helpers.PrintAndPanicErr("error applying price updaters update: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	fmt.Printf("Price updaters updated!\n\n")
}

func applyRouterRampUpdates(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client,
	router *router.Router, 
	onRamps []router.RouterOnRamp, 
	offRampsToAdd []router.RouterOffRamp,
	offRampsToRemove []router.RouterOffRamp,
) {
	chainID := getChainID(chainClient)
	tx, err := router.ApplyRampUpdates(transactor, onRamps, offRampsToRemove, offRampsToAdd)
	helpers.PrintAndPanicErr("error applying router ramp updates: %v", err)
	helpers.ConfirmTXMined(context.Background(), chainClient, tx, chainID)
	fmt.Printf("Router ramp updates applied!\n\n")
}