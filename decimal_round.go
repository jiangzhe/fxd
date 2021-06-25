// Rounding of fixed-point decimal
//
// There are many rounding modes but mysql only supports RoundHalfUp.
// Currently we only support RoundHalfUp for simplicity.
package fxd

type DecRoundMode uint8

const (
	DecRoundHalfUp DecRoundMode = iota // round half away from zero, this is default behavior of MySQL decimal
	// DecRoundHalfEven                     // round half to even value
	// DecRoundCeiling                      // round away from zero if positive, round to zero if negative
	// DecRoundFloor                        // round to zero if positive, round away from zero if negative
	// DecRoundUp                           // round away from zero is positive, round away from zero if negative
	// DecRoundDown                         // round to zero is positive, round to zero if negative
	// DecRoundHalfDown                     // round half to from zero
)

const HalfUnit = Unit / 2

// Round rounds this decimal with provided frac.
// frac can be negative to round the integral part.
// (This is identical to MySQL/Oracle behavior)
// NaN and Infinity values are not considered in this method.
// If you want to not update the current value and store the rounded
// value to a new decimal, use RoundTo() method.
//
// NOTE: Current round mode is always RoundHalfUp, which is the only behavior of MySQL
// If more round modes are required, all methods in this file
// should be reconsidered carefully!
func (fd *FixedDecimal) Round(frac int) {
	thisFrac := int(fd.Frac())
	if frac >= thisFrac { // round precision is larger than current decimal's precision
		return
	}
	intgUnits := fd.IntgUnits()
	fracUnits := fd.FracUnits()
	var trunc int // how many digits should be truncated/rounded up
	var carry int32
	if frac < 0 { // round integral part
		roundIntgUnits := getUnits(-frac + 1)
		if roundIntgUnits > intgUnits { // integral round precision is larger than current value
			fd.SetZero() // round to zero, no matter positive or negative
			return
		}
		copy(fd.lsu[:intgUnits], fd.lsu[fracUnits:fracUnits+intgUnits]) // remove fractional units
		for i := intgUnits + 1; i < intgUnits+fracUnits; i++ {
			fd.lsu[i] = 0 // reset higher units to zero
		}
		fracUnits = 0
		trunc = -frac
	} else {
		roundFracUnits := getUnits(frac + 1) // how many units we need to keep for rounding
		if roundFracUnits < fracUnits {      // this decimal has more frac units, drop them
			if mod9(frac) == 0 { // at edge of one unit
				if unitGreaterEqualHalf(fd.lsu[fracUnits-roundFracUnits-1]) { // check rounding on second unit
					carry = 1
				}
				roundFracUnits-- // drop checked unit because carry is analyzed
			}
			copy(fd.lsu[:roundFracUnits+intgUnits], fd.lsu[fracUnits-roundFracUnits:fracUnits+intgUnits])
			for i := roundFracUnits + intgUnits; i < fracUnits+intgUnits; i++ {
				fd.lsu[i] = 0 // reset higher units to zero
			}
			fracUnits = roundFracUnits
		} else if roundFracUnits > fracUnits { // rounding precision larger than current decimal frac units
			return
		} else if mod9(frac) == 0 { // roundFracUnits == fracUnits and at edge of one uint
			if unitGreaterEqualHalf(fd.lsu[0]) { // check rounding on least significant unit
				carry = 1
			}
			copy(fd.lsu[:fracUnits+intgUnits-1], fd.lsu[1:fracUnits+intgUnits]) // drop checked unit
			for i := roundFracUnits + intgUnits - 1; i < fracUnits+intgUnits; i++ {
				fd.lsu[i] = 0
			}
			fracUnits--
		}
		trunc = fracUnits*DigitsPerUnit - frac // fractional digits to remove
	}
	roundIdx := div9(trunc) // which unit to start rounding
	roundPos := mod9(trunc) // within one unit, which position to start rounding
	roundHalfUp(fd, intgUnits, fracUnits, roundIdx, roundPos, carry)
}

func (fd *FixedDecimal) RoundTo(result *FixedDecimal, frac int) {
	thisFrac := int(fd.Frac())
	if frac >= thisFrac { // round precision is larger than or equal to current decimal's precision
		*result = *fd
		return
	}
	intgUnits := fd.IntgUnits()
	fracUnits := fd.FracUnits()
	var trunc int // how many digits should be truncated/rounded up
	var carry int32
	if frac < 0 { // round integral part
		roundIntgUnits := getUnits(-frac + 1)
		if roundIntgUnits > intgUnits { // integral round precision is larger than current value
			result.SetZero() // round to zero, no matter positive or negative
			return
		}
		copy(result.lsu[:intgUnits], fd.lsu[fracUnits:fracUnits+intgUnits]) // remove fractional units
		for i := intgUnits + 1; i < MaxUnits; i++ {
			result.lsu[i] = 0 // reset higher units to zero
		}
		fracUnits = 0
		trunc = -frac
	} else {
		roundFracUnits := getUnits(frac + 1) // how many units we need to keep for rounding
		if roundFracUnits < fracUnits {      // this decimal has more frac units, drop them
			if mod9(frac) == 0 { // at edge of one unit
				if unitGreaterEqualHalf(fd.lsu[fracUnits-roundFracUnits-1]) { // check rounding on second unit
					carry = 1
				}
				roundFracUnits-- // drop checked unit because carry is analyzed
			}
			copy(result.lsu[:roundFracUnits+intgUnits], fd.lsu[fracUnits-roundFracUnits:fracUnits+intgUnits])
			for i := roundFracUnits + intgUnits; i < MaxUnits; i++ {
				result.lsu[i] = 0
			}
			fracUnits = roundFracUnits
		} else if roundFracUnits > fracUnits { // rounding precision larger than current decimal frac units
			*result = *fd
			return
		} else if mod9(frac) == 0 { // roundFracUnits == fracUnits and at edge of one uint
			if unitGreaterEqualHalf(fd.lsu[0]) { // check rounding on least significant unit
				carry = 1
			}
			copy(result.lsu[:fracUnits+intgUnits-1], fd.lsu[1:fracUnits+intgUnits]) // drop checked unit
			for i := roundFracUnits + intgUnits - 1; i < MaxUnits; i++ {
				result.lsu[i] = 0
			}
			fracUnits--
		} else {
			*result = *fd
		}
		trunc = fracUnits*DigitsPerUnit - frac
	}
	roundIdx := div9(trunc) // which unit to start rounding
	roundPos := mod9(trunc) // within one unit, which position to start rounding
	roundHalfUp(result, intgUnits, fracUnits, roundIdx, roundPos, carry)
}

func roundHalfUp(fd *FixedDecimal, intgUnits, fracUnits, roundIdx, roundPos int, carry int32) {
	clearIdx := roundIdx // where we need to clear the units below
	if roundPos != 0 {
		switch roundPos { // unroll the code to let compiler optimize arithmetic with const values
		case 1:
			v := fd.lsu[roundIdx]
			r := v % 10
			v -= r
			if r >= 5 { // round up
				v += 10
			}
			if v >= Unit { // carry to higher unit
				v -= Unit
				carry = 1
			}
			fd.lsu[roundIdx] = v
		case 2:
			v := fd.lsu[roundIdx]
			r := v % 100
			v -= r
			if r >= 50 { // round up
				v += 100
			}
			if v >= Unit { // carry to higher unit
				v -= Unit
				carry = 1
			}
			fd.lsu[roundIdx] = v
		case 3:
			v := fd.lsu[roundIdx]
			r := v % 1_000
			v -= r
			if r >= 500 { // round up
				v += 1_000
			}
			if v >= Unit {
				v -= Unit
				carry = 1
			}
			fd.lsu[roundIdx] = v
		case 4:
			v := fd.lsu[roundIdx]
			r := v % 10_000
			v -= r
			if r >= 5_000 {
				v += 10_000
			}
			if v >= Unit {
				v -= Unit
				carry = 1
			}
			fd.lsu[roundIdx] = v
		case 5:
			v := fd.lsu[roundIdx]
			r := v % 100_000
			v -= r
			if r >= 50_000 {
				v += 100_000
			}
			if v >= Unit {
				v -= Unit
				carry = 1
			}
			fd.lsu[roundIdx] = v
		case 6:
			v := fd.lsu[roundIdx]
			r := v % 1_000_000
			v -= r
			if r >= 500_000 {
				v += 1_000_000
			}
			if v >= Unit {
				v -= Unit
				carry = 1
			}
			fd.lsu[roundIdx] = v
		case 7:
			v := fd.lsu[roundIdx]
			r := v % 10_000_000
			v -= r
			if r >= 5_000_000 {
				v += 10_000_000
			}
			if v >= Unit {
				v -= Unit
				carry = 1
			}
			fd.lsu[roundIdx] = v
		case 8:
			v := fd.lsu[roundIdx]
			r := v % 100_000_000
			v -= r
			if r >= 50_000_000 {
				v += 100_000_000
			}
			if v >= Unit {
				v -= Unit
				carry = 1
			}
			fd.lsu[roundIdx] = v
		default:
			panic("unreachable")
		}
		roundIdx++
	}
	endIdx := intgUnits + fracUnits
	for ; carry > 0 && roundIdx < endIdx; roundIdx++ {
		fd.lsu[roundIdx], carry = addWithCarry(fd.lsu[roundIdx], 0, carry)
	}
	if carry > 0 {
		fd.lsu[endIdx] = 1
		intgUnits++
	}

	// clear all units below
	for i := 0; i < clearIdx; i++ {
		fd.lsu[i] = 0
	}
	neg := fd.IsNeg()
	fd.intg = int8(intgUnits * DigitsPerUnit)
	fd.frac = int8(fracUnits * DigitsPerUnit)
	if neg {
		fd.setNeg()
	}
}
