package keeper

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/types"
)

func TestEncoding(t *testing.T) {
	var (
		addr1 = sdk.AccAddress([]byte{0, 1, 0, 1})
		addr2 = sdk.AccAddress([]byte{0, 1, 0, 1, 0, 2, 0, 2})
	)
	valAddr := make(sdk.ValAddress, types.SDKAddrLen)
	valAddr[0] = 12
	valAddr2 := make(sdk.ValAddress, types.SDKAddrLen)
	valAddr2[1] = 123

	bankMsg := &banktypes.MsgSend{
		FromAddress: addr2.String(),
		ToAddress:   addr1.String(),
		Amount: sdk.Coins{
			sdk.NewInt64Coin("uatom", 12345),
			sdk.NewInt64Coin("utgd", 54321),
		},
	}
	bankMsgBin, err := proto.Marshal(bankMsg)
	require.NoError(t, err)

	content, err := codectypes.NewAnyWithValue(types.StoreCodeProposalFixture())
	require.NoError(t, err)

	proposalMsg := &govtypes.MsgSubmitProposal{
		Proposer:       addr1.String(),
		InitialDeposit: sdk.NewCoins(sdk.NewInt64Coin("uatom", 12345)),
		Content:        content,
	}
	proposalMsgBin, err := proto.Marshal(proposalMsg)
	require.NoError(t, err)

	cases := map[string]struct {
		sender             sdk.AccAddress
		srcMsg             wasmvmtypes.CosmosMsg
		srcContractIBCPort string
		transferPortSource types.ICS20TransferPortSource
		// set if valid
		output []sdk.Msg
		// set if invalid
		isError bool
	}{
		"stargate encoded bank msg": {
			sender: addr2,
			srcMsg: wasmvmtypes.CosmosMsg{
				Stargate: &wasmvmtypes.StargateMsg{
					TypeURL: "/cosmos.bank.v1beta1.MsgSend",
					Value:   bankMsgBin,
				},
			},
			output: []sdk.Msg{bankMsg},
		},
		"stargate encoded msg with any type": {
			sender: addr2,
			srcMsg: wasmvmtypes.CosmosMsg{
				Stargate: &wasmvmtypes.StargateMsg{
					TypeURL: "/cosmos.gov.v1beta1.MsgSubmitProposal",
					Value:   proposalMsgBin,
				},
			},
			output: []sdk.Msg{proposalMsg},
		},
		"stargate encoded invalid typeUrl": {
			sender: addr2,
			srcMsg: wasmvmtypes.CosmosMsg{
				Stargate: &wasmvmtypes.StargateMsg{
					TypeURL: "/cosmos.bank.v2.MsgSend",
					Value:   bankMsgBin,
				},
			},
			isError: true,
		},
	}
	encodingConfig := MakeEncodingConfig(t)
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var ctx sdk.Context
			encoder := DefaultEncoders(encodingConfig.Marshaler, tc.transferPortSource)
			res, err := encoder.Encode(ctx, tc.sender, tc.srcContractIBCPort, tc.srcMsg)
			if tc.isError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.output, res)
			}
		})
	}
}

func TestConvertWasmCoinToSdkCoin(t *testing.T) {
	specs := map[string]struct {
		src    wasmvmtypes.Coin
		expErr bool
		expVal sdk.Coin
	}{
		"all good": {
			src: wasmvmtypes.Coin{
				Denom:  "foo",
				Amount: "1",
			},
			expVal: sdk.NewCoin("foo", sdk.NewIntFromUint64(1)),
		},
		"negative amount": {
			src: wasmvmtypes.Coin{
				Denom:  "foo",
				Amount: "-1",
			},
			expErr: true,
		},
		"denom to short": {
			src: wasmvmtypes.Coin{
				Denom:  "f",
				Amount: "1",
			},
			expErr: true,
		},
		"invalid demum char": {
			src: wasmvmtypes.Coin{
				Denom:  "&fff",
				Amount: "1",
			},
			expErr: true,
		},
		"not a number amount": {
			src: wasmvmtypes.Coin{
				Denom:  "foo",
				Amount: "bar",
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotVal, gotErr := convertWasmCoinToSdkCoin(spec.src)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expVal, gotVal)
		})
	}

}
