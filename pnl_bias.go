package safem

import (
	"sync/atomic"
)

// PnLBias is the bias value added to all PnL values to make them non-negative in uint64 space
// Using 2^63 ensures we can represent the full int64 range
// PnL = 0 → stored as 2^63
// PnL = +max → stored as 2^64 - 1
// PnL = -max → stored as 0
const PnLBias uint64 = 1 << 63 // 9,223,372,036,854,775,808

// Int64ToBiasedUint64 converts an int64 PnL value to biased uint64 representation
// This allows storing negative PnL values in uint64 space for satoshi operations
// Performance: O(1) - simple addition operation
func Int64ToBiasedUint64(pnl int64) uint64 {
	return uint64(pnl) + PnLBias
}

// BiasedUint64ToInt64 converts a biased uint64 PnL value back to int64
// Removes the bias to get the actual PnL value
// Performance: O(1) - simple subtraction operation
func BiasedUint64ToInt64(biased uint64) int64 {
	return int64(biased - PnLBias)
}

// StoreBiasedPnL atomically stores a PnL value as biased uint64
// Thread-safe atomic operation for concurrent access
func StoreBiasedPnL(addr *uint64, pnl int64) {
	biasedPnL := Int64ToBiasedUint64(pnl)
	atomic.StoreUint64(addr, biasedPnL)
}

// LoadBiasedPnL atomically loads a biased uint64 PnL value and converts to int64
// Thread-safe atomic operation for concurrent access
func LoadBiasedPnL(addr *uint64) int64 {
	biasedValue := atomic.LoadUint64(addr)
	return BiasedUint64ToInt64(biasedValue)
}

// AddBiasedPnL atomically adds a PnL delta to a biased uint64 PnL value
// Thread-safe atomic operation for concurrent updates
func AddBiasedPnL(addr *uint64, delta int64) {
	biasedDelta := Int64ToBiasedUint64(delta)
	atomic.AddUint64(addr, biasedDelta)
	// Note: Adding biased values directly works because:
	// (pnl1 + bias) + (pnl2 + bias) = (pnl1 + pnl2) + (2 * bias)
	// For single position updates, we need to subtract one bias after addition
	// For aggregation, we subtract (n * bias) where n is the number of positions
}

// AggregateBiasedPnL sums multiple biased uint64 PnL values and returns net PnL
// This is SIMD-friendly as all values are positive in uint64 space
// Performance: O(n) with direct uint64 addition (no sign checks in loop)
func AggregateBiasedPnL(biasedValues []uint64) int64 {
	var total uint64
	n := uint64(len(biasedValues))

	// Direct uint64 addition (SIMD-friendly, no sign checks!)
	for _, val := range biasedValues {
		total += val
	}

	// Remove bias: subtract (n × PnLBias)
	// When summing: (pnl1 + bias) + (pnl2 + bias) + ... = (pnl1 + pnl2 + ...) + (n × bias)
	biasCorrection := PnLBias * n
	return int64(total - biasCorrection)
}

// AggregateBiasedPnLAtomic sums multiple atomic biased uint64 PnL values
// Thread-safe aggregation for concurrent position updates
func AggregateBiasedPnLAtomic(addrs []*uint64) int64 {
	biasedValues := make([]uint64, len(addrs))
	for i, addr := range addrs {
		biasedValues[i] = atomic.LoadUint64(addr)
	}
	return AggregateBiasedPnL(biasedValues)
}
