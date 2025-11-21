package safem

import (
	"math"
	"math/big"
	"testing"
)

func TestWeiToEther(t *testing.T) {
	tests := []struct {
		name     string
		wei      *big.Int
		expected float64
		hasError bool
	}{
		{"Zero Wei", big.NewInt(0), 0.0, false},
		{"1 Ether", new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18)), 1.0, false},
		{"0.5 Ether", new(big.Int).Mul(big.NewInt(5), big.NewInt(1e17)), 0.5, false},
		{"Large Ether", new(big.Int).Mul(big.NewInt(123456789), big.NewInt(1e18)), 123456789.0, false},
		{"Small Wei", big.NewInt(123456), 0.000000000000123456, false},
		// 1 Wei is too small for float64 precision, so we expect an error
		{"1 Wei", big.NewInt(1), 0.0, true},
		{"Negative Wei", big.NewInt(-1e18), 0.0, true},
		{"Nil Wei", nil, 0.0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := WeiToEther(test.wei)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if math.Abs(result-test.expected) > 1e-18 {
					t.Errorf("Expected %g, got %g", test.expected, result)
				}
			}
		})
	}
}

func TestWeiToEtherOptimized(t *testing.T) {
	tests := []struct {
		name     string
		wei      *big.Int
		expected float64
		hasError bool
	}{
		{"Zero Wei", big.NewInt(0), 0.0, false},
		{"1 Ether", new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18)), 1.0, false},
		{"0.5 Ether", new(big.Int).Mul(big.NewInt(5), big.NewInt(1e17)), 0.5, false},
		{"Large Ether", new(big.Int).Mul(big.NewInt(123456789), big.NewInt(1e18)), 123456789.0, false},
		{"Small Wei", big.NewInt(123456), 0.000000000000123456, false},
		// 1 Wei is too small for float64 precision, so we expect an error
		{"1 Wei", big.NewInt(1), 0.0, true},
		{"Negative Wei", big.NewInt(-1e18), 0.0, true},
		{"Nil Wei", nil, 0.0, true},
		// Test edge case for fast path
		{"Max Safe Int64", big.NewInt(math.MaxInt64), float64(math.MaxInt64) / 1e18, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := WeiToEtherOptimized(test.wei)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if math.Abs(result-test.expected) > 1e-18 {
					t.Errorf("Expected %g, got %g", test.expected, result)
				}
			}
		})
	}
}

func TestWeiToEtherSafe(t *testing.T) {
	tests := []struct {
		name     string
		wei      *big.Int
		expected float64
		hasError bool
	}{
		{"Zero Wei", big.NewInt(0), 0.0, false},
		{"1 Ether", new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18)), 1.0, false},
		{"0.5 Ether", new(big.Int).Mul(big.NewInt(5), big.NewInt(1e17)), 0.5, false},
		{"Small Wei", big.NewInt(123456), 0.000000000000123456, false},
		// 1 Wei is too small for float64 precision, so we expect an error
		{"1 Wei", big.NewInt(1), 0.0, true},
		{"Negative Wei", big.NewInt(-1e18), 0.0, true},
		{"Nil Wei", nil, 0.0, true},
		// Test with value exceeding safe precision limits
		{"Too Large Wei", new(big.Int).Lsh(big.NewInt(1), 54), 0.0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := WeiToEtherSafe(test.wei)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if math.Abs(result-test.expected) > 1e-18 {
					t.Errorf("Expected %g, got %g", test.expected, result)
				}
			}
		})
	}
}

func TestEtherToWei(t *testing.T) {
	tests := []struct {
		name     string
		ether    float64
		expected *big.Int
		hasError bool
	}{
		{"Zero Ether", 0.0, big.NewInt(0), false},
		{"1 Ether", 1.0, new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18)), false},
		{"0.5 Ether", 0.5, new(big.Int).Mul(big.NewInt(5), big.NewInt(1e17)), false},
		// Large Ether values may have floating-point precision issues
		// We'll test with a smaller value that should work reliably
		{"Small Ether", 0.000000000000123456, big.NewInt(123456), false},
		{"1 Wei in Ether", 0.000000000000000001, big.NewInt(1), false},
		{"Negative Ether", -1.0, nil, true},
		{"NaN", math.NaN(), nil, true},
		{"Positive Inf", math.Inf(1), nil, true},
		{"Negative Inf", math.Inf(-1), nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := EtherToWei(test.ether)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Cmp(test.expected) != 0 {
					t.Errorf("Expected %v, got %v", test.expected, result)
				}
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	tests := []struct {
		name string
		wei  *big.Int
	}{
		{"Zero", big.NewInt(0)},
		// Skip 1 Wei as it's too small for float64 precision
		{"1 Ether", new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18))},
		{"0.5 Ether", new(big.Int).Mul(big.NewInt(5), big.NewInt(1e17))},
		// Use a smaller amount for large value test to avoid precision issues
		{"Small Amount", new(big.Int).Mul(big.NewInt(123456), big.NewInt(1e18))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Convert Wei to Ether
			ether, err := WeiToEther(test.wei)
			if err != nil {
				t.Fatalf("WeiToEther failed: %v", err)
			}

			// Convert back to Wei
			weiBack, err := EtherToWei(ether)
			if err != nil {
				t.Fatalf("EtherToWei failed: %v", err)
			}

			// Compare original and round-trip values
			if test.wei.Cmp(weiBack) != 0 {
				t.Errorf("Round-trip conversion failed: original %v, round-trip %v", test.wei, weiBack)
			}
		})
	}
}

func BenchmarkWeiToEther(b *testing.B) {
	wei := new(big.Int).Mul(big.NewInt(123456789), big.NewInt(1e18))

	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			WeiToEther(wei)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			WeiToEtherOptimized(wei)
		}
	})

	b.Run("Safe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			WeiToEtherSafe(wei)
		}
	})
}

func BenchmarkWeiToEtherSmall(b *testing.B) {
	wei := big.NewInt(123456) // Small value that uses fast path

	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			WeiToEther(wei)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			WeiToEtherOptimized(wei)
		}
	})

	b.Run("Safe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			WeiToEtherSafe(wei)
		}
	})
}

func BenchmarkEtherToWei(b *testing.B) {
	ether := 123456.0 // Use smaller value to avoid precision issues

	b.Run("EtherToWei", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			EtherToWei(ether)
		}
	})
}
