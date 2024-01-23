package utils

import (
	"fmt"
	"math/big"
)

// CheckAndParseBigInt parses a string into a big.Int.
func CheckAndParseBigInt(value string) (*big.Int, error) {
	if value == "" {
		return big.NewInt(0), nil
	}
	result, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, fmt.Errorf("invalid int string: %s", value)
	}

	return result, nil
}

// ParseBigInt parses a string into a big.Int. The function panics if the string is invalid.
func ParseBigInt(value string) *big.Int {
	result, err := CheckAndParseBigInt(value)
	if err != nil {
		panic(err)
	}

	return result
}

// CheckAndParseBigFloat parses a string into a big.Float.
func CheckAndParseBigFloat(value string) (*big.Float, error) {
	result, ok := new(big.Float).SetString(value)
	if !ok {
		return nil, fmt.Errorf("invalid float string: %s", value)
	}

	return result, nil
}

// ParseBigFloat parses a string into a big.Float, returning 0 if the string is invalid.
func ParseBigFloat(value string) *big.Float {
	if value == "" {
		return big.NewFloat(0)
	}
	result, err := CheckAndParseBigFloat(value)
	if err != nil {
		panic(err)
	}

	return result
}

// Equals returns true if `x` and `y` are equal.
func Equals(x, y *big.Int) bool {
	return x.Cmp(y) == 0
}

// GreaterThan returns true if `x` is greater than `y`.
func GreaterThan(x, y *big.Int) bool {
	return x.Cmp(y) > 0
}

// GreaterThanOrEqual returns true if `x` is greater than or equal to `y`.
func GreaterThanOrEqual(x, y *big.Int) bool {
	return x.Cmp(y) >= 0
}

// LessThan returns true if `x` is less than `y`.
func LessThan(x, y *big.Int) bool {
	return x.Cmp(y) < 0
}

// LessThanOrEqual returns true if `x` is less than or equal to `y`.
func LessThanOrEqual(x, y *big.Int) bool {
	return x.Cmp(y) <= 0
}

// FloatGreaterThan returns true if `x` is greater than `y`.
func FloatGreaterThan(x, y *big.Float) bool {
	return x.Cmp(y) > 0
}

// FloatGreaterThanOrEqual returns true if `x` is greater than or equal to `y`.
func FloatGreaterThanOrEqual(x, y *big.Float) bool {
	return x.Cmp(y) >= 0
}

// FloatLessThan returns true if `x` is less than `y`.
func FloatLessThan(x, y *big.Float) bool {
	return x.Cmp(y) < 0
}

// FloatLessThanOrEqual returns true if `x` is less than or equal to `y`.
func FloatLessThanOrEqual(x, y *big.Float) bool {
	return x.Cmp(y) <= 0
}
