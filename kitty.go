package main

import (
  "kitty/proxy"
  "kitty/httpmgr"
  "log"
  "net"
)

const (
  localAddr = "127.0.0.1:6000"
  htppMgrAddr = "127.0.0.1:6001"
  proxyConfFileName = "./conf/proxy.conf"
)

func main() {
  proxyMgr := new(proxy.ProxyMgr)
  if err := proxyMgr.Init(proxyConfFileName); err != nil {
    log.Fatalf("proxy mgr init err: %v", err)
  } else {
    ch := make(chan error)
    go sock5(proxyMgr, ch)
    go httpMgr(proxyMgr, ch)

    err := <-ch
    log.Fatalf("proxy exit with err: %v", err)
  }
}

func sock5(proxyMgr *proxy.ProxyMgr, errChan chan error) {

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

func httpMgr(proxyMgr *proxy.ProxyMgr, errchan chan error) {
  httpmgr.StartHttpServer(proxyMgr, htppMgrAddr)
}
