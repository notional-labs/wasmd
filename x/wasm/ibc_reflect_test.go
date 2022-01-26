package wasm_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	keeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	osmosisapp "github.com/osmosis-labs/osmosis/app"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
)

func SetupICAChannel(path *ibctesting.Path, controllerPortID string) *ibctesting.Path {
	path.EndpointA.ChannelConfig.PortID = controllerPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = icatypes.VersionPrefix
	path.EndpointB.ChannelConfig.Version = icatypes.VersionPrefix

	return path
}

func SetupIBCChannel(path *ibctesting.Path, wasmTransferPortID string) *ibctesting.Path {
	path.EndpointA.ChannelConfig.PortID = wasmTransferPortID
	path.EndpointB.ChannelConfig.PortID = "transfer"
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"

	return path
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

func CopyPath(path *ibctesting.Path) *ibctesting.Path {
	clone := ibctesting.NewPath(path.EndpointA.Chain, path.EndpointB.Chain)
	clone.EndpointA.ClientID = path.EndpointA.ClientID
	clone.EndpointB.ClientID = path.EndpointB.ClientID
	clone.EndpointA.ConnectionID = path.EndpointA.ConnectionID
	clone.EndpointB.ConnectionID = path.EndpointB.ConnectionID
	clone.EndpointA.ConnectionConfig = path.EndpointA.ConnectionConfig
	clone.EndpointB.ConnectionConfig = path.EndpointB.ConnectionConfig
	clone.EndpointA.ClientConfig = path.EndpointA.ClientConfig
	clone.EndpointB.ClientConfig = path.EndpointB.ClientConfig

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

func TestIBCReflectContract(t *testing.T) {
	// create 2 test chain
	var (
		coordinator, appA, appB = ibctesting.NewCoordinator(t)
		chainA                  = coordinator.GetChain(ibctesting.GetChainID(0))
		chainB                  = coordinator.GetChain(ibctesting.GetChainID(1))
	)
	coordinator.CommitBlock(chainA, chainB)
	// store code
	codeID := chainA.StoreCodeFile("./keeper/testdata/cw20_ics20.wasm").CodeID

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
	path := ibctesting.NewPath(chainA, chainB)
	coordinator.SetupConnections(path)
	// path2 have same conn and client as path
	path2 := CopyPath(path)

	// contract msg for creating ica port
	icaMsg := &types.MsgCreateICAPortForSmartContract{
		Sender: chainA.SenderAccount.GetAddress().String(),

		ContractAddress: sendContractAddr.String(),

		ConnectionId: path.EndpointA.ConnectionID,
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
	controllerPortID, _ := types.GenerateICAPortID(sendContractAddr.String(), path.EndpointA.ConnectionID, path.EndpointB.ConnectionID)
	icaPath := SetupICAChannel(path, controllerPortID)
	coordinator.CreateChannels(icaPath)

	// set up ibc transfer chann for contract
	ibcPortID := keeper.PortIDForContract(sendContractAddr)
	ibcPath := SetupIBCChannel(path2, ibcPortID)
	coordinator.CreateChannels(ibcPath)

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
