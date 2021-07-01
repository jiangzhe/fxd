package fxd

import (
	"fmt"
	"testing"
)

func TestDecimalFromString(t *testing.T) {
	var fd FixedDecimal
	for _, s := range []string{
		"0",
		"1",
		"-1",
		"123",
		"123456789012345",
		"0.1",
		"0.123",
		"1.0",
		"-1.0",
		"1E1",
		"1E+2",
		"1.0E-2",
		"1.0e10",
		"1.0e02",
	} {
		if err := fd.FromBytesString([]byte(s), true); err != nil {
			t.Fatal("failed")
		}
		fmt.Printf("input=%v, result=%v\n", s, fd)
	}

	if _, err := DecimalFromAsciiString("100"); err != nil {
		t.Fatal("failed")
	}
	if _, err := DecimalFromBytesString([]byte("100")); err != nil {
		t.Fatal("failed")
	}
}

func TestDecimalFromStringError(t *testing.T) {
	var fd FixedDecimal
	for _, s := range []string{
		"",
		"abc",
		".",
		".a",
		".NaN",
		"N",
		"Nb",
		"Na",
		"Nab",
		"NaNx",
		"0x",
		"0E",
		".1e+",
		".1e-",
		".1e+f",
		".1e12345",
		".1e-200",
		"1234567890123456789012345678901234567890123456789012345678901234567890",
	} {
		if err := fd.FromBytesString([]byte(s), true); err == nil {
			t.Fatal("failed")
		}
	}
}

func TestDecimalFromInt64(t *testing.T) {
	fd := DecimalZero()
	if fd.ToString(-1) != "0" {
		t.Fatal("failed")
	}
	fd = DecimalOne()
	if fd.ToString(-1) != "1" {
		t.Fatal("failed")
	}
	fd = DecimalFromInt64(42)
	if fd.ToString(-1) != "42" {
		t.Fatal("failed")
	}
	for _, n := range []int64{
		0, 1, -1, 100, 1 << 32, (1 << 63) - 1, -(1 << 63),
	} {
		fd.FromInt64(n, true)
		fmt.Printf("input=%v, result=%v\n", n, fd)
	}
}

func TestDecimalToString(t *testing.T) {
	type tscase struct {
		input    string
		frac     int
		expected string
	}

	var fd FixedDecimal
	for _, c := range []tscase{
		{"0", -1, "0"},
		{"1", -1, "1"},
		{"-1", -1, "-1"},
		{"+1", -1, "1"},
		{"123", -1, "123"},
		{"123456789012345", -1, "123456789012345"},
		{"0.0", -1, "0.0"},
		{"0.100", -1, "0.100"},
		{"0.1", -1, "0.1"},
		{"0.12345678901234567890", -1, "0.12345678901234567890"},
		{"0.123", -1, "0.123"},
		{"1.0", -1, "1.0"},
		{"-1.0", -1, "-1.0"},
		{"1e0", -1, "1"},
		{"1E1", -1, "10"},
		{"1E+2", -1, "100"},
		{"1.0E-2", -1, "0.010"},
		{"1.0e10", -1, "10000000000"},
		{"1.2345e20", -1, "123450000000000000000"},
		{"5.4433e4", -1, "54433"},
		{"5.4433e3", -1, "5443.3"},
		{"5.4433e2", -1, "544.33"},
		{"5.4433e1", -1, "54.433"},
		{"5.4433e0", -1, "5.4433"},
		{"5.4433e-1", -1, "0.54433"},
		{"5.4433e-2", -1, "0.054433"},
		{"5.4433e-3", -1, "0.0054433"},
		{"5.4433e-5", -1, "0.000054433"},
		{"5.4433e-6", -1, "0.0000054433"},
		{"5.4433e-7", -1, "0.00000054433"},
		{"5.4433e-8", -1, "0.000000054433"},
		{"5.4433e-9", -1, "0.0000000054433"},
		{"5.4433e-10", -1, "0.00000000054433"},
		{"5.4433e-11", -1, "0.000000000054433"},
		{"5.4433e-20", -1, "0.000000000000000000054433"},
		{"Inf", -1, "Infinity"},
		{"NaN", -1, "NaN"},
		{"123456789.123456789", 0, "123456789"},
		{"123456789.123456789", 1, "123456789.1"},
		{"123456789.123456789", 2, "123456789.12"},
		{"123456789.123456789", 3, "123456789.123"},
		{"123456789.123456789", 4, "123456789.1234"},
		{"123456789.123456789", 5, "123456789.12345"},
		{"123456789.123456789", 6, "123456789.123456"},
		{"123456789.123456789", 7, "123456789.1234567"},
		{"123456789.123456789", 8, "123456789.12345678"},
		{"123456789.123456789", 9, "123456789.123456789"},
		{"123456789.123456789", 10, "123456789.1234567890"},
		{"1.021", 2, "1.02"},
		{"1.00021", 2, "1.00"},
	} {
		if err := fd.FromBytesString([]byte(c.input), true); err != nil {
			t.Fatalf("failed %v", err)
		}
		actual := fd.ToString(c.frac)
		if actual != c.expected {
			t.Fatalf("result mismatch: actual=%v, expected=%v", actual, c.expected)
		}
	}
}

func TestDecimalAdd(t *testing.T) {
	type tcase struct {
		input1, input2, expected string
	}
	var fd1 FixedDecimal
	var fd2 FixedDecimal
	var fd3 FixedDecimal
	for _, c := range []tcase{
		{"0", "0", "0"},
		{"0", "1", "1"},
		{"1", "1", "2"},
		{"1", "-1", "0"},
		{"-1", "1", "0"},
		{"-1", "-100", "-101"},
		{"5", "5", "10"},
		{"1.0", "0", "1.0"},
		{"1.0", "0.0", "1.0"},
		{"-1.0", "0.01", "-0.99"},
		{"-0.3", "1.27", "0.97"},
		{"-0.3", "0.2", "-0.1"},
		{"-0.01", "0.001", "-0.009"},
		{"-123", "0.1", "-122.9"},
		{"1", "-12.5", "-11.5"},
		{"-5.0", "5.0", "0"},
		{"1.0", "0.1", "1.1"},
		{"1.01", "0.1", "1.11"},
		{"1.00000000001", "1000.01", "1001.01000000001"},
		{"1.234567890", "0.0000000001", "1.2345678901"},
		{"-1.234567890", "0.0000000001", "-1.2345678899"},
		{"10000000000", "1", "10000000001"},
		{"1", "10000000000", "10000000001"},
		{"999999999", "1", "1000000000"},
		{"1", "999999999", "1000000000"},
		{"999999999999999999", "1", "1000000000000000000"},
		{"1", "999999999999999999", "1000000000000000000"},
	} {
		if err := fd1.FromAsciiString(c.input1, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := fd2.FromAsciiString(c.input2, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := DecimalAdd(&fd1, &fd2, &fd3); err != nil {
			t.Fatalf("failed %v", err)
		}
		actual := fd3.ToString(-1)
		fmt.Printf("%v+%v=%v, intg=%v, frac=%v\n", fd1.ToString(-1), fd2.ToString(-1), actual, fd3.Intg(), fd3.Frac())
		if actual != c.expected {
			t.Fatalf("result mismatch")
		}
	}
}

func TestDecimalSub(t *testing.T) {
	type tcase struct {
		input1, input2, expected string
	}
	var fd1 FixedDecimal
	var fd2 FixedDecimal
	var fd3 FixedDecimal
	for _, c := range []tcase{
		{"0", "0", "0"},
		{"0", "1", "-1"},
		{"0", "-1", "1"},
		{"-1", "0", "-1"},
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"1", "2", "-1"},
		{"2", "1", "1"},
		{"1", "-1", "2"},
		{"-1", "1", "-2"},
		{"-1", "-100", "99"},
		{"1.0", "0", "1.0"},
		{"1.0", "0.0", "1.0"},
		{"-1.0", "0.01", "-1.01"},
		{"-0.3", "1.27", "-1.57"},
		{"-0.3", "-0.2", "-0.1"},
		{"-0.3", "0.2", "-0.5"},
		{"-0.01", "0.001", "-0.011"},
		{"-123", "0.1", "-123.1"},
		{"1", "-12.5", "13.5"},
		{"-5.0", "5.0", "-10.0"},
		{"1.0", "0.1", "0.9"},
		{"1.01", "0.1", "0.91"},
		{"1.00000000001", "1000.01", "-999.00999999999"},
		{"1.234567890", "0.0000000001", "1.2345678899"},
		{"-1.234567890", "0.0000000001", "-1.2345678901"},
		{"1000000000", "1", "999999999"},
		{"1", "1000000000", "-999999999"},
	} {
		if err := fd1.FromAsciiString(c.input1, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := fd2.FromAsciiString(c.input2, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := DecimalSub(&fd1, &fd2, &fd3); err != nil {
			t.Fatalf("failed %v", err)
		}
		actual := fd3.ToString(-1)
		fmt.Printf("%v-%v=%v\n", fd1.ToString(-1), fd2.ToString(-1), actual)
		if actual != c.expected {
			t.Fatalf("result mismatch")
		}
	}
}

func TestDecimalMul(t *testing.T) {
	type tcase struct {
		input1, input2, expected string
	}
	var fd1 FixedDecimal
	var fd2 FixedDecimal
	var fd3 FixedDecimal
	for _, c := range []tcase{
		{"0", "0", "0"},
		{"0", "1", "0"},
		{"1", "0", "0"},
		{"1", "1", "1"},
		{"1", "2", "2"},
		{"2", "1", "2"},
		{"1", "-1", "-1"},
		{"-1", "1", "-1"},
		{"-1", "-100", "100"},
		{"1.0", "0", "0"},
		{"1.0", "0.0", "0.00"},
		{"-1.0", "0.01", "-0.010"},
		{"-0.3", "1.27", "-0.381"},
		{"-0.3", "-0.2", "0.06"},
		{"-0.3", "0.2", "-0.06"},
		{"-0.01", "0.001", "-0.00001"},
		{"-0.10", "0.001", "-0.00010"},
		{"-123", "0.1", "-12.3"},
		{"1", "-12.5", "-12.5"},
		{"-5.0", "5.0", "-25.00"},
		{"1.0", "0.1", "0.10"},
		{"1.01", "0.1", "0.101"},
		{"1.00000000001", "1000.01", "1000.0100000100001"},
		{"1.234567890", "0.0000000001", "0.0000000001234567890"},
		{"-1.234567890", "0.0000000001", "-0.0000000001234567890"},
	} {
		if err := fd1.FromAsciiString(c.input1, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := fd2.FromAsciiString(c.input2, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := DecimalMul(&fd1, &fd2, &fd3); err != nil {
			t.Fatalf("failed %v", err)
		}
		actual := fd3.ToString(-1)
		fmt.Printf("%v*%v=%v\n", fd1.ToString(-1), fd2.ToString(-1), actual)
		if actual != c.expected {
			t.Fatalf("result mismatch")
		}
	}
}

func TestDecimalDiv(t *testing.T) {
	type tcase struct {
		input1, input2, expected string
	}
	var fd1 FixedDecimal
	var fd2 FixedDecimal
	var fd3 FixedDecimal
	for _, c := range []tcase{
		{"0", "1", "0"},
		{"1", "1", "1.000000000"}, // incremental frac is already multiple of 9
		{"1", "1", "1.000000000"},
		{"1", "2", "0.500000000"},
		{"2", "1", "2.000000000"},
		{"1", "-1", "-1.000000000"},
		{"-1", "1", "-1.000000000"},
		{"-1", "-100", "0.010000000"},
		{"100", "1", "100.000000000"},
		{"100", "100", "1.000000000"},
		{"0.000000002", "1", "0.000000002000000000"},
		{"1.0", "2", "0.500000000"},
		{"1.0", "2.0", "0.500000000000000000"}, // two fractional units
		{"-1", "0.01", "-100.000000000"},
		{"0.27", "0.3", "0.900000000000000000"},
		{"-0.3", "-0.2", "1.500000000000000000"},
		{"0.3", "0.7", "0.428571428571428571"},
		{"0.6", "0.9", "0.666666666666666666"},
		{"-0.3", "0.2", "-1.500000000000000000"},
		{"1000000000.1", "7", "142857142.871428571"},
		{"1000000000.1", "9", "111111111.122222222"},
		{"101000000000.1", "7", "14428571428.585714285"},
		{"101000000000.1", "7.1", "14225352112.690140845070422535"},
		{"101000000000.1", "5", "20200000000.020000000"},
		{"101000000000.1", "5.0", "20200000000.020000000000000000"},
		{"100.10000000001", "7", "14.300000000001428571"},
		{"100.10000000001", "7.0", "14.300000000001428571428571428"},
		{"100.1", "7.0000000001", "14.299999999795714285717204081"},
		{"205.6", "9.5000000001", "21.642105262930083102495472809"},
		{"2000000005.1", "7.5000000001", "266666667.343111111102091851851972108"},
		{"1.2", "0.7", "1.714285714285714285"},
		{"1.22", "0.77", "1.584415584415584415"},
		{"1.222", "0.777", "1.572715572715572715"},
		{"1.2222", "0.7777", "1.571557155715571557"},
		{"1.22222", "0.77777", "1.571441428700001285"},
		{"1.222222", "0.777777", "1.571429857144142858"},
		{"1.2222222", "0.7777777", "1.571428700000012857"},
		{"1.22222222", "0.77777777", "1.571428584285714414285715571"},
		{"1.222222222", "0.777777777", "1.571428572714285715571428572"},
		{"9.8", "1", "9.800000000"},
		{"98.7", "1.2", "82.250000000000000000"},
		{"987.6", "12.3", "80.292682926829268292"},
		{"9876.5", "123.4", "80.036466774716369529"},
		{"98765.4", "1234.5", "80.004374240583232077"},
		{"987654.3", "12345.6", "80.000510303265940902"},
		{"9876543.2", "123456.7", "80.000058320042573631"},
		{"98765432.1", "1234567.8", "80.000006561000538002"},
		{"987654321.1", "12345678.9", "80.000000737100006707"},
		{"987654321.12", "12345678.99", "80.000000155520000281"},
		{"987654321.123", "12345678.998", "80.000000103923000120"},
		{"987654321.1234", "12345678.9987", "80.000000099419400109"},
		{"987654321.12345", "12345678.99876", "80.000000099034650108"},
		{"987654321.123456", "12345678.998765", "80.000000099002736108"},
		{"987654321.1234567", "12345678.9987654", "80.000000099000200808"},
		{"987654321.12345678", "12345678.99876543", "80.000000099000012888900031007"},
		{"987654321.123456789", "12345678.998765432", "80.000000099000000657900001515"},
		{"-9.8", "1", "-9.800000000"},
		{"-98.7", "1.2", "-82.250000000000000000"},
		{"-987.6", "12.3", "-80.292682926829268292"},
		{"-9876.5", "123.4", "-80.036466774716369529"},
		{"-98765.4", "1234.5", "-80.004374240583232077"},
		{"-987654.3", "12345.6", "-80.000510303265940902"},
		{"-9876543.2", "123456.7", "-80.000058320042573631"},
		{"-98765432.1", "1234567.8", "-80.000006561000538002"},
		{"-987654321.1", "12345678.9", "-80.000000737100006707"},
		{"-987654321.12", "12345678.99", "-80.000000155520000281"},
		{"-987654321.123", "12345678.998", "-80.000000103923000120"},
		{"-987654321.1234", "12345678.9987", "-80.000000099419400109"},
		{"-987654321.12345", "12345678.99876", "-80.000000099034650108"},
		{"-987654321.123456", "12345678.998765", "-80.000000099002736108"},
		{"-987654321.1234567", "12345678.9987654", "-80.000000099000200808"},
		{"-987654321.12345678", "12345678.99876543", "-80.000000099000012888900031007"},
		{"-987654321.123456789", "12345678.998765432", "-80.000000099000000657900001515"},
		{"0.170511", "-353390023.459963", "-0.000000000482500887"},
		{"0.170511", "-353390023", "-0.000000000482500888"},
		{"0.1", "300000000", "0.000000000"},
		{"0.1", "300000000.0", "0.000000000333333333"},
		{"0.1", "3000000000", "0.000000000"},
		{"0.1", "3000000000.0", "0.000000000033333333"},
		{"0.0000000001", "300000000", "0.000000000000000000"},
		{"0.0000000001", "300000000.0", "0.000000000000000000333333333"},
		{"0.0000000001", "3000000000", "0.000000000000000000"},
		{"0.0000000001", "3000000000.0", "0.000000000000000000033333333"},
		{"1", "300000000", "0.000000003"},
		{"1", "300000000.0", "0.000000003"},
		{"1", "3000000000", "0.000000000"},
		{"1", "3000000000.0", "0.000000000"},
		{"1.0", "300000000", "0.000000003"},
		{"1.0", "300000000.0", "0.000000003333333333"},
		{"1.0", "3000000000", "0.000000000"},
		{"1.0", "3000000000.0", "0.000000000333333333"},
		{"0.4", "0.000000003", "133333333.333333333333333333"},
		{"0.4", "0.0000000003", "1333333333.333333333333333333333333333"},
		{"0.2", "0.000000003", "66666666.666666666666666666"},
		{"0.2", "0.0000000003", "666666666.666666666666666666666666666"},
		{"400000000", "300000000", "1.333333333"},
		{"400000000.0", "300000000.0", "1.333333333333333333"},
		{"4000000000", "3000000000", "1.333333333"},
		{"4000000000.0", "3000000000.0", "1.333333333333333333"},
		{"200000000", "300000000", "0.666666666"},
		{"200000000.0", "300000000.0", "0.666666666666666666"},
		{"2000000000", "3000000000", "0.666666666"},
		{"2000000000.0", "3000000000.0", "0.666666666666666666"},
		{"400000000", "0.000000003", "133333333333333333.333333333333333333"},
		{"4000000000", "0.000000003", "1333333333333333333.333333333333333333"},
		{"1", "500000000.1", "0.000000001"},
	} {
		if err := fd1.FromAsciiString(c.input1, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := fd2.FromAsciiString(c.input2, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := DecimalDiv(&fd1, &fd2, &fd3, 4); err != nil {
			t.Fatalf("failed %v", err)
		}
		actual := fd3.ToString(-1)
		fmt.Printf("fd1=%v, fd2=%v\n", fd1, fd2)
		fmt.Printf("%v/%v=%v\n", fd1.ToString(-1), fd2.ToString(-1), actual)
		if actual != c.expected {
			t.Fatalf("result mismatch")
		}
	}
}

func TestDecimalCompare(t *testing.T) {
	type tcase struct {
		input1, input2 string
		expected       int
	}
	var fd1 FixedDecimal
	var fd2 FixedDecimal
	for _, c := range []tcase{
		{"0", "0", 0},
		{"0", "1", -1},
		{"1", "0", 1},
		{"-1", "-1", 0},
		{"1", "-1", 1},
		{"-1", "1", -1},
		{"-1", "-2", 1},
		{"1", "1", 0},
		{"2", "1", 1},
		{"1", "2", -1},
		{"1.0", "1", 0},
		{"1", "1.0", 0},
		{"1.000", "1.00", 0},
		{"1.000000000000", "1.00000", 0},
		{"1.01", "1", 1},
		{"1", "1.01", -1},
		{"1.02", "1.01", 1},
		{"1.01", "1.02", -1},
		{"1000000000", "999999999", 1},
		{"999999999", "1000000000", -1},
		{"1.0000000000000000000000000001", "1", 1},
		{"1", "1.0000000000000000000000000001", -1},
		{"1.0000000010000000000000000001", "1.0000000010000000000000000001", 0},
		{"1.0000000010000000000000000001", "1.0000000010000000000000000002", -1},
		{"1.0000000010000000000000000002", "1.0000000010000000000000000001", 1},
		{"1.0000000010000000010000000001", "1.0000000010000000010000000001", 0},
		{"1.0000000010000000010000000001", "1.0000000010000000010000000002", -1},
		{"1.0000000010000000010000000002", "1.0000000010000000010000000001", 1},
		{"100000000000000000000000000", "100000000000000000000000000", 0},
		{"100000000000000000000000001", "100000000000000000000000000", 1},
		{"100000000000000000000000000", "100000000000000000000000001", -1},
	} {
		if err := fd1.FromAsciiString(c.input1, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := fd2.FromAsciiString(c.input2, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		actual := fd1.Compare(&fd2)
		fmt.Printf("(%v)compare(%v) = %v\n", fd1.ToString(-1), fd2.ToString(-1), actual)
		if actual != c.expected {
			t.Fatalf("result mismatch")
		}
	}
}

func TestDecimalArithError(t *testing.T) {
	var fd1, fd2, fd3 FixedDecimal
	var err error
	_ = fd1.FromAsciiString("1", true)
	_ = fd2.FromAsciiString("0", true)
	// div by zero
	if err = DecimalDiv(&fd1, &fd2, &fd3, DivIncrFrac); err == nil {
		t.Fatal("failed")
	}
	// mod by zero
	if err = DecimalMod(&fd1, &fd2, &fd3); err == nil {
		t.Fatal("failed")
	}
	// mul overflow
	_ = fd1.FromAsciiString("1e41", true)
	_ = fd2.FromAsciiString("1e40", true)
	if err = DecimalMul(&fd1, &fd2, &fd3); err == nil {
		t.Fatal("failed")
	}
}

func TestDecimalArithNaN(t *testing.T) {
	var fd1, fd2, fd3 FixedDecimal
	var err error
	_ = fd1.FromAsciiString("NaN", true)
	_ = fd2.FromAsciiString("1", true)
	// add NaN
	if err = DecimalAddAny(&fd1, &fd2, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalAddAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	// sub NaN
	if err = DecimalSubAny(&fd1, &fd2, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalSubAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	// mul NaN
	if err = DecimalMulAny(&fd1, &fd2, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalMulAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	// div NaN
	if err = DecimalDivAny(&fd1, &fd2, &fd3, DivIncrFrac); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalDivAny(&fd2, &fd1, &fd3, DivIncrFrac); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	// mod NaN
	if err = DecimalModAny(&fd1, &fd2, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalModAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsNaN() {
		t.Fatal("failed")
	}

	_ = fd1.FromAsciiString("1", true)
	_ = fd2.FromAsciiString("1", true)
	if err = DecimalAddAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalSubAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalMulAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalDivAny(&fd2, &fd1, &fd3, DivIncrFrac); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsNaN() {
		t.Fatal("failed")
	}
	if err = DecimalModAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsNaN() {
		t.Fatal("failed")
	}
}

func TestDecimalArithInf(t *testing.T) {
	var fd1, fd2, fd3 FixedDecimal
	var err error
	_ = fd1.FromAsciiString("Inf", true)
	_ = fd2.FromAsciiString("1", true)
	// add Inf
	if err = DecimalAddAny(&fd1, &fd2, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalAddAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	// sub Inf
	if err = DecimalSubAny(&fd1, &fd2, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalSubAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	// mul Inf
	if err = DecimalMulAny(&fd1, &fd2, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalMulAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	// div Inf
	if err = DecimalDivAny(&fd1, &fd2, &fd3, DivIncrFrac); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalDivAny(&fd2, &fd1, &fd3, DivIncrFrac); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	// mod Inf
	if err = DecimalModAny(&fd1, &fd2, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalModAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if !fd3.IsInf() {
		t.Fatal("failed")
	}

	_ = fd1.FromAsciiString("1", true)
	_ = fd2.FromAsciiString("1", true)
	if err = DecimalAddAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalSubAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalMulAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalDivAny(&fd2, &fd1, &fd3, DivIncrFrac); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsInf() {
		t.Fatal("failed")
	}
	if err = DecimalModAny(&fd2, &fd1, &fd3); err != nil {
		t.Fatal("failed")
	}
	if fd3.IsInf() {
		t.Fatal("failed")
	}
}

func TestDecimalRound(t *testing.T) {
	type tcase struct {
		input1   string
		frac     int
		expected string
	}
	var fd1 FixedDecimal
	var fd2 FixedDecimal
	for _, c := range []tcase{
		{"0", 0, "0"},
		{"0", 1, "0.0"},
		{"1", -1, "0"},
		{"1", -2, "0"},
		{"1.00", 1, "1.0"},
		{"1.01", 1, "1.0"},
		{"1.02", 1, "1.0"},
		{"1.03", 1, "1.0"},
		{"1.04", 1, "1.0"},
		{"1.05", 1, "1.1"},
		{"1.06", 1, "1.1"},
		{"1.07", 1, "1.1"},
		{"1.08", 1, "1.1"},
		{"1.09", 1, "1.1"},
		{"1.050", 1, "1.1"},
		{"1.051", 1, "1.1"},
		{"1.052", 1, "1.1"},
		{"1.053", 1, "1.1"},
		{"1.054", 1, "1.1"},
		{"1.055", 1, "1.1"},
		{"1.056", 1, "1.1"},
		{"1.057", 1, "1.1"},
		{"1.058", 1, "1.1"},
		{"1.059", 1, "1.1"},
		{"1.040", 1, "1.0"},
		{"1.041", 1, "1.0"},
		{"1.042", 1, "1.0"},
		{"1.043", 1, "1.0"},
		{"1.044", 1, "1.0"},
		{"1.045", 1, "1.0"},
		{"1.046", 1, "1.0"},
		{"1.047", 1, "1.0"},
		{"1.048", 1, "1.0"},
		{"1.049", 1, "1.0"},
		{"1.0000000000", 9, "1.000000000"},
		{"1.0000000001", 9, "1.000000000"},
		{"1.0000000002", 9, "1.000000000"},
		{"1.0000000003", 9, "1.000000000"},
		{"1.0000000004", 9, "1.000000000"},
		{"1.0000000005", 9, "1.000000001"},
		{"1.0000000006", 9, "1.000000001"},
		{"1.0000000007", 9, "1.000000001"},
		{"1.0000000008", 9, "1.000000001"},
		{"1.0000000009", 9, "1.000000001"},
		{"1.0000000090", 9, "1.000000009"},
		{"1.0000000091", 9, "1.000000009"},
		{"1.0000000092", 9, "1.000000009"},
		{"1.0000000093", 9, "1.000000009"},
		{"1.0000000094", 9, "1.000000009"},
		{"1.0000000095", 9, "1.000000010"},
		{"1.0000000096", 9, "1.000000010"},
		{"1.0000000097", 9, "1.000000010"},
		{"1.0000000098", 9, "1.000000010"},
		{"1.0000000099", 9, "1.000000010"},
		{"999999999.0", 0, "999999999"},
		{"999999999.1", 0, "999999999"},
		{"999999999.2", 0, "999999999"},
		{"999999999.3", 0, "999999999"},
		{"999999999.4", 0, "999999999"},
		{"999999999.5", 0, "1000000000"},
		{"999999999.6", 0, "1000000000"},
		{"999999999.7", 0, "1000000000"},
		{"999999999.8", 0, "1000000000"},
		{"999999999.9", 0, "1000000000"},
		{"999999999.99990", 4, "999999999.9999"},
		{"999999999.99991", 4, "999999999.9999"},
		{"999999999.99992", 4, "999999999.9999"},
		{"999999999.99993", 4, "999999999.9999"},
		{"999999999.99994", 4, "999999999.9999"},
		{"999999999.99995", 4, "1000000000.0000"},
		{"999999999.99996", 4, "1000000000.0000"},
		{"999999999.99997", 4, "1000000000.0000"},
		{"999999999.99998", 4, "1000000000.0000"},
		{"999999999.99999", 4, "1000000000.0000"},
		{"999999999999999999.0", 0, "999999999999999999"},
		{"999999999999999999.1", 0, "999999999999999999"},
		{"999999999999999999.2", 0, "999999999999999999"},
		{"999999999999999999.3", 0, "999999999999999999"},
		{"999999999999999999.4", 0, "999999999999999999"},
		{"999999999999999999.5", 0, "1000000000000000000"},
		{"999999999999999999.6", 0, "1000000000000000000"},
		{"999999999999999999.7", 0, "1000000000000000000"},
		{"999999999999999999.8", 0, "1000000000000000000"},
		{"999999999999999999.9", 0, "1000000000000000000"},
		{"999999999999999999.90", 1, "999999999999999999.9"},
		{"999999999999999999.91", 1, "999999999999999999.9"},
		{"999999999999999999.92", 1, "999999999999999999.9"},
		{"999999999999999999.93", 1, "999999999999999999.9"},
		{"999999999999999999.94", 1, "999999999999999999.9"},
		{"999999999999999999.95", 1, "1000000000000000000.0"},
		{"999999999999999999.96", 1, "1000000000000000000.0"},
		{"999999999999999999.97", 1, "1000000000000000000.0"},
		{"999999999999999999.98", 1, "1000000000000000000.0"},
		{"999999999999999999.99", 1, "1000000000000000000.0"},
		{"0.9876543210", 10, "0.9876543210"},
		{"0.9876543210", 9, "0.987654321"},
		{"0.9876543210", 8, "0.98765432"},
		{"0.9876543210", 7, "0.9876543"},
		{"0.9876543210", 6, "0.987654"},
		{"0.9876543210", 5, "0.98765"},
		{"0.9876543210", 4, "0.9877"},
		{"0.9876543210", 3, "0.988"},
		{"0.9876543210", 2, "0.99"},
		{"0.9876543210", 1, "1.0"},
		{"0.9876543210", 0, "1"},
		{"123456789123456789", -1, "123456789123456790"},
		{"123456789123456789", -2, "123456789123456800"},
		{"123456789123456789", -3, "123456789123457000"},
		{"123456789123456789", -4, "123456789123460000"},
		{"123456789123456789", -5, "123456789123500000"},
		{"123456789123456789", -6, "123456789123000000"},
		{"123456789123456789", -7, "123456789120000000"},
		{"123456789123456789", -8, "123456789100000000"},
		{"123456789123456789", -9, "123456789000000000"},
		{"999999999999999999", -1, "1000000000000000000"},
		{"0.999999999", 9, "0.999999999"},
		{"0.999999999", 8, "1.00000000"},
		{"0.999999999", 7, "1.0000000"},
		{"0.999999999", 6, "1.000000"},
		{"0.999999999", 5, "1.00000"},
		{"0.999999999", 4, "1.0000"},
		{"0.999999999", 3, "1.000"},
		{"0.999999999", 2, "1.00"},
		{"0.999999999", 1, "1.0"},
		{"0.999999999", 0, "1"},
	} {
		if err := fd1.FromAsciiString(c.input1, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		fd1.RoundTo(&fd2, c.frac)
		actual := fd2.ToString(-1)
		fmt.Printf("(%v).Round(%v) = %v\n", fd1.ToString(-1), c.frac, actual)
		if actual != c.expected {
			t.Fatalf("result mismatch")
		}
		fd1.Round(c.frac)
		actual = fd1.ToString(-1)
		fmt.Printf("self.Round(%v) = %v\n", c.frac, actual)
		if actual != c.expected {
			t.Fatalf("result mismatch")
		}
	}
}

func TestDecimalFormat(t *testing.T) {
	type tscase struct {
		input    string
		frac     int
		expected string
	}

	var fd FixedDecimal
	for _, c := range []tscase{
		{"0", 0, "0"},
		{"1", 0, "1"},
		{"-1", 0, "-1"},
		{"1", 0, "1"},
		{"123", 0, "123"},
		{"123456789012345", 0, "123456789012345"},
		{"0.0", 0, "0"},
		{"0.0", 1, "0.0"},
		{"0.0", 2, "0.00"},
		{"0.0", 3, "0.000"},
		{"0.0", 4, "0.0000"},
		{"0.0", 5, "0.00000"},
		{"0.0", 6, "0.000000"},
		{"0.0", 7, "0.0000000"},
		{"0.0", 8, "0.00000000"},
		{"0.0", 9, "0.000000000"},
		{"0.0", 10, "0.0000000000"},
		{"0.0", 11, "0.00000000000"},
		{"0.0", 12, "0.000000000000"},
		{"0.123456789123456789", -1, "0.123456789123456789"},
		{"0.123456789123456789", 0, "0"},
		{"0.123456789123456789", 1, "0.1"},
		{"0.123456789123456789", 2, "0.12"},
		{"0.123456789123456789", 3, "0.123"},
		{"0.123456789123456789", 4, "0.1234"},
		{"0.123456789123456789", 5, "0.12345"},
		{"0.123456789123456789", 6, "0.123456"},
		{"0.123456789123456789", 7, "0.1234567"},
		{"0.123456789123456789", 8, "0.12345678"},
		{"0.123456789123456789", 9, "0.123456789"},
		{"0.123456789123456789", 10, "0.1234567891"},
		{"0.123456789123456789", 11, "0.12345678912"},
		{"0.123456789123456789", 12, "0.123456789123"},
		{"0.123456789123456789", 13, "0.1234567891234"},
		{"0.123456789123456789", 14, "0.12345678912345"},
		{"0.123456789123456789", 15, "0.123456789123456"},
		{"0.123456789123456789", 16, "0.1234567891234567"},
		{"0.123456789123456789", 17, "0.12345678912345678"},
		{"0.123456789123456789", 18, "0.123456789123456789"},
		{"123456789123456789", -1, "123456789123456789"},
		{"123456789123456789", 0, "123456789123456789"},
		{"123456789123456789", 1, "123456789123456789.0"},
		{"123456789123456789", 2, "123456789123456789.00"},
		{"123456789123456789", 3, "123456789123456789.000"},
		{"123456789123456789", 4, "123456789123456789.0000"},
		{"123456789123456789", 5, "123456789123456789.00000"},
		{"123456789123456789", 6, "123456789123456789.000000"},
		{"123456789123456789", 7, "123456789123456789.0000000"},
		{"123456789123456789", 8, "123456789123456789.00000000"},
		{"123456789123456789", 9, "123456789123456789.000000000"},
		{"123456789123456789", 10, "123456789123456789.0000000000"},
		{"123456789123456789", 11, "123456789123456789.00000000000"},
		{"123456789123456789", 12, "123456789123456789.000000000000"},
		{"123456789123456789", 13, "123456789123456789.0000000000000"},
		{"123456789123456789", 14, "123456789123456789.00000000000000"},
		{"123456789123456789", 15, "123456789123456789.000000000000000"},
		{"123456789123456789", 16, "123456789123456789.0000000000000000"},
		{"123456789123456789", 17, "123456789123456789.00000000000000000"},
		{"123456789123456789", 18, "123456789123456789.000000000000000000"},
		{"123456789123456789", 19, "123456789123456789.0000000000000000000"},
	} {
		if err := fd.FromAsciiString(c.input, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		actual := fd.ToString(c.frac)
		if actual != c.expected {
			t.Fatalf("result mismatch: actual=%v, expected=%v", actual, c.expected)
		}
	}
}

func TestDecimalNormal(t *testing.T) {
	var dec FixedDecimal
	if err := dec.FromAsciiString("-100", false); err != nil {
		t.Fatal("failed")
	}
	dec.setNormal()
	if !dec.IsNeg() {
		t.Fatal("failed")
	}
	dec.SetOne()
	if dec.IsSpecial() {
		t.Fatal("failed")
	}
	if dec.ToString(-1) != "1" {
		t.Fatal("failed")
	}
	dec.FromInt64(100, true)
	if dec.ToString(-1) != "100" {
		t.Fatal("failed")
	}
}

func TestDecimalMod(t *testing.T) {
	type tcase struct {
		input1, input2, expected string
	}
	var fd1 FixedDecimal
	var fd2 FixedDecimal
	var fd3 FixedDecimal
	for _, c := range []tcase{
		{"0", "1", "0"},
		{"1", "1", "0"},
		{"1", "2", "1"},
		{"2", "1", "0"},
		{"1000000001", "2", "1"},
		{"-1000000001", "2", "-1"},
		{"-1", "2", "-1"},
		{"-1", "-2", "-1"},
		{"1", "-2", "1"},
		{"-1", "-100", "-1"},
		{"100", "3", "1"},
		{"100", "1001", "100"},
		{"0.2", "1", "0.2"},
		{"0.02", "1", "0.02"},
		{"0.002", "1", "0.002"},
		{"0.0002", "1", "0.0002"},
		{"0.00002", "1", "0.00002"},
		{"0.000002", "1", "0.000002"},
		{"0.0000002", "1", "0.0000002"},
		{"0.00000002", "1", "0.00000002"},
		{"0.000000002", "1", "0.000000002"},
		{"0.2", "1.0", "0.2"},
		{"0.2", "1.00", "0.20"},
		{"0.2", "1.000", "0.200"},
		{"0.2", "1.0000", "0.2000"},
		{"0.2", "1.00000", "0.20000"},
		{"0.2", "1.000000", "0.200000"},
		{"0.2", "1.0000000", "0.2000000"},
		{"0.2", "1.00000000", "0.20000000"},
		{"0.2", "1.000000000", "0.200000000"},
		{"-0.2", "1.0", "-0.2"},
		{"-0.2", "1.00", "-0.20"},
		{"-0.2", "1.000", "-0.200"},
		{"-0.2", "1.0000", "-0.2000"},
		{"-0.2", "1.00000", "-0.20000"},
		{"-0.2", "1.000000", "-0.200000"},
		{"-0.2", "1.0000000", "-0.2000000"},
		{"-0.2", "1.00000000", "-0.20000000"},
		{"-0.2", "1.000000000", "-0.200000000"},
		{"-0.3", "-0.2", "-0.1"},
		{"0.3", "0.2", "0.1"},
		{"0.3", "-0.2", "0.1"},
		{"-0.3", "0.2", "-0.1"},
		{"-0.3", "-0.7", "-0.3"},
		{"0.3", "-0.7", "0.3"},
		{"-0.3", "0.7", "-0.3"},
		{"0.3", "0.7", "0.3"},
		{"1000000000.1", "7", "6.1"},
		{"1000000000.1", "9", "1.1"},
		{"1000000000.1", "9.00", "1.10"},
		{"100.10000000001", "7", "2.10000000001"},
		{"101000000000.1", "7.1", "4.9"},
		{"101000000000.1", "5", "0.1"},
		{"101000000000.1", "5.291", "0.201"},
		{"100.1", "7.0000000001", "2.0999999986"},
		{"205.6", "9.5000000001", "6.0999999979"},
		{"2000000005.1", "7.5000000001", "2.5733333333"},
		{"1.2", "0.7", "0.5"},
		{"1.22", "0.77", "0.45"},
		{"1.222", "0.777", "0.445"},
		{"1.2222", "0.7777", "0.4445"},
		{"1.22222", "0.77777", "0.44445"},
		{"1.222222", "0.777777", "0.444445"},
		{"1.2222222", "0.7777777", "0.4444445"},
		{"1.22222222", "0.77777777", "0.44444445"},
		{"1.222222222", "0.777777777", "0.444444445"},
		{"9.8", "1", "0.8"},
		{"98.7", "1.2", "0.3"},
		{"987.6", "1.23", "1.14"},
		{"9876.5", "1.234", "0.798"},
		{"98765.4", "1.2345", "0.4620"},
		{"987654.3", "1.23456", "0.12720"},
		{"9876543.2", "1.234567", "1.027165"},
		{"98765432.1", "1.2345678", "0.6925932"},
		{"987654321.1", "1.23456789", "0.45802477"},
		{"987654321.12", "1.234567899", "0.685432101"},
		{"987654321.123", "1.2345678998", "0.0484321002"},
		{"987654321.1234", "1.23456789987", "1.22740000000"},
		{"987654321.12345", "1.234567899876", "1.222650000000"},
		{"987654321.123456", "1.2345678998765", "1.2222560000000"},
		{"987654321.1234567", "1.23456789987654", "1.22222470000000"},
		{"987654321.12345678", "1.234567899876543", "1.222222380000000"},
		{"987654321.123456789", "1.2345678998765432", "1.2222222290000000"},
		{"-9.8", "1", "-0.8"},
		{"-98.7", "1.2", "-0.3"},
		{"-987.6", "1.23", "-1.14"},
		{"-9876.5", "1.234", "-0.798"},
		{"-98765.4", "1.2345", "-0.4620"},
		{"-987654.3", "1.23456", "-0.12720"},
		{"-9876543.2", "1.234567", "-1.027165"},
		{"-98765432.1", "1.2345678", "-0.6925932"},
		{"-987654321.1", "1.23456789", "-0.45802477"},
		{"-987654321.12", "1.234567899", "-0.685432101"},
		{"-987654321.123", "1.2345678998", "-0.0484321002"},
		{"-987654321.1234", "1.23456789987", "-1.22740000000"},
		{"-987654321.12345", "1.234567899876", "-1.222650000000"},
		{"-987654321.123456", "1.2345678998765", "-1.2222560000000"},
		{"-987654321.1234567", "1.23456789987654", "-1.22222470000000"},
		{"-987654321.12345678", "1.234567899876543", "-1.222222380000000"},
		{"-987654321.123456789", "1.2345678998765432", "-1.2222222290000000"},
		{"0.170511", "-353390023.459963", "0.170511"},
		{"-353390023.459963", "0.170511", "-0.060946"},
		{"0.170511", "-353390023", "0.170511"},
		{"-353390023", "0.170511", "-0.112516"},
		{"0.4", "0.000000003", "0.000000001"},
		{"0.4", "0.0000000003", "0.0000000001"},
		{"0.2", "0.000000003", "0.000000002"},
		{"0.2", "0.0000000003", "0.0000000002"},
		{"1000000000000000001", "70298007", "68215565"},
		{"1000000000000000001", "0.70298007", "0.07924142"},
		{"1000000000000000001", "500000000.1", "300000001.1"},
		{"0.1", "0.20000000001", "0.10000000000"},
	} {
		if err := fd1.FromAsciiString(c.input1, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := fd2.FromAsciiString(c.input2, true); err != nil {
			t.Fatalf("failed %v", err)
		}
		if err := DecimalMod(&fd1, &fd2, &fd3); err != nil {
			t.Fatalf("failed %v", err)
		}
		actual := fd3.ToString(-1)
		fmt.Printf("fd1=%v, fd2=%v\n", fd1, fd2)
		fmt.Printf("%v%%%v=%v\n", fd1.ToString(-1), fd2.ToString(-1), actual)
		if actual != c.expected {
			t.Fatalf("result mismatch")
		}
	}
}
