package ibctesting

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/proto" //nolint
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/CosmWasm/wasmd/x/wasm/types"
)

var wasmIdent = []byte("\x00\x61\x73\x6D")

// SeedNewContractInstance stores some wasm code and instantiates a new contract on this chain.
// This method can be called to prepare the store with some valid CodeInfo and ContractInfo. The returned
// Address is the contract address for this instance. Test should make use of this data and/or use NewIBCContractMockWasmer
// for using a contract mock in Go.
func (chain *TestChain) SeedNewContractInstance() sdk.AccAddress {
	pInstResp := chain.StoreCode(append(wasmIdent, rand.Bytes(10)...))
	codeID := pInstResp.CodeID

	anyAddressStr := chain.SenderAccount.GetAddress().String()
	initMsg := []byte(fmt.Sprintf(`{"verifier": %q, "beneficiary": %q}`, anyAddressStr, anyAddressStr))
	return chain.InstantiateContract(codeID, initMsg)
}

func (chain *TestChain) StoreCodeFile(filename string) types.MsgStoreCodeResponse {
	wasmCode, err := ioutil.ReadFile(filename)
	require.NoError(chain.T, err)
	if strings.HasSuffix(filename, "wasm") { // compress for gas limit
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, err := gz.Write(wasmCode)
		require.NoError(chain.T, err)
		err = gz.Close()
		require.NoError(chain.T, err)
		wasmCode = buf.Bytes()
	}
	return chain.StoreCode(wasmCode)
}

func (chain *TestChain) StoreCode(byteCode []byte) types.MsgStoreCodeResponse {
	storeMsg := &types.MsgStoreCode{
		Sender:       chain.SenderAccount.GetAddress().String(),
		WASMByteCode: byteCode,
	}
	r, err := chain.SendMsgs(storeMsg)
	require.NoError(chain.T, err)
	protoResult := chain.ParseSDKResultData(r)
	require.Len(chain.T, protoResult.Data, 1)
	// unmarshal protobuf response from data
	var pInstResp types.MsgStoreCodeResponse
	require.NoError(chain.T, pInstResp.Unmarshal(protoResult.Data[0].Data))
	require.NotEmpty(chain.T, pInstResp.CodeID)
	return pInstResp
}

func (c *TestChain) InstantiateContract(codeID uint64, msg []byte) sdk.AccAddress {
	instantiateMsg := &types.MsgInstantiateContract{
		Sender: c.SenderAccount.GetAddress().String(),
		Admin:  c.SenderAccount.GetAddress().String(),
		CodeID: codeID,
		Label:  "ibc-test",
		Msg:    msg,
		Funds:  sdk.Coins{TestCoin},
	}

	r, err := c.SendMsgs(instantiateMsg)
	require.NoError(c.T, err)
	protoResult := c.ParseSDKResultData(r)
	require.Len(c.T, protoResult.Data, 1)

	var pExecResp types.MsgInstantiateContractResponse
	require.NoError(c.T, pExecResp.Unmarshal(protoResult.Data[0].Data))
	a, err := sdk.AccAddressFromBech32(pExecResp.Address)
	require.NoError(c.T, err)
	return a
}

// func (c *TestChain) Execute(instantiateMsg types.MsgInstantiateContractResponse)

// SmartQuery This will serialize the query message and submit it to the contract.
// The response is parsed into the provided interface.
// Usage: SmartQuery(addr, QueryMsg{Foo: 1}, &response)
func (chain *TestChain) SmartQuery(contractAddr string, queryMsg []byte) ([]byte, error) {

	req := types.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: queryMsg,
	}
	reqBin, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}

	// TODO: what is the query?
	res := chain.App.Query(abci.RequestQuery{
		Path: "/cosmwasm.wasm.v1.Query/SmartContractState",
		Data: reqBin,
	})

	if res.Code != 0 {
		return nil, fmt.Errorf("query failed: (%d) %s", res.Code, res.Log)
	}

	// unpack protobuf
	var resp types.QuerySmartContractStateResponse
	err = proto.Unmarshal(res.Value, &resp)
	if err != nil {
		return nil, err
	}
	// unpack json content
	return resp.Data, nil
}

func (c *TestChain) ExecuteContract(contractAddr string, msg []byte) []byte {
	executeMsg := &types.MsgExecuteContract{
		Sender:   c.SenderAccount.GetAddress().String(),
		Contract: contractAddr,
		Msg:      msg,
		Funds:    sdk.Coins{TestCoin},
	}

	r, err := c.SendMsgs(executeMsg)
	if err != nil {
		panic(err)
	}
	require.NoError(c.T, err)
	protoResult := c.ParseSDKResultData(r)
	require.Len(c.T, protoResult.Data, 1)

	var pExecResp types.MsgExecuteContractResponse
	require.NoError(c.T, pExecResp.Unmarshal(protoResult.Data[0].Data))
	return pExecResp.Data
}

func (chain *TestChain) ParseSDKResultData(r *sdk.Result) sdk.TxMsgData {
	var protoResult sdk.TxMsgData
	require.NoError(chain.T, proto.Unmarshal(r.Data, &protoResult))
	return protoResult
}
