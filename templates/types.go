package templates

import "strings"

//go:generate stringer -type=TokenType
type TokenType int

const (
	Unknown TokenType = iota
	FT
	NFT
)

func (s TokenType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *TokenType) UnmarshalText(text []byte) error {
	*s = StatusFromText(string(text))
	return nil
}

func StatusFromText(text string) TokenType {
	switch strings.ToLower(text) {
	default:
		return Unknown
	case "ft":
		return FT
	case "nft":
		return NFT
	}
}
