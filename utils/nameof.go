package utils

import (
	"fmt"

	"github.com/deamchain/deam-v2-base/hash"
	"github.com/deamchain/deam-v2-base/inter/idx"
)

// NameOf returns human readable string representation.
func NameOf(p idx.ValidatorID) string {
	if name := hash.GetNodeName(p); len(name) > 0 {
		return name
	}

	return fmt.Sprintf("%d", p)
}
