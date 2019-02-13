package xbinary

import (
	"testing"
)

func BenchmarkMarshalUint(b *testing.B) {
	var bb [30]byte
	buf := bb[:]
	ui := uint(1347598723405981734)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MarshalUInt(ui, buf)
	}
}

func BenchmarkMarshalInt64(b *testing.B) {
	var bb [30]byte
	buf := bb[:]
	ui := int64(1347598723405981734)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MarshalInt64(ui, buf)
	}
}

func BenchmarkMarshalBytes(b *testing.B) {
	var bb [30]byte
	buf := bb[:]
	bbb := []byte("test string")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MarshalBytes(bbb, buf)
	}
}

func BenchmarkUnmarshalBytes(b *testing.B) {
	var bb [30]byte
	buf := bb[:]
	bbb := []byte("test string")
	MarshalBytes(bbb, buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalBytes(buf, true)
	}
}

func BenchmarkUnmarshalBytesNoCopy(b *testing.B) {
	var bb [30]byte
	buf := bb[:]
	bbb := []byte("test string")
	MarshalBytes(bbb, buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalBytes(buf, false)
	}
}

func TestMarshalUInt(t *testing.T) {
	var b [3]byte
	buf := b[:]
	idx, err := MarshalUInt(0, buf)
	if idx != 1 || buf[0] != 0 || err != nil {
		t.Fatal("Unexpected idx=", idx, " buf=", buf, ", err=", err)
	}

	idx, err = MarshalUInt(25, buf)
	if idx != 1 || buf[0] != 25 || err != nil {
		t.Fatal("Unexpected idx=", idx, " buf=", buf, ", err=", err)
	}

	idx, v, err := UnmarshalUInt(buf)
	if v != 25 || idx != 1 || err != nil {
		t.Fatal("Unexpected idx=", idx, " v=", v, ", err=", err)
	}

	idx, err = MarshalUInt(129, buf)
	if idx != 2 || buf[0] != 129 || buf[1] != 1 || err != nil {
		t.Fatal("Unexpected idx=", idx, " buf=", buf, ", err=", err)
	}

	idx, err = MarshalUInt(32565, buf)
	if idx != 3 || err != nil {
		t.Fatal("Unexpected idx=", idx, " buf=", buf, ", err=", err)
	}

	idx, v, err = UnmarshalUInt(buf)
	if v != 32565 || idx != 3 || err != nil {
		t.Fatal("Unexpected idx=", idx, " v=", v, ", err=", err)
	}

	idx, err = MarshalUInt(3256512341234, buf)
	if err == nil {
		t.Fatal("Expecting err != nil, but idx=", idx, " buf=", buf)
	}
}

func TestMarshalBytes(t *testing.T) {
	const str = "abcasdfadfasd"
	bstr := []byte(str)
	var b [20]byte
	buf := b[:]

	idx, err := MarshalBytes(bstr, buf)
	if idx != 1+len(str) || err != nil {
		t.Fatal("idx=", idx, ", err=", err)
	}

	idx2, bts, err := UnmarshalBytes(buf, false)
	if idx != idx2 || string(bts) != str || err != nil {
		t.Fatal("idx2=", idx2, " bts=", string(bts), ", err=", err)
	}

	idx, err = MarshalBytes(bstr[:0], buf)
	if idx != 1 {
		t.Fatal("empty bytes should be 1 byte in result length")
	}
}

func TestMarshalByte(t *testing.T) {
	testMarshalInts(t, 254, 1, func(v int, b []byte) (int, error) {
		return MarshalByte(byte(v), b)
	}, func(b []byte) (int, int, error) {
		i, v, err := UnmarshalByte(b)
		return i, int(v), err
	})
}

func TestMarshalUint16(t *testing.T) {
	testMarshalInts(t, 254, 2, func(v int, b []byte) (int, error) {
		return MarshalUint16(uint16(v), b)
	}, func(b []byte) (int, int, error) {
		i, v, err := UnmarshalUint16(b)
		return i, int(v), err
	})
}

func TestMarshalUint32(t *testing.T) {
	testMarshalInts(t, 223454, 4, func(v int, b []byte) (int, error) {
		return MarshalUint32(uint32(v), b)
	}, func(b []byte) (int, int, error) {
		i, v, err := UnmarshalUint32(b)
		return i, int(v), err
	})
}

func TestMarshalInt64(t *testing.T) {
	testMarshalInts(t, 2542341, 8, func(v int, b []byte) (int, error) {
		return MarshalInt64(int64(v), b)
	}, func(b []byte) (int, int, error) {
		i, v, err := UnmarshalInt64(b)
		return i, int(v), err
	})
}

func TestSizeUint(t *testing.T) {
	testSizeUint(t, bit7, 2)
	testSizeUint(t, bit14, 3)
	testSizeUint(t, bit21, 4)
	testSizeUint(t, bit28, 5)
	testSizeUint(t, bit35, 6)
	testSizeUint(t, bit42, 7)
	testSizeUint(t, bit49, 8)
	testSizeUint(t, bit56, 9)
	testSizeUint(t, bit63, 10)
}

func testSizeUint(t *testing.T, val uint64, sz int) {
	if SizeUint(val) != sz || SizeUint(val-1) != sz-1 {
		t.Fatal("for ", val, " expecting size ", sz, ", but it is ", SizeUint(val))
	}

	var b [20]byte
	buf := b[:]
	idx, _ := MarshalUInt(uint(val), buf)
	if idx != sz {
		t.Fatal("MarshalUInt returns another value for val=", val, " idx=", idx, ", but sz=", sz)
	}

	idx, _ = MarshalUInt(uint(val-1), buf)
	if idx != sz-1 {
		t.Fatal("MarshalUInt returns another value for val=", val, " idx=", idx, ", but sz=", sz-1)
	}
}

func testMarshalInts(t *testing.T, v, elen int, mf func(v int, b []byte) (int, error), uf func(b []byte) (int, int, error)) {
	var b [20]byte
	buf := b[:]

	idx, err := mf(v, buf)
	if idx != elen || err != nil {
		t.Fatal("expected len=", elen, ", but idx=", idx, ", buf=", buf, ", err=", err)
	}

	idx2, val, err := uf(buf)
	if idx != idx2 || v != val || err != nil {
		t.Fatal("unmarshal idx=", idx2, " val=", val, ", but expected=", v, ", err=", err)
	}
}
