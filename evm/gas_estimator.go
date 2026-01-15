package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// GasEstimator provides gas estimation functionality for EVM contract calls.
type GasEstimator struct {
	client       Client         // Ethereum client for blockchain interaction
	contractAddr common.Address // Address of the smart contract
	abi          *abi.ABI       // Contract ABI for encoding function calls
}

// NewGasEstimator creates a new EVM gas estimator.
func NewGasEstimator(client Client, contractAddr common.Address, abi *abi.ABI) *GasEstimator {
	return &GasEstimator{
		client:       client,
		contractAddr: contractAddr,
		abi:          abi,
	}
}

// EstimateGasLimit estimates the gas limit required for a function call.
func (e *GasEstimator) EstimateGasLimit(ctx context.Context, method string, args ...any) (uint64, error) {
	// Pack the function call data
	data, err := e.abi.Pack(method, args...)
	if err != nil {
		return 0, err
	}

	// Create the call message for gas estimation
	msg := ethereum.CallMsg{
		To:   &e.contractAddr, // Use contract address
		Data: data,
	}

	gasLimit, err := e.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, err
	}

	return gasLimit, nil
}

// EstimateGasCost estimates the total gas cost in wei for a function call.
func (e *GasEstimator) EstimateGasCost(ctx context.Context, method string, args ...any) (*big.Int, error) {
	// Get gas limit
	gasLimit, err := e.EstimateGasLimit(ctx, method, args...)
	if err != nil {
		return big.NewInt(0), err
	}

	// Get current gas price
	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		return big.NewInt(0), err
	}

	// Calculate total cost: gasLimit * gasPrice
	totalCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)

	return totalCost, nil
}
