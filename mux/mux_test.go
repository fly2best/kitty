package mux

import (
  "testing"
  "net"
  "fmt"
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
  conn, _ := muxServer.Accept()
  fmt.Printf("accept %s\n", conn)

  for {
    byteBuf := make([]byte, 1000)
    n, _ := conn.Read(byteBuf)
    msg := string(byteBuf[:n])
    fmt.Printf("clinet %s, msg %s\n", conn, msg)
  }

}

func client() {
  connClient, _ := getClientConn()
  muxClient, _ := NewMuxerClient(connClient)
  fmt.Println(muxClient)
  conn, _ := muxClient.OpenConn("localtest")
  conn.Write(([]byte)("hello"))
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
