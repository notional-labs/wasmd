package wasm_test

import (
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	keeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
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

func NewIBCChannel(path *ibctesting.Path, wasmTransferPortID string) *ibctesting.Path {
	path.EndpointA.ChannelConfig.PortID = wasmTransferPortID
	path.EndpointB.ChannelConfig.PortID = "transfer"
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"

	return path
}

func MakeBalancesString(balances map[string]uint64) string {
	balancesString := `"balances" : [ `
	for address, amount := range balances {
		balanceString := fmt.Sprintf(`{"amount" : %d,"address": "%s"},`, amount, address)
		balancesString += balanceString
	}
	l := len(balancesString)
	return balancesString[:l-1] + "]"
}

func TestIBCReflectContract(t *testing.T) {
	// scenario:
	//  chain A: ibc_reflect_send.wasm
	//  chain B: reflect.wasm + ibc_reflect.wasm
	//
	//  Chain A "ibc_reflect_send" sends a IBC packet "on channel connect" event to chain B "ibc_reflect"
	//  "ibc_reflect" sends a submessage to "reflect" which is returned as submessage.

	var (
		coordinator, appA, appB = ibctesting.NewCoordinator(t)
		chainA                  = coordinator.GetChain(ibctesting.GetChainID(0))
		chainB                  = coordinator.GetChain(ibctesting.GetChainID(1))
	)
	coordinator.CommitBlock(chainA, chainB)

	balances := map[string]uint64{

		"channel-1/swap":      99999999,
		"channel-1/join-pool": 99999999,
		"channel-2/swap":      99999999,
		"channel-2/join-pool": 99999999,
	}

	balances[chainA.SenderAccount.GetAddress().String()] = 999999999

	initMsg := []byte(fmt.Sprintf(`{"default_timeout": 999, %s}`, MakeBalancesString(balances)))
	fmt.Println(string(initMsg))
	codeID := chainA.StoreCodeFile("./keeper/testdata/cw20_ics20.wasm").CodeID
	sendContractAddr := chainA.InstantiateContract(codeID, initMsg)

	path := ibctesting.NewPath(chainA, chainB)
	path2 := ibctesting.NewPath(chainA, chainB)
	coordinator.SetupConnections(path)
	coordinator.SetupConnections(path2)

	icaMsg := &types.MsgCreateICAPortForSmartContract{
		Sender: chainA.SenderAccount.GetAddress().String(),

		ContractAddress: sendContractAddr.String(),

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
	icaAddress := pInstResp.IcaAddress
	controllerPortID, _ := types.GenerateICAPortID(sendContractAddr.String(), path.EndpointA.ConnectionID, path.EndpointB.ConnectionID)
	icaPath := NewICAChannel(path, controllerPortID)
	fmt.Println(icaPath.EndpointA.ConnectionID)
	fmt.Println(sendContractAddr.String())
	coordinator.CreateChannels(icaPath)

	// ctxA := chainA.GetContext()
	ctxB := chainB.GetContext()
	fmt.Println(appB.ICAHostKeeper.GetInterchainAccountAddress(ctxB, "wasm.0.0.cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"))
	fmt.Println(icaAddress)

	q := []byte(`{"channel":{"id":"channel-0"}}`)
	res, err := chainA.SmartQuery("cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr", q)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(res))

	ibcPortID := keeper.PortIDForContract(sendContractAddr)
	ibcPath := NewIBCChannel(path2, ibcPortID)
	coordinator.CreateChannels(ibcPath)

	icaAddr := "cosmos1nqtakv2fxs9mdgm98u2hzh3gj3xcn6a9599pzqnk9m2cf2haauqsrwyh79"
	ibcTx := []byte(`{"transfer": {"channel": "channel-1","remote_address": "cosmos1nqtakv2fxs9mdgm98u2hzh3gj3xcn6a9599pzqnk9m2cf2haauqsrwyh79"}}`)

	res = chainA.ExecuteContract(sendContractAddr.String(), ibcTx)
	appA.GetBaseApp()
	pack := &channeltypes.Packet{}
	err = pack.Unmarshal(res)
	if err != nil {
		fmt.Println(err.Error())
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	// path2.EndpointA.UpdateClient()
	// path2.EndpointB.UpdateClient()
	// coordinator.CommitBlock(chainB)
	// coordinator.CommitBlock(chainA)
	// path2.EndpointA.UpdateClient()
	// path2.EndpointB.UpdateClient()
	// coordinator.CommitBlock(chainB)
	// coordinator.CommitBlock(chainA)

	err = path2.RelayPacket(*pack, ack.Acknowledgement())
	if err != nil {
		panic(err)
	}
	icaAccAddr, _ := sdk.AccAddressFromBech32(icaAddr)

	fmt.Println(appB.BankKeeper.GetAllBalances(ctxB, icaAccAddr))
	icaTx := []byte(`{"swap" :{"": }}`)

	res = chainA.ExecuteContract(sendContractAddr.String(), icaTx)

	ack = channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	params := icahosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{"/osmosis.gamm.v1beta1.MsgSwapExactAmountIn",
			"/osmosis.gamm.v1beta1.MsgJoinPool",
			"/cosmos.bank.v1beta1.MsgSend",
		},
	}
	appB.ICAHostKeeper.SetParams(chainB.GetContext(), params)
	appB.Commit()
	chainB.NextBlock()
	fmt.Println(appB.ICAHostKeeper.GetAllowMessages(chainB.GetContext()))
	err = path.RelayPacket(*UmarshalPacket(res), ack.Acknowledgement())
	if err != nil {
		panic(err)
	}

}

func UmarshalPacket(bz []byte) *channeltypes.Packet {
	pack := &channeltypes.Packet{}
	err := pack.Unmarshal(bz)
	if err != nil {
		fmt.Println(err.Error())
	}
	return pack
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
