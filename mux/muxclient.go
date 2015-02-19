package mux

import (
  "io"
  "bytes"
)

type MuxerClient struct {
  Muxer
}

func NewMuxerClient(ioc io.ReadWriteCloser)(muxerClient *MuxerClient, err error) {
  muxerClient = new(MuxerClient)
  muxerClient.conn = ioc
  muxerClient.inChan = make(chan []byte)
  muxerClient.outChanMap = make(map[string]chan []byte)
  muxerClient.outBufMap = make(map[string] *bytes.Buffer)

  // write
  go func (){
    for {
      bytes := <- muxerClient.inChan
      muxerClient.conn.Write(bytes)
    }
  }()

  //read
  go func () {
    for {
      clientId, dataBytes, err := muxerClient.readFromConn()
      if err == nil {
	go func () {
	  if ch, ok := muxerClient.outChanMap[clientId]; ok {
	    ch <-dataBytes
	  }
	}()
      }
    }
  }()
  return
}
