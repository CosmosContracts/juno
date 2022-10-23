package types

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// String implements fmt.Stringer interface
func (d Denom) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

// Equal implements equal interface
func (d Denom) Equal(d1 *Denom) bool {
	return d.BaseDenom == d1.BaseDenom &&
		d.SymbolDenom == d1.SymbolDenom &&
		d.Exponent == d1.Exponent
}

// DenomList is array of Denom
type DenomList []Denom

// String implements fmt.Stringer interface
func (dl DenomList) String() (out string) {
	for _, d := range dl {
		out += d.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// Contains checks whether or not a SymbolDenom (e.g. UMEE) is in the DenomList
func (dl DenomList) Contains(symbolDenom string) bool {
	for _, d := range dl {
		if strings.EqualFold(d.SymbolDenom, symbolDenom) {
			return true
		}
	}
	return false
}
