package types

import "math/big"

type TokenType interface {
	GetFullName() string
}

type CoinInfo struct {
	Name      string
	Decimals  int
	Symbol    string
	TokenType TokenType
}

type Coin struct {
	Value *big.Int
}
