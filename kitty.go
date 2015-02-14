package main

import (
  "kitty/proxy"
  "log"
  "net"
)

const (
  localAddr         = "127.0.0.1:6000"
  proxyConfFileName = "proxy.conf"
)

func main() {

  if proxyMgr, err := proxy.LoadProxyMgrFromFile(proxyConfFileName); err != nil {
    log.Fatalf("load proxy conf %s err: %v", proxyConfFileName, err)
  } else {
    ln, err := net.Listen("tcp", localAddr)
    if err != nil {
      log.Fatalf("listen to %s err: %v", localAddr, err)
    }

    for {
      conn, err := ln.Accept()
      if err != nil {
	log.Printf("accept err: %v", localAddr, err)
      }
      go proxy.Sock5serve(conn, proxyMgr)
    }
  }
}
