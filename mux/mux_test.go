package mux

import (
  "testing"
  "net"
  "fmt"
  "log"
  "bytes"
  "time"
)

const (
  clientNum= 1000
  msgNumPerClient = 1000
)

func TestMux(t *testing.T) {

  go server(clientNum)
  go client(clientNum)

  ch := make(chan string)
  <- ch
}

func server(clientCount int) {
  connServer, _ := getServerConn()
  muxServer, _ := NewMuxerServer(connServer)
  fmt.Println(muxServer)
  for {

    conn, _ := muxServer.Accept()
    clientCount = clientCount - 1
    log.Printf("accept %s, there will be %d more\n", conn, clientCount)
    go func() {
      buf := new(bytes.Buffer)

      for {
	byteBuf := make([]byte, 1000)
	if n, err := conn.Read(byteBuf); err == nil {
	  buf.Write(byteBuf[:n])
	  log.Printf("conn %s receive %s\n", conn, string(byteBuf[:n]))
	} else{
	  break
	}
      }

      log.Printf("conn %s closed\n", conn)

      bufExpected := new(bytes.Buffer)
      msgsToSend := getClientMsg(conn.String())
      for _, msg := range(msgsToSend) {
	bufExpected.Write(([]byte)(msg))
      }

      strRev := string(buf.Bytes())
      strExpected := string(bufExpected.Bytes())

      if strRev != strExpected {
	log.Printf("check conn %s msg error\n strExpected: %s\n strRev:%s\n", conn, strExpected, strRev)
      } else {
	log.Printf("check conn %s msg ok", conn)
      }
    }()
  }

}

func client(clientCount int) {
  connClient, _ := getClientConn()
  muxClient, _ := NewMuxerClient(connClient)
  fmt.Println(muxClient)

  for i := 0; i < clientCount; i++ {
    clientId := fmt.Sprintf("client %d", i)
    go func(clientId string) {
      conn, _ := muxClient.OpenConn(clientId)
      defer conn.Close()
      msgs := getClientMsg(clientId)
      for _, msg := range(msgs) {
	conn.Write(([]byte)(msg))
      }

      time.Sleep(3000 * time.Millisecond)
    }(clientId)
  }
}

func getClientMsg(clientId string)(msgToSend []string) {
  msgToSend = make([]string, msgNumPerClient)
  for i := 0; i < msgNumPerClient ; i++ {
    msgToSend[i] = fmt.Sprintf("hello server, this is %s, msg %d", clientId, i)
  }
  return
}

func getServerConn() (conn net.Conn, err error) {
  l, er := net.Listen("tcp", "127.0.0.1:6060")
  if er != nil {
    fmt.Println("Error listening:", err.Error())
    err = er
    return
  }
  // Close the listener when the application closes.
  defer l.Close()

  // Listen for an incoming connection.
  conn, err = l.Accept()
  return
}

func getClientConn() (conn net.Conn, err error) {
  conn, err = net.Dial("tcp", "127.0.0.1:6060")
  return
}
