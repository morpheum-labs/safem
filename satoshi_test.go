package safem

import (
	"math"
	"math/big"
	"runtime"
	"testing"
)

func TestDecimalToSatoshi(t *testing.T) {
	tests := []struct {
		name     string
		decimal  float64
		expected string
	}{
		{"Zero", 0.0, "0"},
		{"One", 1.0, "100000000"},
		{"Small decimal", 0.00000001, "1"},
		{"Standard value", 50000.0, "5000000000000"},
		{"Decimal value", 123.456789, "12345678900"},
		{"Large value", 1000000.0, "100000000000000"},
		{"Precise decimal", 0.12345678, "12345678"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DecimalToSatoshi(test.decimal)
			if result != test.expected {
				t.Errorf("DecimalToSatoshi(%f) = %s, expected %s", test.decimal, result, test.expected)
			}
		})
	}
}

func TestDecimalToSatoshiEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		decimal  float64
		expected string
	}{
		{"NaN", math.NaN(), "0"},
		{"Positive Infinity", math.Inf(1), "0"},
		{"Negative Infinity", math.Inf(-1), "0"},
		{"Negative value", -100.0, "0"}, // Should handle gracefully
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DecimalToSatoshi(test.decimal)
			if result != test.expected {
				t.Errorf("DecimalToSatoshi(%f) = %s, expected %s", test.decimal, result, test.expected)
			}
		})
	}
}

func TestDecimalToSatoshiBigInt(t *testing.T) {
	tests := []struct {
		name     string
		decimal  float64
		expected string
		hasError bool
	}{
		{"Zero", 0.0, "0", false},
		{"One", 1.0, "100000000", false},
		{"Standard value", 50000.0, "5000000000000", false},
		{"Decimal value", 123.456789, "12345678900", false},
		{"NaN", math.NaN(), "", true},
		{"Positive Infinity", math.Inf(1), "", true},
		{"Negative Infinity", math.Inf(-1), "", true},
		{"Negative value", -100.0, "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := DecimalToSatoshiBigInt(test.decimal)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.String() != test.expected {
					t.Errorf("DecimalToSatoshiBigInt(%f) = %s, expected %s", test.decimal, result.String(), test.expected)
				}
			}
		})
	}
}

func TestSatoshiToDecimal(t *testing.T) {
	tests := []struct {
		name       string
		satoshiStr string
		expected   float64
		hasError   bool
		tolerance  float64
	}{
		{"Zero", "0", 0.0, false, 0},
		{"One satoshi", "1", 0.00000001, false, 1e-9},
		{"Standard value", "5000000000000", 50000.0, false, 0.01},
		{"Decimal value", "12345678900", 123.456789, false, 1e-6},
		{"Large value", "100000000000000", 1000000.0, false, 0.01},
		{"Empty string", "", 0.0, true, 0},
		{"Invalid string", "abc", 0.0, true, 0},
		{"Negative value", "-100000000", 0.0, true, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := SatoshiToDecimal(test.satoshiStr)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				diff := math.Abs(result - test.expected)
				if diff > test.tolerance {
					t.Errorf("SatoshiToDecimal(%s) = %f, expected %f (diff: %f)", test.satoshiStr, result, test.expected, diff)
				}
			}
		})
	}
}

func TestWeiToSatoshi(t *testing.T) {
	tests := []struct {
		name     string
		weiStr   string
		expected string
		hasError bool
	}{
		{"Zero Wei", "0", "0", false},
		{"1 Ether in Wei", "1000000000000000000", "100000000", false},
		{"Standard conversion", "50000000000000000000000", "5000000000000", false},
		{"Small Wei", "100000000000000000", "10000000", false},
		{"Empty string", "", "", true},
		{"Invalid string", "abc", "", true},
		{"Negative Wei", "-1000000000000000000", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := WeiToSatoshi(test.weiStr)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != test.expected {
					t.Errorf("WeiToSatoshi(%s) = %s, expected %s", test.weiStr, result, test.expected)
				}
			}
		})
	}
}

func TestSatoshiToWei(t *testing.T) {
	tests := []struct {
		name       string
		satoshiStr string
		expected   string
		hasError   bool
	}{
		{"Zero Satoshi", "0", "0", false},
		{"1 Ether in Satoshi", "100000000", "1000000000000000000", false},
		{"Standard conversion", "5000000000000", "50000000000000000000000", false},
		{"Small Satoshi", "10000000", "100000000000000000", false},
		{"Empty string", "", "", true},
		{"Invalid string", "abc", "", true},
		{"Negative Satoshi", "-100000000", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := SatoshiToWei(test.satoshiStr)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != test.expected {
					t.Errorf("SatoshiToWei(%s) = %s, expected %s", test.satoshiStr, result, test.expected)
				}
			}
		})
	}
}

func TestSatoshiRoundTripConversion(t *testing.T) {
	tests := []struct {
		name    string
		decimal float64
	}{
		{"Zero", 0.0},
		{"One", 1.0},
		{"Standard value", 50000.0},
		{"Decimal value", 123.456789},
		{"Small value", 0.00000001},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Convert decimal to satoshi
			satoshiStr := DecimalToSatoshi(test.decimal)

			// Convert back to decimal
			decimal, err := SatoshiToDecimal(satoshiStr)
			if err != nil {
				t.Fatalf("SatoshiToDecimal failed: %v", err)
			}

			// Check round-trip accuracy (allow small floating-point errors)
			diff := math.Abs(decimal - test.decimal)
			tolerance := test.decimal * 1e-8 // 8 decimal places precision
			if tolerance < 1e-8 {
				tolerance = 1e-8
			}
			if diff > tolerance {
				t.Errorf("Round-trip conversion failed: original %f, round-trip %f (diff: %f)", test.decimal, decimal, diff)
			}
		})
	}
}

func TestWeiSatoshiRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		weiStr string
	}{
		{"Zero Wei", "0"},
		{"1 Ether", "1000000000000000000"},
		{"Standard value", "50000000000000000000000"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Convert Wei to Satoshi
			satoshiStr, err := WeiToSatoshi(test.weiStr)
			if err != nil {
				t.Fatalf("WeiToSatoshi failed: %v", err)
			}

			// Convert back to Wei
			weiBack, err := SatoshiToWei(satoshiStr)
			if err != nil {
				t.Fatalf("SatoshiToWei failed: %v", err)
			}

			// Compare original and round-trip values
			weiBig := new(big.Int)
			weiBig.SetString(test.weiStr, 10)

			weiBackBig := new(big.Int)
			weiBackBig.SetString(weiBack, 10)

			if weiBig.Cmp(weiBackBig) != 0 {
				t.Errorf("Round-trip conversion failed: original %s, round-trip %s", test.weiStr, weiBack)
			}
		})
	}
}

func TestNormalizeToSatoshi(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
		hasError bool
	}{
		{"Decimal string", "50000.0", "5000000000000", false},
		{"Float64", 50000.0, "5000000000000", false},
		{"Float32", float32(50000.0), "5000000000000", false},
		{"Int64", int64(50000), "5000000000000", false},
		{"Int", 50000, "5000000000000", false},
		{"Already satoshi string", "5000000000000", "5000000000000", false},
		{"Wei string (auto-detect)", "50000000000000000000000", "5000000000000", false},
		{"BigInt satoshi", big.NewInt(5000000000000), "5000000000000", false},
		{"BigInt wei (auto-detect)", func() *big.Int { b, _ := new(big.Int).SetString("50000000000000000000000", 10); return b }(), "5000000000000", false},
		{"Empty string", "", "", true},
		{"Invalid string", "abc", "", true},
		{"Nil big.Int", (*big.Int)(nil), "", true},
		{"Unsupported type", []int{1, 2, 3}, "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NormalizeToSatoshi(test.value)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != test.expected {
					t.Errorf("NormalizeToSatoshi(%v) = %s, expected %s", test.value, result, test.expected)
				}
			}
		})
	}
}

func TestBatchDecimalToSatoshi(t *testing.T) {
	decimals := []float64{0.0, 1.0, 50000.0, 123.456789}
	expected := []string{"0", "100000000", "5000000000000", "12345678900"}

	results := BatchDecimalToSatoshi(decimals)

	if len(results) != len(expected) {
		t.Fatalf("BatchDecimalToSatoshi returned %d results, expected %d", len(results), len(expected))
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("BatchDecimalToSatoshi[%d] = %s, expected %s", i, result, expected[i])
		}
	}
}

func TestBatchSatoshiToDecimal(t *testing.T) {
	satoshiStrs := []string{"0", "100000000", "5000000000000", "12345678900"}
	expected := []float64{0.0, 1.0, 50000.0, 123.456789}

	results, err := BatchSatoshiToDecimal(satoshiStrs)
	if err != nil {
		t.Fatalf("BatchSatoshiToDecimal failed: %v", err)
	}

	if len(results) != len(expected) {
		t.Fatalf("BatchSatoshiToDecimal returned %d results, expected %d", len(results), len(expected))
	}

	for i, result := range results {
		diff := math.Abs(result - expected[i])
		tolerance := expected[i] * 1e-6
		if tolerance < 1e-6 {
			tolerance = 1e-6
		}
		if diff > tolerance {
			t.Errorf("BatchSatoshiToDecimal[%d] = %f, expected %f (diff: %f)", i, result, expected[i], diff)
		}
	}
}

func TestBatchSatoshiToDecimalError(t *testing.T) {
	satoshiStrs := []string{"0", "invalid", "5000000000000"}
	_, err := BatchSatoshiToDecimal(satoshiStrs)
	if err == nil {
		t.Errorf("Expected error for invalid satoshi string")
	}
}

// Memory leak test: Verify that sync.Pool is being used correctly
func TestMemoryLeak(t *testing.T) {
	// Force GC before test
	runtime.GC()

	// Get initial memory stats
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Perform many conversions (should reuse pooled objects)
	for i := 0; i < 10000; i++ {
		_ = DecimalToSatoshi(float64(i))
		_, _ = SatoshiToDecimal("5000000000000")
		_, _ = WeiToSatoshi("50000000000000000000000")
		_, _ = SatoshiToWei("5000000000000")
	}

	// Force GC after operations
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory allocated
	allocated := m2.TotalAlloc - m1.TotalAlloc

	// With proper pooling, memory allocation should be minimal
	// Allow some overhead for initial allocations
	maxExpectedAlloc := uint64(10 * 1024 * 1024) // 10MB max
	if allocated > maxExpectedAlloc {
		t.Errorf("Potential memory leak: allocated %d bytes (expected < %d)", allocated, maxExpectedAlloc)
	}
}

// Performance benchmarks
func BenchmarkDecimalToSatoshi(b *testing.B) {
	decimal := 50000.0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DecimalToSatoshi(decimal)
	}
}

func BenchmarkDecimalToSatoshiBigInt(b *testing.B) {
	decimal := 50000.0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecimalToSatoshiBigInt(decimal)
	}
}

func BenchmarkSatoshiToDecimal(b *testing.B) {
	satoshiStr := "5000000000000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SatoshiToDecimal(satoshiStr)
	}
}

func BenchmarkWeiToSatoshi(b *testing.B) {
	weiStr := "50000000000000000000000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = WeiToSatoshi(weiStr)
	}
}

func BenchmarkSatoshiToWei(b *testing.B) {
	satoshiStr := "5000000000000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SatoshiToWei(satoshiStr)
	}
}

func BenchmarkNormalizeToSatoshi(b *testing.B) {
	value := 50000.0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NormalizeToSatoshi(value)
	}
}

func BenchmarkBatchDecimalToSatoshi(b *testing.B) {
	decimals := make([]float64, 1000)
	for i := range decimals {
		decimals[i] = float64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BatchDecimalToSatoshi(decimals)
	}
}

func BenchmarkBatchSatoshiToDecimal(b *testing.B) {
	satoshiStrs := make([]string, 1000)
	for i := range satoshiStrs {
		satoshiStrs[i] = "5000000000000"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BatchSatoshiToDecimal(satoshiStrs)
	}
}
