package wasm_test

import (
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

func NewICAChannel(path *ibctesting.Path, controllerPortID string) *ibctesting.Path {
	path.EndpointA.ChannelConfig.PortID = controllerPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = icatypes.VersionPrefix
	path.EndpointB.ChannelConfig.Version = icatypes.VersionPrefix

	return path
}

func TestIBCReflectContract(t *testing.T) {
	// scenario:
	//  chain A: ibc_reflect_send.wasm
	//  chain B: reflect.wasm + ibc_reflect.wasm
	//
	//  Chain A "ibc_reflect_send" sends a IBC packet "on channel connect" event to chain B "ibc_reflect"
	//  "ibc_reflect" sends a submessage to "reflect" which is returned as submessage.

	var (
		coordinator = ibctesting.NewCoordinator(t, 2)
		chainA      = coordinator.GetChain(ibctesting.GetChainID(0))
		chainB      = coordinator.GetChain(ibctesting.GetChainID(1))
	)
	coordinator.CommitBlock(chainA, chainB)

	initMsg := []byte(`{"default_timeout": 999}`)
	codeID := chainA.StoreCodeFile("./keeper/testdata/cw20_ics20.wasm").CodeID
	sendContractAddr := chainA.InstantiateContract(codeID, initMsg).String()

	path := ibctesting.NewPath(chainA, chainB)
	coordinator.SetupConnections(path)

	icaMsg := &types.MsgCreateICAPortForSmartContract{
		Sender: chainA.SenderAccount.GetAddress().String(),

		ContractAddress: sendContractAddr,

		ConnectionId: path.EndpointA.ConnectionID,
	}
	fmt.Println(path.EndpointA.ConnectionID)
	result, err := chainA.SendMsgs(icaMsg)
	if err != nil {
		fmt.Println(err.Error())
	}
	protoResult := chainA.ParseSDKResultData(result)
	// unmarshal protobuf response from data
	var pInstResp types.MsgCreateICAPortForSmartContractResponse
	pInstResp.Unmarshal(protoResult.Data[0].Data)
	// icaAddress := pInstResp.IcaAddress
	controllerPortID, _ := types.GenerateICAPortID(sendContractAddr, path.EndpointA.ConnectionID, path.EndpointB.ConnectionID)

	icaPath := NewICAChannel(path, controllerPortID)
	fmt.Println(icaPath.EndpointA.ConnectionID)

	coordinator.CreateChannels(icaPath)

	// chainA.App.BankKeeper

	// bindPortMsg := wasmtypes.MsgCreateICAPortForSmartContract{
	// 	ContractAddress: sendContractAddr,
	// 	Sender: ,
	// }

	// reflectContractAddr := chainB.InstantiateContract(codeID, initMsg)
	// var (
	// 	sourcePortID      = chainA.ContractInfo(sendContractAddr).IBCPortID
	// 	counterpartPortID = chainB.ContractInfo(reflectContractAddr).IBCPortID
	// )

	// flip instantiation so that we do not run into https://github.com/cosmos/cosmos-sdk/issues/8334

	// TODO: query both contracts directly to ensure they have registered the proper connection
	// (and the chainB has created a reflect contract)

	// there should be one packet to relay back and forth (whoami)
	// TODO: how do I find the packet that was previously sent by the smart contract?
	// Coordinator.RecvPacket requires channeltypes.Packet as input?
	// Given the source (portID, channelID), we should be able to count how many packets are pending, query the data
	// and submit them to the other side (same with acks). This is what the real relayer does. I guess the test framework doesn't?

	// Update: I dug through the code, especially channel.Keeper.SendPacket, and it only writes a commitment
	// only writes I see: https://github.com/cosmos/cosmos-sdk/blob/31fdee0228bd6f3e787489c8e4434aabc8facb7d/x/ibc/core/04-channel/keeper/packet.go#L115-L116
	// commitment is hashed packet: https://github.com/cosmos/cosmos-sdk/blob/31fdee0228bd6f3e787489c8e4434aabc8facb7d/x/ibc/core/04-channel/types/packet.go#L14-L34
	// how is the relayer supposed to get the original packet data??
	// eg. ibctransfer doesn't store the packet either: https://github.com/cosmos/cosmos-sdk/blob/master/x/ibc/applications/transfer/keeper/relay.go#L145-L162
	// ... or I guess the original packet data is only available in the event logs????
	// https://github.com/cosmos/cosmos-sdk/blob/31fdee0228bd6f3e787489c8e4434aabc8facb7d/x/ibc/core/04-channel/keeper/packet.go#L121-L132

	// ensure the expected packet was prepared, and relay it

	// err := coordinator.RelayAndAckPendingPackets(chainA, chainB, clientA, clientB)
	// require.NoError(t, err)

	// // let's query the source contract and make sure it registered an address
	// query := ReflectSendQueryMsg{Account: &AccountQuery{ChannelID: channelA.ID}}
	// var account AccountResponse
	// err = chainA.SmartQuery(sendContractAddr.String(), query, &account)
	// require.NoError(t, err)
	// require.NotEmpty(t, account.RemoteAddr)
	// require.Empty(t, account.RemoteBalance)

	// // close channel
	// coordinator.CloseChannel(chainA, chainB, channelA, channelB)

	// // let's query the source contract and make sure it registered an address
	// account = AccountResponse{}
	// err = chainA.SmartQuery(sendContractAddr.String(), query, &account)
	// require.Error(t, err)
	// assert.Contains(t, err.Error(), "not found")

	// _ = clientB
}

type ReflectSendQueryMsg struct {
	Admin        *struct{}     `json:"admin,omitempty"`
	ListAccounts *struct{}     `json:"list_accounts,omitempty"`
	Account      *AccountQuery `json:"account,omitempty"`
}

type AccountQuery struct {
	ChannelID string `json:"channel_id"`
}

type AccountResponse struct {
	LastUpdateTime uint64            `json:"last_update_time,string"`
	RemoteAddr     string            `json:"remote_addr"`
	RemoteBalance  wasmvmtypes.Coins `json:"remote_balance"`
}
