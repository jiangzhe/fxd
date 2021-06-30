package fxd

func (fd *FixedDecimal) Compare(rhs *FixedDecimal) int {
	lneg := fd.IsNeg()
	rneg := rhs.IsNeg()
	if lneg { // left is negative
		if rneg { // right is negative too
			return cmpAbs(rhs, fd) // swap and compare absolute value
		}
		return -1
	}
	// left is positive
	if rneg { // right is negative
		return 1
	}
	// both are positive
	return cmpAbs(fd, rhs)
}

func cmpAbs(lhs *FixedDecimal, rhs *FixedDecimal) int {
	liu, lfu := lhs.IntgUnits(), lhs.FracUnits()
	riu, rfu := rhs.IntgUnits(), rhs.FracUnits()
	return cmpAbsLsu(liu, lfu, &lhs.lsu, riu, rfu, &rhs.lsu)
}

func cmpAbsLsu(liu, lfu int, llsu *[MaxUnits]int32, riu, rfu int, rlsu *[MaxUnits]int32) int {
	// compare integral parts
	for liu > 0 && liu > riu { // lhs has more integral units
		if llsu[lfu+liu-1] > 0 {
			return 1
		}
		liu--
	}
	for riu > 0 && riu > liu { // rhs has more integral units
		if rlsu[rfu+riu-1] > 0 {
			return -1
		}
		riu--
	}
	// both have identical number of integral units: liu==riu
	i := liu + lfu - 1
	j := riu + rfu - 1
	for ; i >= 0 && j >= 0; i, j = i-1, j-1 {
		lv := llsu[i]
		rv := rlsu[j]
		if lv > rv {
			return 1
		}
		if lv < rv {
			return -1
		}
	}
	for ; i >= 0; i-- { // lhs still has units
		if llsu[i] > 0 {
			return 1
		}
	}
	for ; j >= 0; j-- { // rhs still has units
		if rlsu[j] > 0 {
			return -1
		}
	}
	// both has no units so they are same
	return 0
}
