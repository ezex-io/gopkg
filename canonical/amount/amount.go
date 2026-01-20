package amount

import "math/big"

// Amount is a canonical representation of an amount of a coin.
// The value is canonical and the smallest unit of the coin is 10^-9.
type Amount struct {
	// value is the amount of the coin in non-floating format.
	// The value is canonical and the smallest unit of the coin is 10^-9.
	value int64
	// coin is the name of the coin, e.g. "BTC", "ETH", "PAC", etc.
	// coin is used to identify the coin.
	coin string
}

func NewAmount(value int64, coin string) Amount {
	return Amount{
		value: value,
		coin:  coin,
	}
}

func FromBitcoinBtc(btc float64) Amount {
	return NewAmount(int64(btc*1e8), "BTC")
}

func FromBitcoinSatoshi(satoshi int64) Amount {
	return NewAmount(satoshi, "BTC")
}

func FromEthereumEth(ether float64) Amount {
	return NewAmount(int64(ether*1e9), "ETH")
}

func FromEthereumWei(wei *big.Int) Amount {
	return NewAmount(new(big.Int).Div(wei, big.NewInt(1e9)).Int64(), "ETH")
}

func FromPactusPac(pac float64) Amount {
	return NewAmount(int64(pac*1e9), "PAC")
}

func FromPactusNanoPac(nanoPac int64) Amount {
	return NewAmount(nanoPac, "PAC")
}

func (a *Amount) Coin() string {
	return a.coin
}

func (a *Amount) ToBitcoinBtc() float64 {
	return float64(a.value) / 1e8
}

func (a *Amount) ToBitcoinSatoshi() int64 {
	return a.value
}

func (a *Amount) ToEthereumEth() float64 {
	return float64(a.value) / 1e9
}

func (a *Amount) ToEthereumWei() *big.Int {
	return new(big.Int).Mul(big.NewInt(a.value), big.NewInt(1e9))
}

func (a *Amount) ToPactusPac() float64 {
	return float64(a.value) / 1e9
}

func (a *Amount) ToPactusNanoPac() int64 {
	return a.value
}
