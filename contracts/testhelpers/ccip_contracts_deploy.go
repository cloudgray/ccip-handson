package testhelpers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	chainsel "github.com/smartcontractkit/chain-selectors"
	helpers "touchstone.com/ccip/handson/common"
	"touchstone.com/ccip/handson/contracts/generated/arm_proxy_contract"
	"touchstone.com/ccip/handson/contracts/generated/commit_store"
	"touchstone.com/ccip/handson/contracts/generated/commit_store_helper"
	"touchstone.com/ccip/handson/contracts/generated/evm_2_evm_offramp"
	"touchstone.com/ccip/handson/contracts/generated/evm_2_evm_onramp"
	"touchstone.com/ccip/handson/contracts/generated/link_token_interface"
	"touchstone.com/ccip/handson/contracts/generated/lock_release_token_pool"
	"touchstone.com/ccip/handson/contracts/generated/lock_release_token_pool_1_0_0"
	"touchstone.com/ccip/handson/contracts/generated/maybe_revert_message_receiver"
	"touchstone.com/ccip/handson/contracts/generated/mock_arm_contract"
	"touchstone.com/ccip/handson/contracts/generated/price_registry"
	"touchstone.com/ccip/handson/contracts/generated/router"
	"touchstone.com/ccip/handson/contracts/generated/weth9"
	"touchstone.com/ccip/handson/multienv"
)

func deployCCIPInfraContracts(
	env multienv.Env,
	chainID uint64,
) *Common {
	transactor := env.Transactors[chainID]
	chainClient := env.Clients[chainID]

	arm, srcARMProxy := deployArmWithProxy(transactor, chainClient)
	weth9 := deployWeth9(transactor, chainClient)
	linkToken := deployLinkToken(transactor, chainClient)
	weth9Pool := deployLockReleaseTokenPool_1_0_0(transactor, chainClient, weth9.Address(), srcARMProxy.Address())
	customToken := deployLinkTokenInterfaceImpl(transactor, chainClient) // Just re-use this, it's an ERC20.
	priceRegistry := deployPriceRegistry(transactor, chainClient, linkToken.Address(), weth9.Address())

	return &Common{
		ChainID:           chainID,
		ChainSelector:     mustGetChainByEvmID(chainID).Selector,
		User:              transactor,
		Chain:             chainClient,
		LinkToken:         linkToken,
		LinkTokenPool:     nil,
		CustomPool:        nil,
		CustomToken:       customToken,
		ARM:               arm,
		ARMProxy:          srcARMProxy,
		PriceRegistry:     priceRegistry,
		WrappedNative:     weth9,
		WrappedNativePool: weth9Pool,
	}
}

func deployCCIPLaneSourceContracts(src *Common, destChainID uint64) *SourceChain {
	router := deployRouter(src.User, src.Chain, src.WrappedNative.Address(), src.ARMProxy.Address())
	linkTokenPool := deployLockReleaseTokenPool(src.User, src.Chain, src.LinkToken.Address(), src.ARMProxy.Address(), router.Address())
	src.LinkTokenPool = linkTokenPool

	destChainSelector := mustGetChainByEvmID(destChainID).Selector
	onRamp := deployEVM2EVMOnRamp(src, router.Address(), destChainSelector)

	return &SourceChain{
		Common: *src,
		Router: router,
		OnRamp: onRamp,
	}
}

func deployCCIPLaneDestinationContracts(dest *Common, srcChain *SourceChain) *DestinationChain {
	router := deployRouter(dest.User, dest.Chain, dest.WrappedNative.Address(), dest.ARMProxy.Address())
	linkTokenPool := deployLockReleaseTokenPool(dest.User, dest.Chain, dest.LinkToken.Address(), dest.ARMProxy.Address(), router.Address())
	dest.LinkTokenPool = linkTokenPool

	srcChainSelector := srcChain.ChainSelector
	onRampAddress := srcChain.OnRamp.Address()
	commitStore, commitStoreHelper := deployCommitStore(
		dest.User, 
		dest.Chain, 
		dest.ChainSelector, 
		srcChainSelector, 
		onRampAddress, 
		dest.ARMProxy.Address())

	offRamp := deployEVM2EVMOffRamp(dest, commitStore.Address(), srcChain)
	
	revertingMessageReceiver1 := deployRevertingMessageReceiver(dest.User, dest.Chain)
	revertingMessageReceiver2 := deployRevertingMessageReceiver(dest.User, dest.Chain)

	return &DestinationChain{
		Common: *dest,
		CommitStoreHelper: commitStoreHelper,
		CommitStore:       commitStore,
		Router: router,
		OffRamp: offRamp,
		Receivers: []MaybeRevertReceiver{
			{Receiver: revertingMessageReceiver1, Strict: false}, 
			{Receiver: revertingMessageReceiver2, Strict: true},
		},
	}
}

func deployArmWithProxy(transactor *bind.TransactOpts, chainClient *ethclient.Client) (
	*mock_arm_contract.MockARMContract,
	*arm_proxy_contract.ARMProxyContract,
) {
	chainID := getChainID(chainClient)

	armAddress, tx, arm, err := mock_arm_contract.DeployMockARMContract(
		transactor,
		chainClient,
	)
	helpers.PrintAndPanicErr("error deploying mock arm contract: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)
	
	_, tx, armProxy, err := arm_proxy_contract.DeployARMProxyContract(
		transactor,
		chainClient,
		armAddress,
	)
	helpers.PrintAndPanicErr("error deploying arm proxy contract: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)

	return arm, armProxy
}

func deployLinkToken(transactor *bind.TransactOpts, chainClient *ethclient.Client) (*link_token_interface.LinkToken) {
	chainID := getChainID(chainClient)

	_, tx, linkToken, err := link_token_interface.DeployLinkToken(transactor, chainClient)
	helpers.PrintAndPanicErr("error deploying link token: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)

	return linkToken
}

func deployWeth9(transactor *bind.TransactOpts, chainClient *ethclient.Client) (*weth9.WETH9) {
	chainID := getChainID(chainClient)

	weth9Address, tx, _, err := weth9.DeployWETH9(transactor, chainClient)
	helpers.PrintAndPanicErr("error deploying weth9: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)

	weth9, err := weth9.NewWETH9(weth9Address, chainClient)
	helpers.PrintAndPanicErr("error creating weth9: %v", err)

	return weth9
}

func deployRouter(transactor *bind.TransactOpts, chainClient *ethclient.Client, weth9 common.Address, armProxy common.Address) (*router.Router) {
	chainID := getChainID(chainClient)
	
	_, tx, router, err := router.DeployRouter(transactor, chainClient, weth9, armProxy)
	helpers.PrintAndPanicErr("error deploying router: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)

	return router
}

func deployLockReleaseTokenPool_1_0_0(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client, 
	token common.Address, 
	armProxy common.Address,
) (*lock_release_token_pool_1_0_0.LockReleaseTokenPool) {
	chainID := getChainID(chainClient)

	_, tx, tokenPool, err := lock_release_token_pool_1_0_0.DeployLockReleaseTokenPool(
		transactor,
		chainClient,
		token,
		[]common.Address{},
		armProxy,
	)
	helpers.PrintAndPanicErr("error deploying lock release token pool: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)

	return tokenPool
}

func deployLockReleaseTokenPool(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client, 
	token common.Address, 
	armProxy common.Address, 
	router common.Address,
) (*lock_release_token_pool.LockReleaseTokenPool) {
	chainID := getChainID(chainClient)

	_, tx, tokenPool, err := lock_release_token_pool.DeployLockReleaseTokenPool(
		transactor,
		chainClient,
		token,
		[]common.Address{},
		armProxy,
		true,
		router,
	)
	helpers.PrintAndPanicErr("error deploying lock release token pool: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)

	return tokenPool
}

func deployLinkTokenInterfaceImpl(transactor *bind.TransactOpts, chainClient *ethclient.Client) (*link_token_interface.LinkToken) {
	linkTokenAddress, _, _, err := link_token_interface.DeployLinkToken(transactor, chainClient)
	helpers.PrintAndPanicErr("error deploying link token: %v", err)

	linkToken, err := link_token_interface.NewLinkToken(linkTokenAddress, chainClient)
	helpers.PrintAndPanicErr("error creating link token: %v", err)

	return linkToken
}

func deployPriceRegistry(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client, 
	linkToken common.Address, 
	weth9 common.Address,
) (*price_registry.PriceRegistry) {
	chainID := getChainID(chainClient)

	_, tx, priceRegistry, err := price_registry.DeployPriceRegistry(
		transactor,
		chainClient,
		nil,
		[]common.Address{linkToken, weth9},
		60*60*24*14, // two weeks
	)
	helpers.PrintAndPanicErr("error deploying price registry: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)

	return priceRegistry
}

func deployEVM2EVMOnRamp(
	src *Common,
	router common.Address, 
	destChainSelector uint64,
) (*evm_2_evm_onramp.EVM2EVMOnRamp) {
	chainID := getChainID(src.Chain)

	_, tx, onRamp, err := evm_2_evm_onramp.DeployEVM2EVMOnRamp(
		src.User,
		src.Chain,
		evm_2_evm_onramp.EVM2EVMOnRampStaticConfig{
			LinkToken:         src.LinkToken.Address(),
			ChainSelector:     src.ChainSelector,
			DestChainSelector: destChainSelector,
			DefaultTxGasLimit: 200_000,
			MaxNopFeesJuels:   big.NewInt(0).Mul(big.NewInt(100_000_000), big.NewInt(1e18)),
			PrevOnRamp:        common.HexToAddress(""),
			ArmProxy:          src.ARMProxy.Address(),
		},
		evm_2_evm_onramp.EVM2EVMOnRampDynamicConfig{
			Router:                            router,
			MaxNumberOfTokensPerMsg:           5,
			DestGasOverhead:                   350_000,
			DestGasPerPayloadByte:             16,
			DestDataAvailabilityOverheadGas:   33_596,
			DestGasPerDataAvailabilityByte:    16,
			DestDataAvailabilityMultiplierBps: 6840, // 0.684
			PriceRegistry:                     src.PriceRegistry.Address(),
			MaxDataBytes:                      1e5,
			MaxPerMsgGasLimit:                 4_000_000,
		},
		[]evm_2_evm_onramp.InternalPoolUpdate{
			{
				Token: src.LinkToken.Address(),
				Pool:  src.LinkTokenPool.Address(),
			},
			{
				Token: src.WrappedNative.Address(),
				Pool:  src.WrappedNativePool.Address(),
			},
		},
		evm_2_evm_onramp.RateLimiterConfig{
			IsEnabled: true,
			Capacity:  LinkUSDValue(100),
			Rate:      LinkUSDValue(1),
		},
		[]evm_2_evm_onramp.EVM2EVMOnRampFeeTokenConfigArgs{
			{
				Token:                      src.LinkToken.Address(),
				NetworkFeeUSDCents:         1_00,
				GasMultiplierWeiPerEth:     1e18,
				PremiumMultiplierWeiPerEth: 9e17,
				Enabled:                    true,
			},
			{
				Token:                      src.WrappedNative.Address(),
				NetworkFeeUSDCents:         1_00,
				GasMultiplierWeiPerEth:     1e18,
				PremiumMultiplierWeiPerEth: 1e18,
				Enabled:                    true,
			},
		},
		[]evm_2_evm_onramp.EVM2EVMOnRampTokenTransferFeeConfigArgs{
			{
				Token:             src.LinkToken.Address(),
				MinFeeUSDCents:    50,           // $0.5
				MaxFeeUSDCents:    1_000_000_00, // $ 1 million
				DeciBps:           5_0,          // 5 bps
				DestGasOverhead:   34_000,
				DestBytesOverhead: 0,
			},
		},
		[]evm_2_evm_onramp.EVM2EVMOnRampNopAndWeight{},
	)
	helpers.PrintAndPanicErr("error deploying onramp: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), src.Chain, tx, chainID)

	return onRamp
}

func deployCommitStore(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client, 
	chainSelector uint64, 
	srcChainSelector uint64, 
	onRamp common.Address, 
	armProxy common.Address,
) (*commit_store.CommitStore, *commit_store_helper.CommitStoreHelper) {
	chainID := getChainID(chainClient)

	commitStoreAddress, tx, commitStoreHelper, err := commit_store_helper.DeployCommitStoreHelper(
		transactor,
		chainClient,
		commit_store_helper.CommitStoreStaticConfig{
			ChainSelector:       chainSelector,
			SourceChainSelector: srcChainSelector,
			OnRamp:              onRamp,
			ArmProxy:            armProxy,
		},
	)
	helpers.PrintAndPanicErr("error deploying commit store: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, chainID)

	commitStore, err := commit_store.NewCommitStore(commitStoreAddress, chainClient)
	helpers.PrintAndPanicErr("error creating commit store: %v", err)

	return commitStore, commitStoreHelper
}

func deployEVM2EVMOffRamp(
	dest *Common,
	commitStore common.Address, 
	srcChain *SourceChain,
) (*evm_2_evm_offramp.EVM2EVMOffRamp) {
	chainID := getChainID(dest.Chain)

	_, tx, offRamp, err := evm_2_evm_offramp.DeployEVM2EVMOffRamp(
		dest.User,
		dest.Chain,
		evm_2_evm_offramp.EVM2EVMOffRampStaticConfig{
			CommitStore:         commitStore,
			ChainSelector:       dest.ChainSelector,
			SourceChainSelector: srcChain.ChainSelector,
			OnRamp:              srcChain.OnRamp.Address(),
			PrevOffRamp:         common.HexToAddress(""),
			ArmProxy:            dest.ARMProxy.Address(),
		},
		[]common.Address{srcChain.LinkToken.Address(), srcChain.WrappedNative.Address()},
		[]common.Address{dest.LinkTokenPool.Address(), dest.WrappedNativePool.Address()},
		evm_2_evm_offramp.RateLimiterConfig{
			IsEnabled: true,
			Capacity:  LinkUSDValue(100),
			Rate:      LinkUSDValue(1),
		},
	)
	helpers.PrintAndPanicErr("error deploying offramp: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), dest.Chain, tx, chainID)

	return offRamp
}

func deployRevertingMessageReceiver(
	transactor *bind.TransactOpts, 
	chainClient *ethclient.Client,
) *maybe_revert_message_receiver.MaybeRevertMessageReceiver {
	_, tx, revertingMessageReceiver, err := maybe_revert_message_receiver.DeployMaybeRevertMessageReceiver(transactor, chainClient, false)
	helpers.PrintAndPanicErr("error deploying reverting message receiver: %v", err)
	helpers.ConfirmContractDeployed(context.Background(), chainClient, tx, getChainID(chainClient))

	return revertingMessageReceiver
}

func getChainID(chainClient *ethclient.Client) int64 {
	chainID, err := chainClient.ChainID(context.Background())
	helpers.PrintAndPanicErr("error getting chain id: %v", err)

	return chainID.Int64()
}

func validateEnv(env multienv.Env, chainID uint64, websocket bool) {
	_, ok := env.Clients[chainID]
	if !ok {
		panic("SrcChain client not found")
	}

	_, ok = env.Transactors[chainID]
	if !ok {
		panic("SrcChain transactor not found")
	}

	if websocket {
		_, ok = env.WSURLs[chainID]
		if !ok {
			panic("SrcChain websocket URL not found")
		}
	}
}

func mustGetChainByEvmID(evmChainID uint64) chainsel.Chain {
	ch, exists := chainsel.ChainByEvmChainID(evmChainID)
	if !exists {
		helpers.PanicErr(fmt.Errorf("chain id %d doesn't exist in chain-selectors - forgot to add?", evmChainID))
	}
	return ch
}