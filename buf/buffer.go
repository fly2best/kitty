package mux

import (
  "bytes"
  "sync"
)

//todo: improve buf mem utilization

type Buffer struct {
  buf bytes.Buffer
  m sync.Mutex
  cond *sync.Cond
  closed bool
}

func (b *Buffer) Write(p []byte) (n int, err error) {
  b.m.Lock()
  defer b.m.Unlock()
  n, err =  b.buf.Write(p)
  if n != 0 {
    b.cond.Signal()
  }
  return
}

func (b *Buffer) Read(p []byte) (n int, err error) {
  b.m.Lock()
  defer b.m.Unlock()

  for b.buf.Len() == 0  && !b.closed {
    b.buf.Reset()
    b.cond.Wait()
  }

  // when buf had data, read it before return EOF
  if b.buf.Len() != 0 {
    n, err = b.buf.Read(p)
  } else {
    // b.closed = true
    n = 0
  }
  return
}

func (b *Buffer) Close() ( err error) {
  b.m.Lock()
  defer b.m.Unlock()
  b.closed = true
  b.cond.Signal()
  return
}

func NewBuffer() (buf *Buffer) {
  buf = new(Buffer)
  buf.cond = sync.NewCond(&buf.m)
  buf.closed = false
  return
}

