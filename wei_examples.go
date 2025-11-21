package safem

import (
	"fmt"
	"math/big"
)

// ============================================================================
// Wei Conversion Examples and Use Cases
// ============================================================================
//
// This file demonstrates practical usage patterns for Wei (1e18) conversions
// in Ethereum blockchain operations. Wei is the smallest unit of Ether,
// similar to how satoshi is the smallest unit of Bitcoin.
//
// USAGE SCENARIOS:
// 1. Token Transfer Operations
// 2. Balance Display and Formatting
// 3. Gas Fee Calculations
// 4. Smart Contract Interactions
// 5. API Request/Response Formatting
//
// ============================================================================

// ExampleWeiToEther_Basic demonstrates basic Wei to Ether conversion
// Use Case: Displaying user balance in human-readable format
func ExampleWeiToEther_Basic() {
	// User's balance in Wei (from blockchain)
	balanceWei := big.NewInt(1500000000000000000) // 1.5 ETH

	// Convert to Ether for display
	balanceEth, err := WeiToEther(balanceWei)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Balance: %.18f ETH\n", balanceEth)
	// Output: Balance: 1.500000000000000000 ETH
}

// ExampleWeiToEtherOptimized_HighFrequency demonstrates optimized conversion
// Use Case: High-frequency operations like orderbook processing
func ExampleWeiToEtherOptimized_HighFrequency() {
	// Processing many small Wei values in orderbook
	orderAmounts := []*big.Int{
		big.NewInt(100000000000000000), // 0.1 ETH
		big.NewInt(50000000000000000),  // 0.05 ETH
		big.NewInt(25000000000000000),  // 0.025 ETH
	}

	fmt.Println("Order amounts in Ether:")
	for i, amountWei := range orderAmounts {
		amountEth, err := WeiToEtherOptimized(amountWei)
		if err != nil {
			fmt.Printf("Order %d: Error - %v\n", i, err)
			continue
		}
		fmt.Printf("Order %d: %.18f ETH\n", i, amountEth)
	}
	// Output:
	// Order amounts in Ether:
	// Order 0: 0.100000000000000000 ETH
	// Order 1: 0.050000000000000000 ETH
	// Order 2: 0.025000000000000000 ETH
}

// ExampleWeiToEtherSafe_Critical demonstrates safe conversion for critical operations
// Use Case: Settlement systems, accounting, regulatory reporting
func ExampleWeiToEtherSafe_Critical() {
	// Critical financial calculation - must be precise
	settlementAmountWei, _ := new(big.Int).SetString("1000000000000000000000", 10) // 1000 ETH

	// Use safe version for maximum precision
	settlementAmountEth, err := WeiToEtherSafe(settlementAmountWei)
	if err != nil {
		fmt.Printf("Settlement error: %v\n", err)
		return
	}

	fmt.Printf("Settlement amount: %.18f ETH\n", settlementAmountEth)
	// Output: Settlement amount: 1000.000000000000000000 ETH
}

// ExampleEtherToWei_UserInput demonstrates converting user input to Wei
// Use Case: Processing user input for token transfers
func ExampleEtherToWei_UserInput() {
	// User enters amount in Ether (from UI or API)
	userInputEth := 2.5

	// Convert to Wei for blockchain transaction
	amountWei, err := EtherToWei(userInputEth)
	if err != nil {
		fmt.Printf("Conversion error: %v\n", err)
		return
	}

	fmt.Printf("User input: %.1f ETH\n", userInputEth)
	fmt.Printf("Transaction amount: %s Wei\n", amountWei.String())
	// Output:
	// User input: 2.5 ETH
	// Transaction amount: 2500000000000000000 Wei
}

// ExampleEtherToWei_OrderSubmission demonstrates order submission
// Use Case: Converting order quantities for DEX order submission
func ExampleEtherToWei_OrderSubmission() {
	// Order details from user
	orderQuantityEth := 0.1 // 0.1 ETH
	orderPriceEth := 2000.0 // 2000 ETH per token

	// Convert to Wei for EIP-712 signature
	quantityWei, err := EtherToWei(orderQuantityEth)
	if err != nil {
		fmt.Printf("Quantity conversion error: %v\n", err)
		return
	}

	priceWei, err := EtherToWei(orderPriceEth)
	if err != nil {
		fmt.Printf("Price conversion error: %v\n", err)
		return
	}

	fmt.Printf("Order Quantity: %s Wei (%.1f ETH)\n", quantityWei.String(), orderQuantityEth)
	fmt.Printf("Order Price: %s Wei (%.1f ETH)\n", priceWei.String(), orderPriceEth)
	// Output:
	// Order Quantity: 100000000000000000 Wei (0.1 ETH)
	// Order Price: 2000000000000000000000 Wei (2000.0 ETH)
}

// ExampleWeiToEther_BalanceCheck demonstrates balance checking
// Use Case: Checking if user has sufficient balance for transaction
func ExampleWeiToEther_BalanceCheck() {
	// User's current balance (from blockchain)
	userBalanceWei := big.NewInt(5000000000000000000) // 5 ETH

	// Required amount for transaction
	requiredEth := 2.5
	requiredWei, err := EtherToWei(requiredEth)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Check if sufficient balance
	if userBalanceWei.Cmp(requiredWei) >= 0 {
		userBalanceEth, _ := WeiToEther(userBalanceWei)
		fmt.Printf("✅ Sufficient balance: %.18f ETH (required: %.1f ETH)\n",
			userBalanceEth, requiredEth)
	} else {
		userBalanceEth, _ := WeiToEther(userBalanceWei)
		fmt.Printf("❌ Insufficient balance: %.18f ETH (required: %.1f ETH)\n",
			userBalanceEth, requiredEth)
	}
	// Output: ✅ Sufficient balance: 5.000000000000000000 ETH (required: 2.5 ETH)
}

// ExampleWeiToEther_GasFeeCalculation demonstrates gas fee calculations
// Use Case: Calculating and displaying gas fees
func ExampleWeiToEther_GasFeeCalculation() {
	// Gas price in Wei (from network)
	gasPriceWei := big.NewInt(20000000000) // 20 Gwei
	gasLimit := uint64(21000)              // Standard transfer

	// Calculate total gas fee
	totalGasWei := new(big.Int).Mul(gasPriceWei, big.NewInt(int64(gasLimit)))

	// Convert to Ether for display
	totalGasEth, err := WeiToEther(totalGasWei)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Gas Price: %s Wei\n", gasPriceWei.String())
	fmt.Printf("Gas Limit: %d\n", gasLimit)
	fmt.Printf("Total Gas Fee: %.18f ETH\n", totalGasEth)
	// Output:
	// Gas Price: 20000000000 Wei
	// Gas Limit: 21000
	// Total Gas Fee: 0.000420000000000000 ETH
}

// ExampleEtherToWei_TokenTransfer demonstrates token transfer preparation
// Use Case: Preparing token transfer transaction
func ExampleEtherToWei_TokenTransfer() {
	// Transfer details
	recipient := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
	amountEth := 1.5

	// Convert to Wei
	amountWei, err := EtherToWei(amountEth)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Prepare transaction (pseudo-code)
	fmt.Printf("Preparing transfer:\n")
	fmt.Printf("  To: %s\n", recipient)
	fmt.Printf("  Amount: %s Wei (%.1f ETH)\n", amountWei.String(), amountEth)
	// Output:
	// Preparing transfer:
	//   To: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
	//   Amount: 1500000000000000000 Wei (1.5 ETH)
}

// ExampleWeiToEther_BatchProcessing demonstrates batch processing
// Use Case: Processing multiple balances efficiently
func ExampleWeiToEther_BatchProcessing() {
	// Multiple user balances from database
	balance1, _ := new(big.Int).SetString("1000000000000000000", 10)  // 1 ETH
	balance2, _ := new(big.Int).SetString("500000000000000000", 10)   // 0.5 ETH
	balance3, _ := new(big.Int).SetString("2500000000000000000", 10)  // 2.5 ETH
	balance4, _ := new(big.Int).SetString("10000000000000000000", 10) // 10 ETH
	balancesWei := []*big.Int{balance1, balance2, balance3, balance4}

	fmt.Println("User Balances:")
	for i, balanceWei := range balancesWei {
		// Use optimized version for performance
		balanceEth, err := WeiToEtherOptimized(balanceWei)
		if err != nil {
			fmt.Printf("User %d: Error - %v\n", i, err)
			continue
		}
		fmt.Printf("User %d: %.18f ETH\n", i, balanceEth)
	}
	// Output:
	// User Balances:
	// User 0: 1.000000000000000000 ETH
	// User 1: 0.500000000000000000 ETH
	// User 2: 2.500000000000000000 ETH
	// User 3: 10.000000000000000000 ETH
}

// ExampleWeiToEther_ErrorHandling demonstrates error handling
// Use Case: Robust error handling in production code
func ExampleWeiToEther_ErrorHandling() {
	// Test cases with potential errors
	testCases := []struct {
		name  string
		wei   *big.Int
		valid bool
	}{
		{"Valid amount", big.NewInt(1000000000000000000), true},
		{"Zero", big.NewInt(0), true},
		{"Negative (invalid)", big.NewInt(-1), false},
		{"Nil (invalid)", nil, false},
		{"Very large", new(big.Int).Lsh(big.NewInt(1), 200), true},
	}

	for _, tc := range testCases {
		eth, err := WeiToEther(tc.wei)
		if err != nil {
			if tc.valid {
				fmt.Printf("❌ %s: Unexpected error - %v\n", tc.name, err)
			} else {
				fmt.Printf("✅ %s: Correctly rejected - %v\n", tc.name, err)
			}
		} else {
			if tc.valid {
				fmt.Printf("✅ %s: %.18f ETH\n", tc.name, eth)
			} else {
				fmt.Printf("❌ %s: Should have been rejected\n", tc.name)
			}
		}
	}
}

// ExampleEtherToWei_PrecisionHandling demonstrates precision handling
// Use Case: Handling floating-point precision issues
func ExampleEtherToWei_PrecisionHandling() {
	// Small amounts that might have precision issues
	smallAmounts := []float64{
		0.000000000000000001, // 1 Wei
		0.0000000000000001,   // 100 Wei
		0.000000000001,       // 1 Gwei
		0.000000001,          // 1 nanoETH
	}

	fmt.Println("Small amount conversions:")
	for _, amountEth := range smallAmounts {
		amountWei, err := EtherToWei(amountEth)
		if err != nil {
			fmt.Printf("  %.18f ETH: Error - %v\n", amountEth, err)
		} else {
			fmt.Printf("  %.18f ETH: %s Wei\n", amountEth, amountWei.String())
		}
	}
}

// ExampleWeiToEther_APIDisplay demonstrates API response formatting
// Use Case: Formatting blockchain data for API responses
func ExampleWeiToEther_APIDisplay() {
	// Blockchain data in Wei
	blockchainData := map[string]*big.Int{
		"balance":   big.NewInt(1500000000000000000),
		"pending":   big.NewInt(500000000000000000),
		"available": big.NewInt(1000000000000000000),
	}

	// Convert to Ether for API response
	apiResponse := make(map[string]float64)
	for key, valueWei := range blockchainData {
		valueEth, err := WeiToEther(valueWei)
		if err != nil {
			fmt.Printf("Error converting %s: %v\n", key, err)
			continue
		}
		apiResponse[key] = valueEth
	}

	fmt.Println("API Response (Ether):")
	for key, value := range apiResponse {
		fmt.Printf("  %s: %.18f ETH\n", key, value)
	}
	// Output:
	// API Response (Ether):
	//   balance: 1.500000000000000000 ETH
	//   pending: 0.500000000000000000 ETH
	//   available: 1.000000000000000000 ETH
}
