package common

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Amount can be used to convert from fixed point to integer.
// Amount assumes the smallest divisible unit is 1e-8
type Amount int64

func (a Amount) AsInt64() int64 {
	return int64(a)
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(AmountToString(int64(a))), nil
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	v, err := StringToAmount(string(data))
	if err != nil {
		return err
	}
	*a = Amount(v)
	return nil
}

func FloatToAmount(f float64) int64 {
	return int64(f * 1e8)
}

// StringToAmount takes a number in decimal form, and converts it to
// an integer * 1e8. Does not accept negative values
func StringToAmount(amt string) (int64, error) {
	valid := regexp.MustCompile(`^([0-9]+)?(\.[0-9]+)?$`)
	if !valid.MatchString(amt) {
		return 0, fmt.Errorf("improper string")
	}

	var total int64 = 0

	dot := regexp.MustCompile(`\.`)
	pieces := dot.Split(amt, 2)
	whole, _ := strconv.Atoi(pieces[0])
	total += int64(whole) * 1e8

	if len(pieces) > 1 {
		if len(pieces[1]) > 8 {
			return 0, fmt.Errorf("factoids are only subdivisible up to 1e-8, trim back on the number of decimal places.")
		}

		a := regexp.MustCompile(`(0*)([0-9]+)$`)

		as := a.FindStringSubmatch(pieces[1])
		part, _ := strconv.Atoi(as[0])
		power := len(as[1]) + len(as[2])
		total += int64(part * 1e8 / int(math.Pow10(power)))
	}

	return total, nil
}

// AmountToString converts a uint64 amount into a fixed point
// number represented as a string
func AmountToString(i int64) string {
	d := i / 1e8
	r := i % 1e8
	ds := fmt.Sprintf("%d", d)
	rs := fmt.Sprintf("%08d", r)
	rs = strings.TrimRight(rs, "0")
	if len(rs) > 0 {
		ds = ds + "."
	}
	return fmt.Sprintf("%s%s", ds, rs)
}
