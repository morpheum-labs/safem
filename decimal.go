package safem

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Multiprecision decimal numbers.
// For floating-point formatting only; not general purpose.
// Only operations are assign and (binary) left/right shift.
// Can do binary floating point in multiprecision decimal precisely
// because 2 divides 10; cannot do decimal floating point
// in multiprecision binary precisely.

import (
	"strings"
)

var (
	ln10 = newConstApproximation(strLn10)
)

type constApproximation struct {
	exact          Decimal
	approximations []Decimal
}

func newConstApproximation(value string) constApproximation {
	parts := strings.Split(value, ".")
	coeff, fractional := parts[0], parts[1]

	coeffLen := len(coeff)
	maxPrecision := len(fractional)

	var approximations []Decimal
	for p := 1; p < maxPrecision; p *= 2 {
		r := RequireFromString(value[:coeffLen+p])
		approximations = append(approximations, r)
	}

	return constApproximation{
		RequireFromString(value),
		approximations,
	}
}

// Returns the smallest approximation available that's at least as precise
// as the passed precision (places after decimal point), i.e. Floor[ log2(precision) ] + 1
func (c constApproximation) withPrecision(precision int32) Decimal {
	i := 0

	if precision >= 1 {
		i++
	}

	for precision >= 16 {
		precision /= 16
		i += 4
	}

	for precision >= 2 {
		precision /= 2
		i++
	}

	if i >= len(c.approximations) {
		return c.exact
	}

	return c.approximations[i]
}

type floatInfo struct {
	mantbits uint
	expbits  uint
	bias     int
}

var float32info = floatInfo{23, 8, -127}
var float64info = floatInfo{52, 11, -1023}

// decimal represents a floating-point number in decimal format.
// It is used internally for floating-point to decimal conversion.
type decimal struct {
	d     [800]byte // digits, big-endian representation
	nd    int       // number of digits used
	dp    int       // decimal point
	neg   bool      // negative flag
	trunc bool      // discarded nonzero digits beyond d[:nd]
}

// Assign sets the decimal to the value of the given uint64.
func (a *decimal) Assign(v uint64) {
	var buf [24]byte
	w := len(buf)
	for v > 0 {
		w--
		buf[w] = byte(v%10) + '0'
		v /= 10
	}
	a.nd = 0
	for w < len(buf) {
		a.d[a.nd] = buf[w]
		a.nd++
		w++
	}
	a.dp = a.nd
	a.neg = false
	a.trunc = false
}

// Shift shifts the decimal left or right by the given amount.
func (a *decimal) Shift(k int) {
	switch {
	case a.nd == 0:
		// nothing to do: a == 0
	case k > 0:
		for k > 0 {
			var r byte
			if a.nd > 0 {
				r = a.d[0] % 10
			}
			a.ShiftLeft(1)
			if r != 0 {
				a.trunc = true
			}
			k--
		}
	case k < 0:
		for k < 0 {
			if a.nd == 0 {
				a.dp = 0
				return
			}
			a.ShiftRight(1)
			k++
		}
	}
}

// ShiftLeft shifts the decimal left by k digits.
func (a *decimal) ShiftLeft(k int) {
	if a.nd == 0 {
		return
	}
	if k <= 0 {
		return
	}
	if a.nd+k > cap(a.d) {
		// overflow
		a.nd = 0
		a.dp = 0
		return
	}
	copy(a.d[k:], a.d[:a.nd])
	for i := 0; i < k; i++ {
		a.d[i] = '0'
	}
	a.nd += k
	a.dp += k
}

// ShiftRight shifts the decimal right by k digits.
func (a *decimal) ShiftRight(k int) {
	if a.nd == 0 {
		return
	}
	if k <= 0 {
		return
	}
	if k >= a.nd {
		a.nd = 0
		a.dp = 0
		return
	}
	copy(a.d[:a.nd-k], a.d[k:a.nd])
	a.nd -= k
	a.dp -= k
}

// Round rounds the decimal to the given number of digits.
func (a *decimal) Round(nd int) {
	if nd >= a.nd || nd < 0 {
		return
	}
	if nd == 0 {
		a.nd = 0
		return
	}
	if a.d[nd] >= '5' {
		a.RoundUp(nd)
	} else {
		a.RoundDown(nd)
	}
}

// RoundUp rounds the decimal up to the given number of digits.
func (a *decimal) RoundUp(nd int) {
	if nd < 0 || nd >= a.nd {
		return
	}
	for i := nd - 1; i >= 0; i-- {
		if a.d[i] < '9' {
			a.d[i]++
			a.nd = i + 1
			return
		}
	}
	// all digits are 9s
	a.d[0] = '1'
	a.nd = 1
	a.dp++
}

// RoundDown rounds the decimal down to the given number of digits.
func (a *decimal) RoundDown(nd int) {
	if nd >= a.nd {
		return
	}
	a.nd = nd
}

// roundShortest rounds d (= mant * 2^exp) to the shortest number of digits
// that will let the original floating point value be precisely reconstructed.
func roundShortest(d *decimal, mant uint64, exp int, flt *floatInfo) {
	// If mantissa is zero, the number is zero; stop now.
	if mant == 0 {
		d.nd = 0
		return
	}

	// Compute upper and lower such that any decimal number
	// between upper and lower (possibly inclusive)
	// will round to the original floating point number.

	// We may see at once that the number is already shortest.
	//
	// Suppose d is not denormal, so that 2^exp <= d < 10^dp.
	// The closest shorter number is at least 10^(dp-nd) away.
	// The lower/upper bounds computed below are at distance
	// at most 2^(exp-mantbits).
	//
	// So the number is already shortest if 10^(dp-nd) > 2^(exp-mantbits),
	// or equivalently log2(10)*(dp-nd) > exp-mantbits.
	// It is true if 332/100*(dp-nd) >= exp-mantbits (log2(10) > 3.32).
	minexp := flt.bias + 1 // minimum possible exponent
	if exp > minexp && 332*(d.dp-d.nd) >= 100*(exp-int(flt.mantbits)) {
		// The number is already shortest.
		return
	}

	// d = mant << (exp - mantbits)
	// Next highest floating point number is mant+1 << exp-mantbits.
	// Our upper bound is halfway between, mant*2+1 << exp-mantbits-1.
	upper := new(decimal)
	upper.Assign(mant*2 + 1)
	upper.Shift(exp - int(flt.mantbits) - 1)

	// d = mant << (exp - mantbits)
	// Next lowest floating point number is mant-1 << exp-mantbits,
	// unless mant-1 drops the significant bit and exp is not the minimum exp,
	// in which case the next lowest is mant*2-1 << exp-mantbits-1.
	// Either way, call it mantlo << explo-mantbits.
	// Our lower bound is halfway between, mantlo*2+1 << explo-mantbits-1.
	var mantlo uint64
	var explo int
	if mant > 1<<flt.mantbits || exp == minexp {
		mantlo = mant - 1
		explo = exp
	} else {
		mantlo = mant*2 - 1
		explo = exp - 1
	}
	lower := new(decimal)
	lower.Assign(mantlo*2 + 1)
	lower.Shift(explo - int(flt.mantbits) - 1)

	// The upper and lower bounds are possible outputs only if
	// the original mantissa is even, so that IEEE round-to-even
	// would round to the original mantissa and not the neighbors.
	inclusive := mant%2 == 0

	// As we walk the digits we want to know whether rounding up would fall
	// within the upper bound. This is tracked by upperdelta:
	//
	// If upperdelta == 0, the digits of d and upper are the same so far.
	//
	// If upperdelta == 1, we saw a difference of 1 between d and upper on a
	// previous digit and subsequently only 9s for d and 0s for upper.
	// (Thus rounding up may fall outside the bound, if it is exclusive.)
	//
	// If upperdelta == 2, then the difference is greater than 1
	// and we know that rounding up falls within the bound.
	var upperdelta uint8

	// Now we can figure out the minimum number of digits required.
	// Walk along until d has distinguished itself from upper and lower.
	for ui := 0; ; ui++ {
		// lower, d, and upper may have the decimal points at different
		// places. In this case upper is the longest, so we iterate from
		// ui==0 and start li and mi at (possibly) -1.
		mi := ui - upper.dp + d.dp
		if mi >= d.nd {
			break
		}
		li := ui - upper.dp + lower.dp
		l := byte('0') // lower digit
		if li >= 0 && li < lower.nd {
			l = lower.d[li]
		}
		m := byte('0') // middle digit
		if mi >= 0 {
			m = d.d[mi]
		}
		u := byte('0') // upper digit
		if ui < upper.nd {
			u = upper.d[ui]
		}

		// Okay to round down (truncate) if lower has a different digit
		// or if lower is inclusive and is exactly the result of rounding
		// down (i.e., and we have reached the final digit of lower).
		okdown := l != m || inclusive && li+1 == lower.nd

		switch {
		case upperdelta == 0 && m+1 < u:
			// Example:
			// m = 12345xxx
			// u = 12347xxx
			upperdelta = 2
		case upperdelta == 0 && m != u:
			// Example:
			// m = 12345xxx
			// u = 12346xxx
			upperdelta = 1
		case upperdelta == 1 && (m != '9' || u != '0'):
			// Example:
			// m = 1234598x
			// u = 1234600x
			upperdelta = 2
		}
		// Okay to round up if upper has a different digit and either upper
		// is inclusive or upper is bigger than the result of rounding up.
		okup := upperdelta > 0 && (inclusive || upperdelta > 1 || ui+1 < upper.nd)

		// If it's okay to do either, then round to the nearest one.
		// If it's okay to do only one, do it.
		switch {
		case okdown && okup:
			d.Round(mi + 1)
			return
		case okdown:
			d.RoundDown(mi + 1)
			return
		case okup:
			d.RoundUp(mi + 1)
			return
		}
	}
}
