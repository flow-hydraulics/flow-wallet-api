package template_strings

import "fmt"

func GetByName(name string) (string, error) {
	switch name {
	default:
		return "", fmt.Errorf("could not find template for %s", name)
	case "FUSD":
		return FUSD, nil
	case "ExampleNFT":
		return ExampleNFT, nil
	}
}
