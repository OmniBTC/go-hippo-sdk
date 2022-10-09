package types

type TokenType struct {
	Symbol string
}

type CoinInfo struct {
	// TODO
	TokenType TokenType
}

func (t TokenType) FullName() string {
	panic("todo")
	return ""
}

func (t TokenType) ToTypeTag() string {
	panic("todo")
	return ""
}
