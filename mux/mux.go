package mux

import (
  "bytes"
  "encoding/binary"
  "io"
  "log"
  "sync"
)

type Muxer struct {
  conn io.ReadWriteCloser
  inChan chan []byte

  outChanMap map[string]chan []byte
  outBufMap map[string] *bytes.Buffer
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
  // msgLen := int32(4 + len(clientIdBytes) + 4 + len(p))
  // binary.Write(buf, binary.BigEndian, msgLen)
  binary.Write(buf, binary.BigEndian, int32(len(clientIdBytes)))
  binary.Write(buf, binary.BigEndian, clientIdBytes)
  binary.Write(buf, binary.BigEndian, int32(len(p)))
  binary.Write(buf, binary.BigEndian, p)
  muxer.inChan <- buf.Bytes()
  return
}

func (muxer *Muxer)Read(clientId string, p []byte) (n int, err error) {
  buf := muxer.outBufMap[clientId]
  // log.Printf("Muxer.Read client:%s, len:%d\n", clientId, buf.Len())
  if buf.Len() != 0 {
    return buf.Read(p)
  } else {
    bytesBuf := <- muxer.outChanMap[clientId]
    buf.Write(bytesBuf)
    return buf.Read(p)
  }
}

func (muxer *Muxer)readFromConn() (clientId string, dataBytes []byte, err error) {

  // var msgLen uint32
  // err = binary.Read(muxer.conn, binary.BigEndian, &msgLen)
  // if err != nil {
  // log.Printf("read error %v.\n", err)
  // return
  // }
  // log.Printf("readFromConn msgLen:%d\n", msgLen)

  var clientIdLen uint32
  err = binary.Read(muxer.conn, binary.BigEndian, &clientIdLen)
  if err != nil {
    log.Printf("read error %v.\n", err)
    return
  }
  log.Printf("readFromConn clientIdLen:%d\n", clientIdLen)

  clientIdBytes := make([]byte, clientIdLen)
  err = binary.Read(muxer.conn, binary.BigEndian, clientIdBytes)
  if err != nil {
    log.Printf("read error %v.\n", err)
    return
  }

  clientId = string(clientIdBytes)
  log.Printf("readFromConn from %s\n", clientId)

  var dataLen uint32
  err = binary.Read(muxer.conn, binary.BigEndian, &dataLen)
  if err != nil {
    log.Printf("read error %v.\n", err)
    return
  }
  log.Printf("readFromConn dataLen:%d\n", dataLen)
  dataBytes = make([]byte, dataLen)
  err = binary.Read(muxer.conn, binary.BigEndian, dataBytes)
  if err != nil {
    log.Printf("read error %v.\n", err)
    return
  }
  log.Printf("readFromConn dataBytes:%s\n", string(dataBytes))

  return
}

func (muxer *Muxer)OpenConn(clientId string) (conn *Conn, err error) {
  conn = new(Conn)
  conn.id = clientId
  conn.muxer = muxer

  muxer.Lock()
  defer muxer.Unlock()
  muxer.outChanMap[clientId] = make(chan []byte)
  muxer.outBufMap[clientId] =  new(bytes.Buffer)

  return
}

func (muxer *Muxer)CloseConn(clientId string) (err error) {
  muxer.Lock()
  defer muxer.Unlock()

  delete(muxer.outChanMap, clientId)
  delete(muxer.outBufMap, clientId)
  return
}

