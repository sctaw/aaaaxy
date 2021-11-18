// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package math

import (
	"fmt"
	"math"
)

type fixedUnderlying = int64
type Fixed fixedUnderlying

const (
	fixedBits = 12
	FixedOne Fixed = 1<<fixedBits
)

func NewFixed(i int) Fixed {
	return Fixed(i) * FixedOne
}

func NewFixedInt64(i fixedUnderlying) Fixed {
	return Fixed(i) * FixedOne
}

func NewFixedFloat64(f float64) Fixed {
	return Fixed(math.RoundToEven(f * float64(FixedOne)))
}

func (f Fixed) Mul(g Fixed) Fixed {
	return g.MulFrac(g, FixedOne)
}

func (f Fixed) MulFrac(n, d Fixed) Fixed {
	return Fixed(MulFracInt64(fixedUnderlying(f), fixedUnderlying(n), fixedUnderlying(d)))
}

func (f Fixed) Div(g Fixed) Fixed {
	return g.MulFrac(FixedOne, g)
}

func (f Fixed) Rint() int {
	return int((f + FixedOne / 2) >> fixedBits)
}

func (f Fixed) Float64() float64 {
	return float64(f) * (1.0 / float64(FixedOne))
}

func (f Fixed) String() string {
	return fmt.Sprintf("%d.0x%03x", fixedUnderlying(f>>fixedBits), fixedUnderlying(f&(FixedOne-1)))
}

func (f Fixed) Sqrt() Fixed {
	if f == 0 {
		return 0
	}

	// Compute a wild guess using the FPU.
	guess := NewFixedFloat64(math.Sqrt(f.Float64()))

	// Want unique number s so that, where delta=0.5:
	//   s-delta <= 4096*sqrt(f/4096) < s+delta
	// Square everything; assumes s-delta >= 0. Thus the check above.
	//   s^2 - s <= 4096 * f - 1/4 < s^2 + s
	//   s^2 - s < 4096 * f <= s^2 + s

	// In practice these loops tend to execute only once.

	goal := f << fixedBits
	// fixes := 0
	s := guess
	for s*s-s >= goal {
		s--
		// fixes++
	}
	for s*s+s < goal {
		s++
		// fixes++
	}
	// if fixes > 16 {
	// log.Fatalf("too many fixes for Sqrt(%v): %v fixes, guess was %v, result is %v", f, fixes, guess, s)
	// }

	return s
}
