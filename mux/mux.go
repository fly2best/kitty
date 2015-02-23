package mux

import (
  "bytes"
  "encoding/binary"
  "io"
  "log"
  "sync"
  "kitty/buf"
  "fmt"
)

type Muxer struct {
  conn io.ReadWriteCloser
  inChan chan []byte

  outBufMap map[string] *buf.Buffer
  sync.Mutex
}

func NewMuxer(ioc io.ReadWriteCloser)(muxer *Muxer, err error) {
  muxer = new(Muxer)
  muxer.conn = ioc
  return
}

func (muxer *Muxer)Write(clientId string, p []byte) (n int, err error) {
  buf := new(bytes.Buffer)
  clientIdBytes := ([]byte)(clientId)
  binary.Write(buf, binary.BigEndian, int32(len(clientIdBytes)))
  binary.Write(buf, binary.BigEndian, clientIdBytes)
  binary.Write(buf, binary.BigEndian, int32(len(p)))
  binary.Write(buf, binary.BigEndian, p)
  if len(p) < 15 {
    log.Printf("Muxer Write start. client %s msg:%v", clientId, p)
  }
  muxer.inChan <- buf.Bytes()
  if len(p) < 15 {
    log.Printf("Muxer Write done. client %s  msg:%v done", clientId, p)
  }
  n = len(p)
  return
}

func (muxer *Muxer)Read(clientId string, p []byte) (n int, err error) {
  if buf, ok := muxer.outBufMap[clientId]; ok {
    return buf.Read(p)
  } else {
    err = io.EOF
    return
  }
}

func (muxer *Muxer)readFromConn() (clientId string, dataBytes []byte, err error) {
  var clientIdLen uint32
  err = binary.Read(muxer.conn, binary.BigEndian, &clientIdLen)
  if err != nil {
    err = fmt.Errorf("read clientId len error. %v", err)
    return
  }
  log.Printf("readFromConn clientIdLen:%d\n", clientIdLen)

  clientIdBytes := make([]byte, clientIdLen)
  err = binary.Read(muxer.conn, binary.BigEndian, clientIdBytes)
  if err != nil {
    err = fmt.Errorf("read clientId error. %v", err)
    return
  }

  clientId = string(clientIdBytes)
  log.Printf("readFromConn from %s\n", clientId)

  var dataLen uint32
  err = binary.Read(muxer.conn, binary.BigEndian, &dataLen)
  if err != nil {
    err = fmt.Errorf("read date len error. %v", err)
    return
  }
  log.Printf("readFromConn dataLen:%d\n", dataLen)

  if dataLen == 0 {
    muxer.CloseLocalConn(clientId)
    return
  }

  dataBytes = make([]byte, dataLen)
  err = binary.Read(muxer.conn, binary.BigEndian, dataBytes)
  if err != nil {
    err = fmt.Errorf("read date error. %v", err)
    return
  }
  // log.Printf("readFromConn %s dataBytes:%v\n", clientId,  dataBytes)
  return
}

func (muxer *Muxer)OpenConn(clientId string) (conn *Conn, err error) {
  conn = new(Conn)
  conn.id = clientId
  conn.muxer = muxer

  muxer.Lock()
  defer muxer.Unlock()
  muxer.outBufMap[clientId] = buf.NewBuffer()
  return
}

func (muxer *Muxer)CloseConn(clientId string) (err error) {

  //close remote conn
  bytes := make([]byte, 0)
  muxer.Write(clientId, bytes)

  muxer.CloseLocalConn(clientId)
  return
}

func (muxer *Muxer)CloseLocalConn(clientId string) (err error) {
  muxer.Lock()
  defer muxer.Unlock()

  if buf, ok := muxer.outBufMap[clientId]; ok {
    buf.Close()
  }

  delete(muxer.outBufMap, clientId)
  return
}

func (muxer *Muxer)writeToConnReadBuf(clientId string, dataBytes []byte)(n int,err error) {
  if buf, ok := muxer.outBufMap[clientId]; ok {
    n, err = buf.Write(dataBytes)
  } else {
    err = fmt.Errorf("writeToConnReadBuf cannot find client:%s", clientId)
  }
  return
}
