package monetary

import (
	"errors"
	"math/big"
	"strings"
)

type decimal struct {
	unscaled *big.Int
	scale    int
}

//nolint:gocyclo // Exact lexical decimal validation is intentionally explicit.
func parseDecimal(value string) (decimal, error) {
	if value == "" {
		return decimal{}, errors.New("empty decimal")
	}
	sign := 1
	if value[0] == '-' || value[0] == '+' {
		if value[0] == '-' {
			sign = -1
		}
		value = value[1:]
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" || (len(parts) == 2 && parts[1] == "") {
		return decimal{}, errors.New("invalid decimal")
	}
	for _, part := range parts {
		for _, r := range part {
			if r < '0' || r > '9' {
				return decimal{}, errors.New("invalid decimal")
			}
		}
	}
	scale := 0
	digits := parts[0]
	if len(parts) == 2 {
		scale = len(parts[1])
		digits += parts[1]
	}
	unscaled, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return decimal{}, errors.New("invalid decimal")
	}
	if sign < 0 {
		unscaled.Neg(unscaled)
	}
	return decimal{unscaled: unscaled, scale: scale}, nil
}

func zeroDecimal(scale int) decimal {
	return decimal{unscaled: new(big.Int), scale: scale}
}

func (d decimal) rescale(scale int) decimal {
	if scale <= d.scale {
		return decimal{unscaled: new(big.Int).Set(d.unscaled), scale: d.scale}
	}
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale-d.scale)), nil)
	return decimal{unscaled: new(big.Int).Mul(d.unscaled, factor), scale: scale}
}

func (d decimal) add(other decimal) decimal {
	scale := max(d.scale, other.scale)
	left := d.rescale(scale)
	right := other.rescale(scale)
	return decimal{unscaled: new(big.Int).Add(left.unscaled, right.unscaled), scale: scale}
}

func (d decimal) subtract(other decimal) decimal {
	scale := max(d.scale, other.scale)
	left := d.rescale(scale)
	right := other.rescale(scale)
	return decimal{unscaled: new(big.Int).Sub(left.unscaled, right.unscaled), scale: scale}
}

func (d decimal) abs() decimal {
	return decimal{unscaled: new(big.Int).Abs(d.unscaled), scale: d.scale}
}

func (d decimal) cmp(other decimal) int {
	scale := max(d.scale, other.scale)
	return d.rescale(scale).unscaled.Cmp(other.rescale(scale).unscaled)
}

func (d decimal) stringAtLeast(minScale int) string {
	scale := max(d.scale, minScale)
	value := d.rescale(scale).unscaled
	sign := ""
	if value.Sign() < 0 {
		sign = "-"
		value = new(big.Int).Abs(value)
	}
	digits := value.String()
	if scale == 0 {
		return sign + digits
	}
	if len(digits) <= scale {
		digits = strings.Repeat("0", scale-len(digits)+1) + digits
	}
	cut := len(digits) - scale
	return sign + digits[:cut] + "." + digits[cut:]
}
