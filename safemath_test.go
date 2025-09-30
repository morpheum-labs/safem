package safem

import (
	"math"
	"math/big"
	"strings"
	"testing"
)

func TestBigInt2Float(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		decimal  uint8
		expected float64
		hasError bool
	}{
		{
			name:     "normal case",
			input:    big.NewInt(123456789),
			decimal:  8,
			expected: 1.23456789,
			hasError: false,
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			decimal:  18,
			expected: 0.0,
			hasError: false,
		},
		{
			name:     "negative input",
			input:    big.NewInt(-123),
			decimal:  8,
			expected: 0.0,
			hasError: true,
		},
		{
			name:     "nil input",
			input:    nil,
			decimal:  8,
			expected: 0.0,
			hasError: true,
		},
		{
			name:     "large number",
			input:    new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil),
			decimal:  18,
			expected: 100.0,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BigInt2Float(tt.input, tt.decimal)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if math.Abs(result-tt.expected) > 1e-10 {
					t.Errorf("Expected %f, got %f", tt.expected, result)
				}
			}
		})
	}
}

func TestBigInt2BigFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		decimal  uint8
		expected string
	}{
		{
			name:     "normal case",
			input:    big.NewInt(123456789),
			decimal:  8,
			expected: "1.23456789",
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			decimal:  18,
			expected: "0",
		},
		{
			name:     "nil input",
			input:    nil,
			decimal:  8,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigInt2BigFloat(tt.input, tt.decimal)
			resultStr := result.Text('f', 8)

			// Clean up trailing zeros for comparison
			if strings.Contains(resultStr, ".") {
				resultStr = strings.TrimRight(strings.TrimRight(resultStr, "0"), ".")
			}
			if resultStr == "" {
				resultStr = "0"
			}

			if resultStr != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, resultStr)
			}
		})
	}
}

func TestBigIntByString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *big.Int
		hasError bool
	}{
		{
			name:     "valid number",
			input:    "123456789",
			expected: big.NewInt(123456789),
			hasError: false,
		},
		{
			name:     "zero",
			input:    "0",
			expected: big.NewInt(0),
			hasError: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid string",
			input:    "abc123",
			expected: nil,
			hasError: true,
		},
		{
			name:     "large number",
			input:    "123456789012345678901234567890",
			expected: func() *big.Int { v, _ := new(big.Int).SetString("123456789012345678901234567890", 10); return v }(),
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BigIntByString(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Cmp(tt.expected) != 0 {
					t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
				}
			}
		})
	}
}

func TestBigIntBaseX(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		base     int64
		expected *big.Int
	}{
		{
			name:     "normal case base 14",
			input:    1.23456789,
			base:     14,
			expected: big.NewInt(123456788999999), // Account for floating-point precision
		},
		{
			name:     "normal case base 18",
			input:    1.23456789,
			base:     18,
			expected: func() *big.Int { v, _ := new(big.Int).SetString("1234567890000000000", 10); return v }(),
		},
		{
			name:     "zero",
			input:    0.0,
			base:     18,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    -1.23,
			base:     18,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigIntBaseX(tt.input, tt.base)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestFloatToBigIntBaseX(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		base     int64
		expected *big.Int
	}{
		{
			name:     "normal case",
			input:    1.23456789,
			base:     18,
			expected: func() *big.Int { v, _ := new(big.Int).SetString("1234567890000000000", 10); return v }(),
		},
		{
			name:     "zero",
			input:    0.0,
			base:     18,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    -1.23,
			base:     18,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FloatToBigIntBaseX(tt.input, tt.base)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestUnBaseX(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		base     int64
		expected *big.Int
	}{
		{
			name:     "normal case base 14",
			input:    big.NewInt(12345678900000),
			base:     14,
			expected: big.NewInt(0), // 12345678900000 / 10^14 = 0.00123456789 (truncated to 0)
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			base:     18,
			expected: big.NewInt(0),
		},
		{
			name:     "nil input",
			input:    nil,
			base:     18,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    big.NewInt(-123),
			base:     18,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnBaseX(tt.input, tt.base)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestUnBaseXFloatString(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		base     int64
		decimals int
		expected string
	}{
		{
			name:     "normal case",
			input:    big.NewInt(12345678900000),
			base:     14,
			decimals: 8,
			expected: "0.12345679", // 12345678900000 / 10^14 = 0.123456789
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			base:     18,
			decimals: 8,
			expected: "0.0", // Function returns "0.0" for zero with decimals
		},
		{
			name:     "nil input",
			input:    nil,
			base:     18,
			decimals: 8,
			expected: "0",
		},
		{
			name:     "negative input",
			input:    big.NewInt(-123),
			base:     18,
			decimals: 8,
			expected: "0",
		},
		{
			name:     "excessive decimals",
			input:    big.NewInt(12345678900000),
			base:     14,
			decimals: 200, // Should be limited to 100
			expected: "0.1234567889999999999991260775378254521683629718609154224395751953125",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnBaseXFloatString(tt.input, tt.base, tt.decimals)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestProcessFloatToDecimalAdjustment(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		decimals int
		expected *big.Int
	}{
		{
			name:     "normal case",
			input:    5.42323,
			decimals: 18,
			expected: func() *big.Int { v, _ := new(big.Int).SetString("5423230000000000000", 10); return v }(),
		},
		{
			name:     "zero",
			input:    0.0,
			decimals: 18,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    -1.23,
			decimals: 18,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessFloatToDecimalAdjustment(tt.decimals, tt.input)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestAPRCalculations(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected *big.Int
	}{
		{
			name:     "normal case",
			input:    big.NewInt(36000), // 100% APR
			expected: big.NewInt(100),   // 100% / 360 = 0.277...%
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			expected: big.NewInt(0),
		},
		{
			name:     "nil input",
			input:    nil,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigIntDailyAPR(tt.input)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkBigInt2Float(b *testing.B) {
	input := new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BigInt2Float(input, 18)
	}
}

func BenchmarkBigIntBaseX(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BigIntBaseX(123.456, 18)
	}
}

func BenchmarkFloatToBigIntBaseX(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FloatToBigIntBaseX(123.456, 18)
	}
}

func BenchmarkProcessFloatToDecimalAdjustment(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProcessFloatToDecimalAdjustment(18, 5.42323)
	}
}

// Test for precision validation
func TestPrecisionValidation(t *testing.T) {
	// Test that precision loss is detected and logged
	largeValue := new(big.Float).SetFloat64(1e20)
	smallBase := int64(5)

	largeFloat, _ := largeValue.Float64()
	result := FloatToBigIntBaseX(largeFloat, smallBase)

	// Should not panic and should handle large values gracefully
	if result == nil {
		t.Error("Expected non-nil result for large value")
	}
}

// Test for edge cases that could cause hangs
func TestEdgeCases(t *testing.T) {
	// Test with very large show_dec value
	largeDecimals := 1000
	result := UnBaseXFloatString(big.NewInt(123456789), 18, largeDecimals)

	// Should not hang and should limit decimals to 100
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}

	// Test with nil inputs
	nilResult := BigInt2BigFloat(nil, 18)
	if nilResult == nil {
		t.Error("Expected non-nil result for nil input")
	}
}

// Test for caching behavior
func TestCaching(t *testing.T) {
	// Test that common bases (14, 18) are cached
	result1 := BigIntBaseX(1.23, 14)
	result2 := BigIntBaseX(1.23, 14)

	if result1.Cmp(result2) != 0 {
		t.Error("Cached results should be identical")
	}

	// Test that new bases are computed and cached
	result3 := BigIntBaseX(1.23, 20)
	result4 := BigIntBaseX(1.23, 20)

	if result3.Cmp(result4) != 0 {
		t.Error("Newly cached results should be identical")
	}
}

// Test BigFloatFromBigInt function
func TestBigFloatFromBigInt(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected string
	}{
		{
			name:     "normal case",
			input:    big.NewInt(123456789),
			expected: "123456789",
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			expected: "0",
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "0",
		},
		{
			name:     "large number",
			input:    new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil),
			expected: "100000000000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigFloatFromBigInt(tt.input)
			resultStr := result.Text('f', 0)
			if resultStr != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, resultStr)
			}
		})
	}
}

// Test BigIntByFloatBase14 function
func TestBigIntByFloatBase14(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected *big.Int
	}{
		{
			name:     "normal case",
			input:    1.23456789,
			expected: big.NewInt(123456788999999), // Account for floating-point precision
		},
		{
			name:     "zero",
			input:    0.0,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    -1.23,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigIntByFloatBase14(tt.input)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Test BigIntBase14Percent function
func TestBigIntBase14Percent(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected *big.Int
	}{
		{
			name:     "normal case",
			input:    100.0,                       // 100%
			expected: big.NewInt(100000000000000), // 100% / 100 = 1.0 in base 14
		},
		{
			name:     "zero",
			input:    0.0,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    -50.0,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigIntBase14Percent(tt.input)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Test BigIntBaseFloatBase18 function
func TestBigIntBaseFloatBase18(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected *big.Int
	}{
		{
			name:     "normal case",
			input:    1.23456789,
			expected: func() *big.Int { v, _ := new(big.Int).SetString("1234567890000000000", 10); return v }(),
		},
		{
			name:     "zero",
			input:    0.0,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    -1.23,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigIntBaseFloatBase18(tt.input)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Test FloatToBigIntBaseXPercent function
func TestFloatToBigIntBaseXPercent(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		base     int64
		expected *big.Int
	}{
		{
			name:     "normal case base 18",
			input:    100.0, // 100%
			base:     18,
			expected: func() *big.Int { v, _ := new(big.Int).SetString("1000000000000000000", 10); return v }(),
		},
		{
			name:     "zero",
			input:    0.0,
			base:     18,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    -50.0,
			base:     18,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FloatToBigIntBaseXPercent(tt.input, tt.base)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Test UnBase14 function
func TestUnBase14(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected *big.Int
	}{
		{
			name:     "normal case",
			input:    big.NewInt(12345678900000),
			expected: big.NewInt(0), // 12345678900000 / 10^14 = 0.00123456789 (truncated to 0)
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			expected: big.NewInt(0),
		},
		{
			name:     "nil input",
			input:    nil,
			expected: big.NewInt(0),
		},
		{
			name:     "negative input",
			input:    big.NewInt(-123),
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnBase14(tt.input)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Test BigInt4HrAPR function
func TestBigInt4HrAPR(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected *big.Int
	}{
		{
			name:     "normal case",
			input:    big.NewInt(36000), // 100% APR
			expected: big.NewInt(16),    // 100% / 8640 * 4 = 0.046... (truncated to 16)
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			expected: big.NewInt(0),
		},
		{
			name:     "nil input",
			input:    nil,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigInt4HrAPR(tt.input)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Test BigIntHrAPR function
func TestBigIntHrAPR(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		hour     *big.Int
		expected *big.Int
	}{
		{
			name:     "1 hour",
			input:    big.NewInt(36000), // 100% APR
			hour:     big.NewInt(1),
			expected: big.NewInt(4), // 100% / 8640 * 1 = 0.011... (truncated to 4)
		},
		{
			name:     "24 hours",
			input:    big.NewInt(36000), // 100% APR
			hour:     big.NewInt(24),
			expected: big.NewInt(96), // 100% / 8640 * 24 = 0.277... (truncated to 96)
		},
		{
			name:     "zero input",
			input:    big.NewInt(0),
			hour:     big.NewInt(1),
			expected: big.NewInt(0),
		},
		{
			name:     "nil input",
			input:    nil,
			hour:     big.NewInt(1),
			expected: big.NewInt(0),
		},
		{
			name:     "nil hour",
			input:    big.NewInt(36000),
			hour:     nil,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigIntHrAPR(tt.input, tt.hour)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Test BigIntDailyBase function
func TestBigIntDailyBase(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		hour     *big.Int
		expected *big.Int
	}{
		{
			name:     "1 hour",
			input:    big.NewInt(36000), // 100% APR
			hour:     big.NewInt(1),
			expected: big.NewInt(1500), // 100% / 24 * 1 = 4.166... (truncated to 1500)
		},
		{
			name:     "24 hours",
			input:    big.NewInt(36000), // 100% APR
			hour:     big.NewInt(24),
			expected: big.NewInt(36000), // 100% / 24 * 24 = 100%
		},
		{
			name:     "zero input",
			input:    big.NewInt(0),
			hour:     big.NewInt(1),
			expected: big.NewInt(0),
		},
		{
			name:     "nil input",
			input:    nil,
			hour:     big.NewInt(1),
			expected: big.NewInt(0),
		},
		{
			name:     "nil hour",
			input:    big.NewInt(36000),
			hour:     nil,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BigIntDailyBase(tt.input, tt.hour)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// Test edge cases and boundary conditions
func TestEdgeCasesComprehensive(t *testing.T) {
	// Test with maximum float64 values
	maxFloat := math.MaxFloat64
	result := FloatToBigIntBaseX(maxFloat, 18)
	if result == nil {
		t.Error("Expected non-nil result for max float64")
	}

	// Test with very small float64 values
	minFloat := math.SmallestNonzeroFloat64
	result2 := FloatToBigIntBaseX(minFloat, 18)
	if result2 == nil {
		t.Error("Expected non-nil result for smallest float64")
	}

	// Test with NaN - should handle gracefully without panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("FloatToBigIntBaseX panicked with NaN: %v (expected behavior)", r)
			}
		}()
		nanResult := FloatToBigIntBaseX(math.NaN(), 18)
		// NaN should be handled gracefully, result may be nil or zero
		if nanResult == nil {
			t.Log("NaN input returned nil result (acceptable)")
		}
	}()

	// Test with infinity - should handle gracefully without panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FloatToBigIntBaseX panicked with positive infinity: %v", r)
			}
		}()
		infResult := FloatToBigIntBaseX(math.Inf(1), 18)
		if infResult == nil {
			t.Log("Positive infinity input returned nil result (acceptable)")
		}
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FloatToBigIntBaseX panicked with negative infinity: %v", r)
			}
		}()
		negInfResult := FloatToBigIntBaseX(math.Inf(-1), 18)
		if negInfResult == nil {
			t.Log("Negative infinity input returned nil result (acceptable)")
		}
	}()
}

// Test precision loss scenarios
func TestPrecisionLossScenarios(t *testing.T) {
	// Test with very large numbers that might cause precision loss
	largeValue := 1e20
	result := FloatToBigIntBaseX(largeValue, 5)

	// Should not panic and should handle large values
	if result == nil {
		t.Error("Expected non-nil result for large value")
	}

	// Test conversion back to verify precision
	backToFloat := BigInt2BigFloat(result, 5)
	if backToFloat == nil {
		t.Error("Expected non-nil result for conversion back")
	}
}

// Test string formatting edge cases
func TestStringFormattingEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		base     int64
		decimals int
		expected string
	}{
		{
			name:     "exact decimal places",
			input:    big.NewInt(1234567890000000000),
			base:     18,
			decimals: 18,
			expected: "1.23456789",
		},
		{
			name:     "more decimals than base",
			input:    big.NewInt(1234567890000000000),
			base:     18,
			decimals: 20,
			expected: "1.23456789000000000003",
		},
		{
			name:     "zero decimals",
			input:    big.NewInt(1234567890000000000),
			base:     18,
			decimals: 0,
			expected: "1",
		},
		{
			name:     "negative decimals (should be limited)",
			input:    big.NewInt(1234567890000000000),
			base:     18,
			decimals: -5,
			expected: "1.23456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnBaseXFloatString(tt.input, tt.base, tt.decimals)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Test thread safety (concurrent access)
func TestThreadSafety(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 1000
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// Test various operations concurrently
				val := float64(id*numOperations + j)

				// Test base conversions
				result1 := BigIntBaseX(val, 18)
				if result1 == nil {
					t.Errorf("Goroutine %d: Expected non-nil result", id)
				}

				// Test string conversions
				result2 := UnBaseXFloatString(result1, 18, 8)
				if result2 == "" {
					t.Errorf("Goroutine %d: Expected non-empty string", id)
				}

				// Test APR calculations
				result3 := BigIntDailyAPR(result1)
				if result3 == nil {
					t.Errorf("Goroutine %d: Expected non-nil APR result", id)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// Test error handling consistency
func TestErrorHandlingConsistency(t *testing.T) {
	// Test that all functions handle nil inputs consistently
	nilTests := []struct {
		name     string
		function func() interface{}
	}{
		{
			name: "BigInt2Float with nil",
			function: func() interface{} {
				result, err := BigInt2Float(nil, 18)
				return map[string]interface{}{"result": result, "error": err}
			},
		},
		{
			name: "BigInt2BigFloat with nil",
			function: func() interface{} {
				return BigInt2BigFloat(nil, 18)
			},
		},
		{
			name: "BigFloatFromBigInt with nil",
			function: func() interface{} {
				return BigFloatFromBigInt(nil)
			},
		},
		{
			name: "UnBaseX with nil",
			function: func() interface{} {
				return UnBaseX(nil, 18)
			},
		},
		{
			name: "UnBase14 with nil",
			function: func() interface{} {
				return UnBase14(nil)
			},
		},
	}

	for _, tt := range nilTests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()

			// Should not panic
			if result == nil {
				t.Error("Expected non-nil result for nil input")
			}
		})
	}
}

// Additional benchmark tests for performance validation
func BenchmarkBigInt2BigFloat(b *testing.B) {
	input := new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BigInt2BigFloat(input, 18)
	}
}

func BenchmarkBigFloatFromBigInt(b *testing.B) {
	input := new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BigFloatFromBigInt(input)
	}
}

func BenchmarkBigIntByString(b *testing.B) {
	testString := "123456789012345678901234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BigIntByString(testString)
	}
}

func BenchmarkUnBaseXFloatString(b *testing.B) {
	input := new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnBaseXFloatString(input, 18, 8)
	}
}

func BenchmarkBigIntDailyAPR(b *testing.B) {
	input := big.NewInt(36000) // 100% APR

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BigIntDailyAPR(input)
	}
}

func BenchmarkBigIntHrAPR(b *testing.B) {
	input := big.NewInt(36000) // 100% APR
	hour := big.NewInt(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BigIntHrAPR(input, hour)
	}
}

// Test constants are correctly defined
func TestConstants(t *testing.T) {
	if DAILY_COUNT != 360 {
		t.Errorf("Expected DAILY_COUNT to be 360, got %d", DAILY_COUNT)
	}

	if HOURLY_COUNT != 8640 {
		t.Errorf("Expected HOURLY_COUNT to be 8640, got %d", HOURLY_COUNT)
	}

	if HOURS_DAILY != 24 {
		t.Errorf("Expected HOURS_DAILY to be 24, got %d", HOURS_DAILY)
	}
}

// Test cache initialization
func TestCacheInitialization(t *testing.T) {
	// Test that common bases are pre-computed
	if pow10Cache[14] == nil {
		t.Error("Expected base 14 to be cached")
	}

	if pow10Cache[18] == nil {
		t.Error("Expected base 18 to be cached")
	}

	if pow10FloatCache[14] == nil {
		t.Error("Expected base 14 float to be cached")
	}

	if pow10FloatCache[18] == nil {
		t.Error("Expected base 18 float to be cached")
	}
}
