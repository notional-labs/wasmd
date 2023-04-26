package backend

// LimitingTunables is a custom struct that allows you to set a memory limit.
// After adjusting the memory limits, it delegates all other logic
// to the base tunables.
type LimitingTunables struct {
	// Limit is the maximum allowed size for a linear memory (in Wasm pages, 65 KiB each).
	// Since wazero ensures there is only none or one memory, this is practically
	// an upper limit for the guest memory.
	Limit uint32
}

// NewLimitingTunables creates a new LimitingTunables instance.
func NewLimitingTunables(limit uint32) *LimitingTunables {
	return &LimitingTunables{Limit: limit}
}

// MemoryType represents the memory type with minimum and maximum limits.
type MemoryType struct {
	Minimum uint32
	Maximum *uint32
}

// MemoryError represents a memory-related error.
type MemoryError struct {
	Message string
}

func (e *MemoryError) Error() string {
	return e.Message
}

// AdjustMemory takes in input memory type as requested by the guest and sets
// a maximum if missing. The resulting memory type is final if
// valid. However, this can produce invalid types, such that
// ValidateMemory must be called before creating the memory.
func (lt *LimitingTunables) AdjustMemory(requested *MemoryType) *MemoryType {
	adjusted := &MemoryType{
		Minimum: requested.Minimum,
		Maximum: requested.Maximum,
	}

	if adjusted.Maximum == nil {
		adjusted.Maximum = new(uint32)
		*adjusted.Maximum = lt.Limit
	}

	return adjusted
}

// ValidateMemory ensures the given memory type does not exceed the memory limit.
// Call this after adjusting the memory.
func (lt *LimitingTunables) ValidateMemory(ty *MemoryType) error {
	if ty.Minimum > lt.Limit {
		return &MemoryError{Message: "Minimum exceeds the allowed memory limit"}
	}

	if ty.Maximum != nil {
		if *ty.Maximum > lt.Limit {
			return &MemoryError{Message: "Maximum exceeds the allowed memory limit"}
		}
	} else {
		return &MemoryError{Message: "Maximum unset"}
	}

	return nil
}

// Tunables is an interface that provides methods to customize the memory and table creation.
type Tunables interface {
	MemoryStyle(memory *MemoryType) MemoryStyle
	TableStyle(table *TableType) TableStyle
	CreateHostMemory(ty *MemoryType, style MemoryStyle) (Memory, error)
	CreateVMMemory(ty *MemoryType, style MemoryStyle) (Memory, error)
	CreateHostTable(ty *TableType, style TableStyle) (Table, error)
	CreateVMTable(ty *TableType, style TableStyle) (Table, error)
}

// MemoryStyle represents the style of a memory.
type MemoryStyle struct{}

// TableStyle represents the style of a table.
type TableStyle struct{}

// Memory is an interface representing a memory object.
type Memory interface{}

// Table is an interface representing a table object.
type Table interface{}

// TableType represents a table type with a minimum and maximum limit.
type TableType struct {
	Minimum uint32
	Maximum *uint32
}

// Implement Tunables interface for LimitingTunables.
func (lt *LimitingTunables) MemoryStyle(memory *MemoryType) MemoryStyle {
	adjusted := lt.AdjustMemory(memory)
	// Replace 'base' with your base tunables implementation.
	return base.MemoryStyle(adjusted)
}

func (lt *LimitingTunables) TableStyle(table *TableType) TableStyle {
	// Replace 'base' with your base tunables implementation.
	return base.TableStyle(table)
}

func (lt *LimitingTunables) CreateHostMemory(ty *MemoryType, style MemoryStyle) (Memory, error) {
	adjusted := lt.AdjustMemory(ty)
	err := lt.ValidateMemory(adjusted)
	if err != nil {
		return nil, err
	}
	// Replace 'base' with your base tunables implementation.
	return base.CreateHostMemory(adjusted, style)
}

func (lt *LimitingTunables) CreateVMMemory(ty *MemoryType, style MemoryStyle) (Memory, error) {
	adjusted := lt.AdjustMemory(ty)
	err := lt.ValidateMemory(adjusted)
	if err != nil {
		return nil, err
	}
	// Replace 'base' with your base tunables implementation.
	return base.CreateVMMemory(adjusted, style)
}

func (lt *LimitingTunables) CreateHostTable(ty *TableType, style TableStyle) (Table, error) {
	// Replace 'base' with your base tunables implementation.
	return base.CreateHostTable(ty, style)
}

func (lt *LimitingTunables) CreateVMTable(ty *TableType, style TableStyle) (Table, error) {
	// Replace 'base' with your base tunables implementation.
	return base.CreateVMTable(ty, style)
}
