package mux

import (
  "testing"
  "net"
  "fmt"
  "log"
)

func TestMux(t *testing.T) {
  go server()
  go client()

  ch := make(chan string)
  <- ch
}

func server() {
  connServer, _ := getServerConn()
  muxServer, _ := NewMuxerServer(connServer)
  fmt.Println(muxServer)

  for {
    conn, _ := muxServer.Accept()
    log.Printf("accept %s\n", conn)
    go func() {
      for {
	byteBuf := make([]byte, 1000)
	n, _ := conn.Read(byteBuf)
	msg := string(byteBuf[:n])
	log.Printf("server receive client:%s, msg:%s\n", conn, msg)
      }
    }()
  }

}

func client() {
  connClient, _ := getClientConn()
  muxClient, _ := NewMuxerClient(connClient)
  fmt.Println(muxClient)
  for i := 0; i < 100; i++ {
    clientId := fmt.Sprintf("client %d", i)
    go func(clientId string) {
      conn, _ := muxClient.OpenConn(clientId)
      defer conn.Close()
      for j := 0; j < 100; j++ {
	msg := fmt.Sprintf("hello server, this is %s, msg %d", clientId, j)
	log.Printf("client clientId:%s send msg:%s\n", clientId, msg)
	conn.Write(([]byte)(msg))
      }
    }(clientId)
  }
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
