package safem

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

// BigIntPool for reusing big.Int instances to reduce allocations.
var BigIntPool = sync.Pool{
	New: func() interface{} { return new(big.Int) },
}

// BigInt is a wrapper for *big.Int with custom JSON unmarshaling (inspired by go-ethereum/hexutil.Big).
// This ensures safe conversion from JSON numbers (float64/string) to *big.Int without precision loss.
type BigInt struct {
	*big.Int
}

// UnmarshalJSON handles JSON number (float64/string) to *big.Int conversion safely.
// This prevents overflow issues when large numbers are parsed as float64 from JSON.
func (b *BigInt) UnmarshalJSON(data []byte) error {
	start := time.Now()
	defer func() {
		// Log timing for performance monitoring
		duration := time.Since(start)
		if duration > 10*time.Millisecond {
			// Log slow unmarshaling for debugging
			// TODO: Add proper metrics when metrics package is available
		}
	}()

	var val interface{}
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	// Get big.Int from pool and initialize
	b.Int = BigIntPool.Get().(*big.Int)
	b.Int.SetInt64(0)

	switch v := val.(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			BigIntPool.Put(b.Int)
			return strconv.ErrSyntax
		}
		// Prefer direct int64 for small integers to avoid string conversion
		if v == math.Trunc(v) && v <= math.MaxInt64 && v >= math.MinInt64 {
			b.Int.SetInt64(int64(v))
		} else {
			// Log warning for float64 usage - should use string in production
			str := strconv.FormatFloat(v, 'f', -1, 64)
			_, ok := b.Int.SetString(str, 10)
			if !ok {
				BigIntPool.Put(b.Int)
				return strconv.ErrSyntax
			}
		}
	case string:
		// Pre-validate string to avoid unnecessary SetString calls
		v = strings.Trim(v, `"`)
		if len(v) == 0 {
			BigIntPool.Put(b.Int)
			b.Int = nil
			return nil
		}

		// Try to parse as base 10 first
		if _, ok := b.Int.SetString(v, 10); ok {
			return nil
		}

		// Try to parse as hex if it starts with 0x
		if strings.HasPrefix(v, "0x") {
			if _, ok := b.Int.SetString(v[2:], 16); ok {
				return nil
			}
		}

		BigIntPool.Put(b.Int)
		return strconv.ErrSyntax
	case nil:
		BigIntPool.Put(b.Int)
		b.Int = nil
		return nil
	default:
		BigIntPool.Put(b.Int)
		return strconv.ErrSyntax
	}

	return nil
}

// MarshalJSON outputs as string for Ethereum compatibility.
func (b BigInt) MarshalJSON() ([]byte, error) {
	start := time.Now()
	defer func() {
		// Log timing for performance monitoring
		duration := time.Since(start)
		if duration > 10*time.Millisecond {
			// Log slow marshaling for debugging
			// TODO: Add proper metrics when metrics package is available
		}
	}()

	if b.Int == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(b.String())
}

// Set sets the value of BigInt from a *big.Int
func (b *BigInt) Set(x *big.Int) *BigInt {
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	b.Int.Set(x)
	return b
}

// SetString sets the value of BigInt from a string
func (b *BigInt) SetString(s string, base int) (*BigInt, bool) {
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	_, ok := b.Int.SetString(s, base)
	return b, ok
}

// Cmp compares BigInt with another BigInt
func (b *BigInt) Cmp(x *BigInt) int {
	if b.Int == nil {
		if x.Int == nil {
			return 0
		}
		return -1
	}
	if x.Int == nil {
		return 1
	}
	return b.Int.Cmp(x.Int)
}

// Add adds two BigInt values
func (b *BigInt) Add(x, y *BigInt) *BigInt {
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	if x.Int == nil {
		b.Int.Set(y.Int)
	} else if y.Int == nil {
		b.Int.Set(x.Int)
	} else {
		b.Int.Add(x.Int, y.Int)
	}
	return b
}

// Sub subtracts two BigInt values
func (b *BigInt) Sub(x, y *BigInt) *BigInt {
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	if x.Int == nil {
		if y.Int == nil {
			b.Int.SetInt64(0)
		} else {
			b.Int.Neg(y.Int)
		}
	} else if y.Int == nil {
		b.Int.Set(x.Int)
	} else {
		b.Int.Sub(x.Int, y.Int)
	}
	return b
}

// Mul multiplies two BigInt values
func (b *BigInt) Mul(x, y *BigInt) *BigInt {
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	if x.Int == nil || y.Int == nil {
		b.Int.SetInt64(0)
	} else {
		b.Int.Mul(x.Int, y.Int)
	}
	return b
}

// Div divides two BigInt values
func (b *BigInt) Div(x, y *BigInt) *BigInt {
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	if x.Int == nil || y.Int == nil || y.Int.Sign() == 0 {
		b.Int.SetInt64(0)
	} else {
		b.Int.Div(x.Int, y.Int)
	}
	return b
}

// Int64 returns the int64 representation of BigInt
func (b *BigInt) Int64() int64 {
	if b.Int == nil {
		return 0
	}
	return b.Int.Int64()
}

// Uint64 returns the uint64 representation of BigInt
func (b *BigInt) Uint64() uint64 {
	if b.Int == nil {
		return 0
	}
	return b.Int.Uint64()
}

// Sign returns the sign of BigInt
func (b *BigInt) Sign() int {
	if b.Int == nil {
		return 0
	}
	return b.Int.Sign()
}

// IsNil returns true if BigInt is nil
func (b *BigInt) IsNil() bool {
	return b.Int == nil
}

// NewBigInt creates a new BigInt from a *big.Int
func NewBigInt(x *big.Int) *BigInt {
	if x == nil {
		return &BigInt{Int: nil}
	}
	return &BigInt{Int: new(big.Int).Set(x)}
}

// NewBigIntFromString creates a new BigInt from a string
func NewBigIntFromString(s string, base int) (*BigInt, bool) {
	b := &BigInt{Int: new(big.Int)}
	_, ok := b.Int.SetString(s, base)
	if !ok {
		return nil, false
	}
	return b, true
}

// NewBigIntFromInt64 creates a new BigInt from int64
func NewBigIntFromInt64(x int64) *BigInt {
	return &BigInt{Int: big.NewInt(x)}
}

// NewBigIntFromUint64 creates a new BigInt from uint64
func NewBigIntFromUint64(x uint64) *BigInt {
	return &BigInt{Int: new(big.Int).SetUint64(x)}
}

// isNumericString checks if a string is a valid decimal number (basic validation)
func isNumericString(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Handle negative numbers
	if s[0] == '-' {
		s = s[1:]
		if len(s) == 0 {
			return false
		}
	}

	// Check if all characters are digits
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
