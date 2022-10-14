package types

type TokenType struct {
	StructTag StructTag
}

type CoinInfo struct {
	Name      string
	Decimals  int
	Symbol    string
	TokenType TokenType
}

func (t TokenType) FullName() string {
	return t.StructTag.GetFullName()
}

func (t TokenType) ToTypeTag() string {
	return t.FullName()
}
