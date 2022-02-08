package wasm_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	keeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	osmosisapp "github.com/osmosis-labs/osmosis/app"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
)

var (
	ContractAddr    sdk.AccAddress
	ContractIcaAddr sdk.AccAddress
)

func SetupICAChannel(path1 *ibctesting.Path, controllerPortID string) *ibctesting.Path {
	path1.EndpointA.ChannelConfig.PortID = controllerPortID
	path1.EndpointB.ChannelConfig.PortID = icatypes.PortID
	path1.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path1.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path1.EndpointA.ChannelConfig.Version = icatypes.VersionPrefix
	path1.EndpointB.ChannelConfig.Version = icatypes.VersionPrefix

	return path1
}

func SetupIBCChannel(path1 *ibctesting.Path, wasmTransferPortID string) *ibctesting.Path {
	path1.EndpointA.ChannelConfig.PortID = wasmTransferPortID
	path1.EndpointB.ChannelConfig.PortID = "transfer"
	path1.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path1.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path1.EndpointA.ChannelConfig.Version = "ics20-1"
	path1.EndpointB.ChannelConfig.Version = "ics20-1"

	return path1
}

func MakeInitMsg(balances map[string]uint64) []byte {
	balancesString := `"balances" : [ `
	for address, amount := range balances {
		balanceString := fmt.Sprintf(`{"amount" : %d,"address": "%s"},`, amount, address)
		balancesString += balanceString
	}
	l := len(balancesString)

	balancesString = balancesString[:l-1] + "]"
	return []byte(fmt.Sprintf(`{"default_timeout": 999, %s}`, balancesString))
}

func MakeBalanceQ(address string) []byte {
	return []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, address))
}

func MakeSwapTx(chann, in_denom, pool_id, exact_amount_out, remote_address, in_amount, out_denom string) []byte {
	msgStr := fmt.Sprintf(`{"swap" :{"in_denom": "%s", "in_amount": "%s", "channel": "%s", "remote_address": "%s", "pool_id": %s, "exact_amount_out": "%s", "out_denom": "%s"}}`,
		in_denom,
		in_amount,
		chann,
		remote_address,
		pool_id,
		exact_amount_out,
		out_denom,
	)
	return []byte(msgStr)
}

func CopyPath(path1 *ibctesting.Path) *ibctesting.Path {
	clone := ibctesting.NewPath(path1.EndpointA.Chain, path1.EndpointB.Chain)
	clone.EndpointA.ClientID = path1.EndpointA.ClientID
	clone.EndpointB.ClientID = path1.EndpointB.ClientID
	clone.EndpointA.ConnectionID = path1.EndpointA.ConnectionID
	clone.EndpointB.ConnectionID = path1.EndpointB.ConnectionID
	clone.EndpointA.ConnectionConfig = path1.EndpointA.ConnectionConfig
	clone.EndpointB.ConnectionConfig = path1.EndpointB.ConnectionConfig
	clone.EndpointA.ClientConfig = path1.EndpointA.ClientConfig
	clone.EndpointB.ClientConfig = path1.EndpointB.ClientConfig

	clone.EndpointA.Counterparty = clone.EndpointB
	clone.EndpointB.Counterparty = clone.EndpointA

	return clone
}

func FundOsmoForAcc(app *osmosisapp.OsmosisApp, osmoChain *ibctesting.TestChain, accAddr sdk.AccAddress) {
	ctx := osmoChain.GetContext()
	app.BankKeeper.MintCoins(ctx, gammtypes.ModuleName, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewIntFromUint64(99999999999999))))
	app.BankKeeper.SendCoinsFromModuleToAccount(ctx, gammtypes.ModuleName, accAddr, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewIntFromUint64(99999999999999))))
	app.Commit()
	osmoChain.NextBlock()
}

func MakeIBCmsg(chann, amount, toAddress string) []byte {
	msgString := fmt.Sprintf(`{"transfer": {"channel": "%s","remote_address": "%s", "amount": %s}}`, chann, toAddress, amount)
	return []byte(msgString)
}

// send contract token to an address on a foreign chain using ics20 packet
func ContractIBCTransfer(contractAddress, toAddress string, fromChain *ibctesting.TestChain, ibcPath *ibctesting.Path) {
	ibcTx := MakeIBCmsg("channel-1", "999999999", toAddress)
	res := fromChain.ExecuteContract(contractAddress, ibcTx)
	pack := &channeltypes.Packet{}
	err := pack.Unmarshal(res)
	if err != nil {
		panic(err)
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	err = ibcPath.RelayPacket(*pack, ack.Acknowledgement())
	if err != nil {
		panic(err)
	}
}

func CreateICAPortForContract(contractAddress string, connectionId string, contractChain *ibctesting.TestChain) {

}

// set up ibc channel and ica channel with the same connection for contract
func SetUpIbcAndICAChannelPathForContract() {

}

func UmarshalPacket(bz []byte) *channeltypes.Packet {
	pack := &channeltypes.Packet{}
	err := pack.Unmarshal(bz)
	if err != nil {
		fmt.Println(err.Error())
	}
	return pack
}

func TestIBCReflectContract(t *testing.T) {
	// create 2 test chain
	var (
		coordinator, appA, appB = ibctesting.NewCoordinator(t)
		chainA                  = coordinator.GetChain(ibctesting.GetChainID(0))
		chainB                  = coordinator.GetChain(ibctesting.GetChainID(1))
	)
	coordinator.CommitBlock(chainA, chainB)
	// store code
	codeID := chainA.StoreCodeFile("./keeper/testdata/gamm_contract.wasm").CodeID

	// create balances to instantiate contract
	balances := map[string]uint64{

		"channel-1/swap":      99999999,
		"channel-1/join-pool": 99999999,
		"channel-2/swap":      99999999,
		"channel-2/join-pool": 99999999,
	}
	balances[chainA.SenderAccount.GetAddress().String()] = 9999999999999999

	// instantiate contract
	initMsg := MakeInitMsg(balances)
	sendContractAddr := chainA.InstantiateContract(codeID, initMsg)

	// create paths and set up ibc conn between 2 chain
	path1 := ibctesting.NewPath(chainA, chainB)
	coordinator.SetupConnections(path1)
	// path2 have same conn and client as path1
	path2 := CopyPath(path1)

	q := []byte(`{"ran": {"address": "channel-1/swap"}}`)
	ran, err := chainA.SmartQuery(sendContractAddr.String(), q)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(ran))

	// contract msg for creating ica port
	icaMsg := &types.MsgCreateICAPortForSmartContract{
		Sender: chainA.SenderAccount.GetAddress().String(),

		ContractAddress: sendContractAddr.String(),

		ConnectionId: path1.EndpointA.ConnectionID,
	}
	result, err := chainA.SendMsgs(icaMsg)
	if err != nil {
		panic(err)
	}
	protoResult := chainA.ParseSDKResultData(result)
	var pInstResp types.MsgCreateICAPortForSmartContractResponse
	pInstResp.Unmarshal(protoResult.Data[0].Data)
	icaAddr := pInstResp.IcaAddress

	// set up ica chann for contract
	controllerPortID, _ := types.GenerateICAPortID(sendContractAddr.String(), path1.EndpointA.ConnectionID, path1.EndpointB.ConnectionID)
	icaPath := SetupICAChannel(path1, controllerPortID)
	coordinator.CreateChannels(icaPath)

	// set up ibc transfer chann for contract
	ibcPortID := keeper.PortIDForContract(sendContractAddr)
	ibcPath := SetupIBCChannel(path2, ibcPortID)
	coordinator.CreateChannels(ibcPath)

	// ibc transfer contract token from chainA to
	ibcTx := MakeIBCmsg("channel-1", "999999999", icaAddr)
	res := chainA.ExecuteContract(sendContractAddr.String(), ibcTx)
	appA.GetBaseApp()
	pack := &channeltypes.Packet{}
	err = pack.Unmarshal(res)
	if err != nil {
		panic(err)
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	err = path2.RelayPacket(*pack, ack.Acknowledgement())
	if err != nil {
		panic(err)
	}

	ibcTx = MakeIBCmsg("channel-1", "999999999", chainB.SenderAccount.GetAddress().String())
	res = chainA.ExecuteContract(sendContractAddr.String(), ibcTx)
	appA.GetBaseApp()
	pack = &channeltypes.Packet{}
	err = pack.Unmarshal(res)
	if err != nil {
		panic(err)
	}

	ack = channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	err = path2.RelayPacket(*pack, ack.Acknowledgement())
	if err != nil {
		panic(err)
	}

	poolParams := gammtypes.PoolParams{
		SwapFee: sdk.MustNewDecFromStr("0.003"),
		ExitFee: sdk.MustNewDecFromStr("0.001"),
	}

	asset1 := gammtypes.PoolAsset{
		Token:  sdk.NewCoin("ibc/1A757F169E3BB799B531736E060340FF68F37CBCEA881A147D83F84F7D87E828", sdk.NewIntFromUint64(999)),
		Weight: sdk.NewIntFromUint64(99999),
	}
	asset2 := gammtypes.PoolAsset{
		Token:  sdk.NewCoin("stake", sdk.NewIntFromUint64(999)),
		Weight: sdk.NewIntFromUint64(99999),
	}

	FundOsmoForAcc(appB, chainB, chainB.SenderAccount.GetAddress())
	assets := []gammtypes.PoolAsset{asset1, asset2}
	poolId, err := appB.GAMMKeeper.CreatePool(chainB.GetContext(), chainB.SenderAccount.GetAddress(), poolParams, assets, chainB.SenderAccount.GetAddress().String())
	if err != nil {
		panic(err)
	}

	//

	// test := "tr"

	icaTx := MakeSwapTx("channel-0", "ibc/1A757F169E3BB799B531736E060340FF68F37CBCEA881A147D83F84F7D87E828", strconv.Itoa(int(poolId)), "69", chainB.SenderAccount.GetAddress().String(), "100", "stake")

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
	err = path1.RelayPacket(*UmarshalPacket(res), ack.Acknowledgement())
	if err != nil {
		panic(err)
	}

}
