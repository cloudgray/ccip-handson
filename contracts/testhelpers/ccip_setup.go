package testhelpers

import (
	"touchstone.com/ccip/handson/multienv"
)

func SetupCCIPLane(env multienv.Env, srcChainID, destChainID uint64) CCIPContracts {
	validateEnv(env, srcChainID, false)
	src := deployCCIPInfraContracts(env, srcChainID)
	srcChain := deployCCIPLaneSourceContracts(src, destChainID)
	updateCCIPSrcContracts(srcChain, destChainID)

	validateEnv(env, destChainID, false)
	dest := deployCCIPInfraContracts(env, destChainID)
	destChain := deployCCIPLaneDestinationContracts(dest, srcChain)
	updateCCIPDestContracts(destChain, srcChain)

	return CCIPContracts{
		Source: *srcChain,
		Dest:   *destChain,
	}
}
