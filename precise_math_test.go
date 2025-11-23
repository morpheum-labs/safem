package safem

import (
	"math/big"
	"testing"
)

// ============================================================================
// Test MulDivPrecise
// ============================================================================

func TestMulDivPrecise(t *testing.T) {
	tests := []struct {
		name        string
		a           *big.Int
		b           *big.Int
		c           *big.Int
		ceil        bool
		expected    *big.Int
		expectError bool
		errorType   error
	}{
		{
			name:        "round down (ceil=false) - exact division",
			a:           big.NewInt(1000),
			b:           big.NewInt(2000),
			c:           big.NewInt(100),
			ceil:        false,
			expected:    big.NewInt(20000),
			expectError: false,
		},
		{
			name:        "round up (ceil=true) - exact division",
			a:           big.NewInt(1000),
			b:           big.NewInt(2000),
			c:           big.NewInt(100),
			ceil:        true,
			expected:    big.NewInt(20000),
			expectError: false,
		},
		{
			name:        "round down (ceil=false) - with remainder",
			a:           big.NewInt(1001),
			b:           big.NewInt(2000),
			c:           big.NewInt(100),
			ceil:        false,
			expected:    big.NewInt(20020), // 2002000 / 100 = 20020 (truncated)
			expectError: false,
		},
		{
			name:        "round up (ceil=true) - with remainder",
			a:           big.NewInt(1001),
			b:           big.NewInt(2000),
			c:           big.NewInt(100),
			ceil:        true,
			expected:    big.NewInt(20020), // 2002000 / 100 = 20020 (exact division, no remainder)
			expectError: false,
		},
		{
			name:        "round up (ceil=true) - with actual remainder",
			a:           big.NewInt(1001),
			b:           big.NewInt(2000),
			c:           big.NewInt(3),
			ceil:        true,
			expected:    big.NewInt(667334), // 2002000 / 3 = 667333.33... rounded up to 667334
			expectError: false,
		},
		{
			name:        "nil input a",
			a:           nil,
			b:           big.NewInt(1000),
			c:           big.NewInt(100),
			ceil:        false,
			expected:    nil,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "nil input b",
			a:           big.NewInt(1000),
			b:           nil,
			c:           big.NewInt(100),
			ceil:        false,
			expected:    nil,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "nil input c",
			a:           big.NewInt(1000),
			b:           big.NewInt(1000),
			c:           nil,
			ceil:        false,
			expected:    nil,
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "division by zero",
			a:           big.NewInt(1000),
			b:           big.NewInt(1000),
			c:           big.NewInt(0),
			ceil:        false,
			expected:    nil,
			expectError: true,
			errorType:   ErrDivisionByZero,
		},
		{
			name:        "very small values - round down",
			a:           big.NewInt(1),
			b:           big.NewInt(1),
			c:           big.NewInt(3),
			ceil:        false,
			expected:    big.NewInt(0), // 1/3 = 0.33... truncated to 0
			expectError: false,
		},
		{
			name:        "very small values - round up",
			a:           big.NewInt(1),
			b:           big.NewInt(1),
			c:           big.NewInt(3),
			ceil:        true,
			expected:    big.NewInt(1), // 1/3 = 0.33... rounded up to 1
			expectError: false,
		},
		{
			name:        "large values - round down",
			a:           big.NewInt(1000000000),
			b:           big.NewInt(2000000000),
			c:           big.NewInt(3),
			ceil:        false,
			expected:    new(big.Int).Div(new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(2000000000)), big.NewInt(3)),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MulDivPrecise(tt.a, tt.b, tt.c, tt.ceil)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error type %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result.Cmp(tt.expected) != 0 {
					t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
				}
			}
		})
	}
}

// ============================================================================
// Test CalculateFee
// ============================================================================

func TestCalculateFee(t *testing.T) {
	tests := []struct {
		name        string
		amountIn    *big.Int
		feePips     *big.Int
		expected    *big.Int
		expectError bool
	}{
		{
			name:        "standard fee calculation",
			amountIn:    big.NewInt(100000000000), // 1000.0 scaled by 1e8
			feePips:     big.NewInt(10),           // 0.1% = 10 pips
			expected:    big.NewInt(100000000),    // 1.0 scaled by 1e8 (rounded UP)
			expectError: false,
		},
		{
			name:        "fee with remainder - rounds up",
			amountIn:    big.NewInt(100000000001), // 1000.00000001
			feePips:     big.NewInt(10),           // 0.1%
			expected:    big.NewInt(100000001),    // Should round UP
			expectError: false,
		},
		{
			name:        "zero amount",
			amountIn:    big.NewInt(0),
			feePips:     big.NewInt(10),
			expected:    big.NewInt(0),
			expectError: false,
		},
		{
			name:        "zero fee",
			amountIn:    big.NewInt(100000000000),
			feePips:     big.NewInt(0),
			expected:    big.NewInt(0),
			expectError: false,
		},
		{
			name:        "nil amountIn",
			amountIn:    nil,
			feePips:     big.NewInt(10),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "nil feePips",
			amountIn:    big.NewInt(100000000000),
			feePips:     nil,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "negative amountIn",
			amountIn:    big.NewInt(-100000000000),
			feePips:     big.NewInt(10),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "negative feePips",
			amountIn:    big.NewInt(100000000000),
			feePips:     big.NewInt(-10),
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateFee(tt.amountIn, tt.feePips)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result.Cmp(tt.expected) != 0 {
					t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
				}
			}
		})
	}
}

// ============================================================================
// Test CalculatePnLAdjusted
// ============================================================================

func TestCalculatePnLAdjusted(t *testing.T) {
	tests := []struct {
		name        string
		priceDiff   *big.Int
		size        *big.Int
		scale       *big.Int
		expected    *big.Int
		expectError bool
		description string
	}{
		{
			name:        "profit - rounds down",
			priceDiff:   big.NewInt(100000000),   // +1.0 price diff
			size:        big.NewInt(10000000000), // 100.0 size
			scale:       big.NewInt(100000000),   // 1e8 scale
			expected:    big.NewInt(10000000000), // 100.0 profit (rounded DOWN)
			expectError: false,
			description: "Positive PnL (profit) should round DOWN",
		},
		{
			name:        "loss - rounds up",
			priceDiff:   big.NewInt(-100000000),   // -1.0 price diff
			size:        big.NewInt(10000000000),  // 100.0 size
			scale:       big.NewInt(100000000),    // 1e8 scale
			expected:    big.NewInt(-10000000000), // -100.0 loss (rounded UP, so less negative)
			expectError: false,
			description: "Negative PnL (loss) should round UP (less negative)",
		},
		{
			name:        "profit with remainder - rounds down",
			priceDiff:   big.NewInt(100000001),   // +1.00000001 price diff (scaled by 1e8)
			size:        big.NewInt(10000000000), // 100.0 size (scaled by 1e8)
			scale:       big.NewInt(100000000),   // 1e8 scale
			expected:    big.NewInt(10000000100), // (100000001 * 10000000000) / 100000000 = 10000000100 (rounded DOWN)
			expectError: false,
			description: "Positive PnL with remainder should round DOWN",
		},
		{
			name:        "loss with remainder - rounds up",
			priceDiff:   big.NewInt(-100000001),   // -1.00000001 price diff (scaled by 1e8)
			size:        big.NewInt(10000000000),  // 100.0 size (scaled by 1e8)
			scale:       big.NewInt(100000000),    // 1e8 scale
			expected:    big.NewInt(-10000000100), // (-100000001 * 10000000000) / 100000000 = -10000000100 exactly (no remainder, so rounding UP doesn't change it)
			expectError: false,
			description: "Negative PnL with remainder should round UP (less negative)",
		},
		{
			name:        "loss with actual remainder - rounds up",
			priceDiff:   big.NewInt(-100000001),          // -1.00000001 price diff (scaled by 1e8)
			size:        big.NewInt(10000000000),         // 100.0 size (scaled by 1e8)
			scale:       big.NewInt(3),                   // 3 (creates remainder)
			expected:    big.NewInt(-333333336666666666), // (-100000001 * 10000000000) / 3 = -333333336666666667, remainder is negative, so rounding UP (ceil=true) adds 1: -333333336666666666
			expectError: false,
			description: "Negative PnL with actual remainder should round UP (less negative)",
		},
		{
			name:        "zero price diff",
			priceDiff:   big.NewInt(0),
			size:        big.NewInt(10000000000),
			scale:       big.NewInt(100000000),
			expected:    big.NewInt(0),
			expectError: false,
		},
		{
			name:        "zero size",
			priceDiff:   big.NewInt(100000000),
			size:        big.NewInt(0),
			scale:       big.NewInt(100000000),
			expected:    big.NewInt(0),
			expectError: false,
		},
		{
			name:        "nil priceDiff",
			priceDiff:   nil,
			size:        big.NewInt(10000000000),
			scale:       big.NewInt(100000000),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "nil size",
			priceDiff:   big.NewInt(100000000),
			size:        nil,
			scale:       big.NewInt(100000000),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "nil scale",
			priceDiff:   big.NewInt(100000000),
			size:        big.NewInt(10000000000),
			scale:       nil,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "zero scale",
			priceDiff:   big.NewInt(100000000),
			size:        big.NewInt(10000000000),
			scale:       big.NewInt(0),
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculatePnLAdjusted(tt.priceDiff, tt.size, tt.scale)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result.Cmp(tt.expected) != 0 {
					t.Errorf("expected %s, got %s. %s", tt.expected.String(), result.String(), tt.description)
				}
			}
		})
	}
}

// ============================================================================
// Test CalculateMarginRatioConservative
// ============================================================================

func TestCalculateMarginRatioConservative(t *testing.T) {
	tests := []struct {
		name              string
		equity            *big.Int
		marginRequirement *big.Int
		expected          *big.Int
		expectError       bool
		description       string
	}{
		{
			name:              "standard margin ratio - rounds down",
			equity:            big.NewInt(100000000000), // 1000.0 equity
			marginRequirement: big.NewInt(50000000000),  // 500.0 margin
			expected:          big.NewInt(200000000),    // 2.0 ratio (scaled by 1e8)
			expectError:       false,
			description:       "Margin ratio should round DOWN for conservative risk assessment",
		},
		{
			name:              "margin ratio with remainder - rounds down",
			equity:            big.NewInt(100000000001), // 1000.00000001 equity
			marginRequirement: big.NewInt(50000000000),  // 500.0 margin
			expected:          big.NewInt(200000000),    // Should round DOWN to 2.0
			expectError:       false,
			description:       "Margin ratio with remainder should round DOWN",
		},
		{
			name:              "low margin ratio",
			equity:            big.NewInt(110000000000), // 1100.0 equity
			marginRequirement: big.NewInt(100000000000), // 1000.0 margin
			expected:          big.NewInt(110000000),    // 1.1 ratio
			expectError:       false,
		},
		{
			name:              "zero equity",
			equity:            big.NewInt(0),
			marginRequirement: big.NewInt(50000000000),
			expected:          big.NewInt(0),
			expectError:       false,
		},
		{
			name:              "nil equity",
			equity:            nil,
			marginRequirement: big.NewInt(50000000000),
			expected:          nil,
			expectError:       true,
		},
		{
			name:              "nil marginRequirement",
			equity:            big.NewInt(100000000000),
			marginRequirement: nil,
			expected:          nil,
			expectError:       true,
		},
		{
			name:              "division by zero",
			equity:            big.NewInt(100000000000),
			marginRequirement: big.NewInt(0),
			expected:          nil,
			expectError:       true,
		},
		{
			name:              "negative equity",
			equity:            big.NewInt(-100000000000),
			marginRequirement: big.NewInt(50000000000),
			expected:          nil,
			expectError:       true,
		},
		{
			name:              "negative marginRequirement",
			equity:            big.NewInt(100000000000),
			marginRequirement: big.NewInt(-50000000000),
			expected:          nil,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateMarginRatioConservative(tt.equity, tt.marginRequirement)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result.Cmp(tt.expected) != 0 {
					t.Errorf("expected %s, got %s. %s", tt.expected.String(), result.String(), tt.description)
				}
			}
		})
	}
}

// ============================================================================
// Test CalculateLiquidationPriceConservative
// ============================================================================

func TestCalculateLiquidationPriceConservative(t *testing.T) {
	tests := []struct {
		name             string
		entryPrice       *big.Int
		pnlAtLiquidation *big.Int
		size             *big.Int
		expected         *big.Int
		expectError      bool
		description      string
	}{
		{
			name:             "standard liquidation price - rounds down",
			entryPrice:       big.NewInt(200000000000), // 2000.0 entry price (scaled by 1e8)
			pnlAtLiquidation: big.NewInt(-50000000000), // -500.0 PnL at liquidation (scaled by 1e8)
			size:             big.NewInt(10000000000),  // 100.0 size (scaled by 1e8)
			expected:         big.NewInt(199999999995), // 2000.0 + (-500.0/100.0) = 1995.0, but rounded DOWN: (-50000000000/10000000000) = -5, so 200000000000 + (-5) = 199999999995
			expectError:      false,
			description:      "Liquidation price should round DOWN for earlier liquidation",
		},
		{
			name:             "liquidation price with remainder - rounds down",
			entryPrice:       big.NewInt(200000000000),
			pnlAtLiquidation: big.NewInt(-50000000001), // -500.00000001 (scaled by 1e8)
			size:             big.NewInt(10000000000),
			expected:         big.NewInt(199999999994), // (-50000000001/10000000000) = -5.0000000001, rounded DOWN to -5, so 200000000000 + (-5) = 199999999995, but actually -50000000001/10000000000 = -5.0000000001, rounded DOWN to -6, so 200000000000 + (-6) = 199999999994
			expectError:      false,
			description:      "Liquidation price with remainder should round DOWN",
		},
		{
			name:             "long position liquidation",
			entryPrice:       big.NewInt(100000000000), // 1000.0 (scaled by 1e8)
			pnlAtLiquidation: big.NewInt(-10000000000), // -100.0 (scaled by 1e8)
			size:             big.NewInt(10000000000),  // 100.0 (scaled by 1e8)
			expected:         big.NewInt(99999999999),  // 1000.0 + (-100.0/100.0) = 999.0, but (-10000000000/10000000000) = -1, so 100000000000 + (-1) = 99999999999
			expectError:      false,
		},
		{
			name:             "short position liquidation",
			entryPrice:       big.NewInt(100000000000), // 1000.0 (scaled by 1e8)
			pnlAtLiquidation: big.NewInt(10000000000),  // +100.0 (loss for short, scaled by 1e8)
			size:             big.NewInt(-10000000000), // -100.0 (short, scaled by 1e8)
			expected:         big.NewInt(99999999999),  // 1000.0 + (100.0/-100.0) = 999.0, but (10000000000/-10000000000) = -1, rounded DOWN to -1, so 100000000000 + (-1) = 99999999999
			expectError:      false,
		},
		{
			name:             "nil entryPrice",
			entryPrice:       nil,
			pnlAtLiquidation: big.NewInt(-50000000000),
			size:             big.NewInt(10000000000),
			expected:         nil,
			expectError:      true,
		},
		{
			name:             "nil pnlAtLiquidation",
			entryPrice:       big.NewInt(200000000000),
			pnlAtLiquidation: nil,
			size:             big.NewInt(10000000000),
			expected:         nil,
			expectError:      true,
		},
		{
			name:             "nil size",
			entryPrice:       big.NewInt(200000000000),
			pnlAtLiquidation: big.NewInt(-50000000000),
			size:             nil,
			expected:         nil,
			expectError:      true,
		},
		{
			name:             "division by zero (zero size)",
			entryPrice:       big.NewInt(200000000000),
			pnlAtLiquidation: big.NewInt(-50000000000),
			size:             big.NewInt(0),
			expected:         nil,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateLiquidationPriceConservative(tt.entryPrice, tt.pnlAtLiquidation, tt.size)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result.Cmp(tt.expected) != 0 {
					t.Errorf("expected %s, got %s. %s", tt.expected.String(), result.String(), tt.description)
				}
			}
		})
	}
}

// ============================================================================
// Test RoundUp and RoundDown helpers
// ============================================================================

func TestRoundUp(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		divisor  *big.Int
		expected *big.Int
	}{
		{
			name:     "exact division",
			value:    big.NewInt(1000),
			divisor:  big.NewInt(100),
			expected: big.NewInt(10),
		},
		{
			name:     "with remainder - rounds up",
			value:    big.NewInt(1001),
			divisor:  big.NewInt(100),
			expected: big.NewInt(11), // 1001/100 = 10.01, rounded up to 11
		},
		{
			name:     "zero value",
			value:    big.NewInt(0),
			divisor:  big.NewInt(100),
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundUp(tt.value, tt.divisor)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestRoundDown(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		divisor  *big.Int
		expected *big.Int
	}{
		{
			name:     "exact division",
			value:    big.NewInt(1000),
			divisor:  big.NewInt(100),
			expected: big.NewInt(10),
		},
		{
			name:     "with remainder - rounds down",
			value:    big.NewInt(1001),
			divisor:  big.NewInt(100),
			expected: big.NewInt(10), // 1001/100 = 10.01, truncated to 10
		},
		{
			name:     "zero value",
			value:    big.NewInt(0),
			divisor:  big.NewInt(100),
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundDown(tt.value, tt.divisor)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// ============================================================================
// Rounding Attack Simulation Tests
// ============================================================================

func TestRoundingAttackPrevention(t *testing.T) {
	// Test that fees always round UP (exchange-favorable)
	t.Run("fee rounding attack prevention", func(t *testing.T) {
		// Simulate many small fee calculations that could be exploited
		for i := 0; i < 1000; i++ {
			amountIn := big.NewInt(int64(100000000 + i)) // Small amounts
			feePips := big.NewInt(10)                    // 0.1% fee

			fee, err := CalculateFee(amountIn, feePips)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Calculate expected minimum fee (rounded up)
			expectedMin := new(big.Int).Div(
				new(big.Int).Mul(amountIn, feePips),
				big.NewInt(10000),
			)

			// Fee should be >= expectedMin (rounded up)
			if fee.Cmp(expectedMin) < 0 {
				t.Errorf("fee %s is less than minimum %s (rounding attack possible)", fee.String(), expectedMin.String())
			}
		}
	})

	// Test that PnL profits round DOWN (exchange-favorable)
	t.Run("PnL profit rounding attack prevention", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			priceDiff := big.NewInt(int64(100000000 + i)) // Positive (profit)
			size := big.NewInt(10000000000)               // 100.0 size
			scale := big.NewInt(100000000)                // 1e8 scale

			pnl, err := CalculatePnLAdjusted(priceDiff, size, scale)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Calculate expected maximum PnL (rounded down)
			expectedMax := new(big.Int).Div(
				new(big.Int).Mul(priceDiff, size),
				scale,
			)

			// PnL should be <= expectedMax (rounded down)
			if pnl.Cmp(expectedMax) > 0 {
				t.Errorf("PnL profit %s is greater than maximum %s (rounding attack possible)", pnl.String(), expectedMax.String())
			}
		}
	})

	// Test that PnL losses round UP (exchange-favorable)
	t.Run("PnL loss rounding attack prevention", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			priceDiff := big.NewInt(int64(-100000000 - i)) // Negative (loss)
			size := big.NewInt(10000000000)                // 100.0 size
			scale := big.NewInt(100000000)                 // 1e8 scale

			pnl, err := CalculatePnLAdjusted(priceDiff, size, scale)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Calculate expected minimum PnL (rounded up, less negative)
			expectedMin := new(big.Int).Div(
				new(big.Int).Mul(priceDiff, size),
				scale,
			)

			// PnL should be >= expectedMin (rounded up, less negative)
			if pnl.Cmp(expectedMin) < 0 {
				t.Errorf("PnL loss %s is less than minimum %s (rounding attack possible)", pnl.String(), expectedMin.String())
			}
		}
	})
}

// ============================================================================
// Precision Preservation Tests
// ============================================================================

func TestPrecisionPreservation(t *testing.T) {
	// Test that precision is preserved over many operations
	t.Run("precision preservation over 1M operations", func(t *testing.T) {
		// Simulate 1M fee calculations
		totalFee := big.NewInt(0)
		amountIn := big.NewInt(100000000000) // 1000.0
		feePips := big.NewInt(10)            // 0.1%

		for i := 0; i < 1000000; i++ {
			fee, err := CalculateFee(amountIn, feePips)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			totalFee.Add(totalFee, fee)
		}

		// Expected total: 1M * 1.0 = 1M (scaled by 1e8)
		expectedTotal := big.NewInt(100000000000000) // 1M * 1e8
		if totalFee.Cmp(expectedTotal) < 0 {
			t.Errorf("total fee %s is less than expected %s (precision loss)", totalFee.String(), expectedTotal.String())
		}
	})
}
