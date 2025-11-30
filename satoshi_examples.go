package safem

import (
	"fmt"
	"math/big"
)

// ============================================================================
// Satoshi Conversion Examples and Use Cases
// ============================================================================
//
// This file demonstrates practical usage patterns for Satoshi (1e8) conversions
// in Morphcore financial exchange operations. Satoshi is the 8-decimal precision
// format used for prices, quantities, and amounts in the exchange.
//
// USAGE SCENARIOS:
// 1. Order Submission and Payload Processing
// 2. Cross-Chain Operations (Wei ↔ Satoshi)
// 3. Price and Quantity Formatting
// 4. API Request/Response Handling
// 5. EIP-712 Signature Payload Construction
//
// ============================================================================

// ExampleDecimalToSatoshi_OrderSubmission demonstrates order submission
// Use Case: Converting address input to satoshi format for EIP-712 payload
func ExampleDecimalToSatoshi_OrderSubmission() {
	// Address input from UI
	orderPrice := 50000.0 // $50,000
	orderQuantity := 0.1  // 0.1 BTC

	// Convert to satoshi format for payload
	priceSatoshi := DecimalToSatoshi(orderPrice)
	quantitySatoshi := DecimalToSatoshi(orderQuantity)

	fmt.Printf("Order Details:\n")
	fmt.Printf("  Price: %s satoshi (%.2f USD)\n", priceSatoshi, orderPrice)
	fmt.Printf("  Quantity: %s satoshi (%.1f BTC)\n", quantitySatoshi, orderQuantity)
	// Output:
	// Order Details:
	//   Price: 5000000000000 satoshi (50000.00 USD)
	//   Quantity: 10000000 satoshi (0.1 BTC)
}

// ExampleSatoshiToDecimal_Display demonstrates display formatting
// Use Case: Converting satoshi values back to decimal for UI display
func ExampleSatoshiToDecimal_Display() {
	// Values from API in satoshi format
	priceSatoshi := "5000000000000" // 50,000 USD
	quantitySatoshi := "10000000"   // 0.1 BTC

	// Convert to decimal for display
	priceDecimal, err := SatoshiToDecimal(priceSatoshi)
	if err != nil {
		fmt.Printf("Price conversion error: %v\n", err)
		return
	}

	quantityDecimal, err := SatoshiToDecimal(quantitySatoshi)
	if err != nil {
		fmt.Printf("Quantity conversion error: %v\n", err)
		return
	}

	fmt.Printf("Display Values:\n")
	fmt.Printf("  Price: $%.2f\n", priceDecimal)
	fmt.Printf("  Quantity: %.8f BTC\n", quantityDecimal)
	// Output:
	// Display Values:
	//   Price: $50000.00
	//   Quantity: 0.10000000 BTC
}

// ExampleWeiToSatoshi_CrossChain demonstrates cross-chain conversion
// Use Case: Converting Ethereum Wei values to Morphcore Satoshi format
func ExampleWeiToSatoshi_CrossChain() {
	// Ethereum token amount in Wei (from Ethereum chain)
	ethereumAmountWei := "50000000000000000000000" // 50,000 tokens in Wei (string format)

	// Convert to Morphcore Satoshi format
	morphcoreAmountSatoshi, err := WeiToSatoshi(ethereumAmountWei)
	if err != nil {
		fmt.Printf("Conversion error: %v\n", err)
		return
	}

	fmt.Printf("Cross-Chain Conversion:\n")
	fmt.Printf("  Ethereum Wei: %s\n", ethereumAmountWei)
	fmt.Printf("  Morphcore Satoshi: %s\n", morphcoreAmountSatoshi)
	// Output:
	// Cross-Chain Conversion:
	//   Ethereum Wei: 50000000000000000000000
	//   Morphcore Satoshi: 5000000000000
}

// ExampleSatoshiToWei_CrossChain demonstrates reverse cross-chain conversion
// Use Case: Converting Morphcore Satoshi to Ethereum Wei for bridge operations
func ExampleSatoshiToWei_CrossChain() {
	// Morphcore amount in Satoshi
	morphcoreAmountSatoshi := "5000000000000" // 50,000 tokens

	// Convert to Ethereum Wei format for bridge
	ethereumAmountWei, err := SatoshiToWei(morphcoreAmountSatoshi)
	if err != nil {
		fmt.Printf("Conversion error: %v\n", err)
		return
	}

	fmt.Printf("Bridge Conversion:\n")
	fmt.Printf("  Morphcore Satoshi: %s\n", morphcoreAmountSatoshi)
	fmt.Printf("  Ethereum Wei: %s\n", ethereumAmountWei)
	// Output:
	// Bridge Conversion:
	//   Morphcore Satoshi: 5000000000000
	//   Ethereum Wei: 50000000000000000000000
}

// ExampleNormalizeToSatoshi_Universal demonstrates universal input normalization
// Use Case: Accepting various input formats and normalizing to satoshi
func ExampleNormalizeToSatoshi_Universal() {
	// Various input formats from different sources
	bigIntSatoshi := big.NewInt(5000000000000)
	bigIntWei, _ := new(big.Int).SetString("50000000000000000000000", 10)
	inputs := []interface{}{
		"50000.0",                 // Decimal string
		50000.0,                   // Float64
		"5000000000000",           // Already satoshi
		"50000000000000000000000", // Wei format (auto-detected)
		bigIntSatoshi,             // *big.Int (satoshi)
		bigIntWei,                 // *big.Int (wei, auto-detected)
	}

	fmt.Println("Input Normalization:")
	for i, input := range inputs {
		satoshi, err := NormalizeToSatoshi(input)
		if err != nil {
			fmt.Printf("  Input %d: Error - %v\n", i, err)
			continue
		}
		fmt.Printf("  Input %d (%T): %s satoshi\n", i, input, satoshi)
	}
	// Output:
	// Input Normalization:
	//   Input 0 (string): 5000000000000 satoshi
	//   Input 1 (float64): 5000000000000 satoshi
	//   Input 2 (string): 5000000000000 satoshi
	//   Input 3 (string): 5000000000000 satoshi
	//   Input 4 (*big.Int): 5000000000000 satoshi
	//   Input 5 (*big.Int): 5000000000000 satoshi
}

// ExampleDecimalToSatoshiBigInt_Precision demonstrates high-precision conversion
// Use Case: Critical financial calculations requiring exact precision
func ExampleDecimalToSatoshiBigInt_Precision() {
	// Critical calculation - must be exact
	criticalAmount := 12345.67890123

	// Convert to *big.Int for exact calculations
	amountSatoshiBig, err := DecimalToSatoshiBigInt(criticalAmount)
	if err != nil {
		fmt.Printf("Conversion error: %v\n", err)
		return
	}

	fmt.Printf("High-Precision Conversion:\n")
	fmt.Printf("  Decimal: %.8f\n", criticalAmount)
	fmt.Printf("  Satoshi (big.Int): %s\n", amountSatoshiBig.String())
	// Output:
	// High-Precision Conversion:
	//   Decimal: 12345.67890123
	//   Satoshi (big.Int): 1234567890123
}

// ExampleSatoshiToDecimalBigFloat_MaxPrecision demonstrates maximum precision
// Use Case: Audit trails, regulatory reporting, exact calculations
func ExampleSatoshiToDecimalBigFloat_MaxPrecision() {
	// Large satoshi value
	satoshiStr := "12345678901234567890"

	// Convert to *big.Float for maximum precision
	decimalBigFloat, err := SatoshiToDecimalBigFloat(satoshiStr)
	if err != nil {
		fmt.Printf("Conversion error: %v\n", err)
		return
	}

	fmt.Printf("Maximum Precision Conversion:\n")
	fmt.Printf("  Satoshi: %s\n", satoshiStr)
	fmt.Printf("  Decimal (big.Float): %s\n", decimalBigFloat.Text('f', 8))
	// Output:
	// Maximum Precision Conversion:
	//   Satoshi: 12345678901234567890
	//   Decimal (big.Float): 123456789012.34567890
}

// ExampleBatchDecimalToSatoshi_Orderbook demonstrates batch processing
// Use Case: Converting orderbook prices/quantities efficiently
func ExampleBatchDecimalToSatoshi_Orderbook() {
	// Orderbook price levels
	prices := []float64{
		50000.0,
		50001.0,
		50002.0,
		50003.0,
		50004.0,
	}

	// Batch convert to satoshi
	pricesSatoshi := BatchDecimalToSatoshi(prices)

	fmt.Println("Orderbook Prices (Satoshi):")
	for i, price := range prices {
		fmt.Printf("  Level %d: %s satoshi (%.2f USD)\n", i, pricesSatoshi[i], price)
	}
	// Output:
	// Orderbook Prices (Satoshi):
	//   Level 0: 5000000000000 satoshi (50000.00 USD)
	//   Level 1: 5000100000000 satoshi (50001.00 USD)
	//   Level 2: 5000200000000 satoshi (50002.00 USD)
	//   Level 3: 5000300000000 satoshi (50003.00 USD)
	//   Level 4: 5000400000000 satoshi (50004.00 USD)
}

// ExampleBatchSatoshiToDecimal_API demonstrates batch API response formatting
// Use Case: Converting multiple satoshi values for API responses
func ExampleBatchSatoshiToDecimal_API() {
	// API data in satoshi format
	satoshiValues := []string{
		"5000000000000",
		"10000000",
		"250000000000",
		"1234567890123",
	}

	// Batch convert to decimal
	decimalValues, err := BatchSatoshiToDecimal(satoshiValues)
	if err != nil {
		fmt.Printf("Batch conversion error: %v\n", err)
		return
	}

	fmt.Println("API Response (Decimal):")
	for i, satoshi := range satoshiValues {
		fmt.Printf("  Value %d: %s satoshi = %.8f\n", i, satoshi, decimalValues[i])
	}
	// Output:
	// API Response (Decimal):
	//   Value 0: 5000000000000 satoshi = 50000.00000000
	//   Value 1: 10000000 satoshi = 0.10000000
	//   Value 2: 250000000000 satoshi = 2500.00000000
	//   Value 3: 1234567890123 satoshi = 12345.67890123
}

// ExampleNormalizeToSatoshi_PayloadProcessing demonstrates payload processing
// Use Case: Processing EIP-712 payload with various input formats
func ExampleNormalizeToSatoshi_PayloadProcessing() {
	// Payload data from different sources
	payloadData := map[string]interface{}{
		"price":    "50000.0",               // Address input (decimal string)
		"quantity": 0.1,                     // Address input (float64)
		"amount":   "5000000000000",         // Already satoshi
		"fee":      "500000000000000000000", // Wei format (from Ethereum)
	}

	// Normalize all values to satoshi
	normalizedPayload := make(map[string]string)
	for key, value := range payloadData {
		satoshi, err := NormalizeToSatoshi(value)
		if err != nil {
			fmt.Printf("Error normalizing %s: %v\n", key, err)
			continue
		}
		normalizedPayload[key] = satoshi
	}

	fmt.Println("Normalized Payload (Satoshi):")
	for key, satoshi := range normalizedPayload {
		fmt.Printf("  %s: %s\n", key, satoshi)
	}
	// Output:
	// Normalized Payload (Satoshi):
	//   price: 5000000000000
	//   quantity: 10000000
	//   amount: 5000000000000
	//   fee: 50000000000
}

// ExampleWeiToSatoshi_TokenBridge demonstrates token bridge operations
// Use Case: Bridging tokens from Ethereum to Morphcore
func ExampleWeiToSatoshi_TokenBridge() {
	// Address wants to bridge tokens from Ethereum
	ethereumBalanceWei := "1000000000000000000000" // 1000 tokens in Wei

	// Convert to Morphcore format
	morphcoreBalanceSatoshi, err := WeiToSatoshi(ethereumBalanceWei)
	if err != nil {
		fmt.Printf("Bridge conversion error: %v\n", err)
		return
	}

	fmt.Printf("Token Bridge:\n")
	fmt.Printf("  Ethereum: %s Wei\n", ethereumBalanceWei)
	fmt.Printf("  Morphcore: %s Satoshi\n", morphcoreBalanceSatoshi)
	// Output:
	// Token Bridge:
	//   Ethereum: 1000000000000000000000 Wei
	//   Morphcore: 100000000000 Satoshi
}

// ExampleSatoshiToWei_Withdrawal demonstrates withdrawal to Ethereum
// Use Case: Withdrawing tokens from Morphcore to Ethereum
func ExampleSatoshiToWei_Withdrawal() {
	// Address wants to withdraw from Morphcore
	morphcoreBalanceSatoshi := "100000000000" // 1000 tokens

	// Convert to Wei for Ethereum transaction
	ethereumAmountWei, err := SatoshiToWei(morphcoreBalanceSatoshi)
	if err != nil {
		fmt.Printf("Withdrawal conversion error: %v\n", err)
		return
	}

	fmt.Printf("Withdrawal to Ethereum:\n")
	fmt.Printf("  Morphcore: %s Satoshi\n", morphcoreBalanceSatoshi)
	fmt.Printf("  Ethereum: %s Wei\n", ethereumAmountWei)
	// Output:
	// Withdrawal to Ethereum:
	//   Morphcore: 100000000000 Satoshi
	//   Ethereum: 1000000000000000000000 Wei
}

// ExampleDecimalToSatoshi_PriceFormatting demonstrates price formatting
// Use Case: Formatting prices for different markets
func ExampleDecimalToSatoshi_PriceFormatting() {
	// Market prices
	marketPrices := map[string]float64{
		"BTC-USD":  50000.0,
		"ETH-USD":  3000.0,
		"SOL-USD":  100.0,
		"USDC-USD": 1.0,
	}

	fmt.Println("Market Prices (Satoshi):")
	for market, price := range marketPrices {
		priceSatoshi := DecimalToSatoshi(price)
		fmt.Printf("  %s: %s satoshi ($%.2f)\n", market, priceSatoshi, price)
	}
	// Output:
	// Market Prices (Satoshi):
	//   BTC-USD: 5000000000000 satoshi ($50000.00)
	//   ETH-USD: 300000000000 satoshi ($3000.00)
	//   SOL-USD: 10000000000 satoshi ($100.00)
	//   USDC-USD: 100000000 satoshi ($1.00)
}

// ExampleNormalizeToSatoshi_ErrorHandling demonstrates error handling
// Use Case: Robust error handling in production code
func ExampleNormalizeToSatoshi_ErrorHandling() {
	// Test cases with potential errors
	testCases := []struct {
		name  string
		value interface{}
		valid bool
	}{
		{"Valid decimal string", "50000.0", true},
		{"Valid float64", 50000.0, true},
		{"Valid satoshi string", "5000000000000", true},
		{"Valid wei string", "50000000000000000000000", true},
		{"Empty string", "", false},
		{"Invalid string", "invalid", false},
		{"Negative float", -100.0, false},
		{"Nil big.Int", (*big.Int)(nil), false},
	}

	fmt.Println("Error Handling Tests:")
	for _, tc := range testCases {
		satoshi, err := NormalizeToSatoshi(tc.value)
		if err != nil {
			if tc.valid {
				fmt.Printf("  ❌ %s: Unexpected error - %v\n", tc.name, err)
			} else {
				fmt.Printf("  ✅ %s: Correctly rejected - %v\n", tc.name, err)
			}
		} else {
			if tc.valid {
				fmt.Printf("  ✅ %s: %s satoshi\n", tc.name, satoshi)
			} else {
				fmt.Printf("  ❌ %s: Should have been rejected\n", tc.name)
			}
		}
	}
}
