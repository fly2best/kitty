package main

import (
  "log"
  "net"
  "io"
  "kitty/mux"
)

const (
  muxerAddr = "127.0.0.1:7000"
  sock5Addr = "127.0.0.1:7070"
)

func main() {

  l, err := net.Listen("tcp", muxerAddr)
  if err != nil {
    log.Println("Error listening:", err.Error())
    return
  }
  log.Printf("kittyd listening: %s", muxerAddr)
  defer l.Close()

  // Listen for an incoming connection.
  for {
    conn, err := l.Accept()
    if err == nil {
      muxerServer, _ := mux.NewMuxerServer(conn)
      go muxerServe(muxerServer)
    } else {
      log.Printf("accept error! %v", err)
    }
  }
  return
}

func muxerServe(muxerServer *mux.MuxerServer) {
  for {
    conn, _ := muxerServer.Accept()
    log.Printf("accept conn %s", conn)
    socks5Conn, err := net.Dial("tcp", sock5Addr)
    if err != nil {
      log.Printf("muxerSever conn to sock5 proxy error %v", err)
    } else {

      go func () {
	defer conn.Close()
	defer socks5Conn.Close()

	n, err := io.Copy(conn, socks5Conn)
	log.Printf("copy sock5proxy to conn %v end, %d bytes written. err: %v", conn , n, err)
      }()

      go func () {
	defer conn.Close()
	defer socks5Conn.Close()

	n, err := io.Copy(socks5Conn, conn)
	log.Printf("copy conn %v to sock5proxy end, %d bytes written. err: %v", conn, n, err)
      }()
    }
  }
}
