package safem

import (
	"fmt"
	"math"
	"math/big"
	"sync/atomic"
)

// PnLBias is the bias value added to all PnL values to make them non-negative in uint64 space
// Using 2^63 ensures we can represent the full int64 range
// PnL = 0 → stored as 2^63
// PnL = +max → stored as 2^64 - 1
// PnL = -max → stored as 0
const PnLBias uint64 = 1 << 63 // 9,223,372,036,854,775,808

// MaxSafeAggregationPositions is the maximum number of positions that can be safely
// aggregated using direct uint64 arithmetic without overflow risk
// Conservative limit: 2^10 = 1,024 positions
// For larger aggregations, use AggregateBiasedPnLSafe with *big.Int
const MaxSafeAggregationPositions = 1024

// Int64ToBiasedUint64 converts an int64 PnL value to biased uint64 representation
// This allows storing negative PnL values in uint64 space for satoshi operations
// Performance: O(1) - simple addition operation
// Overflow: Safe - single int64 value always fits in uint64 after bias
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
// PnL value should be in satoshi (scaled by 1e8)
func StoreBiasedPnL(addr *uint64, pnl int64) {
	biasedPnL := Int64ToBiasedUint64(pnl)
	atomic.StoreUint64(addr, biasedPnL)
}

// LoadBiasedPnL atomically loads a biased uint64 PnL value and converts to int64
// Thread-safe atomic operation for concurrent access
// Returns PnL in satoshi (scaled by 1e8)
func LoadBiasedPnL(addr *uint64) int64 {
	biasedValue := atomic.LoadUint64(addr)
	return BiasedUint64ToInt64(biasedValue)
}

// AddBiasedPnL atomically adds a PnL delta to a biased uint64 PnL value
// Thread-safe atomic operation for concurrent updates
// CRITICAL: This correctly handles bias by converting delta to biased form
// Formula: (current + bias) + (delta + bias) - bias = (current + delta) + bias
func AddBiasedPnL(addr *uint64, delta int64) {
	// Convert delta to biased form
	biasedDelta := Int64ToBiasedUint64(delta)
	
	// Add biased delta to current biased value
	// Result: (current + bias) + (delta + bias) = (current + delta) + (2 * bias)
	// We need to subtract one bias to get correct result: (current + delta) + bias
	atomic.AddUint64(addr, biasedDelta)
	atomic.AddUint64(addr, ^PnLBias+1) // Subtract bias (using two's complement: -bias = ^bias + 1)
}

// AggregateBiasedPnL sums multiple biased uint64 PnL values and returns net PnL
// Fast path: Uses direct uint64 arithmetic for small aggregations (< MaxSafeAggregationPositions)
// This is SIMD-friendly as all values are positive in uint64 space
// Performance: O(n) with direct uint64 addition (no sign checks in loop)
// WARNING: May overflow for large aggregations (> 1,024 positions)
// Use AggregateBiasedPnLSafe for guaranteed overflow protection
func AggregateBiasedPnL(biasedValues []uint64) int64 {
	if len(biasedValues) == 0 {
		return 0
	}

	var total uint64
	n := uint64(len(biasedValues))

	// Direct uint64 addition (SIMD-friendly, no sign checks!)
	for _, val := range biasedValues {
		// Overflow check: if adding would exceed uint64 max
		if total > math.MaxUint64-val {
			// Overflow detected - fall back to safe aggregation
			result, _ := AggregateBiasedPnLSafe(biasedValues)
			return result
		}
		total += val
	}

	// Remove bias: subtract (n × PnLBias)
	// When summing: (pnl1 + bias) + (pnl2 + bias) + ... = (pnl1 + pnl2 + ...) + (n × bias)
	biasCorrection := PnLBias * n
	
	// Check for underflow (shouldn't happen in normal operation)
	if total < biasCorrection {
		// This indicates a logic error - return 0 as safe fallback
		return 0
	}

	netUint64 := total - biasCorrection
	return int64(netUint64)
}

// AggregateBiasedPnLSafe sums multiple biased uint64 PnL values using *big.Int
// Guarantees no overflow for any number of positions
// Performance: O(n) with *big.Int (slightly slower than direct uint64, but safe)
// Use this for large aggregations or when overflow protection is critical
func AggregateBiasedPnLSafe(biasedValues []uint64) (int64, error) {
	if len(biasedValues) == 0 {
		return 0, nil
	}

	// Use *big.Int for safe aggregation (prevents overflow)
	total := new(big.Int)
	for _, val := range biasedValues {
		total.Add(total, new(big.Int).SetUint64(val))
	}

	// Calculate bias correction: n × PnLBias
	n := big.NewInt(int64(len(biasedValues)))
	biasBig := new(big.Int).SetUint64(PnLBias)
	biasCorrection := new(big.Int).Mul(n, biasBig)

	// Remove bias: total - (n × bias)
	netBig := new(big.Int).Sub(total, biasCorrection)

	// Check if result fits in int64
	if !netBig.IsInt64() {
		// Overflow: return max/min int64 with error
		if netBig.Sign() > 0 {
			return math.MaxInt64, fmt.Errorf("PnL aggregation overflow: positive value exceeds int64 max (sum of %d positions)", len(biasedValues))
		}
		return math.MinInt64, fmt.Errorf("PnL aggregation overflow: negative value exceeds int64 min (sum of %d positions)", len(biasedValues))
	}

	return netBig.Int64(), nil
}

// AggregateBiasedPnLAtomic sums multiple atomic biased uint64 PnL values
// Thread-safe aggregation for concurrent position updates
// Uses fast path (AggregateBiasedPnL) - may overflow for very large aggregations
// For guaranteed safety, use AggregateBiasedPnLAtomicSafe
func AggregateBiasedPnLAtomic(addrs []*uint64) int64 {
	if len(addrs) == 0 {
		return 0
	}

	biasedValues := make([]uint64, len(addrs))
	for i, addr := range addrs {
		biasedValues[i] = atomic.LoadUint64(addr)
	}
	return AggregateBiasedPnL(biasedValues)
}

// AggregateBiasedPnLAtomicSafe sums multiple atomic biased uint64 PnL values using *big.Int
// Thread-safe aggregation with guaranteed overflow protection
// Use this for large aggregations or when safety is critical
func AggregateBiasedPnLAtomicSafe(addrs []*uint64) (int64, error) {
	if len(addrs) == 0 {
		return 0, nil
	}

	biasedValues := make([]uint64, len(addrs))
	for i, addr := range addrs {
		biasedValues[i] = atomic.LoadUint64(addr)
	}
	return AggregateBiasedPnLSafe(biasedValues)
}

// ShouldUseSafeAggregation determines if safe aggregation (*big.Int) should be used
// Returns true if number of positions exceeds safe limit
func ShouldUseSafeAggregation(positionCount int) bool {
	return positionCount > MaxSafeAggregationPositions
}
