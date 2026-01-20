package amount_test

import (
	"math/big"
	"testing"

	"github.com/ezex-io/gopkg/canonical/amount"
	"github.com/stretchr/testify/assert"
)

func TestAmountBitcoin(t *testing.T) {
	tests := []struct {
		name    string
		btc     float64
		satoshi int64
	}{
		{
			name:    "1 Satoshi",
			btc:     0.000_000_01,
			satoshi: 1,
		},
		{
			name:    "1 BTC",
			btc:     1.0,
			satoshi: 1_000_000_00,
		},
		{
			name:    "1.23 BTC",
			btc:     1.23,
			satoshi: 1_230_000_00,
		},
		{
			name:    "123.456 BTC",
			btc:     123.456,
			satoshi: 123_456_000_00,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amt1 := amount.FromBitcoinBtc(tt.btc)
			amt2 := amount.FromBitcoinSatoshi(tt.satoshi)

			assert.Equal(t, "BTC", amt1.Coin())
			assert.Equal(t, tt.btc, amt1.ToBitcoinBtc())
			assert.Equal(t, tt.satoshi, amt1.ToBitcoinSatoshi())

			assert.Equal(t, amt1.Coin(), amt2.Coin())
			assert.Equal(t, amt1.ToBitcoinSatoshi(), amt2.ToBitcoinSatoshi())
			assert.Equal(t, amt1.ToBitcoinBtc(), amt2.ToBitcoinBtc())
		})
	}
}

func TestAmountEthereum(t *testing.T) {
	tests := []struct {
		name string
		eth  float64
		wei  *big.Int
	}{
		{
			name: "1 GWei",
			eth:  1e-9,
			wei:  big.NewInt(1e9),
		},
		{
			name: "1 ETH",
			eth:  1.0,
			wei:  big.NewInt(1e18),
		},
		{
			name: "1.23 ETH",
			eth:  1.23,
			wei:  big.NewInt(1.23e18),
		},
		{
			name: "123.456 ETH",
			eth:  123.456,
			wei:  new(big.Int).Mul(big.NewInt(123456), big.NewInt(1e15)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amt1 := amount.FromEthereumEth(tt.eth)
			amt2 := amount.FromEthereumWei(tt.wei)

			assert.Equal(t, "ETH", amt1.Coin())
			assert.Equal(t, tt.eth, amt1.ToEthereumEth())
			assert.Equal(t, tt.wei, amt1.ToEthereumWei())

			assert.Equal(t, amt1.Coin(), amt2.Coin())
			assert.Equal(t, amt1.ToEthereumWei(), amt2.ToEthereumWei())
			assert.Equal(t, amt1.ToEthereumEth(), amt2.ToEthereumEth())
		})
	}
}

func TestAmountEthereumWeiSmall(t *testing.T) {
	tests := []struct {
		name   string
		weiIn  *big.Int
		weiOut *big.Int
	}{
		{
			weiIn:  big.NewInt(111_222_333_444),
			weiOut: big.NewInt(111_000_000_000),
		},
	}

	for _, tt := range tests {
		amt := amount.FromEthereumWei(tt.weiIn)
		assert.Equal(t, tt.weiOut, amt.ToEthereumWei())
	}
}

func TestAmountPactus(t *testing.T) {
	tests := []struct {
		name    string
		pac     float64
		nanoPac int64
	}{
		{
			name:    "1 Nano PAC",
			pac:     0.000_000_001,
			nanoPac: 1,
		},
		{
			name:    "1 PAC",
			pac:     1.0,
			nanoPac: 1_000_000_000,
		},
		{
			name:    "1.23 PAC",
			pac:     1.23,
			nanoPac: 1_230_000_000,
		},
		{
			name:    "123.456 PAC",
			pac:     123.456,
			nanoPac: 123_456_000_000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amt1 := amount.FromPactusPac(tt.pac)
			amt2 := amount.FromPactusNanoPac(tt.nanoPac)

			assert.Equal(t, "PAC", amt1.Coin())
			assert.Equal(t, tt.pac, amt1.ToPactusPac())
			assert.Equal(t, tt.nanoPac, amt1.ToPactusNanoPac())

			assert.Equal(t, amt1.Coin(), amt2.Coin())
			assert.Equal(t, amt1.ToPactusNanoPac(), amt2.ToPactusNanoPac())
			assert.Equal(t, amt1.ToPactusPac(), amt2.ToPactusPac())
		})
	}
}
