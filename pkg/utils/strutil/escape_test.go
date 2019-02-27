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

func TestEscapeUnescapeSimple(t *testing.T) {
	escaper := NewStringEscaper("%",
		"/", "\\", "`", "*", "|", ";", "\"", "'", ":")

	exp := ""
	act := escaper.Unescape(escaper.Escape(exp))
	assert.Equal(t, exp, act)

	exp = "%"
	act = escaper.Unescape(escaper.Escape(exp))
	assert.Equal(t, exp, act)

	exp = "%/"
	act = escaper.Unescape(escaper.Escape(exp))
	assert.Equal(t, exp, act)

	exp = "this% /is//simple\\`test``*|to tr%y 'escape;'''::/unescape %1%22%12%%"
	act = escaper.Unescape(escaper.Escape(exp))
	assert.Equal(t, exp, act)
}

func TestEscapeUnescapeMultiple(t *testing.T) {
	escaper := NewStringEscaper("^&*",
		"a*bc", "&oioi", "^", "&", "*")

	exp := ""
	act := escaper.Unescape(escaper.Unescape(escaper.Escape(escaper.Escape(exp))))
	assert.Equal(t, exp, act)

	exp = "a*bc"
	act = escaper.Unescape(escaper.Unescape(escaper.Escape(escaper.Escape(exp))))
	assert.Equal(t, exp, act)

	exp = "a*bc&oioi"
	act = escaper.Unescape(escaper.Unescape(escaper.Escape(escaper.Escape(exp))))
	assert.Equal(t, exp, act)

	exp = "this ^&*^^&* is multiple ^&*1^&*11 test &oi&oioioi ** to try escape a^&**bc ^&&* unescape ^&*^&*"
	act = escaper.Unescape(escaper.Unescape(escaper.Escape(escaper.Escape(exp))))
	assert.Equal(t, exp, act)
}
