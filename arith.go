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
		return nil
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
	resultNeg := lhs.IsNeg() != rhs.IsNeg()
	if err := divAbs(lhs, rhs, result, incrFrac); err != nil {
		return err
	}
	if resultNeg {
		result.setNegAndCheckZero()
	}
	return nil
}

func DecimalModAny(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	if lhs.IsNaN() || rhs.IsNaN() {
		result.setNaN()
		return nil
	}
	if lhs.IsInf() || rhs.IsInf() {
		result.setInf()
		return nil
	}
	return DecimalMod(lhs, rhs, result)
}

// DecimalMod modulos two normal decimals.
func DecimalMod(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	resultNeg := lhs.IsNeg()
	if err := modAbs(lhs, rhs, result); err != nil {
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
	result.Reset() // always clear result first
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
		if carry > 0 {
			result.lsu[resultIdx] = carry
		}
	} else if intgUnitDiff < 0 { // rhs has more intgSeg
		stop = rhsIdx - intgUnitDiff
		for rhsIdx < stop {
			result.lsu[resultIdx], carry = addWithCarry(rhs.lsu[rhsIdx], 0, carry)
			resultIdx++
			rhsIdx++
		}
		if carry > 0 {
			result.lsu[resultIdx] = carry
		}
	} else if carry != 0 { // no more inteSeg but carry is non-zero
		result.lsu[resultIdx] = carry
	}

	resultIntgNonZero := maxInt(liu, riu) + int(carry)
	result.frac = maxInt8(lhs.frac, rhs.frac) // unify as maximum frac
	resultFracUnits := result.FracUnits()
	for ; resultIntgNonZero >= 0; resultIntgNonZero-- {
		if result.lsu[resultIntgNonZero+resultFracUnits] > 0 {
			break
		}
	}
	if resultIntgNonZero >= 0 {
		result.intg = int8((resultIntgNonZero + 1) * DigitsPerUnit) // expand to multiple of DigitsPerUnit
	} else {
		result.intg = 0
	}

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
	result.Reset() // always clear result first
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
	resultIntgNonZero := maxInt(liu, riu)
	result.frac = maxInt8(lhs.frac, rhs.frac) // unify as maximum frac
	resultFracUnits := result.FracUnits()
	for ; resultIntgNonZero >= 0; resultIntgNonZero-- {
		if result.lsu[resultIntgNonZero+resultFracUnits] > 0 {
			break
		}
	}
	if resultIntgNonZero >= 0 {
		result.intg = int8((resultIntgNonZero + 1) * DigitsPerUnit) // expand to multiple of DigitsPerUnit
	} else {
		result.intg = 0
	}
	return neg, nil
}

// mulAbs multiplies two decimals' absolute values.
// The result precision is extended to the maximum possible one, until reaching
// the limitation of MaxUnits or MaxFracUnits.
func mulAbs(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	result.Reset() // always clear result first
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

// divAbs divides two decimals' absolute values.
// It's implementation of Knuth's Algorithm 4.3.1 D, with support on frational numbers.
func divAbs(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal, incrFrac int) error {
	result.Reset() // always clear result first
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
	for rhsNonZero = riu + rfu - 1; rhsNonZero >= 0; rhsNonZero-- {
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

	// calculate result frac units
	resultFracUnits := getUnits(lhsExtFrac + rhsExtFrac + incrFrac)
	if resultFracUnits > MaxFracUnits { // still exceeds maximum fractional digits
		resultFracUnits = MaxFracUnits
	}
	// calculate result intg units
	resultIntg := (lhsPrec - lhsExtFrac) - (rhsPrec - rhsExtFrac)
	var dividendShift int
	if unitsGreaterEqual(lhs.lsu[:lhsNonZero+1], rhs.lsu[:rhsNonZero+1]) {
		resultIntg++ // one more digit
	} else {
		dividendShift = -1 // dividend should be shift right one unit to start the division
	}
	// now adjust result units based on limitation of maximum precision
	// and determine the start position of result unit.
	var resultIntgUnits int
	var resultStartIdx int // start index of result units for this division
	if resultIntg > 0 {
		resultIntgUnits = getUnits(resultIntg)
		if resultIntgUnits > MaxUnits { // exceeds maximum precision
			return DecErrOverflow
		}
		if resultIntgUnits+resultFracUnits > MaxUnits {
			resultFracUnits = MaxUnits - resultIntgUnits // truncate extra fractional units
		}
		resultStartIdx = resultFracUnits + resultIntgUnits - 1
	} else {
		resultIntgUnits = 0
		resultStartOffset := getUnits(1 - resultIntg)
		resultStartIdx = resultFracUnits - resultStartOffset
		resultIntg = 0
	}
	resultUnits := resultIntgUnits + resultFracUnits

	// here we identify short/long division
	// short division means the divider only has single unit,
	// otherwise, long division
	if rhsNonZero == 0 { // short division
		result.SetZero()
		d := int64(rhs.lsu[rhsNonZero]) // single divider
		var u, q, rem int64
		if dividendShift < 0 {
			rem = int64(lhs.lsu[lhsNonZero])
		}
		for i, j := lhsNonZero+dividendShift, resultStartIdx; j >= 0; i, j = i-1, j-1 { // i is index of lhs, j is index of result
			if i >= 0 {
				u = rem*Unit + int64(lhs.lsu[i])
			} else {
				u = rem * Unit
			}
			q = u / d                // div
			rem = u - q*d            // update remainder
			result.lsu[j] = int32(q) // update result
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
		copy(buf1[resultUnits:resultUnits+lhsNonZero+1], lhs.lsu[:lhsNonZero+1])
		copy(buf2[:rhsNonZero+1], rhs.lsu[:rhsNonZero+1])
	} else {
		var carry int64
		// normalize lhs into buf1
		for i, j = 0, resultUnits; i <= lhsNonZero; i, j = i+1, j+1 {
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
		assertTrue(carry == 0, "carry must be zero in divider normalization")
	}
	vd0 := int64(buf2[rhsNonZero])   // rhs most significant unit
	vd1 := int64(buf2[rhsNonZero-1]) // rhs second significant unit
	for i, j = resultUnits+lhsNonZero+dividendShift, resultStartIdx; j >= 0; i, j = i-1, j-1 {
		// D3. make the guess on u1
		// uidx := j + lhsNonZero + dividendShift
		// u0 := int64(buf1[uidx+1])
		u0 := int64(buf1[i+1])
		var u1 int64
		if i >= 0 {
			u1 = int64(buf1[i])
		}
		v := u0*Unit + u1
		qhat := v / vd0
		rhat := v - qhat*vd0
		// qhat cannot be greater or equal to Unit
		assertTrue(qhat < Unit, "qhat must be less than Unit")
		var u2 int64
		if i > 0 {
			u2 = int64(buf1[i-1])
		}
		for qhat*vd1 > rhat*Unit+u2 { // check if qhat can satisfy next unit
			qhat--      // decrese qhat
			rhat += vd0 // increase rhat
		}
		// D4. multiply and subtract
		var mulV, mulV0, carry int64 // the product, product within current unit, carry of multiplication
		var subV, borrow int32       // the diff and boorow of subtraction
		var k, msIdx int
		for msIdx = i - rhsNonZero; k <= rhsNonZero; k, msIdx = k+1, msIdx+1 {
			mulV = qhat*int64(buf2[k]) + carry // mul
			carry = mulV / Unit                // update carry
			mulV0 = mulV - carry*Unit          // in current unit
			if msIdx < 0 {
				_, borrow = subWithBorrow(0, int32(mulV0), borrow) // sub using 0
			} else {
				subV, borrow = subWithBorrow(buf1[msIdx], int32(mulV0), borrow) // sub
				buf1[msIdx] = subV                                              // update buf1 with result
			}
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
		result.lsu[j] = int32(qhat) // update result
	}
	result.intg = int8(resultIntg)
	result.frac = int8(resultFracUnits * DigitsPerUnit)
	return nil
}

func modAbs(lhs *FixedDecimal, rhs *FixedDecimal, result *FixedDecimal) error {
	result.Reset() // always clear result first
	lhsIntg := int(lhs.Intg())
	liu := getUnits(lhsIntg) // lhs intg units
	lhsFrac := int(lhs.Frac())
	lfu := getUnits(lhsFrac) // lhs frac units
	rhsIntg := int(rhs.Intg())
	riu := getUnits(rhsIntg) // rhs intg units
	rhsFrac := int(rhs.Frac())
	rfu := getUnits(rhsFrac)          // rhs frac units
	lhsExtFrac := lfu * DigitsPerUnit // extended frac with unit size
	rhsExtFrac := rfu * DigitsPerUnit // extended frac with unit size
	// leading non-zero unit of lhs and rhs
	var lhsNonZero, rhsNonZero int
	// check and remove leading zeros in rhs
	for rhsNonZero = riu + rfu - 1; rhsNonZero >= 0; rhsNonZero-- {
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

	cmp := cmpAbsLsu(liu, lfu, &lhs.lsu, riu, rfu, &rhs.lsu)
	if cmp < 0 { // lhs is less than rhs
		if rfu > lfu { // rhs has higher fractional precision
			copy(result.lsu[rfu-lfu:liu+rfu], lhs.lsu[:lfu+liu])
			result.frac = int8(rhsFrac)
			result.intg = int8(lhsIntg)
			return nil
		}
		// lhs has higher fractional precision
		copy(result.lsu[:lfu+liu], lhs.lsu[:lfu+liu])
		result.intg = int8(lhsIntg)
		result.frac = int8(maxInt(lhsFrac, rhsFrac))
		return nil
	}
	if cmp == 0 { // lhs equals to rhs, result is zero
		// align to max frac of both lhs and rhs
		result.intg = 0
		result.frac = int8(maxInt(lhsFrac, rhsFrac))
		return nil
	}

	// digits of lhs from leading non-zero position
	lhsPrec := lhsNonZero*DigitsPerUnit + DigitsPerUnit - unitLeadingZeroes(rhs.lsu[rhsNonZero])

	// calculate result frac units
	remainderFrac := maxInt(lhsFrac, rhsFrac)
	remainderFracUnits := getUnits(remainderFrac)

	// calculate result intg units, but result may have leading zeroes that should be removed
	// at the end
	quotientIntg := (lhsPrec - lhsExtFrac) - (rhsPrec - rhsExtFrac)
	var dividendShift int
	if unitsGreaterEqual(lhs.lsu[:lhsNonZero+1], rhs.lsu[:rhsNonZero+1]) {
		quotientIntg++ // one more digit
	} else {
		dividendShift = -1 // dividend should be shift right one unit to start the division
	}

	// align frac units between lhs and rhs
	var lhsLeftShiftUnits, rhsLeftShiftUnits int
	if lfu < rfu {
		lhsLeftShiftUnits = rfu - lfu
	} else if lfu > rfu {
		rhsLeftShiftUnits = lfu - rfu
	}

	if rhsNonZero == 0 { // identify short division
		d := int64(rhs.lsu[0]) // single divider
		var buf [MaxUnits * 2]int32
		buflen := lhsLeftShiftUnits + lhsNonZero + 1
		copy(buf[lhsLeftShiftUnits:buflen], lhs.lsu[:lhsNonZero+1])
		var u, q, rem int64
		if dividendShift < 0 {
			rem = int64(lhs.lsu[lhsNonZero])
		}
		stop := rhsNonZero + rhsLeftShiftUnits
		var i int // i is lhs index
		for i = buflen - 1 + dividendShift; i >= stop; i-- {
			u = rem*Unit + int64(buf[i])
			q = u / d     // div
			rem = u - q*d // update remainder
		}
		resultNonZero := -1
		if rem > 0 {
			result.lsu[i+1] = int32(rem)
			resultNonZero = i + 1
		}
		for ; i >= 0; i-- { // copy rest of lhs into result
			result.lsu[i] = buf[i]
			if buf[i] > 0 && resultNonZero < 0 {
				resultNonZero = i
			}
		}
		if resultNonZero >= remainderFracUnits {
			result.intg = int8(resultNonZero + 1 - remainderFracUnits)
		} else {
			result.intg = 0
		}
		result.frac = int8(remainderFrac) // keep fraction precision like subtraction
		return nil
	}

	buf1len := lhsNonZero + 1 + lhsLeftShiftUnits
	buf2len := rhsNonZero + 1 + rhsLeftShiftUnits

	// D1. normalization
	normFactor := Unit / (rhs.lsu[rhsNonZero] + 1)
	// normalize buf1 and buf2
	var buf1 [MaxUnits * 2]int32 // store normalized lhs with one extra leading unit
	var buf2 [MaxUnits * 2]int32 // store normalized rhs
	if normFactor == 1 {         // happy path, if rhs.lsu[rhsNonZero] >= Unit / 2
		copy(buf1[lhsLeftShiftUnits:buf1len], lhs.lsu[:lhsNonZero+1])
		copy(buf2[rhsLeftShiftUnits:buf2len], rhs.lsu[:rhsNonZero+1])
	} else {
		var carry int64
		var i int // loop index control
		// normalize lhs into buf1
		for i = 0; i <= lhsNonZero; i++ {
			v := int64(lhs.lsu[i])*int64(normFactor) + carry
			carry = v / Unit
			buf1[i+lhsLeftShiftUnits] = int32(v - carry*Unit)
		}
		buf1[i+lhsLeftShiftUnits] = int32(carry) // additional carry on buf1[buf1len]
		carry = 0
		// normalize rhs into buf2
		for i = 0; i <= rhsNonZero; i++ {
			v := int64(rhs.lsu[i])*int64(normFactor) + carry
			carry = v / Unit
			buf2[i+rhsLeftShiftUnits] = int32(v - carry*Unit)
		}
		assertTrue(carry == 0, "carry must be 0 in divider normalization")
	}
	stop := buf2len - 1                                // stop index
	vd0 := int64(buf2[rhsNonZero+rhsLeftShiftUnits])   // rhs most significant unit
	vd1 := int64(buf2[rhsNonZero+rhsLeftShiftUnits-1]) // rhs second significant unit
	for i := buf1len + dividendShift - 1; i >= stop; i-- {
		// D3. make the guess on u1
		u0 := int64(buf1[i+1])
		var u1 int64
		if i >= 0 {
			u1 = int64(buf1[i])
		}
		v := u0*Unit + u1
		qhat := v / vd0
		rhat := v - qhat*vd0
		assertTrue(qhat < Unit, "qhat must be less than Unit")
		var u2 int64
		if i > 0 {
			u2 = int64(buf1[i-1])
		}
		for qhat*vd1 > rhat*Unit+u2 { // check if qhat can satisfy next unit
			qhat--      // decrese qhat
			rhat += vd0 // increase rhat
		}
		// D4. multiply and subtract
		var mulV, mulV0, carry int64 // the product, product within current unit, carry of multiplication
		var subV, borrow int32       // the diff and boorow of subtraction
		var k, msIdx int
		for msIdx = i - buf2len + 1; k < buf2len; k, msIdx = k+1, msIdx+1 {
			mulV = qhat*int64(buf2[k]) + carry                              // mul
			carry = mulV / Unit                                             // update carry
			mulV0 = mulV - carry*Unit                                       // in current unit
			subV, borrow = subWithBorrow(buf1[msIdx], int32(mulV0), borrow) // sub
			buf1[msIdx] = subV                                              // update buf1 with result
		}
		borrow = buf1[msIdx] - int32(carry) + borrow
		if borrow == -1 { // qhat is larger, cannot satisfy the whole decimal
			// D6. add back (reverse subtract)
			qhat--                                                              // decrease qhat
			borrow = 0                                                          // reset borrow to zero
			for msIdx = i - buf2len + 1; k < buf2len; k, msIdx = k+1, msIdx+1 { // reverse subtract
				buf1[msIdx], borrow = subWithBorrow(0, buf1[msIdx], borrow)
			}
		}
		buf1[msIdx] = 0 // clear buf1 because multiply w/ subtract succeeds
	}
	// now we have remainder in buf1
	assertTrue(buf1[buf1len] == 0, "value must be zero")

	// divide by normFactor, put quotient in result
	var rem int64
	resultNonZero := -1
	for i := buf1len - 1; i >= 0; i-- {
		v := rem*Unit + int64(buf1[i])
		if v == 0 {
			continue
		}
		q := v / int64(normFactor)
		rem = v - q*int64(normFactor) // update remainder
		if q > 0 && resultNonZero < 0 {
			resultNonZero = i
		}
		result.lsu[i] = int32(q) // update result
	}
	// because we multiply lhs and rhs with identical normFactor, the remainder must be zero.
	assertTrue(rem == 0, "remainder must be zero")

	if resultNonZero >= remainderFracUnits {
		result.intg = int8(resultNonZero + 1 - remainderFracUnits)
	} else {
		result.intg = 0
	}
	result.frac = int8(remainderFrac) // keep fraction precision like subtraction
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
