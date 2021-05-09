package main

import (
	"fmt"
	"sort"
)

// Curve :
type Curve interface {
	y(x int) (float32, bool)
	allXs() []int
}

// CurveBase :
type CurveBase struct {
	Values map[int]float32
}

// Stock :
type Stock struct {
	Symbol string
	CurveBase
}

func (curve CurveBase) y(x int) (float32, bool) {
	if value, ok := curve.Values[x]; ok {
		return value, true
	} else {
		return 0, false
	}
}

func (curve CurveBase) allXs() []int {
	return getAllXs(curve.Values)
}

func compare(c1, c2 Curve, comparator func(x1, x2 int, y1, y2 float32) float32, summarizor func(diffs map[int]float32) float32) (float32, int) {
	x1s := c1.allXs()
	x2s := c2.allXs()

	l1, l2 := len(x1s), len(x2s)
	if l1 == 0 || l2 == 0 {
		return 0, l1 + l2
	}

	gapnum := 0
	diffs := make(map[int]float32)

	x1last := x1s[l1-1]
	x2last := x2s[l2-1]
	for x1, x2 := 0, 0; x1 <= x1last && x2 <= x2last; {
		y1, ok1 := c1.y(x1)
		y2, ok2 := c2.y(x2)

		if !ok1 || !ok2 {
			if ok1 || ok2 {
				gapnum++
			}
			if !ok1 {
				x1++
			}
			if !ok2 {
				x2++
			}
			continue
		} else {
			if x1 < x2 {
				gapnum++
				x1++
				continue
			} else if x1 > x2 {
				gapnum++
				x2++
				continue
			}
		}

		diffs[x1] = comparator(x1, x2, y1, y2)
		x1++
		x2++
	}

	return summarizor(diffs), gapnum
}

func slope(in Curve) Curve {
	values := make(map[int]float32)
	xs := in.allXs()

	l := len(xs) - 1

	if l <= 1 {
		return nil
	}

	x1 := xs[0]
	y1, _ := in.y(x1)

	for i := 1; i < l; i++ {
		x2 := xs[i]
		y2, _ := in.y(x2)

		values[x1] = (y2 - y1) / float32(x2-x1)
		x1 = x2
		y1 = y2
	}

	return CurveBase{
		Values: values,
	}
}

func minusComparator(x1, x2 int, y1, y2 float32) float32 {
	return y1 - y2
}

func summarySummarizor(diffs map[int]float32) float32 {
	summary := float32(0)

	for _, diff := range diffs {
		summary += diff
	}

	return summary
}

func compareSlope(c1, c2 Curve) (float32, int) {
	return compare(slope(c1), slope(c2), minusComparator, summarySummarizor)
}

func test(curve Curve) {
	fmt.Println(curve.y(1))
}

func main() {
	market := LoadAllMarketData()
	stock1 := market.GetSmblCurve("GOLD", "Close")
	stock2 := market.GetSmblCurve("AG", "Close")

	diff, gapnum := compareSlope(stock1, stock2)

	fmt.Printf("%f, %d", diff, gapnum)
	/*myStock := Stock{
		Symbol: "My Stock",
		CurveBase: CurveBase{
			Values: map[int]float32{
				1: 1,
				2: 2,
			},
		},
	}*/

	/*if len(stock1.Values) > 0 {
		test(stock1)
	}*/
}

func getAllXs(inMap map[int]float32) []int {
	keys := make([]int, 0, len(inMap))
	for k := range inMap {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	return keys
}
