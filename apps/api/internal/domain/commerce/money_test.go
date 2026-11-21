package commerce

import "testing"

func TestExactTomanArithmetic(t *testing.T) {
	for input, want := range map[string]Amount{"0": 0, "15000000": 150000000, "1.1": 11, "900000000000000": MaxAmount} {
		got, e := ParseToman(input)
		if e != nil || got != want || Toman(got) != input {
			t.Fatal(input, got, e)
		}
	}
	for _, s := range []string{"", "-1", "1e9", "1.12", "NaN", "900000000000000.1", "1,000", " 1", "+1"} {
		if _, e := ParseToman(s); e == nil {
			t.Fatal("invalid amount accepted", s)
		}
	}
	if _, e := Add(MaxAmount, 1); e == nil {
		t.Fatal("overflow")
	}
	if _, e := Multiply(MaxAmount, 2); e == nil {
		t.Fatal("multiply overflow")
	}
}
