package testhelpers

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/smartcontractkit/libocr/offchainreporting/confighelper"
	"touchstone.com/ccip/handson/contracts/generated/arm_proxy_contract"
	"touchstone.com/ccip/handson/contracts/generated/commit_store"
	"touchstone.com/ccip/handson/contracts/generated/commit_store_helper"
	"touchstone.com/ccip/handson/contracts/generated/custom_token_pool"
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
)
var (
	Link        	= func(amount int64) *big.Int { return new(big.Int).Mul(big.NewInt(1e18), big.NewInt(amount)) }
	HundredLink 	= Link(100)
	LinkUSDValue 	= func(amount int64) *big.Int { return new(big.Int).Mul(big.NewInt(1e18), big.NewInt(amount)) }
)

type MaybeRevertReceiver struct {
	Receiver *maybe_revert_message_receiver.MaybeRevertMessageReceiver
	Strict   bool
}

type Common struct {
	ChainID           uint64
	ChainSelector     uint64
	User              *bind.TransactOpts
	Chain             *ethclient.Client
	LinkToken         *link_token_interface.LinkToken
	LinkTokenPool     *lock_release_token_pool.LockReleaseTokenPool
	CustomPool        *custom_token_pool.CustomTokenPool
	CustomToken       *link_token_interface.LinkToken
	WrappedNative     *weth9.WETH9
	WrappedNativePool *lock_release_token_pool_1_0_0.LockReleaseTokenPool
	ARM               *mock_arm_contract.MockARMContract
	ARMProxy          *arm_proxy_contract.ARMProxyContract
	PriceRegistry     *price_registry.PriceRegistry
}

type SourceChain struct {
	Common
	Router *router.Router
	OnRamp *evm_2_evm_onramp.EVM2EVMOnRamp
}

type DestinationChain struct {
	Common

	CommitStoreHelper *commit_store_helper.CommitStoreHelper
	CommitStore       *commit_store.CommitStore
	Router            *router.Router
	OffRamp           *evm_2_evm_offramp.EVM2EVMOffRamp
	Receivers         []MaybeRevertReceiver
}

type OCR2Config struct {
	Signers               []common.Address
	Transmitters          []common.Address
	F                     uint8
	OnchainConfig         []byte
	OffchainConfigVersion uint64
	OffchainConfig        []byte
}

type CCIPContracts struct {
	Source  SourceChain
	Dest    DestinationChain
	Oracles []confighelper.OracleIdentityExtra

	commitOCRConfig, execOCRConfig *OCR2Config
}
