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

package bytes

// Writer struct support io.Writer interface and allows to write data into extendable
// underlying buffer. It uses Pool for arranging new buffers
type Writer struct {
	pool *Pool
	buf  []byte
	pos  int
}

// Reset drops the existing buffer position and initializes the Writer to use p as a Pool
func (w *Writer) Reset(initSz int, p *Pool) {
	w.pool = p
	w.pos = 0
	w.buf = w.pool.Arrange(initSz)
}

// Write is part of io.Writer
func (w *Writer) Write(p []byte) (n int, err error) {
	av := len(w.buf) - w.pos
	if av < len(p) {
		nbs := int(2*len(w.buf) - len(w.buf)/5)
		if nbs-w.pos < len(p) {
			nbs = len(w.buf) + 2*len(p)
		}
		nb := w.pool.Arrange(nbs)
		copy(nb, w.buf[:w.pos])
		w.pool.Release(w.buf)
		w.buf = nb
	}
	n = copy(w.buf[w.pos:], p)
	w.pos += n
	return
}

// Buf returns underlying written buffer
func (w *Writer) Buf() []byte {
	return w.buf[:w.pos]
}

// Close releases resources and makes w unusable
func (w *Writer) Close() error {
	w.pool.Release(w.buf)
	w.buf = nil
	w.pool = nil
	return nil
}
