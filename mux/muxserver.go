package mux

import (
  "io"
  "log"
  "kitty/buf"
)

type MuxerServer struct {
  Muxer
  connChan chan *Conn
}

func NewMuxerServer(ioc io.ReadWriteCloser)(muxerServer *MuxerServer, err error) {
  muxerServer = new(MuxerServer)
  muxerServer.conn = ioc
  muxerServer.inChan = make(chan []byte)
  muxerServer.outBufMap = make(map[string] *buf.Buffer)
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
      log.Printf("NewMuxerSerrver receive from %s\n  msg:%s", clientId, string(dataBytes))
      if err == nil {
	if _, ok := muxerServer.outBufMap[clientId]; !ok {
	  // open conn  passive
	  conn, _ := muxerServer.OpenConn(clientId)
	  log.Printf("NewMuxerSerrver openConn %s\n", clientId)
	  go func() {
	    muxerServer.connChan <- conn
	  }()
	}
	if _, err = muxerServer.writeToConnReadBuf(clientId, dataBytes); err != nil{
	  log.Printf("NewMuxerSerrver writeToConnReadBuf error %s\n", err)
	}
      } else {
	log.Printf("NewMuxerSerrver readFrom conn err %v\n", err)
      }
    }
  }()
  return
}

func (muxerServer *MuxerServer) Accept() (conn *Conn, err error) {
  conn = <- muxerServer.connChan
  return
}
