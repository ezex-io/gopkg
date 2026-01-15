package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
)

type Client interface {
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
}
