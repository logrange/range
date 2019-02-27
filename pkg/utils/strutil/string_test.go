// Copyright 2018-2019 The logrange Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveDups(t *testing.T) {
	e := []string{"a", "b", "c", "d"}
	a := RemoveDups([]string{"a", "a", "b", "c", "b", "c", "d"})

	assert.NotNil(t, a)
	assert.ElementsMatch(t, a, e)
}

func TestSwapEvenOdd(t *testing.T) {
	assert.Equal(t, []string{}, SwapEvenOdd([]string{}))
	assert.Equal(t, []string{"a"}, SwapEvenOdd([]string{"a"}))
	assert.Equal(t, []string{"a", "b"}, SwapEvenOdd([]string{"b", "a"}))
	assert.Equal(t, []string{"a", "b", "c"}, SwapEvenOdd([]string{"b", "a", "c"}))
	assert.Equal(t, []string{"a", "b", "c", "d"}, SwapEvenOdd([]string{"b", "a", "d", "c"}))
}
