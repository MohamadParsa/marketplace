package commerce

import "testing"

func TestExactPercentageAndStableAllocation(t *testing.T) {
	for _, v := range []struct {
		n    Amount
		bps  int
		want Amount
	}{{10001, 5000, 5001}, {MaxAmount, 10000, MaxAmount}, {10400, 1000, 1040}, {1, 5000, 1}} {
		got, e := Percent(v.n, v.bps)
		if e != nil || got != v.want {
			t.Fatalf("percent %v %v", got, e)
		}
	}
	// Latest owner example: cash1000 +4% markup40; 10% coupon104 =>936 toman.
	cash := Amount(10000)
	markup, _ := Percent(cash, 400)
	coupon, _ := Percent(cash+markup, 1000)
	if Toman(cash+markup-coupon) != "936" {
		t.Fatal("coupon base")
	}
	a, e := Allocate(2, []WeightedAmount{{"c", 1}, {"a", 1}, {"b", 1}})
	if e != nil || a["a"] != 1 || a["b"] != 1 || a["c"] != 0 {
		t.Fatal(a, e)
	}
	b, e := Allocate(MaxAmount-1, []WeightedAmount{{"b", MaxAmount / 2}, {"a", MaxAmount / 2}})
	if e != nil || b["a"]+b["b"] != MaxAmount-1 {
		t.Fatal(b, e)
	}
	if _, e = Allocate(1, []WeightedAmount{{"a", 1}, {"a", 1}}); e == nil {
		t.Fatal("duplicate")
	}
	if _, e = Percent(MaxAmount, 10001); e == nil {
		t.Fatal("invalid percentage")
	}
}
