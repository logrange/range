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
package crypto

import (
	"github.com/logrange/range/pkg/utils/strutil"
)

const (
	secretKeyAlphabet = "0123456789QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnazx_^-()@#$%"
	sessionAlphabet   = "0123456789QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnazx"
)

// Makes a session with size characters in the the string. sessionAlphabet
// is used for making the session
func NewSessionId(size int) string {
	return strutil.GetRandomString(size, sessionAlphabet)
}

// Makes a password string
func NewPassword(size int) string {
	return strutil.GetRandomString(size, secretKeyAlphabet)
}
