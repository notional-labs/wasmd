package backend

import (
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/wasi"
)

const (
	maxWASMPages = 65536
	wasmPageSize = 65536
	cost         = 150_000
)

func createWazeroInstance(code []byte, memoryLimit uint64, cost uint64) (*wazero.Instance, error) {
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

	// Apply customized metering cost
	// Note: This is an example; you should modify the code or implement your own metering logic.

	// Instantiate
	instance, err := vm.Instantiate(&wasi.PreopenedFiles{})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func limitToPages(limit uint64) uint32 {
	limitInPages := limit / wasmPageSize
	if limitInPages > maxWASMPages {
		return maxWASMPages
	}
	return uint32(limitInPages)
}
