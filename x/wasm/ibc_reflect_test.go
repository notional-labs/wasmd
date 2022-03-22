package wasm_test

import (
	"fmt"
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v2/testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/stretchr/testify/require"

	wasmibctesting "github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmosis "github.com/osmosis-labs/osmosis/v7/app"
	gammpool "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func MakeFundIbcGammContractTx(channelId, remoteGammContract string) []byte {
	msgStr := fmt.Sprintf(`{"ibc_fund_gamm_contract": {"channel_id": "%s", "remote_gamm_contract_address": "%s"}}`,
		channelId,
		remoteGammContract,
	)
	return []byte(msgStr)
}

func MakeSwapTx(channelID, poolID, inDenom, inAmount, outDenom, minOutAmount, toAddress string) []byte {
	msgStr := fmt.Sprintf(`{"ibc_swap" :{"channel_id": "%s", "pool_id": %s, "in_denom": "%s", "in_amount": "%s", "out_denom": "%s", "min_out_amount": "%s","to_address": "%s"}}`,
		channelID,
		poolID,
		inDenom,
		inAmount,
		outDenom,
		minOutAmount,
		toAddress,
	)
	return []byte(msgStr)
}

func MakeSetIbcDenomTx(ibcDenom, contractChannelID, contractNativeDenom string) []byte {
	msgStr := fmt.Sprintf(`{"set_ibc_denom_for_contract" :{"ibc_denom": "%s", "contract_channel_id": "%s", "contract_native_denom": "%s"}}`,
		ibcDenom,
		contractChannelID,
		contractNativeDenom,
	)
	return []byte(msgStr)
}

func MakeSpotPriceQueryTx(channelID, poolID, inDenom, outDenom, withSwapFee string) []byte {
	msgStr := fmt.Sprintf(`{"spot_price_query": {"channel_id": "%s", "pool_id": %s, "in_denom": "%s", "out_denom": "%s", "with_swap_fee": %s}}`,
		channelID,
		poolID,
		inDenom,
		outDenom,
		withSwapFee,
	)
	return []byte(msgStr)
}

func FundOsmoForAcc(app *osmosis.OsmosisApp, osmoChain *wasmibctesting.TestChain, accAddr sdk.AccAddress) {
	ctx := osmoChain.GetContext()
	app.BankKeeper.MintCoins(ctx, gammtypes.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewIntFromUint64(99999999999999))))
	app.BankKeeper.SendCoinsFromModuleToAccount(ctx, gammtypes.ModuleName, accAddr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewIntFromUint64(99999999999999))))
	app.Commit()
	osmoChain.NextBlock()
}

func CopyPath(path1 *wasmibctesting.Path) *wasmibctesting.Path {
	clone := wasmibctesting.NewPath(path1.EndpointA.Chain, path1.EndpointB.Chain)
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

func TestIBCReflectContract(t *testing.T) {
	// scenario:
	//  chain A: ibc_reflect_send.wasm
	//  chain B: reflect.wasm + ibc_reflect.wasm
	//
	//  Chain A "ibc_reflect_send" sends a IBC packet "on channel connect" event to chain B "ibc_reflect"
	//  "ibc_reflect" sends a submessage to "reflect" which is returned as submessage.

	var (
		coordinator, wasmApp, osmosisApp = wasmibctesting.NewCoordinator(t, 2)
		wasmChain                        = coordinator.GetChain(wasmibctesting.GetChainID(0))
		osmosisChain                     = coordinator.GetChain(wasmibctesting.GetChainID(1))
	)
	coordinator.CommitBlock(wasmChain, osmosisChain)

	initMsg := []byte(`{}`)
	osmosisContractCodeID := osmosisChain.StoreCodeFile("./keeper/testdata/ibc_gamm_osmosis.wasm").CodeID
	osmosisContractAddr := osmosisChain.InstantiateContract(osmosisContractCodeID, initMsg)

	junoContractCodeID := wasmChain.StoreCodeFile("./keeper/testdata/ibc_gamm_juno.wasm").CodeID

	junoContractAddr := wasmChain.InstantiateContract(junoContractCodeID, initMsg)
	var (
		osmosisContractPortID = osmosisChain.ContractInfo(osmosisContractAddr).IBCPortID
		junoCountractPortID   = wasmApp.GetWasmKeeper().GetContractInfo(wasmChain.GetContext(), junoContractAddr).IBCPortID
	)
	coordinator.CommitBlock(wasmChain, osmosisChain)
	coordinator.UpdateTime()

	require.Equal(t, wasmChain.CurrentHeader.Time, osmosisChain.CurrentHeader.Time)
	ibcGammPath := wasmibctesting.NewPath(osmosisChain, wasmChain)
	ibcGammPath.EndpointA.ChannelConfig = &ibctesting.ChannelConfig{
		PortID:  osmosisContractPortID,
		Version: "ibc-gamm-1",
		Order:   channeltypes.UNORDERED,
	}
	ibcGammPath.EndpointB.ChannelConfig = &ibctesting.ChannelConfig{
		PortID:  junoCountractPortID,
		Version: "ibc-gamm-1",
		Order:   channeltypes.UNORDERED,
	}

	coordinator.SetupConnections(ibcGammPath)
	coordinator.CreateChannels(ibcGammPath)

	transferPath := CopyPath(ibcGammPath)
	transferPath.EndpointA.ChannelConfig = &ibctesting.ChannelConfig{
		PortID:  "transfer",
		Version: "ics20-1",
		Order:   channeltypes.UNORDERED,
	}
	transferPath.EndpointB.ChannelConfig = &ibctesting.ChannelConfig{
		PortID:  junoCountractPortID,
		Version: "ics20-1",
		Order:   channeltypes.UNORDERED,
	}
	coordinator.CreateChannels(transferPath)

	fundMsg := MakeFundIbcGammContractTx(transferPath.EndpointB.ChannelID, osmosisContractAddr.String())

	wasmChain.ExecuteContract(junoContractAddr.String(), fundMsg)

	err := coordinator.RelayAndAckPendingPackets(transferPath)
	if err != nil {
		panic(err)
	}

	poolParams := gammpool.PoolParams{
		SwapFee: sdk.MustNewDecFromStr("0.003"),
		ExitFee: sdk.MustNewDecFromStr("0.001"),
	}

	asset1 := gammtypes.PoolAsset{
		Token:  sdk.NewCoin("ibc/1A757F169E3BB799B531736E060340FF68F37CBCEA881A147D83F84F7D87E828", sdk.NewIntFromUint64(200)),
		Weight: sdk.NewIntFromUint64(99999),
	}
	asset2 := gammtypes.PoolAsset{
		Token:  sdk.NewCoin("stake", sdk.NewIntFromUint64(200)),
		Weight: sdk.NewIntFromUint64(99999),
	}
	assets := []gammtypes.PoolAsset{asset1, asset2}
	FundOsmoForAcc(osmosisApp, osmosisChain, osmosisContractAddr)

	setIbcDenomForContractTx := MakeSetIbcDenomTx("ibc/1A757F169E3BB799B531736E060340FF68F37CBCEA881A147D83F84F7D87E828", transferPath.EndpointA.ChannelID, "test")

	_, ev := osmosisChain.ExecuteContract(osmosisContractAddr.String(), setIbcDenomForContractTx)
	for _, i := range ev {
		if i.Type == "wasm" {
			fmt.Println(i.String())
		}
	}

	_, err = osmosisApp.GAMMKeeper.CreateBalancerPool(osmosisChain.GetContext(), osmosisContractAddr, poolParams, assets, osmosisContractAddr.String())
	if err != nil {
		panic(err)
	}
	fmt.Println(osmosisApp.BankKeeper.GetAllBalances(osmosisChain.GetContext(), osmosisContractAddr))

	ibcSwapMsg := MakeSwapTx(ibcGammPath.EndpointB.ChannelID,
		"1",
		"ibc/1A757F169E3BB799B531736E060340FF68F37CBCEA881A147D83F84F7D87E828",
		"10",
		"stake",
		"1",
		osmosisChain.SenderAccount.GetAddress().String(),
	)
	fmt.Println(string(ibcSwapMsg))

	fmt.Println(ibcGammPath.EndpointA.ConnectionID, transferPath.EndpointA.ConnectionID)

	// wasmChain.ExecuteContract(junoContractAddr.String(), ibcSwapMsg)

	// fmt.Println(string(wasmChain.PendingSendPackets[0].Data), "test")

	// err = coordinator.RelayAndAckPendingPackets(ibcGammPath)
	// if err != nil {
	// 	panic(err)
	// }
	coordinator.CommitBlock(wasmChain, osmosisChain)
	coordinator.UpdateTime()
	coordinator.CommitBlock(wasmChain, osmosisChain)
	coordinator.UpdateTime()

	osmosisApp.GRPCQueryRouter().Route("")

	spotPriceQuery := MakeSpotPriceQueryTx(
		ibcGammPath.EndpointB.ChannelID,
		"1",
		"ibc/1A757F169E3BB799B531736E060340FF68F37CBCEA881A147D83F84F7D87E828",
		"stake",
		"true",
	)

	wasmChain.ExecuteContract(junoContractAddr.String(), spotPriceQuery)

	err = coordinator.RelayAndAckPendingPackets(ibcGammPath)

	if err != nil {
		panic(err)
	}

	fmt.Println(osmosisApp.BankKeeper.GetAllBalances(osmosisChain.GetContext(), osmosisChain.SenderAccount.GetAddress()))
	fmt.Println(osmosisApp.BankKeeper.GetAllBalances(osmosisChain.GetContext(), osmosisContractAddr))

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
