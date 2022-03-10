package main

import (
	"fmt"

	"github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
)

func main() {
	sourcePrefix := types.GetDenomPrefix("transfer", "channel-1")
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + "test"

	// construct the denomination trace from the full raw denomination
	denomTrace := types.ParseDenomTrace(prefixedDenom)

	traceHash := denomTrace.IBCDenom()
	fmt.Println(traceHash)

}
