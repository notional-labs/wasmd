package app

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"cosmossdk.io/x/feegrant"

	"github.com/spf13/viper"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm"
)

// SimAppChainID hardcoded chainID for simulation
const SimAppChainID = "simulation-app"

var FlagEnableStreamingValue bool

// Get flags every time the simulator is run
func init() {
	simcli.GetSimulatorFlags()
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestFullAppSimulation(t *testing.T) {
	config, db, _, app := setupSimulationApp(t, "skipping application simulation")
	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		BlockedAddresses(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err := simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

func TestAppImportExport(t *testing.T) {
	config, db, appOptions, app := setupSimulationApp(t, "skipping application import/export simulation")

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		BlockedAddresses(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err := simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	t.Log("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err)

	t.Log("importing genesis...\n")

	newDB, newDir, _, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim-2", "Simulation-2", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, newDB.Close())
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewWasmApp(log.NewNopLogger(), newDB, nil, true, wasm.EnableAllProposals, appOptions, emptyWasmOpts, fauxMerkleModeOpt, baseapp.SetChainID(SimAppChainID))
	require.Equal(t, "WasmApp", newApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Sprintf("%v", r)
			if !strings.Contains(err, "validator set is empty after InitGenesis") {
				panic(r)
			}
			t.Log("Skipping simulation as all validators have been unbonded")
			t.Logf("err: %s stacktrace: %s\n", err, string(debug.Stack()))
		}
	}()

	ctxA := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})
	_, err = newApp.ModuleManager.InitGenesis(ctxB, app.AppCodec(), genesisState)
	require.NoError(t, err)
	err = newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)
	require.NoError(t, err)

	t.Log("comparing stores...")
	// skip certain prefixes
	skipPrefixes := map[string][][]byte{
		stakingtypes.StoreKey: {
			stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
			stakingtypes.HistoricalInfoKey, stakingtypes.UnbondingIDKey, stakingtypes.UnbondingIndexKey,
			stakingtypes.UnbondingTypeKey, stakingtypes.ValidatorUpdatesKey,
		},
		authzkeeper.StoreKey:   {authzkeeper.GrantQueuePrefix},
		feegrant.StoreKey:      {feegrant.FeeAllowanceQueueKeyPrefix},
		slashingtypes.StoreKey: {slashingtypes.ValidatorMissedBlockBitmapKeyPrefix},
	}

	storeKeys := app.GetStoreKeys()
	require.NotEmpty(t, storeKeys)

	for _, appKeyA := range storeKeys {
		// only compare kvstores
		if _, ok := appKeyA.(*storetypes.KVStoreKey); !ok {
			continue
		}

		keyName := appKeyA.Name()
		appKeyB := newApp.GetKey(keyName)

		storeA := ctxA.KVStore(appKeyA)
		storeB := ctxB.KVStore(appKeyB)

		failedKVAs, failedKVBs := simtestutil.DiffKVStores(storeA, storeB, skipPrefixes[keyName])
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare %s", keyName)

		t.Logf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), appKeyA, appKeyB)
		require.Equal(t, 0, len(failedKVAs), simtestutil.GetSimulationLog(keyName, app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, appOptions, app := setupSimulationApp(t, "skipping application simulation after import")

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		BlockedAddresses(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err := simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(true, []string{}, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	newDB, newDir, _, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim-2", "Simulation-2", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, newDB.Close())
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewWasmApp(log.NewNopLogger(), newDB, nil, true, wasm.EnableAllProposals, appOptions, emptyWasmOpts, fauxMerkleModeOpt, baseapp.SetChainID(SimAppChainID))
	require.Equal(t, "WasmApp", newApp.Name())

	_, err = newApp.InitChain(&abci.RequestInitChain{
		ChainId:       SimAppChainID,
		AppStateBytes: exported.AppState,
	})
	require.NoError(t, err)

	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		newApp.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(newApp, newApp.AppCodec(), config),
		BlockedAddresses(),
		config,
		app.AppCodec(),
	)
	require.NoError(t, err)
}

func setupSimulationApp(t *testing.T, msg string) (simtypes.Config, dbm.DB, simtestutil.AppOptionsMap, *WasmApp) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if skip {
		t.Skip(msg)
	}
	require.NoError(t, err, "simulation setup failed")

	t.Cleanup(func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	})

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = dir // ensure a unique folder
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	app := NewWasmApp(logger, db, nil, true, wasm.EnableAllProposals, appOptions, emptyWasmOpts, fauxMerkleModeOpt, baseapp.SetChainID(SimAppChainID))
	require.Equal(t, "WasmApp", app.Name())
	return config, db, appOptions, app
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	appOptions := viper.New()
	appOptions.SetDefault(flags.FlagHome, t.TempDir()) // ensure a unique folder
	appOptions.SetDefault(server.FlagInvCheckPeriod, simcli.FlagPeriodValue)

	if FlagEnableStreamingValue {
		appOptions.SetDefault("streaming.abci.keys", []string{"*"})
		appOptions.SetDefault("streaming.abci.plugin", "abci_v1")
		appOptions.SetDefault("streaming.abci.stop-node-on-err", true)
	}

	for i := 0; i < numSeeds; i++ {
		config.Seed += int64(i)
		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simcli.FlagVerboseValue {
				logger = log.NewTestLogger(t)
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			app := NewWasmApp(logger, db, nil, true, wasm.EnableAllProposals, appOptions, emptyWasmOpts, interBlockCacheOpt(), baseapp.SetChainID(SimAppChainID))

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simtestutil.SimulationOperations(app, app.AppCodec(), config),
				BlockedAddresses(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simtestutil.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
