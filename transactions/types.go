package transactions

import "strings"

//go:generate stringer -type=Type
type Type int

const (
	Unknown Type = iota
	Raw
	FtSetup
	FtDeposit
	FtWithdrawal
	NftSetup
	NftDeposit
	NftWithdrawal
)

func (s Type) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Type) UnmarshalText(text []byte) error {
	*s = StatusFromText(string(text))
	return nil
}

func StatusFromText(text string) Type {
	switch strings.ToLower(text) {
	default:
		return Unknown
	case "general":
		return Raw
	case "ftsetup":
		return FtSetup
	case "ftdeposit":
		return FtDeposit
	case "ftwithdrawal":
		return FtWithdrawal
	case "nftsetup":
		return NftSetup
	case "nftdeposit":
		return NftDeposit
	case "nftwithdrawal":
		return NftWithdrawal
	}
}
