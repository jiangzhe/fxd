package fxd

var DecMaxInt64 FixedDecimal

const MaxInt64 int64 = 9223372036854775807

var DecMinInt64 FixedDecimal

const MinInt64 int64 = -9223372036854775808

func init() {
	DecMaxInt64, _ = DecimalFromAsciiString("9223372036854775807")
	DecMinInt64, _ = DecimalFromAsciiString("-9223372036854775808")
}

func (fd *FixedDecimal) ToInt() int64 {
	tgt := fd
	tgt.Round(0) // round to integral number
	neg := tgt.IsNeg()
	if neg && fd.Compare(&DecMinInt64) <= 0 { // less than or equal to min value
		return MinInt64
	}
	if !neg && fd.Compare(&DecMaxInt64) >= 0 { // greater than or equal to max value
		return MaxInt64
	}
	intgUnits := tgt.IntgUnits()
	var unit int64 = 1
	var sum int64
	for _, v := range tgt.lsu[:intgUnits] {
		sum += int64(v) * unit
		unit *= Unit
	}
	if neg {
		return -sum
	}
	return sum
}
