package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
)

// Bech32HrpToSourceChannelMap defines the contract that must be fulfilled by a bech32 prefix to source
// channel mapper
// The x/bech32ibc keeper is a reference implementation and is expected to satisfy this interface
type Bech32HrpToSourceChannelMap interface {
	GetHrpSourceChannel(ctx sdk.Context, hrp string) (sourceChannel string, err error)
	GetNativeHrp(ctx sdk.Context) (hrp string, err error)
	GetHrpIbcRecord(ctx sdk.Context, hrp string) (bech32ibctypes.HrpIbcRecord, error)
}

// ICS20TransferMsgServer defines the contract that must be fulfilled by an ICS20 msg server
type ICS20TransferMsgServer interface {
	Transfer(goCtx context.Context, msg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error)
}

type TransferKeeper interface {
	// GetPort returns the portID for the transfer module. Used in ExportGenesis
	GetPort(ctx sdk.Context) string
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (_ channeltypes.Channel, found bool)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}
