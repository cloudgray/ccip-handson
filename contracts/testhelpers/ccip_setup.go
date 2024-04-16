package testhelpers

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	helpers "touchstone.com/ccip/handson/common"
	"touchstone.com/ccip/handson/contracts/generated/commit_store"
	"touchstone.com/ccip/handson/multienv"
)

func SetupCCIPLane(env multienv.Env, srcChainID, destChainID uint64) CCIPContracts {
	fmt.Println("\nCCIP Lane setup started!\n")

	fmt.Println("#########################################")
	fmt.Println("#   Setting up Source Chain CCIP Lane   #")
	fmt.Println("#########################################\n")
	validateEnv(env, srcChainID, false)
	src := deployCCIPInfraContracts(env, srcChainID)
	srcChain := deployCCIPLaneSourceContracts(src, destChainID)
	updateCCIPSrcContracts(srcChain, destChainID)

	fmt.Println("##########################################")
	fmt.Println("# Setting up Destination Chain CCIP Lane #")
	fmt.Println("##########################################")
	validateEnv(env, destChainID, false)
	dest := deployCCIPInfraContracts(env, destChainID)
	destChain := deployCCIPLaneDestinationContracts(dest, srcChain)
	updateCCIPDestContracts(destChain, srcChain)

	fmt.Println("\nCCIP Lane setup completed!\n")

	return CCIPContracts{
		Source: *srcChain,
		Dest:   *destChain,
	}
}

func SetCommitStoreConfig(env multienv.Env, chainID uint64, commitStoreAddress common.Address, signers, transmitters []common.Address) {
	transactor := env.Transactors[chainID]
	chainClient := env.Clients[chainID]

	commitStore, err := commit_store.NewCommitStore(commitStoreAddress, chainClient)
	helpers.PrintAndPanicErr("error creating commit store helper: %v", err)

	applyCommitStoreSetOCR2Config(transactor, chainClient, commitStore, signers, transmitters, 1, []byte{}, 1, []byte{})
}