package utils

import "math/big"

// ToUnit number of  to Wei
func ToUnit(value uint64) *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(value), big.NewInt(1e18))
}
