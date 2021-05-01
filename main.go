package main

import "fmt"

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

func compare(c1, c2 Curve, comparator func(x1, x2 int, c1, c2 Curve) float32, summarizor func(diffs map[int]float32) float32) (float32, int) {
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
		_, ok1 := c1.y(x1)
		_, ok2 := c2.y(x2)

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

		diffs[x1] = comparator(x1, x2, c1, c2)
	}

	return summarizor(diffs), gapnum
}

func test(curve Curve) {
	fmt.Println(curve.y(1))
}

func main() {
	myStock := Stock{
		Symbol: "My Stock",
		CurveBase: CurveBase{
			Values: map[int]float32{
				1: 1,
				2: 2,
			},
		},
	}

	test(myStock)
}

func getAllXs(inMap map[int]float32) []int {
	keys := make([]int, 0, len(inMap))
	for k := range inMap {
		keys = append(keys, k)
	}

	return keys
}
