package templates

import "strings"

//go:generate stringer -type=TokenType
type TokenType int

const (
	NotSpecified TokenType = iota
	FT
	NFT
)

func (s TokenType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *TokenType) UnmarshalText(text []byte) error {
	*s = TypeFromText(string(text))
	return nil
}

func TypeFromText(text string) TokenType {
	switch strings.ToLower(text) {
	default:
		return NotSpecified
	case "ft":
		return FT
	case "nft":
		return NFT
	}
}
