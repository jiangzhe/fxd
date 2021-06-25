// Fixed-point decimal
//
// We use 9 4-byte integers to store decimals.
// The largest number we can represent is 10^65-1.
// Each 4-byte integer represents a number from 0-999,999,999,
// in another word, the number is base 1,000,000,000.
package fxd

const MaxDigits = 65
const MaxFrac = 30

const DigitsPerUnit = 9
const Unit = 1_000_000_000
const MaxIntgUnits = (MaxDigits + DigitsPerUnit - 1) / DigitsPerUnit
const MaxFracUnits = (MaxFrac + DigitsPerUnit - 1) / DigitsPerUnit

// DECIMAL(65, 28) requires 9 units to store the number.
// The integral part is 37 digits so requires 5 units.
// The fractional part is 28 digits so requires 4 units.
const MaxUnits = 9
const DoubleMaxUnits = MaxUnits * 2

// FixedDecimal is an implementation of "exact number" type
// defined in SQL standard.
// It directly maps to the MYSQL data type "DECIMAL".
// It differs from floating-point decimal, because its allocation
// size is fixed: 9*4+2 = 38 bytes.
type FixedDecimal struct {
	// integral digit number.
	// maximum is 65, if frac is 0.
	// we use most significant bit to indicates whether it is negative
	intg int8
	// fractional digit number.
	// maximum is 30.
	// we use first 2 bits to indicates INF and NaN
	// 00: normal
	// 01: INF
	// 10: NaN
	frac int8
	lsu  [MaxUnits]int32
}

// IsNeg returns true if this decimal is negative.
func (fd *FixedDecimal) IsNeg() bool {
	return uint8(fd.intg)&0x80 != 0
}

func (fd *FixedDecimal) setNeg() {
	fd.intg |= ^0x7f
}

// set negative flag to this decimal.
// To avoid negating zero value, check and reset if all units are zero.
func (fd *FixedDecimal) setNegAndCheckZero() {
	fd.setNeg()
	if fd.allUnitsZero() {
		fd.SetZero()
	}
}

func (fd *FixedDecimal) setPos() {
	fd.intg &= 0x7f
}

// IsNaN returns true if this decimal is not a number.
func (fd *FixedDecimal) IsNaN() bool {
	return uint8(fd.frac)&0x80 != 0
}

func (fd *FixedDecimal) setNaN() {
	fd.frac |= ^0x7f
}

// IsInf returns true if this decimal is infinity.
func (fd *FixedDecimal) IsInf() bool {
	return uint8(fd.frac)&0x40 != 0
}

func (fd *FixedDecimal) setInf() {
	fd.frac |= 0x40
}

// IsSpecial returns true if this decimal is special (NaN or Inf).
func (fd *FixedDecimal) IsSpecial() bool {
	return uint8(fd.intg)&0x80 != 0 || uint8(fd.frac)&0xcf != 0
}

func (fd *FixedDecimal) setNormal() {
	fd.intg &= 0x7f
	fd.frac &= 0x3f
}

// reset this decimal to zero.
func (fd *FixedDecimal) SetZero() {
	fd.intg = 1 // with negative=false
	fd.frac = 0 // with inf=false, NaN=false
	fd.resetUnits()
}

func (fd *FixedDecimal) SetOne() {
	fd.intg = 1
	fd.frac = 0
	fd.resetUnits()
	fd.lsu[0] = 1
}

// reset units.
// this method is useful especially in case the struct
// is cached/reused due to performance consideration.
func (fd *FixedDecimal) resetUnits() {
	fd.lsu[0] = 0
	fd.lsu[1] = 0
	fd.lsu[2] = 0
	fd.lsu[3] = 0
	fd.lsu[4] = 0
	fd.lsu[5] = 0
	fd.lsu[6] = 0
	fd.lsu[7] = 0
	fd.lsu[8] = 0
}

// IsZero returns true if this decimal is zero.
func (fd *FixedDecimal) IsZero() bool {
	// single value comparison on intg and frac excludes special cases
	return fd.lsu[0] == 0 && fd.intg == 1 && fd.frac == 0
}

// allUnitsZero returns true if all units are zero.
func (fd *FixedDecimal) allUnitsZero() bool {
	return fd.lsu[0] == 0 &&
		fd.lsu[1] == 0 &&
		fd.lsu[2] == 0 &&
		fd.lsu[3] == 0 &&
		fd.lsu[4] == 0 &&
		fd.lsu[5] == 0 &&
		fd.lsu[6] == 0 &&
		fd.lsu[7] == 0 &&
		fd.lsu[8] == 0
}

// Intg returns the integral digit number of this decimal.
func (fd *FixedDecimal) Intg() int8 {
	return fd.intg & 0x7f
}

// IntgUnits returns unit number to store integral digits.
func (fd *FixedDecimal) IntgUnits() int {
	return getUnits(int(fd.Intg()))
}

// Frac returns the fractional digit number of this decimal.
func (fd *FixedDecimal) Frac() int8 {
	return fd.frac & 0x3f
}

// FracUnits returns unit number to store fractional digits.
func (fd *FixedDecimal) FracUnits() int {
	return getUnits(int(fd.Frac()))
}

// FromInt64 set int64 value into this decimal.
// If reset flag is true, always clear all fields first.
func (fd *FixedDecimal) FromInt64(val int64, reset bool) {
	if reset {
		fd.intg = 0
		fd.frac = 0
		fd.resetUnits()
	}
	if val == 0 { // is zero
		fd.intg = 1
		return
	}
	neg := val < 0
	if neg {
		val = -val // convert to positive
	}
	var i int
	for val != 0 {
		q := val / Unit
		r := val - q*Unit
		fd.lsu[i] = int32(r)
		i++
		val = q
	}
	fd.intg = int8(i * DigitsPerUnit) // possible maximum integral digits
	if neg {
		fd.setNeg()
	}
	return
}

// DecimalZero creates a new decimal with zero value.
func DecimalZero() (fd FixedDecimal) {
	fd.intg = 1
	return
}

func DecimalOne() (fd FixedDecimal) {
	fd.intg = 1
	fd.lsu[0] = 1
	return
}

// DecimalFromInt64 creates a new decimal from provided int64.
func DecimalFromInt64(val int64) (fd FixedDecimal) {
	fd.FromInt64(val, false)
	return
}
