package testhelpers

import (
	"fmt"

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
