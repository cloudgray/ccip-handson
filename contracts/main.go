package main

import (
	"flag"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	helpers "touchstone.com/ccip/handson/common"
	"touchstone.com/ccip/handson/multienv"
	"touchstone.com/ccip/handson/testhelpers"
)

func main() {
	switch os.Args[1] {
	case "setup-ccip-lane":
		cmd := flag.NewFlagSet("setup-ccip-lane", flag.ExitOnError)
		srcChainID := cmd.Uint64("src-chain-id", 0, "Source Chain ID")
		destChainID := cmd.Uint64("dest-chain-id", 0, "Destination Chain ID")

		helpers.ParseArgs(cmd, os.Args[2:], "src-chain-id", "dest-chain-id")
		testhelpers.SetupCCIPLane(
			multienv.New(false, false),
			*srcChainID,
			*destChainID,
		)
	case "set-commit-store-config":
		cmd := flag.NewFlagSet("set-commit-store-config", flag.ExitOnError)
		chainID := cmd.Uint64("chain-id", 0, "Chain ID")
		commitStoreAddressHex := cmd.String("commit-store-address", "", "Commit Store Address")
		signersHex := cmd.String("signers", "", "Signers")
		transmittersHex := cmd.String("transmitters", "", "Transmitters")
		
		helpers.ParseArgs(cmd, os.Args[2:], "chain-id", "commit-store-address", "signers", "transmitters")
		commitStoreAddress := common.HexToAddress(*commitStoreAddressHex)
		signers := []common.Address{} 
		for _, signerHex := range strings.Split(*signersHex, ",") {
			signerHex = strings.TrimSpace(signerHex)
			signers = append(signers, common.HexToAddress(signerHex))
		}
		transmitters := []common.Address{}
		for _, transmitterHex := range strings.Split(*transmittersHex, ",") {
			transmitterHex = strings.TrimSpace(transmitterHex)
			transmitters = append(transmitters, common.HexToAddress(transmitterHex))
		}

		testhelpers.SetCommitStoreConfig(
			multienv.New(false, false),
			*chainID,
			commitStoreAddress,
			signers,
			transmitters,
		)
	}
}