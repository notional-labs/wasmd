package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/CosmWasm/wasmd/app/params"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v2/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v2/modules/core/keeper"

	"github.com/CosmWasm/wasmd/x/wasm"
)

// type TestSupport struct {
// 	t   testing.TB
// 	app *WasmApp
// }

func (app *WasmApp) IBCKeeper() *ibckeeper.Keeper {
	return app.ibcKeeper
}

func (app *WasmApp) GetWasmKeeper() wasm.Keeper {
	return app.wasmKeeper
}

func (app *WasmApp) ScopedWasmIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.scopedWasmKeeper
}

func (app *WasmApp) ScopeIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.scopedIBCKeeper
}

func (app *WasmApp) ScopedTransferKeeper() capabilitykeeper.ScopedKeeper {
	return app.scopedTransferKeeper
}

func (app *WasmApp) StakingKeeper() stakingkeeper.Keeper {
	return app.stakingKeeper
}

func (app *WasmApp) BankKeeper() bankkeeper.Keeper {
	return app.bankKeeper
}

func (app *WasmApp) TransferKeeper() ibctransferkeeper.Keeper {
	return app.transferKeeper
}

func (app *WasmApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

func (app *WasmApp) GetTxConfig() client.TxConfig {
	return params.MakeEncodingConfig().TxConfig
}

func (app *WasmApp) GetStakingKeeper() stakingkeeper.Keeper {
	return app.StakingKeeper()
}

func (app *WasmApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper()
}

func (app *WasmApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopeIBCKeeper()
}
