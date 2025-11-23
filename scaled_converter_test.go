package safem

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

// ============================================================================
// Test BigIntToScaledUint64
// ============================================================================

func TestBigIntToScaledUint64(t *testing.T) {
	tests := []struct {
		name        string
		value       *big.Int
		scale       uint64
		expected    uint64
		expectError bool
		errorType   error
	}{
		{
			name:        "normal conversion",
			value:       big.NewInt(2000), // 2000.0 (unscaled)
			scale:       PriceScale,
			expected:    200000000000, // 2000.0 * 1e8
			expectError: false,
		},
		{
			name:        "zero value",
			value:       big.NewInt(0),
			scale:       PriceScale,
			expected:    0,
			expectError: false,
		},
		{
			name:        "nil value",
			value:       nil,
			scale:       PriceScale,
			expected:    0,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "negative value",
			value:       big.NewInt(-100),
			scale:       PriceScale,
			expected:    0,
			expectError: true,
			errorType:   ErrUnderflow,
		},
		{
			name:        "zero scale",
			value:       big.NewInt(100),
			scale:       0,
			expected:    0,
			expectError: true,
			errorType:   ErrInvalidScale,
		},
		{
			name:        "overflow case",
			value:       new(big.Int).SetUint64(math.MaxUint64),
			scale:       PriceScale,
			expected:    0,
			expectError: true,
			errorType:   ErrOverflow,
		},
		{
			name:        "small value with large scale",
			value:       big.NewInt(1),
			scale:       PriceScale,
			expected:    PriceScale,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BigIntToScaledUint64(tt.value, tt.scale)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorType != nil {
					if !errors.Is(err, tt.errorType) {
						t.Errorf("Expected error type %v, got %v", tt.errorType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

// ============================================================================
// Test ScaledUint64ToBigInt
// ============================================================================

func TestScaledUint64ToBigInt(t *testing.T) {
	tests := []struct {
		name     string
		value    uint64
		scale    uint64
		expected *big.Int
	}{
		{
			name:     "normal conversion",
			value:    200000000000, // 2000.0 * 1e8
			scale:    PriceScale,
			expected: big.NewInt(2000),
		},
		{
			name:     "zero value",
			value:    0,
			scale:    PriceScale,
			expected: big.NewInt(0),
		},
		{
			name:     "zero scale",
			value:    100,
			scale:    0,
			expected: big.NewInt(0),
		},
		{
			name:     "small value",
			value:    PriceScale, // 1.0 * 1e8
			scale:    PriceScale,
			expected: big.NewInt(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScaledUint64ToBigInt(tt.value, tt.scale)

			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// ============================================================================
// Test Float64ToScaledUint64
// ============================================================================

func TestFloat64ToScaledUint64(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		scale       uint64
		maxValue    float64
		expected    uint64
		expectError bool
		errorType   error
	}{
		{
			name:        "normal conversion",
			value:       2000.50,
			scale:       PriceScale,
			maxValue:    MaxSafePrice,
			expected:    200050000000,
			expectError: false,
		},
		{
			name:        "NaN value",
			value:       math.NaN(),
			scale:       PriceScale,
			maxValue:    MaxSafePrice,
			expected:    0,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "positive infinity",
			value:       math.Inf(1),
			scale:       PriceScale,
			maxValue:    MaxSafePrice,
			expected:    0,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "negative infinity",
			value:       math.Inf(-1),
			scale:       PriceScale,
			maxValue:    MaxSafePrice,
			expected:    0,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "negative value",
			value:       -100.0,
			scale:       PriceScale,
			maxValue:    MaxSafePrice,
			expected:    0,
			expectError: true,
			errorType:   ErrUnderflow,
		},
		{
			name:        "out of bounds",
			value:       MaxSafePrice + 1,
			scale:       PriceScale,
			maxValue:    MaxSafePrice,
			expected:    0,
			expectError: true,
			errorType:   ErrOutOfBounds,
		},
		{
			name:        "zero value",
			value:       0.0,
			scale:       PriceScale,
			maxValue:    MaxSafePrice,
			expected:    0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Float64ToScaledUint64(tt.value, tt.scale, tt.maxValue)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorType != nil {
					if !errors.Is(err, tt.errorType) {
						t.Errorf("Expected error type %v, got %v", tt.errorType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

// ============================================================================
// Test Price Key Conversions
// ============================================================================

func TestBigIntToPriceKey(t *testing.T) {
	tests := []struct {
		name        string
		priceBig    *big.Int
		expected    uint64
		expectError bool
	}{
		{
			name:        "normal price",
			priceBig:    big.NewInt(2000000000000000000), // 2000.0 with 18 decimals
			expected:    200000000000,                    // 2000.0 * 1e8
			expectError: false,
		},
		{
			name:        "zero price",
			priceBig:    big.NewInt(0),
			expected:    0,
			expectError: false,
		},
		{
			name:        "nil price",
			priceBig:    nil,
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BigIntToPriceKey(tt.priceBig)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestPriceKeyToBigInt(t *testing.T) {
	tests := []struct {
		name     string
		priceKey uint64
		expected *big.Int
	}{
		{
			name:     "normal price key",
			priceKey: 200000000000, // 2000.0 * 1e8
			expected: big.NewInt(2000),
		},
		{
			name:     "zero price key",
			priceKey: 0,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PriceKeyToBigInt(tt.priceKey)

			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestFloat64ToPriceKey(t *testing.T) {
	tests := []struct {
		name        string
		price       float64
		expected    uint64
		expectError bool
	}{
		{
			name:        "normal price",
			price:       2000.50,
			expected:    200050000000,
			expectError: false,
		},
		{
			name:        "zero price",
			price:       0.0,
			expected:    0,
			expectError: false,
		},
		{
			name:        "out of bounds",
			price:       MaxSafePrice + 1,
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Float64ToPriceKey(tt.price)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestPriceKeyToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		priceKey uint64
		expected float64
	}{
		{
			name:     "normal price key",
			priceKey: 200050000000, // 2000.50 * 1e8
			expected: 2000.50,
		},
		{
			name:     "zero price key",
			priceKey: 0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PriceKeyToFloat64(tt.priceKey)

			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestStringToPriceKey(t *testing.T) {
	tests := []struct {
		name        string
		priceStr    string
		expected    uint64
		expectError bool
	}{
		{
			name:        "valid price string",
			priceStr:    "2000000000000000000", // 2000.0 with 18 decimals
			expected:    200000000000,          // 2000.0 * 1e8
			expectError: false,
		},
		{
			name:        "invalid price string",
			priceStr:    "invalid",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty string",
			priceStr:    "",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StringToPriceKey(tt.priceStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestBigIntToPriceKeyFromSatoshi(t *testing.T) {
	tests := []struct {
		name        string
		priceBig    *big.Int
		expected    uint64
		expectError bool
		errorType   error
	}{
		{
			name:        "normal satoshi price",
			priceBig:    big.NewInt(5000000000000), // 50000.00 in satoshi (already scaled)
			expected:    5000000000000,              // No scaling, used directly
			expectError: false,
		},
		{
			name:        "zero satoshi price",
			priceBig:    big.NewInt(0),
			expected:    0,
			expectError: false,
		},
		{
			name:        "nil price",
			priceBig:    nil,
			expected:    0,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "negative price",
			priceBig:    big.NewInt(-100),
			expected:    0,
			expectError: true,
			errorType:   ErrUnderflow,
		},
		{
			name:        "overflow - exceeds uint64",
			priceBig:    new(big.Int).SetUint64(math.MaxUint64).Add(new(big.Int).SetUint64(math.MaxUint64), big.NewInt(1)),
			expected:    0,
			expectError: true,
			errorType:   ErrOverflow,
		},
		{
			name:        "max uint64 value",
			priceBig:    new(big.Int).SetUint64(math.MaxUint64),
			expected:    math.MaxUint64,
			expectError: false,
		},
		{
			name:        "small satoshi value",
			priceBig:    big.NewInt(1), // 0.00000001 in satoshi
			expected:    1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BigIntToPriceKeyFromSatoshi(tt.priceBig)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorType != nil {
					if !errors.Is(err, tt.errorType) {
						t.Errorf("Expected error type %v, got %v", tt.errorType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestBigIntToQuantityKeyFromSatoshi(t *testing.T) {
	tests := []struct {
		name        string
		quantityBig *big.Int
		expected    uint64
		expectError bool
		errorType   error
	}{
		{
			name:        "normal satoshi quantity",
			quantityBig: big.NewInt(100000000), // 1.0 in satoshi (already scaled)
			expected:    100000000,              // No scaling, used directly
			expectError: false,
		},
		{
			name:        "zero satoshi quantity",
			quantityBig: big.NewInt(0),
			expected:    0,
			expectError: false,
		},
		{
			name:        "nil quantity",
			quantityBig: nil,
			expected:    0,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "negative quantity",
			quantityBig: big.NewInt(-100),
			expected:    0,
			expectError: true,
			errorType:   ErrUnderflow,
		},
		{
			name:        "overflow - exceeds uint64",
			quantityBig: new(big.Int).SetUint64(math.MaxUint64).Add(new(big.Int).SetUint64(math.MaxUint64), big.NewInt(1)),
			expected:    0,
			expectError: true,
			errorType:   ErrOverflow,
		},
		{
			name:        "max uint64 value",
			quantityBig: new(big.Int).SetUint64(math.MaxUint64),
			expected:    math.MaxUint64,
			expectError: false,
		},
		{
			name:        "small satoshi value",
			quantityBig: big.NewInt(1), // 0.00000001 in satoshi
			expected:    1,
			expectError: false,
		},
		{
			name:        "large satoshi quantity",
			quantityBig: big.NewInt(2100000000000000), // 21,000,000 BTC in satoshi
			expected:    2100000000000000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BigIntToQuantityKeyFromSatoshi(tt.quantityBig)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorType != nil {
					if !errors.Is(err, tt.errorType) {
						t.Errorf("Expected error type %v, got %v", tt.errorType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

// ============================================================================
// Test Safe Arithmetic Operations
// ============================================================================

func TestAddPriceKeys(t *testing.T) {
	tests := []struct {
		name        string
		a           uint64
		b           uint64
		expected    uint64
		expectError bool
	}{
		{
			name:        "normal addition",
			a:           100000000000,
			b:           50000000000,
			expected:    150000000000,
			expectError: false,
		},
		{
			name:        "zero addition",
			a:           100000000000,
			b:           0,
			expected:    100000000000,
			expectError: false,
		},
		{
			name:        "overflow",
			a:           math.MaxUint64,
			b:           1,
			expected:    0,
			expectError: true,
		},
		{
			name:        "near overflow",
			a:           math.MaxUint64 - 100,
			b:           50,
			expected:    math.MaxUint64 - 50,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AddPriceKeys(tt.a, tt.b)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestSubtractPriceKeys(t *testing.T) {
	tests := []struct {
		name        string
		a           uint64
		b           uint64
		expected    uint64
		expectError bool
	}{
		{
			name:        "normal subtraction",
			a:           150000000000,
			b:           50000000000,
			expected:    100000000000,
			expectError: false,
		},
		{
			name:        "zero subtraction",
			a:           100000000000,
			b:           0,
			expected:    100000000000,
			expectError: false,
		},
		{
			name:        "underflow",
			a:           100,
			b:           200,
			expected:    0,
			expectError: true,
		},
		{
			name:        "equal values",
			a:           100000000000,
			b:           100000000000,
			expected:    0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SubtractPriceKeys(tt.a, tt.b)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestMultiplyPriceKeys(t *testing.T) {
	tests := []struct {
		name        string
		a           uint64
		b           uint64
		expected    uint64
		expectError bool
	}{
		{
			name:        "normal multiplication",
			a:           200000000000, // 2000.0
			b:           1500000000,   // 15.0
			expected:    30000000000,  // 30000.0 / 1e8
			expectError: false,
		},
		{
			name:        "zero multiplication",
			a:           200000000000,
			b:           0,
			expected:    0,
			expectError: false,
		},
		{
			name:        "one multiplication",
			a:           200000000000,
			b:           100000000, // 1.0
			expected:    200000000000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MultiplyPriceKeys(tt.a, tt.b)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestComparePriceKeys(t *testing.T) {
	tests := []struct {
		name     string
		a        uint64
		b        uint64
		expected int
	}{
		{
			name:     "a less than b",
			a:        100000000000,
			b:        200000000000,
			expected: -1,
		},
		{
			name:     "a greater than b",
			a:        200000000000,
			b:        100000000000,
			expected: 1,
		},
		{
			name:     "a equals b",
			a:        100000000000,
			b:        100000000000,
			expected: 0,
		},
		{
			name:     "zero comparison",
			a:        0,
			b:        100000000000,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparePriceKeys(tt.a, tt.b)

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Test Batch Operations
// ============================================================================

func TestBatchBigIntToPriceKeys(t *testing.T) {
	tests := []struct {
		name        string
		prices      []*big.Int
		expected    []uint64
		expectError bool
	}{
		{
			name: "normal batch conversion",
			prices: []*big.Int{
				big.NewInt(2000000000000000000), // 2000.0
				big.NewInt(1500000000000000000), // 1500.0
				big.NewInt(1000000000000000000), // 1000.0
			},
			expected: []uint64{
				200000000000,
				150000000000,
				100000000000,
			},
			expectError: false,
		},
		{
			name:        "nil slice",
			prices:      nil,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "empty slice",
			prices:      []*big.Int{},
			expected:    []uint64{},
			expectError: false,
		},
		{
			name: "invalid price in batch",
			prices: []*big.Int{
				big.NewInt(2000000000000000000),
				nil, // Invalid
				big.NewInt(1000000000000000000),
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BatchBigIntToPriceKeys(tt.prices)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(result) != len(tt.expected) {
					t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				}
				for i := range result {
					if result[i] != tt.expected[i] {
						t.Errorf("Expected[%d] %d, got %d", i, tt.expected[i], result[i])
					}
				}
			}
		})
	}
}

func TestBatchPriceKeysToBigInt(t *testing.T) {
	tests := []struct {
		name     string
		keys     []uint64
		expected []*big.Int
	}{
		{
			name: "normal batch conversion",
			keys: []uint64{
				200000000000,
				150000000000,
				100000000000,
			},
			expected: []*big.Int{
				big.NewInt(2000),
				big.NewInt(1500),
				big.NewInt(1000),
			},
		},
		{
			name:     "nil slice",
			keys:     nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			keys:     []uint64{},
			expected: []*big.Int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BatchPriceKeysToBigInt(tt.keys)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i].Cmp(tt.expected[i]) != 0 {
					t.Errorf("Expected[%d] %s, got %s", i, tt.expected[i].String(), result[i].String())
				}
			}
		})
	}
}

// ============================================================================
// Test Advanced Operations
// ============================================================================

func TestCalculateMarginRatioKey(t *testing.T) {
	tests := []struct {
		name                 string
		equityKey            uint64
		marginRequirementKey uint64
		expected             uint64
		expectError          bool
		errorType            error
	}{
		{
			name:                 "normal ratio calculation",
			equityKey:            120000000000, // 1200.0 * 1e8 (using value scale for equity)
			marginRequirementKey: 100000000000, // 1000.0 * 1e8
			expected:             120000000,    // 1.2 * 1e8
			expectError:          false,
		},
		{
			name:                 "division by zero",
			equityKey:            120000000000,
			marginRequirementKey: 0,
			expected:             0,
			expectError:          true,
			errorType:            ErrDivisionByZero,
		},
		{
			name:                 "ratio less than 1",
			equityKey:            50000000000,  // 500.0 * 1e8
			marginRequirementKey: 100000000000, // 1000.0 * 1e8
			expected:             50000000,    // 0.5 * 1e8
			expectError:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateMarginRatioKey(tt.equityKey, tt.marginRequirementKey)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorType != nil {
					if !errors.Is(err, tt.errorType) {
						t.Errorf("Expected error type %v, got %v", tt.errorType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestIsLiquidatableKey(t *testing.T) {
	tests := []struct {
		name                    string
		marginRatioKey          uint64
		liquidationThresholdKey uint64
		expected                bool
	}{
		{
			name:                    "liquidatable (ratio below threshold)",
			marginRatioKey:          1000000, // 1.0
			liquidationThresholdKey: 1050000, // 1.05
			expected:                true,
		},
		{
			name:                    "not liquidatable (ratio above threshold)",
			marginRatioKey:          1200000, // 1.2
			liquidationThresholdKey: 1050000, // 1.05
			expected:                false,
		},
		{
			name:                    "at threshold (not liquidatable)",
			marginRatioKey:          1050000, // 1.05
			liquidationThresholdKey: 1050000, // 1.05
			expected:                false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLiquidatableKey(tt.marginRatioKey, tt.liquidationThresholdKey)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Test Round-trip Conversions
// ============================================================================

func TestScaledRoundTripConversion(t *testing.T) {
	tests := []struct {
		name  string
		price *big.Int
	}{
		{
			name:  "normal price",
			price: big.NewInt(2000000000000000000), // 2000.0 with 18 decimals
		},
		{
			name:  "small price",
			price: big.NewInt(1000000000000000), // 0.001 with 18 decimals
		},
		{
			name:  "large price",
			price: new(big.Int).Exp(big.NewInt(10), big.NewInt(23), nil), // 100000.0 with 18 decimals
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to key
			key, err := BigIntToPriceKey(tt.price)
			if err != nil {
				t.Fatalf("Failed to convert to key: %v", err)
			}

			// Convert back to big.Int
			result := PriceKeyToBigInt(key)

			// Note: Due to scaling (18 decimals -> 8 decimals), we lose precision
			// So we compare the scaled values
			originalScaled := new(big.Int).Set(tt.price)
			originalScaled.Div(originalScaled, big.NewInt(1e10)) // Remove 10 decimals (18-8)

			resultScaled := new(big.Int).Set(result)
			resultScaled.Mul(resultScaled, big.NewInt(1e8)) // Add 8 decimals

			// Compare within 1e10 tolerance (due to precision loss)
			diff := new(big.Int).Sub(originalScaled, resultScaled)
			diff.Abs(diff)
			tolerance := big.NewInt(1e10)

			if diff.Cmp(tolerance) > 0 {
				t.Errorf("Round-trip conversion failed: original %s, result %s, diff %s",
					tt.price.String(), result.String(), diff.String())
			}
		})
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkBigIntToPriceKey(b *testing.B) {
	priceBig := big.NewInt(2000000000000000000) // 2000.0 with 18 decimals

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BigIntToPriceKey(priceBig)
	}
}

func BenchmarkPriceKeyToBigInt(b *testing.B) {
	priceKey := uint64(200000000000) // 2000.0 * 1e8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PriceKeyToBigInt(priceKey)
	}
}

func BenchmarkFloat64ToPriceKey(b *testing.B) {
	price := 2000.50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Float64ToPriceKey(price)
	}
}

func BenchmarkAddPriceKeys(b *testing.B) {
	a := uint64(100000000000)
	bVal := uint64(50000000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddPriceKeys(a, bVal)
	}
}

func BenchmarkComparePriceKeys(b *testing.B) {
	a := uint64(200000000000)
	bVal := uint64(100000000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComparePriceKeys(a, bVal)
	}
}

func BenchmarkBatchBigIntToPriceKeys(b *testing.B) {
	prices := make([]*big.Int, 1000)
	for i := range prices {
		prices[i] = big.NewInt(int64(2000000000000000000 + i*1000000000000000))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BatchBigIntToPriceKeys(prices)
	}
}
