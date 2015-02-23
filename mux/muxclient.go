package mux

import (
  "io"
  "log"
  "kitty/buf"
)

type MuxerClient struct {
  Muxer
}

func NewMuxerClient(ioc io.ReadWriteCloser)(muxerClient *MuxerClient, err error) {
  muxerClient = new(MuxerClient)
  muxerClient.conn = ioc
  muxerClient.inChan = make(chan []byte)
  muxerClient.outBufMap = make(map[string] *buf.Buffer)

  // write
  go func (){
    for {
      bytes := <- muxerClient.inChan
      n, err := muxerClient.conn.Write(bytes)
      if err != nil {
	log.Printf("muxerClient Write to conn error: %v", err)
      }
      if n != len(bytes) {
	log.Printf("muxerClient Write to conn error. %d != %d", n, len(bytes))
      }
      log.Printf("muxerClient Write to conn done, %d bytes write", n)
    }
  }()

  //read
  go func () {
    for {
      clientId, dataBytes, err := muxerClient.readFromConn()
      // log.Printf("NewMuxerClient receive from %s\n  msg:%v", clientId, dataBytes)
      if err  == nil {
	_, err := muxerClient.writeToConnReadBuf(clientId, dataBytes)
	if  err != nil {
	  log.Printf("muxerClient writeToConnRead Buf errro: %v", err)
	}
      } else {
	log.Printf("MuxerClient read error: %v", err)
	break
      }
    }
  }()
  return
}
