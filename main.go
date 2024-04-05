package main

import (
	"flag"
	"os"

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
	}
}