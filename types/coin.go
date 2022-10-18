package types

type TokenType interface {
	GetFullName() string
}

type CoinInfo struct {
	Name      string
	Decimals  int
	Symbol    string
	TokenType TokenType
}
