package mux

import (
  "io"
  "bytes"
  "log"
)

type MuxerServer struct {
  Muxer
  connChan chan *Conn
}

func NewMuxerServer(ioc io.ReadWriteCloser)(muxerServer *MuxerServer, err error) {
  muxerServer = new(MuxerServer)
  muxerServer.conn = ioc
  muxerServer.inChan = make(chan []byte)
  muxerServer.outChanMap = make(map[string]chan []byte)
  muxerServer.outBufMap = make(map[string] *bytes.Buffer)
  muxerServer.connChan = make(chan *Conn)

  // write
  go func (){
    for {
      bytes := <- muxerServer.inChan
      muxerServer.conn.Write(bytes)
    }
  }()

  //read
  go func () {
    for {
      clientId, dataBytes, err := muxerServer.readFromConn()
      log.Printf("NewMuxerSerrver receive from %s\n", clientId)
      if err == nil {
	if _, ok := muxerServer.outChanMap[clientId]; !ok {
	  // open conn  passive
	  conn, _ := muxerServer.OpenConn(clientId)
	  log.Printf("NewMuxerSerrver openConn %s\n", clientId)
	  go func() {
	    muxerServer.connChan <- conn
	  }()
	}
	//todo: different clientId can be out-of-order, same client should stay in-order
	ch := muxerServer.outChanMap[clientId]
	ch <- dataBytes
      }
    }
  }()
  return
}

func (muxerServer *MuxerServer) Accept() (conn *Conn, err error) {
  conn = <- muxerServer.connChan
  return
}
