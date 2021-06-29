// Serialization and deserialization of fixed-point decimal
package fxd

import (
	"reflect"
	"unsafe"
)

// DecimalFromString parses provided string and fills the value into given decimal.
func DecimalFromAsciiString(s string, result *FixedDecimal) error {
	var bs []byte
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return DecimalFromBytesString(bs, result)
}

// DecimalFromBytesString creates a new FixedPointDecimal given numeric string.
func DecimalFromBytesString(bs []byte, result *FixedDecimal) error {
	exp := 0 // working exponent [assume 0]
	d := 0   // count of digits found in decimal part
	var dotchar int = -1
	var last int = -1
	var cfirst int = 0
	var neg bool

	// todo: check valid operands
	var i int
	var c byte
	var moreToProcess bool
	for i, c = range bs { // input character
		if c >= '0' && c <= '9' { // test for digit char
			last = i
			d++      // count of real digits
			continue // still in decimal part
		}
		if c == '.' && dotchar == -1 {
			dotchar = i
			if i == cfirst {
				cfirst++ // first digit must follow
			}
			continue
		}
		if i == 0 { // first in string
			if c == '-' { // valid - sign
				cfirst++
				neg = true
				continue
			}
			if c == '+' { // valid + sign
				cfirst++
				continue
			}
		}
		// c is not a digit, or a valid '+', '-' or '.'
		moreToProcess = true // indicate more data to process
		break
	}

	if last == -1 { // no digits yet
		if i == len(bs)-1 { // and no more to come
			return DecErrConversionSyntax
		}
		// Infinities and NaNs are possible, here
		if dotchar != -1 { // unless has a dot
			return DecErrConversionSyntax
		}
		result.SetZero() // be optimitic
		if decBiStr(bs[i:], decStrInfinityUpperFull, decStrInfinityLowerFull) || decBiStr(bs[i:], decStrInfinityUpperFull, decStrInfinityLowerAbbr) {
			result.setInf()
			return nil
		}
		// a NaN expected
		if c = bs[i]; c != 'n' && c != 'N' {
			return DecErrConversionSyntax
		}
		i++
		if c = bs[i]; c != 'a' && c != 'A' {
			return DecErrConversionSyntax
		}
		i++
		if c = bs[i]; c != 'n' && c != 'N' {
			return DecErrConversionSyntax
		}
		i++
		// now either nothing, or nnnn payload, expected
		// -> start of integer and skip leading 0s
		for cfirst = i; cfirst < len(bs); cfirst++ {
			if bs[cfirst] != '0' {
				return DecErrConversionSyntax
			}
		}
		if cfirst == len(bs) { // "NaN", maybe with all 0s
			result.setNaN()
			return nil
		}
		// todo: additinal payload after NaN
		// currently does not support this format
		return DecErrConversionSyntax
	} else if moreToProcess { // more to process
		// had some digits; exponent is only valid sequence now
		var nege bool         // 1=negative exponent
		var firstexp int = -1 // -> first sginificant exponent digit
		if c = bs[i]; c != 'e' && c != 'E' {
			return DecErrConversionSyntax
		}
		// Found 'e' or 'E'
		// sign no longer required
		i++ // to (possible) sign
		c = bs[i]
		if c == '-' {
			nege = true
			i++
		} else if c == '+' {
			i++
		}
		if i == len(bs) {
			return DecErrConversionSyntax
		}
		for {
			c = bs[i]
			if c == '0' && i != len(bs)-1 { // strip insignificant zeros
				i++
			} else {
				break
			}
		}
		firstexp = i // save exponent digit place
		for ; i < len(bs); i++ {
			c = bs[i]
			if c < '0' || c > '9' { // not a digit
				return DecErrConversionSyntax
			}
			exp = exp*10 + int(c) - int('0')
		}
		// maximum exponent is 65, with sign, at most 4 chars
		if i >= firstexp+4 {
			return DecErrConversionSyntax
		}
		if (!nege && exp > MaxDigits) || (nege && exp > MaxFrac) {
			return DecErrConversionSyntax
		}
		if nege {
			exp = -exp
		}
	}
	// Here when whole string has been inspected; syntax is good
	// cfirst->first digit(never dot), last->last digit(ditto)

	var fracDigits int
	if dotchar == -1 || last < dotchar { // no dot found or dot is last char
		fracDigits = 0
	} else {
		fracDigits = last - dotchar
	}
	// apply exponent to frac
	frac := fracDigits - exp
	var digits int
	headingZeros := 0 // additional 0s for first unit
	if frac == 0 {    // integer only
		digits = d
	} else if frac > 0 { // have fractional part, should move dot to left
		if d > frac { // have both integral and fractional part
			digits = d
		} else { // only fractional part
			digits = frac
			headingZeros = frac - d
		}
	} else { // no fractional part and need to move dot to right
		digits = d - frac
		frac = 0
	}

	if digits > MaxDigits {
		// currently fail at parsing
		// todo: apply rounding strategy
		return DecErrConversionSyntax
	}

	// units of integral part, and fractional part
	intgUnits := getUnits(digits - frac)
	fracUnits := getUnits(frac)
	up := intgUnits + fracUnits - 1 // lsu unit index, from highest to lowest
	// reset i as parse index
	i = cfirst
	// parse integral part
	if intgUnits > 0 {
		out := 0 // accumulator in unit
		cut := digits - frac
		for ; ; i++ {
			c = bs[i]
			if c == '.' { // ignore '.', this may be caused by exponent normalization
				continue
			}
			out = out*10 + int(c) - int('0')
			cut--
			if cut == 0 {
				break // nothing for this unit
			}
			if i == last { // no more digits to read, adjust out if cut > 0
				break
			}
			if mod9(cut) > 0 {
				continue
			}
			result.lsu[up] = int32(out) // write unit
			up--                        // prepare for unit below
			out = 0
		}
		// input integral digits processed.
		// increment i for frac processing
		i++
		// handle cut > 0, e.g. '1E2', '1E20'
		re := mod9(cut)
		result.lsu[up] = int32(out * pow10[re])
		up--
		out = 0
		cut -= re
		for ; cut > 0; cut -= DigitsPerUnit { // zero with large exponent
			result.lsu[up] = 0
			up--
		}
	}
	if fracUnits > 0 {
		out := 0 // accumulator in unit
		cut := DigitsPerUnit
		for ; headingZeros >= DigitsPerUnit; headingZeros -= DigitsPerUnit { // current unit filled all zeros
			result.lsu[up] = 0
			up--
		}
		cut -= headingZeros
		// parse fractional part
		for ; ; i++ {
			c = bs[i]
			if c == '.' { // ignore '.', this may be caused by exponent normalization
				continue
			}
			cut--
			out += (int(c) - int('0')) * pow10[cut]
			if i == last { // done
				break
			}
			if cut > 0 {
				continue // more for this unit
			}
			result.lsu[up] = int32(out) // write unit
			up--                        // prepare for unit below
			cut = DigitsPerUnit
			out = 0
		}
		result.lsu[up] = int32(out) // write lsu
	}
	result.intg = int8(digits - frac)
	result.frac = int8(frac)
	if neg {
		result.setNeg()
	}
	return nil
}

// ToString converts this decimal to string format.
// frac specify the fractional precision of the output decimal string.
// if frac < 0, will output all fractional digits.
// if frac >= 0, will truncate to frac digits
func (fd *FixedDecimal) ToString(frac int) string {
	return string(fd.AppendStringBuffer(nil, frac))
}

// AppendStringBuffer appends this decimal's formatted string to given buffer.
// if frac < 0, will output all fractional digits.
// if frac >= 0, will truncate to frac digits.
func (fd *FixedDecimal) AppendStringBuffer(buf []byte, frac int) []byte {
	if buf == nil {
		buf = make([]byte, 0, 16)
	}
	if fd.IsNeg() {
		buf = append(buf, '-')
	}
	if fd.IsInf() {
		buf = append(buf, decStrInfinityOutput...)
		return buf
	}
	if fd.IsNaN() {
		buf = append(buf, decStrNaNOutput...)
		return buf
	}

	intgUnits, fracUnits := fd.IntgUnits(), fd.FracUnits()
	var up = intgUnits + fracUnits - 1
	if intgUnits > 0 { // integral part
		eliminateHeadingZeros := true
		for ; up >= fracUnits; up-- { // loop until fracional part
			// for each unit, we divide into 3 3-digit parts and print them each time
			uv := fd.lsu[up]
			if eliminateHeadingZeros && uv == 0 { // rare case, integral part exists but zero
				eliminateHeadingZeros = false
				buf = append(buf, '0')
				continue
			}
			// XXX,xxx,xxx
			if uv >= BinHighUnit {
				high := uv / BinHighUnit
				buf, eliminateHeadingZeros = d3str(int(high), buf, eliminateHeadingZeros)
				uv -= high * BinHighUnit
			} else {
				buf, eliminateHeadingZeros = d3str(0, buf, eliminateHeadingZeros)
			}
			// xxx,XXX,xxx
			if uv >= BinMidUnit {
				mid := uv / BinMidUnit
				buf, eliminateHeadingZeros = d3str(int(mid), buf, eliminateHeadingZeros)
				uv -= mid * BinMidUnit
			} else {
				buf, eliminateHeadingZeros = d3str(0, buf, eliminateHeadingZeros)
			}
			// xxx,xxx,XXX
			buf, eliminateHeadingZeros = d3str(int(uv), buf, eliminateHeadingZeros)
		}
	} else {
		buf = append(buf, '0') // append 0 if no integral part
	}
	if frac == 0 { // no output on fractional digits
		return buf
	}
	if frac < 0 { // frac is not specified
		if fracUnits > 0 { // fractional part exists
			buf = append(buf, '.') // append dot
			fracDigits := int(fd.Frac())
			buf = appendFrac(fd, up, fracDigits, buf)
		}
		return buf
	}

	// frac is specified
	if fracUnits == 0 { // no fractional part
		buf = append(buf, '.')
		for i := 0; i < frac; i++ { // append zeros
			buf = append(buf, '0')
		}
		return buf
	}

	fullFracDigits := fracUnits * DigitsPerUnit
	if frac <= fullFracDigits { // this decimal supports specified frac
		buf = append(buf, '.')
		buf = appendFrac(fd, up, frac, buf)
		return buf
	}

	// this decimal does not have sufficient precision
	buf = append(buf, '.')
	buf = appendFrac(fd, up, fullFracDigits, buf)
	// add extra zeros
	for i := 0; i < frac-fullFracDigits; i++ {
		buf = append(buf, '0')
	}
	return buf
}

func appendFrac(fd *FixedDecimal, up int, fracDigits int, buf []byte) []byte {
	for ; up >= 0; up-- {
		// for each unit, divide into 3 3-digit parts and print.
		// need to check whether the digits exceeds fractional length
		if fracDigits == 0 { // already reached the fractional precision
			break
		}
		uv := fd.lsu[up]
		if uv == 0 { // all zeros
			if fracDigits > DigitsPerUnit {
				for j := 0; j < DigitsPerUnit; j++ {
					buf = append(buf, '0')
				}
				fracDigits -= DigitsPerUnit
				continue
			}
			for j := 0; j < fracDigits; j++ {
				buf = append(buf, '0')
			}
			fracDigits = 0
			break
		}
		// non-zero
		// XXX,xxx,xxx
		if uv >= BinHighUnit {
			if fracDigits == 0 {
				break
			}
			high := uv / BinHighUnit
			startIdx := int(high)*4 + 1
			if fracDigits < 3 {
				buf = append(buf, BIN2CHAR[startIdx:startIdx+fracDigits]...)
				fracDigits = 0
				break
			} else {
				buf = append(buf, BIN2CHAR[startIdx:startIdx+3]...)
				fracDigits -= 3
			}
			uv -= high * BinHighUnit
		} else {
			if fracDigits < 3 {
				buf = append(buf, BIN2CHAR[1:1+fracDigits]...)
				fracDigits = 0
			} else {
				buf = append(buf, BIN2CHAR[1:4]...)
				fracDigits -= 3
			}
		}
		if uv >= BinMidUnit {
			if fracDigits == 0 {
				break
			}
			mid := uv / BinMidUnit
			startIdx := int(mid)*4 + 1
			if fracDigits < 3 {
				buf = append(buf, BIN2CHAR[startIdx:startIdx+fracDigits]...)
				fracDigits = 0
				break
			} else {
				buf = append(buf, BIN2CHAR[startIdx:startIdx+3]...)
				fracDigits -= 3
			}
			uv -= mid * BinMidUnit
		} else {
			if fracDigits < 3 {
				buf = append(buf, BIN2CHAR[1:1+fracDigits]...)
				fracDigits = 0
			} else {
				buf = append(buf, BIN2CHAR[1:4]...)
				fracDigits -= 3
			}
		}
		if fracDigits == 0 {
			break
		}
		startIdx := int(uv)*4 + 1
		if fracDigits < 3 {
			buf = append(buf, BIN2CHAR[startIdx:startIdx+fracDigits]...)
			fracDigits = 0
			break
		} else {
			buf = append(buf, BIN2CHAR[startIdx:startIdx+3]...)
			fracDigits -= 3
		}
	}
	return buf
}
