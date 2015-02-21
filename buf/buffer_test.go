package buf

import (
  "testing"
  "time"
  "bytes"
  "math/rand"
  // "log"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func TestWRBuffer(t *testing.T) {

  buffer := NewBuffer()
  msgSend := "hello buffer wr!"
  buffer.Write(([]byte)(msgSend))

  byteBuf := make([]byte, 1000)
  n, _ := buffer.Read(byteBuf)
  msgRev := string(byteBuf[:n])

  if msgRev != msgSend {
    t.Errorf("msg send:% != msg rev:%\n", msgSend, msgRev)
  }
}

func TestRWBuffer(t *testing.T) {

  buffer := NewBuffer()
  msgSend := "hello buffer rw!"

  go func () {
    time.Sleep(100 * time.Millisecond)
    buffer.Write(([]byte)(msgSend))
  }()

  byteBuf := make([]byte, 1000)
  n, _ := buffer.Read(byteBuf)
  msgRev := string(byteBuf[:n])

  if msgRev != msgSend {
    t.Errorf("msg send:% != msg rev:%\n", msgSend, msgRev)
  }
}

func TestRWCurrent(t *testing.T) {

  buffer := NewBuffer()
  randStr := randSeq(1000000)
  wch := make(chan bool)
  rch := make(chan bool)

  var randStrRev string

  go func () {
    byteToSend := []byte(randStr)
    for i := 0; i < len(byteToSend); i++{
      buffer.Write(byteToSend[i:i+1])
    }
    buffer.Close()
    wch <- true
  }()

  go func () {
    bufferRead := new(bytes.Buffer)
    byteBuf := make([]byte, 100)
    for {
      n, err := buffer.Read(byteBuf)
      if err == nil {
	bufferRead.Write(byteBuf[0:n])
      } else {
	break
      }
    }
    randStrRev = string(bufferRead.Bytes())
    rch <- true
  }()

  <- wch
  <- rch

  if randStr != randStrRev {
    t.Errorf("msg send != msg rev\n")
    t.Errorf("msg send: %s", randStr)
    t.Errorf("msg rev: %s", randStrRev)
  }

}

func randSeq(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}
