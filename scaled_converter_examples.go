package safem

import (
	"fmt"
	"math/big"
)

// ============================================================================
// Scaled Converter Examples and Use Cases
// ============================================================================
//
// This file demonstrates practical usage patterns for scaled uint64 conversions
// in high-performance orderbook operations, risk calculations, and oracle aggregation.
// Scaled converters convert u256 (*big.Int) to u64 (scaled) for fast sorting and
// comparisons while preserving precision for exact calculations.
//
// USAGE SCENARIOS:
// 1. Orderbook Price Level Keys
// 2. Risk Engine Calculations
// 3. Oracle Price Aggregation
// 4. ADL (Auto-Deleveraging) Ranking
// 5. High-Frequency Trading Operations
//
// ============================================================================

// ExampleBigIntToPriceKey_Orderbook demonstrates orderbook price keys
// Use Case: Converting order prices to uint64 keys for fast sorting
func ExampleBigIntToPriceKey_Orderbook() {
	// Order prices in *big.Int (from EIP-712 payload)
	orderPrices := []*big.Int{
		big.NewInt(5000000000000), // 50000.0 USD (satoshi format)
		big.NewInt(5000100000000), // 50001.0 USD
		big.NewInt(5000200000000), // 50002.0 USD
		big.NewInt(4999900000000), // 49999.0 USD
		big.NewInt(5000300000000), // 50003.0 USD
	}

	fmt.Println("Order Prices → Price Keys:")
	for i, priceBig := range orderPrices {
		priceKey, err := BigIntToPriceKey(priceBig)
		if err != nil {
			fmt.Printf("  Order %d: Error - %v\n", i, err)
			continue
		}
		// Convert back to verify
		priceBack := PriceKeyToBigInt(priceKey)
		fmt.Printf("  Order %d: %s → key %d → %s\n", i, priceBig.String(), priceKey, priceBack.String())
	}
	// Output:
	// Order Prices → Price Keys:
	//   Order 0: 5000000000000 → key 5000000000000 → 5000000000000
	//   Order 1: 5000100000000 → key 5000100000000 → 5000100000000
	//   Order 2: 5000200000000 → key 5000200000000 → 5000200000000
	//   Order 3: 4999900000000 → key 4999900000000 → 4999900000000
	//   Order 4: 5000300000000 → key 5000300000000 → 5000300000000
}

// ExampleFloat64ToPriceKey_UserInput demonstrates user input conversion
// Use Case: Converting user input prices to keys for orderbook operations
func ExampleFloat64ToPriceKey_UserInput() {
	// User input prices from UI
	userPrices := []float64{
		50000.0,
		50001.5,
		50002.25,
		49999.75,
	}

	fmt.Println("User Input → Price Keys:")
	for i, price := range userPrices {
		priceKey, err := Float64ToPriceKey(price)
		if err != nil {
			fmt.Printf("  Price %d: Error - %v\n", i, err)
			continue
		}
		// Convert back for display
		priceBack := PriceKeyToFloat64(priceKey)
		fmt.Printf("  Price %d: $%.2f → key %d → $%.2f\n", i, price, priceKey, priceBack)
	}
	// Output:
	// User Input → Price Keys:
	//   Price 0: $50000.00 → key 5000000000000 → $50000.00
	//   Price 1: $50001.50 → key 5000150000000 → $50001.50
	//   Price 2: $50002.25 → key 5000225000000 → $50002.25
	//   Price 3: $49999.75 → key 4999975000000 → $49999.75
}

// ExampleComparePriceKeys_Sorting demonstrates fast price comparison
// Use Case: Sorting orders by price using integer comparison
func ExampleComparePriceKeys_Sorting() {
	// Price keys (already converted)
	priceKeys := []uint64{
		5000200000000, // 50002.0
		5000000000000, // 50000.0
		5000100000000, // 50001.0
		4999900000000, // 49999.0
	}

	fmt.Println("Price Key Comparison:")
	for i := 0; i < len(priceKeys)-1; i++ {
		comparison := ComparePriceKeys(priceKeys[i], priceKeys[i+1])
		price1 := PriceKeyToFloat64(priceKeys[i])
		price2 := PriceKeyToFloat64(priceKeys[i+1])

		var result string
		switch comparison {
		case -1:
			result = fmt.Sprintf("%.2f < %.2f", price1, price2)
		case 0:
			result = fmt.Sprintf("%.2f == %.2f", price1, price2)
		case 1:
			result = fmt.Sprintf("%.2f > %.2f", price1, price2)
		}
		fmt.Printf("  Compare key %d vs %d: %s\n", i, i+1, result)
	}
	// Output:
	// Price Key Comparison:
	//   Compare key 0 vs 1: 50002.00 > 50000.00
	//   Compare key 1 vs 2: 50000.00 < 50001.00
	//   Compare key 2 vs 3: 50001.00 > 49999.00
}

// ExampleBigIntToValueKey_RiskEngine demonstrates risk engine calculations
// Use Case: Converting position values to keys for aggregation
func ExampleBigIntToValueKey_RiskEngine() {
	// Position values in *big.Int
	positionValues := []*big.Int{
		big.NewInt(100000000000),  // $1000 position
		big.NewInt(500000000000),  // $5000 position
		big.NewInt(2500000000000), // $25000 position
		big.NewInt(1000000000000), // $10000 position
	}

	fmt.Println("Position Values → Value Keys:")
	totalValueKey := uint64(0)
	for i, valueBig := range positionValues {
		valueKey, err := BigIntToValueKey(valueBig)
		if err != nil {
			fmt.Printf("  Position %d: Error - %v\n", i, err)
			continue
		}
		// Aggregate values
		totalValueKey, _ = AddValueKeys(totalValueKey, valueKey)
		valueBack := ValueKeyToFloat64(valueKey)
		fmt.Printf("  Position %d: %s → key %d ($%.2f)\n", i, valueBig.String(), valueKey, valueBack)
	}
	totalValue := ValueKeyToFloat64(totalValueKey)
	fmt.Printf("  Total Portfolio Value: $%.2f\n", totalValue)
	// Output:
	// Position Values → Value Keys:
	//   Position 0: 100000000000 → key 100000000000 ($1000.00)
	//   Position 1: 500000000000 → key 500000000000 ($5000.00)
	//   Position 2: 2500000000000 → key 2500000000000 ($25000.00)
	//   Position 3: 1000000000000 → key 1000000000000 ($10000.00)
	//   Total Portfolio Value: $41000.00
}

// ExampleCalculateMarginRatioKey_Liquidation demonstrates liquidation checks
// Use Case: Fast margin ratio calculation for liquidation engine
func ExampleCalculateMarginRatioKey_Liquidation() {
	// User's equity and margin requirement
	equityKey, _ := Float64ToValueKey(10000.0)           // $10,000 equity
	marginRequirementKey, _ := Float64ToValueKey(5000.0) // $5,000 margin requirement

	// Calculate margin ratio
	marginRatioKey, err := CalculateMarginRatioKey(equityKey, marginRequirementKey)
	if err != nil {
		fmt.Printf("Error calculating margin ratio: %v\n", err)
		return
	}

	marginRatio := RatioKeyToFloat64(marginRatioKey)
	fmt.Printf("Margin Ratio Calculation:\n")
	fmt.Printf("  Equity: $%.2f (key: %d)\n", ValueKeyToFloat64(equityKey), equityKey)
	fmt.Printf("  Margin Requirement: $%.2f (key: %d)\n", ValueKeyToFloat64(marginRequirementKey), marginRequirementKey)
	fmt.Printf("  Margin Ratio: %.4f (key: %d)\n", marginRatio, marginRatioKey)

	// Check if liquidatable (threshold: 1.1 = 110%)
	liquidationThresholdKey, _ := Float64ToRatioKey(1.1)
	isLiquidatable := IsLiquidatableKey(marginRatioKey, liquidationThresholdKey)
	if isLiquidatable {
		fmt.Printf("  Status: ❌ LIQUIDATABLE (ratio < 1.1)\n")
	} else {
		fmt.Printf("  Status: ✅ SAFE (ratio >= 1.1)\n")
	}
	// Output:
	// Margin Ratio Calculation:
	//   Equity: $10000.00 (key: 1000000000000)
	//   Margin Requirement: $5000.00 (key: 500000000000)
	//   Margin Ratio: 2.000000 (key: 2000000)
	//   Status: ✅ SAFE (ratio >= 1.1)
}

// ExampleAddPriceKeys_Aggregation demonstrates price aggregation
// Use Case: Aggregating order quantities at price levels
func ExampleAddPriceKeys_Aggregation() {
	// Multiple orders at the same price level
	priceKey, _ := Float64ToPriceKey(50000.0)
	orderQuantities := []uint64{
		1000000000, // 0.01 BTC
		2000000000, // 0.02 BTC
		5000000000, // 0.05 BTC
	}

	fmt.Printf("Aggregating Orders at Price $%.2f (key: %d):\n", PriceKeyToFloat64(priceKey), priceKey)
	totalQuantity := uint64(0)
	for i, qty := range orderQuantities {
		var err error
		totalQuantity, err = AddPriceKeys(totalQuantity, qty)
		if err != nil {
			fmt.Printf("  Order %d: Error aggregating - %v\n", i, err)
			continue
		}
		// Convert quantity key to float64 (using QuantityScale)
		qtyDecimal := ScaledUint64ToFloat64(qty, QuantityScale)
		fmt.Printf("  Order %d: %.8f BTC (key: %d)\n", i, qtyDecimal, qty)
	}
	totalQtyDecimal := ScaledUint64ToFloat64(totalQuantity, QuantityScale)
	fmt.Printf("  Total Quantity: %.8f BTC (key: %d)\n", totalQtyDecimal, totalQuantity)
	// Output:
	// Aggregating Orders at Price $50000.00 (key: 5000000000000):
	//   Order 0: 0.01000000 BTC (key: 1000000000)
	//   Order 1: 0.02000000 BTC (key: 2000000000)
	//   Order 2: 0.05000000 BTC (key: 5000000000)
	//   Total Quantity: 0.08000000 BTC (key: 8000000000)
}

// ExampleMultiplyPriceKeys_PositionValue demonstrates position value calculation
// Use Case: Calculating position value (size * price)
func ExampleMultiplyPriceKeys_PositionValue() {
	// Position details
	positionSizeKey, _ := Float64ToQuantityKey(0.5) // 0.5 BTC
	entryPriceKey, _ := Float64ToPriceKey(50000.0)  // $50,000

	// Calculate position value: size * price
	positionValueKey, err := MultiplyPriceKeys(positionSizeKey, entryPriceKey)
	if err != nil {
		fmt.Printf("Error calculating position value: %v\n", err)
		return
	}

	positionValue := ValueKeyToFloat64(positionValueKey)
	fmt.Printf("Position Value Calculation:\n")
	fmt.Printf("  Size: %.8f BTC (key: %d)\n", ScaledUint64ToFloat64(positionSizeKey, QuantityScale), positionSizeKey)
	fmt.Printf("  Entry Price: $%.2f (key: %d)\n", PriceKeyToFloat64(entryPriceKey), entryPriceKey)
	fmt.Printf("  Position Value: $%.2f (key: %d)\n", positionValue, positionValueKey)
	// Output:
	// Position Value Calculation:
	//   Size: 0.50000000 BTC (key: 50000000)
	//   Entry Price: $50000.00 (key: 5000000000000)
	//   Position Value: $25000.00 (key: 2500000000000)
}

// ExampleBatchBigIntToPriceKeys_OrderbookSnapshot demonstrates batch processing
// Use Case: Converting orderbook snapshot efficiently
func ExampleBatchBigIntToPriceKeys_OrderbookSnapshot() {
	// Orderbook snapshot prices in *big.Int
	snapshotPrices := []*big.Int{
		big.NewInt(5000000000000),
		big.NewInt(5000100000000),
		big.NewInt(5000200000000),
		big.NewInt(5000300000000),
		big.NewInt(5000400000000),
	}

	// Batch convert to price keys
	priceKeys, err := BatchBigIntToPriceKeys(snapshotPrices)
	if err != nil {
		fmt.Printf("Batch conversion error: %v\n", err)
		return
	}

	fmt.Println("Orderbook Snapshot (Price Keys):")
	for i, priceKey := range priceKeys {
		price := PriceKeyToFloat64(priceKey)
		fmt.Printf("  Level %d: key %d ($%.2f)\n", i, priceKey, price)
	}
	// Output:
	// Orderbook Snapshot (Price Keys):
	//   Level 0: key 5000000000000 ($50000.00)
	//   Level 1: key 5000100000000 ($50001.00)
	//   Level 2: key 5000200000000 ($50002.00)
	//   Level 3: key 5000300000000 ($50003.00)
	//   Level 4: key 5000400000000 ($50004.00)
}

// ExampleStringToPriceKey_EIP712 demonstrates EIP-712 payload processing
// Use Case: Converting string prices from EIP-712 payloads to keys
func ExampleStringToPriceKey_EIP712() {
	// Prices from EIP-712 payload (u256 strings)
	eip712Prices := []string{
		"5000000000000", // 50000.0 USD
		"5000100000000", // 50001.0 USD
		"5000200000000", // 50002.0 USD
	}

	fmt.Println("EIP-712 Prices → Price Keys:")
	for i, priceStr := range eip712Prices {
		priceKey, err := StringToPriceKey(priceStr)
		if err != nil {
			fmt.Printf("  Price %d: Error - %v\n", i, err)
			continue
		}
		price := PriceKeyToFloat64(priceKey)
		fmt.Printf("  Price %d: %s → key %d ($%.2f)\n", i, priceStr, priceKey, price)
	}
	// Output:
	// EIP-712 Prices → Price Keys:
	//   Price 0: 5000000000000 → key 5000000000000 ($50000.00)
	//   Price 1: 5000100000000 → key 5000100000000 ($50001.00)
	//   Price 2: 5000200000000 → key 5000200000000 ($50002.00)
}

// ExampleFloat64ToScoreKey_ADL demonstrates ADL ranking
// Use Case: Converting ADL scores to keys for fast sorting
func ExampleFloat64ToScoreKey_ADL() {
	// ADL scores for positions (higher = more profitable to liquidate)
	adlScores := []float64{
		1.25, // Position 1
		1.50, // Position 2
		1.10, // Position 3
		1.75, // Position 4
	}

	fmt.Println("ADL Scores → Score Keys (sorted by priority):")
	scoreKeys := make([]uint64, len(adlScores))
	for i, score := range adlScores {
		scoreKey, err := Float64ToScoreKey(score)
		if err != nil {
			fmt.Printf("  Position %d: Error - %v\n", i, err)
			continue
		}
		scoreKeys[i] = scoreKey
		fmt.Printf("  Position %d: score %.2f → key %d\n", i, score, scoreKey)
	}

	// Sort by score key (descending - highest first)
	fmt.Println("\nSorted by ADL Priority (highest first):")
	// Simple bubble sort for demonstration
	for i := 0; i < len(scoreKeys)-1; i++ {
		for j := i + 1; j < len(scoreKeys); j++ {
			if scoreKeys[i] < scoreKeys[j] {
				scoreKeys[i], scoreKeys[j] = scoreKeys[j], scoreKeys[i]
				adlScores[i], adlScores[j] = adlScores[j], adlScores[i]
			}
		}
	}
	for i, scoreKey := range scoreKeys {
		score := ScoreKeyToFloat64(scoreKey)
		fmt.Printf("  Priority %d: Position %d (score: %.2f, key: %d)\n", i+1, i, score, scoreKey)
	}
	// Output:
	// ADL Scores → Score Keys (sorted by priority):
	//   Position 0: score 1.25 → key 125000000
	//   Position 1: score 1.50 → key 150000000
	//   Position 2: score 1.10 → key 110000000
	//   Position 3: score 1.75 → key 175000000
	//
	// Sorted by ADL Priority (highest first):
	//   Priority 1: Position 0 (score: 1.75, key: 175000000)
	//   Priority 2: Position 1 (score: 1.50, key: 150000000)
	//   Priority 3: Position 2 (score: 1.25, key: 125000000)
	//   Priority 4: Position 3 (score: 1.10, key: 110000000)
}

// ExampleSubtractPriceKeys_OrderCancellation demonstrates order cancellation
// Use Case: Removing order quantity from price level
func ExampleSubtractPriceKeys_OrderCancellation() {
	// Price level with aggregated quantity
	priceKey, _ := Float64ToPriceKey(50000.0)
	totalQuantityKey, _ := Float64ToQuantityKey(1.0) // 1.0 BTC total

	// Cancel order with quantity
	cancelQuantityKey, _ := Float64ToQuantityKey(0.25) // 0.25 BTC

	// Subtract cancelled quantity
	remainingQuantityKey, err := SubtractPriceKeys(totalQuantityKey, cancelQuantityKey)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Order Cancellation:\n")
	fmt.Printf("  Price: $%.2f (key: %d)\n", PriceKeyToFloat64(priceKey), priceKey)
	fmt.Printf("  Total Quantity: %.8f BTC (key: %d)\n", ScaledUint64ToFloat64(totalQuantityKey, QuantityScale), totalQuantityKey)
	fmt.Printf("  Cancelled Quantity: %.8f BTC (key: %d)\n", ScaledUint64ToFloat64(cancelQuantityKey, QuantityScale), cancelQuantityKey)
	fmt.Printf("  Remaining Quantity: %.8f BTC (key: %d)\n", ScaledUint64ToFloat64(remainingQuantityKey, QuantityScale), remainingQuantityKey)
	// Output:
	// Order Cancellation:
	//   Price: $50000.00 (key: 5000000000000)
	//   Total Quantity: 1.00000000 BTC (key: 100000000)
	//   Cancelled Quantity: 0.25000000 BTC (key: 25000000)
	//   Remaining Quantity: 0.75000000 BTC (key: 75000000)
}

// ExampleValidatePriceKey_BoundsChecking demonstrates bounds validation
// Use Case: Validating prices before orderbook operations
func ExampleValidatePriceKey_BoundsChecking() {
	// Test price keys
	testKeys := []uint64{
		5000000000000,     // Valid: $50,000
		MaxPriceKey(),     // Valid: Maximum safe price
		MaxPriceKey() + 1, // Invalid: Exceeds maximum
		0,                 // Valid: Minimum
	}

	fmt.Println("Price Key Validation:")
	for i, key := range testKeys {
		err := ValidatePriceKey(key)
		price := PriceKeyToFloat64(key)
		if err != nil {
			fmt.Printf("  Key %d: ❌ INVALID - $%.2f (key: %d) - %v\n", i, price, key, err)
		} else {
			fmt.Printf("  Key %d: ✅ VALID - $%.2f (key: %d)\n", i, price, key)
		}
	}
}
