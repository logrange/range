package bytes

import (
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"
)

func BenchmarkPoolArrange(b *testing.B) {
	p := &Pool{}

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		p.Release(p.Arrange(rand.Intn(1200000)))
	}
}

func TestGetBucket(t *testing.T) {
	if getBucket(10) != 0 {
		t.Fatal("expected 0, but ", getBucket(10))
	}
	if getBucket(100) != 0 {
		t.Fatal("expected 0, but ", getBucket(10))
	}
	if getBucket(101) != 1 || getBucket(200) != 1 || getBucket(201) != 2 {
		t.Fatal("expected 1, but ", getBucket(200))
	}
}
