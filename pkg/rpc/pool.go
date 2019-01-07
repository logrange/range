package rpc

type (
	pool struct {
	}
)

func (p *pool) arrange(size int) []byte {
	return make([]byte, size)
}

func (p *pool) release(b []byte) {
	// Later
}
