package backend

import (
	"sync"

	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	// WebAssembly linear memory objects have sizes measured in pages. Each page
	// is 65536 (2^16) bytes. In WebAssembly version 1, a linear memory can have at
	// most 65536 pages, for a total of 2^32 bytes (4 gibibytes).
	// https://github.com/WebAssembly/memory64/blob/master/proposals/memory64/Overview.md
	MaxWasmPages uint32 = 65536

	// Cost per operation
	CostPerOperation uint64 = 150_000
)

type Gatekeeper struct {
	sync.Mutex
}

type LimitingTunables struct {
	wasi.Tunables
	gatekeeper *Gatekeeper
}

func cost() uint64 {
	// A flat fee for each operation
	// The target is 1 Teragas per millisecond (see GAS.md).
	//
	// In https://github.com/CosmWasm/cosmwasm/pull/1042 a profiler is developed to
	// identify runtime differences between different Wasm operation, but this is not yet
	// precise enough to derive insights from it.
	return CostPerOperation
}

func createLimitingTunables(gatekeeper *Gatekeeper) *LimitingTunables {
	return &LimitingTunables{
		Tunables:   wasi.NewDefaultTunables(),
		gatekeeper: gatekeeper,
	}
}

func (t *LimitingTunables) AdjustMemorySize(size uint64) uint64 {
	if size > uint64(MaxWasmPages) {
		return uint64(MaxWasmPages)
	}
	return size
}

func (t *LimitingTunables) OnMemorySizeAdjusted(size uint64) {
	t.gatekeeper.Lock()
	defer t.gatekeeper.Unlock()

	// Code to adjust the memory size
}

func createMeteringMiddleware(costFn func() uint64) *wazero.Middleware {
	middleware := wazero.NewMiddleware()

	middleware.OnInstruction(func(instruction *wazero.Instruction) {
		cost := costFn()
		// Code to apply the cost to the execution
	})

	return middleware
}

func createWazeroInstance(
	code []byte,
	memoryLimit uint64,
	middleware *wazero.Middleware,
) (*wazero.Instance, error) {
	// Create Wazero instance
	config := &wazero.Config{
		EnableWasi: true,
	}
	vm, err := wazero.NewVM(code, config)
	if err != nil {
		return nil, err
	}

	// Set memory limit
	vm.Memory.SetSize(memoryLimit)

	// Apply middleware
	vm.Use(middleware)

	// Instantiate
	instance, err := vm.Instantiate(&wasi.PreopenedFiles{})
	if err != nil {
		return nil, err
	}

	return instance, nil
}
