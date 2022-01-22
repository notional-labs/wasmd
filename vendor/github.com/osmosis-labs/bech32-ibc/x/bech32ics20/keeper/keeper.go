package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/osmosis-labs/bech32-ibc/x/bech32ics20/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		bankkeeper.Keeper
		channelKeeper          types.ChannelKeeper
		tk                     types.TransferKeeper
		hrpToChannelMapper     types.Bech32HrpToSourceChannelMap
		ics20TransferMsgServer types.ICS20TransferMsgServer
		cdc                    codec.Codec
		storeKey               sdk.StoreKey
		memKey                 sdk.StoreKey
	}
)

func NewKeeper(
	channelKeeper types.ChannelKeeper,
	bk bankkeeper.Keeper,
	tk types.TransferKeeper,
	hrpToChannelMapper types.Bech32HrpToSourceChannelMap,
	ics20TransferMsgServer types.ICS20TransferMsgServer,
	cdc codec.Codec,
) *Keeper {
	return &Keeper{
		Keeper:                 bk,
		channelKeeper:          channelKeeper,
		tk:                     tk,
		hrpToChannelMapper:     hrpToChannelMapper,
		ics20TransferMsgServer: ics20TransferMsgServer,
		cdc:                    cdc,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", banktypes.ModuleName))
}
