package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const basefeeWiggleMultiplier = 2

type ContractGasEstimator interface {
	ethereum.GasEstimator
	ethereum.GasPricer
	ethereum.GasPricer1559

	// HeaderByNumber returns a block header from the current canonical chain. If
	// number is nil, the latest known header is returned.
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

type GasInfo struct {
	EstimatedGasLimit uint64
	BaseFee           *big.Int // baseFeePerGas
	PriorityFee       *big.Int // maxPriorityFeePerGas
}

// EffectiveGasPrice returns the actual price per gas that will be paid
// assuming maxFeePerGas is sufficiently high.
func (g *GasInfo) EffectiveGasPrice() *big.Int {
	return new(big.Int).Add(g.BaseFee, g.PriorityFee)
}

// MaxFeePerGas returns a safe maxFeePerGas value with headroom
// to tolerate short-term base fee increases.
func (g *GasInfo) MaxFeePerGas() *big.Int {
	return new(big.Int).Add(
		g.PriorityFee,
		new(big.Int).Mul(g.BaseFee, big.NewInt(basefeeWiggleMultiplier)),
	)
}

// EstimateGasCost estimates the total transaction cost using the
// effective gas price (not the max fee).
func (g *GasInfo) EstimateGasCost() *big.Int {
	return new(big.Int).Mul(
		new(big.Int).SetUint64(g.EstimatedGasLimit),
		g.EffectiveGasPrice(),
	)
}

// GasEstimator provides gas estimation functionality for EVM contract calls.
type GasEstimator struct {
	client       ContractGasEstimator
	contractAddr common.Address
	abi          *abi.ABI
}

// NewGasEstimator creates a new EVM gas estimator.
func NewGasEstimator(client ContractGasEstimator, contractAddr common.Address, abi *abi.ABI) *GasEstimator {
	return &GasEstimator{
		client:       client,
		contractAddr: contractAddr,
		abi:          abi,
	}
}

// EstimateGasParams estimates the gas parameters for a contract method call.
// It has 3 RPC calls:
// 1. eth_estimateGas
// 2. eth_maxPriorityFeePerGas
// 3. eth_getBlockByNumber.
func (e *GasEstimator) EstimateGasParams(
	ctx context.Context,
	method string,
	from common.Address,
	args ...any,
) (*GasInfo, error) {
	// ABI-encode the contract method call and its arguments.
	data, err := e.abi.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	// Construct a call message for gas estimation.
	// This simulates a transaction sent from `from` to the contract
	// without broadcasting it to the network.
	msg := ethereum.CallMsg{
		To:   &e.contractAddr,
		From: from,
		Data: data,
	}

	// Ask the client to estimate the gas required for the simulated call.
	gasLimit, err := e.client.EstimateGas(ctx, msg)
	if err != nil {
		return nil, err
	}

	priorityFee, err := e.client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}

	head, err := e.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &GasInfo{
		EstimatedGasLimit: gasLimit,
		BaseFee:           head.BaseFee,
		PriorityFee:       priorityFee,
	}, nil
}

// SuggestGasPrice returns the node's suggested gas price for legacy (pre-EIP-1559)
// transactions, intended to achieve timely inclusion in a block.
//
// This is backed by the `eth_gasPrice` RPC call.
func (e *GasEstimator) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return e.client.SuggestGasPrice(ctx)
}
