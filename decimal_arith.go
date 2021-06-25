// Fixed-point decimal arithmethic
//
// Add, Sub, Mul and Div operators are implemented based
// on Kunth's Algorithm 4.3.1
package fxd

func DecimalAddAny(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	if lhs.IsNaN() || rhs.IsNaN() {
		result.setNaN()
		return nil
	}
	if lhs.IsInf() || rhs.IsInf() {
		result.setInf()
		return nil
	}
	return DecimalAdd(lhs, rhs, result)
}

// DecimalAdd adds two normal decimals.
func DecimalAdd(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	if lhs.IsZero() {
		*result = *rhs
		return nil
	}
	if rhs.IsZero() {
		*result = *lhs
		return nil
	}
	lneg := lhs.IsNeg()
	rneg := rhs.IsNeg()
	if lneg == rneg { // sign is same
		if err := addAbs(lhs, rhs, result); err != nil {
			return err
		}
		if lneg {
			result.setNegAndCheckZero()
		}
		return nil
	}

	var subNeg bool
	var err error
	if subNeg, err = subAbs(lhs, rhs, result); err != nil {
		return err
	}
	if subNeg != lneg {
		result.setNegAndCheckZero()
	}
	return nil
}

func DecimalSubAny(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	if lhs.IsNaN() || rhs.IsNaN() {
		result.setNaN()
		return nil
	}
	if lhs.IsInf() || rhs.IsInf() {
		result.setInf()
		return nil
	}
	return DecimalSub(lhs, rhs, result)
}

// DecimalSub subtracts two normal decimals.
func DecimalSub(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	if lhs.IsZero() {
		*result = *rhs
		if result.IsNeg() {
			result.setPos()
		} else {
			result.setNegAndCheckZero()
		}
		return nil
	}
	if rhs.IsZero() {
		*result = *lhs
		if result.IsNeg() {
			result.setPos()
		} else {
			result.setNegAndCheckZero()
		}
	}
	lneg := lhs.IsNeg()
	rneg := rhs.IsNeg()

	if lneg != rneg { // sign is different
		if err := addAbs(lhs, rhs, result); err != nil {
			return err
		}
		if lneg {
			result.setNegAndCheckZero()
		}
		return nil
	}

	// sign is same
	var subNeg bool
	var err error
	if subNeg, err = subAbs(lhs, rhs, result); err != nil {
		return err
	}
	if subNeg != lneg {
		result.setNegAndCheckZero()
	}
	return nil
}

func DecimalMulAny(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	if lhs.IsNaN() || rhs.IsNaN() {
		result.setNaN()
		return nil
	}
	if lhs.IsInf() || rhs.IsInf() {
		result.setInf()
		return nil
	}
	return DecimalMul(lhs, rhs, result)
}

// DecimalMul multiples two normal decimals.
func DecimalMul(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	if lhs.IsZero() || rhs.IsZero() {
		result.SetZero()
		return nil
	}
	resultNeg := lhs.IsNeg() != rhs.IsNeg()
	if err := mulAbs(lhs, rhs, result); err != nil {
		return err
	}
	if resultNeg {
		result.setNegAndCheckZero()
	}
	return nil
}

func DecimalDivAny(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal, incrFrac int) error {
	if lhs.IsNaN() || rhs.IsNaN() {
		result.setNaN()
		return nil
	}
	if lhs.IsInf() || rhs.IsInf() {
		result.setInf()
		return nil
	}
	return DecimalDiv(lhs, rhs, result, incrFrac)
}

// DecimalDiv divides two normal decimals.
// incrFrac is provided to add more fractional precision. Because all digits are
// stored in units, the final precision will be rounded to multiple of 9 (DigitsPerUnit).
func DecimalDiv(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal, incrFrac int) error {
	if lhs.IsZero() || rhs.IsZero() {
		result.SetZero()
		return nil
	}
	resultNeg := lhs.IsNeg() != rhs.IsNeg()
	if err := divAbs(lhs, rhs, result, incrFrac); err != nil {
		return err
	}
	if resultNeg {
		result.setNegAndCheckZero()
	}
	return nil
}

// addAbs sums two decimals' absolute values.
// Separate units into 3 segments, intgSeg, commonSeg, fracSeg
// lhs:  |  xxxx  |  xxxx.xxxx  |
// rhs:           |  yyyy.yyyy  |  yyyy  |
// ---------------------------------------
//       |intgSeg |  commonSeg  |fracSeg |
//
// fracSeg can be directly added to result.
// commonSeg uses normal addition, taking care of carry.
// intgSeg is added with 0 and carry.
func addAbs(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	result.resetUnits() // always clear result units first
	liu, lfu := lhs.IntgUnits(), lhs.FracUnits()
	riu, rfu := rhs.IntgUnits(), rhs.FracUnits()
	var lhsIdx, rhsIdx, resultIdx int
	fracUnitDiff := lfu - rfu
	if fracUnitDiff > 0 { // lhs has more fracSeg
		for ; lhsIdx < fracUnitDiff; lhsIdx++ {
			result.lsu[lhsIdx] = lhs.lsu[lhsIdx] // copy lhs to result
		}
		resultIdx = lhsIdx
	} else if fracUnitDiff < 0 { // rhs has more fracSeg
		fracUnitDiff = -fracUnitDiff
		for ; rhsIdx < fracUnitDiff; rhsIdx++ {
			result.lsu[rhsIdx] = rhs.lsu[rhsIdx] // copy rhs to result
		}
		resultIdx = rhsIdx
	}

	var carry int32
	stop := resultIdx + minInt(liu, riu) + minInt(lfu, rfu)
	for resultIdx < stop {
		result.lsu[resultIdx], carry = addWithCarry(lhs.lsu[lhsIdx], rhs.lsu[rhsIdx], carry)
		resultIdx++
		lhsIdx++
		rhsIdx++
	}

	intgUnitDiff := liu - riu
	if intgUnitDiff > 0 { // lhs has more intgSeg
		stop = lhsIdx + intgUnitDiff
		for lhsIdx < stop {
			result.lsu[resultIdx], carry = addWithCarry(lhs.lsu[lhsIdx], 0, carry)
			resultIdx++
			lhsIdx++
		}
	} else if intgUnitDiff < 0 { // rhs has more intgSeg
		stop = rhsIdx - intgUnitDiff
		for rhsIdx < stop {
			result.lsu[resultIdx], carry = addWithCarry(rhs.lsu[rhsIdx], 0, carry)
			resultIdx++
			rhsIdx++
		}
	} else if carry != 0 { // no more inteSeg but carry is non-zero
		result.lsu[resultIdx] = carry
	}

	result.frac = maxInt8(lhs.frac, rhs.frac) // unify as maximum frac
	// there is extra cost to calculate exact integral digits so we expand it to intgUnits*DigitsPerUnit
	// (optionally plus carry)
	result.intg = int8(maxInt(liu, riu)*DigitsPerUnit + int(carry))
	return nil
}

// subAbs diff two decimals's absolute values, returns the negative flag.
// Similar to addAbs(), but if lhs is smaller than rhs, we will have
// borrow=-1 at end. Then we can traverse all units and apply subtraction
// again.
//
// Separate units into 3 segments, intgSeg, commonSeg, fracSeg
// lhs:  |  xxxx  |  xxxx.xxxx  |
// rhs:           |  yyyy.yyyy  |  yyyy  |
// ---------------------------------------
//       |intgSeg |  commonSeg  |fracSeg |
func subAbs(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) (bool, error) {
	result.resetUnits() // always clear result units first
	liu, lfu := lhs.IntgUnits(), lhs.FracUnits()
	riu, rfu := rhs.IntgUnits(), rhs.FracUnits()
	var lhsIdx, rhsIdx, resultIdx int
	var borrow int32
	fracUnitDiff := lfu - rfu
	if fracUnitDiff > 0 { // lhs has more fracSeg
		for ; lhsIdx < fracUnitDiff; lhsIdx++ {
			result.lsu[lhsIdx] = lhs.lsu[lhsIdx] // copy lhs to result
		}
		resultIdx = lhsIdx
	} else if fracUnitDiff < 0 { // rhs has more fracSeg
		fracUnitDiff = -fracUnitDiff
		for ; rhsIdx < fracUnitDiff; rhsIdx++ {
			result.lsu[rhsIdx], borrow = subWithBorrow(0, rhs.lsu[rhsIdx], borrow) // subtract rhs with zeros
		}
		resultIdx = rhsIdx
	}

	stop := resultIdx + minInt(liu, riu) + minInt(lfu, rfu)
	for resultIdx < stop {
		result.lsu[resultIdx], borrow = subWithBorrow(lhs.lsu[lhsIdx], rhs.lsu[rhsIdx], borrow)
		resultIdx++
		lhsIdx++
		rhsIdx++
	}

	intgUnitDiff := liu - riu
	if intgUnitDiff > 0 { // lhs has more intgSeg
		stop = lhsIdx + intgUnitDiff
		for lhsIdx < stop {
			result.lsu[resultIdx], borrow = subWithBorrow(lhs.lsu[lhsIdx], 0, borrow)
			resultIdx++
			lhsIdx++
		}
	} else if intgUnitDiff < 0 { // rhs has more intgSeg
		stop = rhsIdx - intgUnitDiff
		for rhsIdx < stop {
			result.lsu[resultIdx], borrow = subWithBorrow(0, rhs.lsu[rhsIdx], borrow)
			resultIdx++
			rhsIdx++
		}
	}
	neg := borrow == -1 // left is smaller than right
	if neg {            // must traverse all result units from lsu and do subtraction again
		borrow = 0
		for i := 0; i < resultIdx; i++ {
			result.lsu[i], borrow = subWithBorrow(0, result.lsu[i], borrow)
		}
	}

	result.frac = maxInt8(lhs.frac, rhs.frac) // unify as maximum frac
	// there is extra cost to calculate exact integral digits so we expand it to intgUnits*DigitsPerUnit
	result.intg = int8(maxInt(liu, riu) * DigitsPerUnit)
	return neg, nil
}

// mulAbs multiplies two decimals' absolute values.
// The result precision is extended to the maximum possible one, until reaching
// the limitation of MaxUnits or MaxFracUnits.
func mulAbs(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	result.resetUnits() // always clear result units first
	// result integral digits should be sum of left and right integral digits
	resultIntgDigits := int(lhs.Intg() + rhs.Intg())
	resultIntgUnits := getUnits(resultIntgDigits)
	// result fractional digits should be sum of left and right fractional digits
	resultFracDigits := int(lhs.Frac() + rhs.Frac())
	resultFracUnits := getUnits(resultFracDigits)
	if resultIntgUnits > MaxUnits { // integral overflow
		return DecErrOverflow
	}
	if resultIntgUnits+resultFracUnits > MaxUnits { // integral+fractional overflow
		resultFracUnits = MaxUnits - resultIntgUnits // fractional truncation required
	}
	if resultFracUnits > MaxFracUnits { // still exceeds maximum fractional digits
		resultFracUnits = MaxFracUnits
	}
	liu, lfu := lhs.IntgUnits(), lhs.FracUnits()
	riu, rfu := rhs.IntgUnits(), rhs.FracUnits()
	// because result fractional part may be truncated, we need to calculate
	// how many units has to be shifted and can be ignored in calculation.
	// 1. If leftIdx + rightIdx - shiftUnits < -1, we can ignore the
	//    result of left.lsu[leftIdx]*right.lsu[rightIdx].
	// 2. If leftIdx + rightIdx - shiftUnits = -1, we need to take care of the carry
	//    and add to result.lsu[0].
	// 3. If leftIdx + rightIdx - shiftUnits >= 0, follow normal calculation.
	shiftUnits := lfu + rfu - resultFracUnits
	var carry int64
	var leftIdx, rightIdx, resultIdx int
	var lv, rv int32
	for rightIdx, rv = range rhs.lsu[:riu+rfu] {
		for leftIdx, lv = range lhs.lsu[:liu+lfu] {
			resultIdx = leftIdx + rightIdx - shiftUnits
			if resultIdx < -1 {
				continue
			}
			if resultIdx == -1 {
				v := int64(lv) * int64(rv)
				if v < Unit {
					continue
				}
				carry = v / Unit // we only need the carry for lsu of result
				continue
			}
			// calculate product and sum with previous result and carry
			v := int64(lv)*int64(rv) + int64(result.lsu[resultIdx]) + carry
			carry = v / Unit
			result.lsu[resultIdx] = int32(v - carry*Unit)
		}
		if resultIdx+1 < MaxUnits {
			result.lsu[resultIdx+1] = int32(carry)
		} else if carry > 0 {
			return DecErrOverflow
		}
		carry = 0
	}
	result.frac = minInt8(int8(resultFracDigits), int8(resultFracUnits)*DigitsPerUnit)
	result.intg = int8(resultIntgUnits) * DigitsPerUnit
	return nil
}

func divAbs(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal, incrFrac int) error {
	result.resetUnits() // always clear result units first
	lhsIntg := int(lhs.Intg())
	liu := getUnits(lhsIntg)
	lhsFrac := int(lhs.Frac())
	lfu := getUnits(lhsFrac)
	lhsExtFrac := lfu * DigitsPerUnit // extended frac with unit size
	rhsIntg := int(rhs.Intg())
	riu := getUnits(rhsIntg)
	rhsFrac := int(rhs.Frac())
	rfu := getUnits(rhsFrac)
	rhsExtFrac := rfu * DigitsPerUnit // extended frac with unit size
	// leading non-zero unit of lhs and rhs
	var lhsNonZero, rhsNonZero int

	// check and remove leading zeros in rhs
	for rhsNonZero = riu + rfu - 1; rhsNonZero >= rfu; rhsNonZero-- {
		if rhs.lsu[rhsNonZero] != 0 {
			break
		}
	}
	if rhsNonZero < 0 { // divider is zero
		return DecErrDivisionByZero
	}
	// digits of rhs from leading non-zero position
	rhsPrec := rhsNonZero*DigitsPerUnit + DigitsPerUnit - unitLeadingZeroes(rhs.lsu[rhsNonZero])

	// check and remove leading zeros in lhs
	for lhsNonZero = liu + lfu - 1; lhsNonZero >= 0; lhsNonZero-- {
		if lhs.lsu[lhsNonZero] != 0 {
			break
		}
	}
	if lhsNonZero < 0 { // dividend is zero
		result.SetZero()
		return nil
	}
	// digits of lhs from leading non-zero position
	lhsPrec := lhsNonZero*DigitsPerUnit + DigitsPerUnit - unitLeadingZeroes(rhs.lsu[rhsNonZero])

	// because we store fractional part in unit, we always extend fractional precision
	// with multiple of 9. Here check if incrPrec is already covered by the natural extension,
	// if so, reset incrPrec to 0.
	incrFrac -= lhsExtFrac - lhsFrac + rhsExtFrac - rhsFrac
	if incrFrac < 0 {
		incrFrac = 0
	}
	// guess the quotient integral digits:
	// if left first non-zero unit is no less than right first non-zero unit,
	// the quotient may probably have one additional integral digit.
	// If left first non-zero unit is smaller than right one,
	// set dividendShift to -1 so that use second unit to start the division.
	resultIntg := (lhsPrec - lhsExtFrac) - (rhsPrec - rhsExtFrac)
	var dividendShift int
	if lhs.lsu[lhsNonZero] >= rhs.lsu[rhsNonZero] {
		resultIntg++
	} else {
		dividendShift = -1 // dividend should be shift by one unit
	}

	var resultIntgUnits int
	if resultIntg > 0 {
		resultIntgUnits = getUnits(resultIntg)
	} else {
		resultIntg = 0
	}
	// calculate result frac
	resultFracUnits := getUnits(lhsExtFrac + rhsExtFrac + incrFrac)
	if resultIntgUnits > MaxUnits { // integral overflow
		return DecErrOverflow
	}
	if resultIntgUnits+resultFracUnits > MaxUnits { // integral+fractional overflow
		resultFracUnits = MaxUnits - resultIntgUnits // fractional truncation required
	}
	if resultFracUnits > MaxFracUnits { // still exceeds maximum fractional digits
		resultFracUnits = MaxFracUnits
	}

	m := resultIntgUnits + resultFracUnits // units of the quotient
	// here we identify short/long division
	// short division means the divider only has single unit,
	// otherwise, long division
	if rhsNonZero == 0 { // short division
		result.SetZero()
		var buf1 [MaxUnits + 1]int32
		copy(buf1[m:m+lhsNonZero+1], lhs.lsu[:lhsNonZero+1])
		d := int64(rhs.lsu[rhsNonZero]) // single divider
		var u, q, rem int64             // remainder
		if dividendShift < 0 {
			rem = int64(buf1[m+lhsNonZero])
		}
		for j := m; j > 0; j-- {
			uidx := j + lhsNonZero + dividendShift
			if uidx >= 0 {
				u = rem*Unit + int64(buf1[uidx])
			} else {
				u = rem * Unit
			}

			q = u / d                  // div
			rem = u - q*d              // update remainder
			result.lsu[j-1] = int32(q) // update result
		}
		result.intg = int8(resultIntg)
		result.frac = int8(resultFracUnits * DigitsPerUnit)
		return nil
	}

	// long division using Knuth's algorithm
	// D1. normalization
	normFactor := Unit / (rhs.lsu[rhsNonZero] + 1)
	// normalize buf1 and buf2
	var buf1 [MaxUnits * 2]int32 // store normalized lhs with one extra leading unit
	var buf2 [MaxUnits]int32     // store normalized rhs
	var i, j int                 // loop index control
	if normFactor == 1 {         // happy path, if rhs.lsu[rhsNonZero] >= Unit / 2
		copy(buf1[m:m+lhsNonZero+1], lhs.lsu[:lhsNonZero+1])
		copy(buf2[:rhsNonZero+1], rhs.lsu[:rhsNonZero+1])
	} else {
		var carry int64
		// normalize lhs into buf1
		for i, j = 0, m; i <= lhsNonZero; i, j = i+1, j+1 {
			v := int64(lhs.lsu[i])*int64(normFactor) + carry
			carry = v / Unit
			buf1[j] = int32(v - carry*Unit)
		}
		buf1[j] = int32(carry)
		carry = 0
		// normalize rhs into buf2
		for i = 0; i <= rhsNonZero; i++ {
			v := int64(rhs.lsu[i])*int64(normFactor) + carry
			carry = v / Unit
			buf2[i] = int32(v - carry*Unit)
		}
		if carry != 0 {
			panic("carry must be 0 in divider normalization")
		}
	}
	vd0 := int64(buf2[rhsNonZero])   // rhs most significant unit
	vd1 := int64(buf2[rhsNonZero-1]) // rhs second significant unit
	for j = m; j > 0; j-- {
		// D3. make the guess on u1
		uidx := j + lhsNonZero + dividendShift
		u0 := int64(buf1[uidx+1])
		var u1 int64
		if uidx >= 0 {
			u1 = int64(buf1[uidx])
		}
		v := u0*Unit + u1
		qhat := v / vd0
		rhat := v - qhat*vd0
		// qhat cannot be greater or equal to Unit
		if qhat >= Unit {
			panic("unreachable")
		}
		var u2 int64
		if uidx > 0 {
			u2 = int64(buf1[uidx-1])
		}
		for qhat*vd1 > rhat*Unit+u2 { // check if qhat can satisfy next unit
			qhat--      // decrese qhat
			rhat += vd0 // increase rhat
		}
		// D4. multiply and subtract
		var mulV, mulV0, carry int64 // the product, product within current unit, carry of multiplication
		var subV, borrow int32       // the diff and boorow of subtraction
		var i, msIdx int
		for msIdx = uidx - rhsNonZero; i <= rhsNonZero; i, msIdx = i+1, msIdx+1 {
			mulV = qhat*int64(buf2[i]) + carry                              // mul
			carry = mulV / Unit                                             // update carry
			mulV0 = mulV - carry*Unit                                       // in current unit
			subV, borrow = subWithBorrow(buf1[msIdx], int32(mulV0), borrow) // sub
			buf1[msIdx] = subV                                              // update buf1 with result
		}
		borrow = buf1[msIdx] - int32(carry) + borrow
		if borrow == -1 { // qhat is larger, cannot satisfy the whole decimal
			// D6. add back (reverse subtract)
			qhat--                            // decrease qhat
			borrow = 0                        // reset borrow to zero
			for i := 0; i < rhsNonZero; i++ { // reverse subtract
				buf1[i+j], borrow = subWithBorrow(0, buf1[i+j], borrow)
			}
		} else {
			buf1[msIdx] = 0 // clear buf1 because multiply w/ subtract succeeds
		}
		result.lsu[j-1] = int32(qhat) // update result
	}
	result.intg = int8(resultIntg)
	result.frac = int8(resultFracUnits * DigitsPerUnit)
	return nil
}

// NOTE: carry can only be 0 or 1
func addWithCarry(a, b, carry int32) (int32, int32) {
	r := a + b + carry
	if r >= Unit {
		return r - Unit, 1
	}
	return r, 0
}

// NOTE: borrow can only be 0 or -1
func subWithBorrow(a, b, borrow int32) (int32, int32) {
	r := a - b + borrow
	if r < 0 {
		return Unit + r, -1
	}
	return r, 0
}
