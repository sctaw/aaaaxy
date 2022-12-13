// Copyright 2022 Google LLC
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

package locale

import (
	"sort"
)

// Lingua identifies a language.
type Lingua string

// Linguas is the set of current languages that are translated "enough".
var Linguas = map[Lingua]struct{}{}

// Name returns the human readable name of a language.
func (l Lingua) Name() string {
	switch l {
	case "":
		return "English"
	case "de":
		return "Deutsch"
	default:
		return string(l)
	}
}

// Linguas returns the languages sorted by humanly expected ordering.
func LinguasSorted() []Lingua {
	ret := []Lingua{""}
	for l := range Linguas {
		ret = append(ret, l)
	}
	sort.Slice(ret, func(a, b int) bool {
		return ret[a].Name() < ret[b].Name()
	})
	return ret
}