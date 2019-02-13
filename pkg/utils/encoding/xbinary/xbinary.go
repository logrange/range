// Copyright 2018 The logrange Authors
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

package xbinary

import (
	"encoding/binary"
	"fmt"
	"github.com/logrange/range/pkg/utils/bytes"
	"io"
)

// MarshalByte writes value v to the buf. Returns number of bytes written or an error, if any
func MarshalByte(v byte, buf []byte) (int, error) {
	if len(buf) < 1 {
		return 0, noBufErr("MarshalByte", len(buf), 1)
	}
	buf[0] = v
	return 1, nil
}

// WriteByte writes value v to writer w. It returns number of bytes written or an error if any.
func WriteByte(v byte, w io.Writer) (int, error) {
	var b [1]byte
	b[0] = v
	return w.Write(b[:])
}

// UnmarshalByte reads next byte value from the buf. Retruns number of bytes read, the value or an error, if any.
func UnmarshalByte(buf []byte) (int, byte, error) {
	if len(buf) == 0 {
		return 0, 0, noBufErr("UnmarshalByte", len(buf), 1)
	}
	return 1, buf[0], nil
}

// MarshalUint16 writes value v to the buf. Returns number of bytes written or an error, if any
func MarshalUint16(v uint16, buf []byte) (int, error) {
	if len(buf) < 2 {
		return 0, noBufErr("MarshalUint16", len(buf), 2)
	}
	binary.BigEndian.PutUint16(buf, v)
	return 2, nil
}

// WriteUint16 writes value v to writer w. It returns number of bytes written or an error if any.
func WriteUint16(v uint16, w io.Writer) (int, error) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	return w.Write(b[:])
}

// UnmarshalUint16 reads next uint16 value from the buf. Retruns number of bytes read, the value or an error, if any.
func UnmarshalUint16(buf []byte) (int, uint16, error) {
	if len(buf) < 2 {
		return 0, 0, noBufErr("UnmarshalUint16", len(buf), 2)
	}
	return 2, binary.BigEndian.Uint16(buf), nil
}

// MarshalUint32 writes value v to the buf. Returns number of bytes written or an error, if any
func MarshalUint32(v uint32, buf []byte) (int, error) {
	if len(buf) < 4 {
		return 0, noBufErr("MarshalUint32", len(buf), 4)
	}
	binary.BigEndian.PutUint32(buf, v)
	return 4, nil
}

// WriteUint32 writes value v to writer w. It returns number of bytes written or an error if any.
func WriteUint32(v uint32, w io.Writer) (int, error) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	return w.Write(b[:])
}

// UnmarshalUint32 reads next uint32 value from the buf. Retruns number of bytes read, the value or an error, if any.
func UnmarshalUint32(buf []byte) (int, uint32, error) {
	if len(buf) < 4 {
		return 0, 0, noBufErr("UnmarshalUint32", len(buf), 4)
	}
	return 4, binary.BigEndian.Uint32(buf), nil
}

// MarshalInt64 writes value v to the buf. Returns number of bytes written or an error, if any
func MarshalInt64(v int64, buf []byte) (int, error) {
	if len(buf) < 8 {
		return 0, noBufErr("MarshalInt64", len(buf), 8)
	}
	binary.BigEndian.PutUint64(buf, uint64(v))
	return 8, nil
}

// WriteInt64 writes value v to writer w. It returns number of bytes written or an error if any.
func WriteInt64(v int64, w io.Writer) (int, error) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(v))
	return w.Write(b[:])
}

// UnmarshalInt64 reads next int64 value from the buf. Retruns number of bytes read, the value or an error, if any.
func UnmarshalInt64(buf []byte) (int, int64, error) {
	if len(buf) < 8 {
		return 0, 0, noBufErr("UnmarshalInt64", len(buf), 8)
	}
	return 8, int64(binary.BigEndian.Uint64(buf)), nil
}

// MarshalUInt writes value v to the buf. Returns number of bytes written or an error, if any
func MarshalUInt(v uint, buf []byte) (int, error) {
	idx := 0
	for {
		if idx == len(buf) {
			return 0, noBufErr("MarshalUInt", len(buf), 8)
		}

		if v > 127 {
			buf[idx] = 128 | (byte(v & 127))
		} else {
			buf[idx] = byte(v)
			return idx + 1, nil
		}
		v = v >> 7
		idx++
	}
}

// WriteUInt writes value v to writer w. It returns number of bytes written or an error if any.
func WriteUInt(v uint, w io.Writer) (int, error) {
	var b [10]byte
	sz, err := MarshalUInt(v, b[:])
	if err != nil {
		return 0, err
	}
	return w.Write(b[:sz])
}

// UnmarshalUInt reads next uint value from the buf. It retruns number of bytes read, the value or an error, if any.
func UnmarshalUInt(buf []byte) (int, uint, error) {
	res := uint(0)
	idx := 0
	shft := uint(0)
	for {
		if idx == len(buf) {
			return 0, 0, noBufErr("UnmarshalUInt", len(buf), 8)
		}

		b := buf[idx]
		res = res | (uint(b&127) << shft)
		if b <= 127 {
			return idx + 1, res, nil
		}
		shft += 7
		idx++
	}
}

const (
	bit7  = 1 << 7
	bit14 = 1 << 14
	bit21 = 1 << 21
	bit28 = 1 << 28
	bit35 = 1 << 35
	bit42 = 1 << 42
	bit49 = 1 << 49
	bit56 = 1 << 56
	bit63 = 1 << 63
)

// SizeUint returns size of encoded uint64 size
func SizeUint(v uint64) int {
	if v >= bit35 {
		if v >= bit49 {
			if v >= bit63 {
				return 10
			}
			if v >= bit56 {
				return 9
			}
			return 8
		}
		if v >= bit42 {
			return 7
		}
		return 6
	}

	if v >= bit21 {
		if v >= bit28 {
			return 5
		}
		return 4
	}

	if v >= bit14 {
		return 3
	}

	if v >= bit7 {
		return 2
	}
	return 1
}

// MarshalBytes writes value v to the buf. Returns number of bytes written or an error, if any
func MarshalBytes(v []byte, buf []byte) (int, error) {
	ln := len(v)

	idx, err := MarshalUInt(uint(ln), buf)
	if err != nil {
		return 0, err
	}

	buf = buf[idx:]
	if len(buf) < ln {
		return 0, noBufErr("MarshalBytes-size-body", ln, len(buf))
	}

	copy(buf[:ln], v)
	return ln + idx, nil
}

// WriteBytes writes value v to writer w. It returns number of bytes written or an error if any.
func WriteBytes(v []byte, w io.Writer) (int, error) {
	n, err := WriteUInt(uint(len(v)), w)
	if err != nil {
		return n, err
	}

	n2, err := w.Write(v)
	return n2 + n, err
}

// UnmarshalBytes reads next []byte value from the buf. It retruns number of bytes read, the value or an error, if any.
func UnmarshalBytes(buf []byte, newBuf bool) (int, []byte, error) {
	idx, uln, err := UnmarshalUInt(buf)
	if err != nil {
		return 0, nil, err
	}

	ln := int(uln)
	if len(buf) < ln+idx {
		return 0, nil, noBufErr("UnmarshalBytes-size-body", ln, len(buf))
	}

	res := buf[idx : idx+ln]
	if newBuf {
		res = bytes.BytesCopy(res)
	}
	return idx + ln, res, nil
}

// SizeBytes returns size of encoded buf
func SizeBytes(buf []byte) int {
	return SizeUint(uint64(len(buf))) + len(buf)
}

// MarshalString writes value v to the buf. Returns number of bytes written or an error, if any
func MarshalString(v string, buf []byte) (int, error) {
	return MarshalBytes(bytes.StringToByteArray(v), buf)
}

// WriteString writes value v to writer w. It returns number of bytes written or an error if any.
func WriteString(v string, w io.Writer) (int, error) {
	return WriteBytes(bytes.StringToByteArray(v), w)
}

// UnmarshalString reads next string value from the buf. It retruns number of bytes read, the value or an error, if any.
func UnmarshalString(buf []byte, newBuf bool) (int, string, error) {
	idx, res, err := UnmarshalBytes(buf, newBuf)
	if err == nil {
		return idx, bytes.ByteArrayToString(res), err
	}
	return idx, "", err
}

// SizeString returns size of encoded v
func SizeString(v string) int {
	return SizeBytes(bytes.StringToByteArray(v))
}

func noBufErr(src string, ln, req int) error {
	return fmt.Errorf("not enough space in the buf: %s requres %d bytes, but actual buf size is %d", src, req, ln)
}
