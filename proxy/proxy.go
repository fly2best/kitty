package proxy

import (
  "net"
  "io"
  "fmt"
  "errors"
  "strconv"
  "log"
  "sync"
)

const (
  sock5Version = uint8(5)
  connCmd = uint8(1)
  atypIPV4 = uint8(1)
  atypDomainName = uint8(3)
  atypIPV6 = uint8(4)
)

var connId = 0
var connIdLock sync.Mutex

var ErrNotSock5Version = errors.New("not sock5 version")
var ErrNotSupportdCmd = errors.New("just support conn cmd now")
var ErrNotSupportdIPV6 = errors.New("not support ipv6 adddr")

func Sock5serve(conn net.Conn, proxyMgr *ProxyMgr) {

  err := handleHandShake(conn)
  if err != nil {
    log.Printf("handle conn handshake err %v", err)
    return
  }

  log.Printf("start to handle request %v", conn)

  err = handleRequest(conn, proxyMgr)
  if err != nil {
    log.Printf("handle conn request err %v", err)
    return
  }

}

func getRemoteConnByDestAddr(host string, port uint16, proxyMgr *ProxyMgr) (conn io.ReadWriteCloser, err error) {

  dstAddr := host + ":" + strconv.Itoa(int(port))
  if forwardAddr, ok := proxyMgr.GetDirectForwardAddr(dstAddr); ok {
    log.Printf("direct forward %s to %s", dstAddr, forwardAddr)
    return net.Dial("tcp", forwardAddr)
  }

  if proxy, matched := proxyMgr.Match(host); matched {
    log.Printf("connect to host %s by proxy %v", host, proxy.Name)
    return connectToSock5Proxy(host, port, proxy, proxyMgr)
  } else {
    // dircet conn
    log.Printf("direct connto to host %s", host)
    return net.Dial("tcp", dstAddr)
  }
}

func handleConnCmd(conn net.Conn, proxyMgr *ProxyMgr) (err error) {

  host, port, err := getDestConnAddr(conn)
  if err != nil {
    return fmt.Errorf("handleConnCmd connect to %s:%d, error:%v", host, port, err)
  } else {
    log.Printf("handleConnCmd connecting to %s:%d", host, port)
  }

  if remoteConn, err := getRemoteConnByDestAddr(host, port, proxyMgr); err != nil {
    log.Printf("getRemoteConnByDestAddr %s:%d error: %v", host, port, err)
    return err
  } else {

    cmdRet := []byte{5, 0, 0, 1, 0, 0, 0, 0, 3, 3}
    conn.Write(cmdRet)

    log.Printf("handleConnCmd start to copy, local:%v, remote:%v", conn, remoteConn)

    go func() {
      defer conn.Close()
      defer remoteConn.Close()
      proxyMgr.RegisterConn(host, conn)
      defer proxyMgr.RemoveConn(host, conn)

      n, err := io.Copy(conn, remoteConn)
      if err == nil {
	log.Printf("localhost to %s closed, %d bytes written", host, n)
      } else {
	log.Printf("copy from localhost to %s err:%v", host, err)
      }
    }()

    go func() {
      defer conn.Close()
      defer remoteConn.Close()

      n, err := io.Copy(remoteConn, conn)
      if err == nil {
	log.Printf("%s to localhost closed, %d bytes written", host, n)
      } else {
	log.Printf("copy %s to localhost err: %v", host, err)
      }
    }()

    return nil
  }
}


func handleHandShake(conn net.Conn) (err error) {
  header := make([]byte, 2)
  _, err = io.ReadAtLeast(conn, header, 2)
  if err != nil {
    return err
  }

  if header[0] != sock5Version {
    return ErrNotSock5Version
  }

  methodsLen := int(header[1])
  methodsBuf := make([]byte, methodsLen)
  _, err = io.ReadAtLeast(conn, methodsBuf, methodsLen)
  if err != nil {
    return fmt.Errorf("read auth methods err: %v", err)
  }

  _, err = conn.Write([]byte{sock5Version , 0})
  return err
}

func handleRequest(conn net.Conn, proxyMgr *ProxyMgr) (err error) {

  requestBuf := make([]byte, 3)
  _, er := io.ReadAtLeast(conn, requestBuf, 3)

  if er != nil {
    return fmt.Errorf("read reques buf err: %s", er)
  }

  if requestBuf[1] != connCmd {
    return fmt.Errorf("request cmd %d not supported, err: %v", requestBuf[1], ErrNotSupportdCmd)
  }

  return handleConnCmd(conn, proxyMgr)
}


func getDestConnAddr(conn net.Conn) (host string, port uint16, err error) {

  atypBuf := []byte{0}
  _, err = io.ReadAtLeast(conn, atypBuf, 1)
  if err != nil {
    return
  }

  atyp := uint8(atypBuf[0])
  switch atyp {
  case atypIPV4:
    ipv4 := make([]byte, 4)
    _, err = io.ReadAtLeast(conn, ipv4, 4)
    if err != nil {
      return
    }
    host = net.IP(ipv4).String()
  case atypDomainName:
    dnLen := make([]byte, 1)
    _, err = io.ReadAtLeast(conn, dnLen, 1)
    if err != nil {
      return
    }
    dn := make([]byte, int(dnLen[0]))
    _, err = io.ReadAtLeast(conn, dn, int(dnLen[0]))
    if err != nil {
      return
    }
    host = string(dn)
  default:
    err = ErrNotSupportdIPV6
    return
  }

  portBuf := []byte{0, 0}
  _, err = io.ReadAtLeast(conn, portBuf, 2)
  if err != nil {
    return
  }

  port = uint16(uint16(portBuf[0]) << 8) | uint16(portBuf[1])

  return
}

func getProxyConn(host string, port uint16, proxy *Proxy, proxyMgr *ProxyMgr) (conn io.ReadWriteCloser, err error) {
  if proxy.ProxyType == "kitty" {
    if muxerClient, ok := proxyMgr.KittyMuxerClientMap[proxy.Name]; ok {
      // localAddr := host + ":" + strconv.Itoa(int(port))

      connIdLock.Lock()
      currConnid := connId
      connId = connId + 1
      connIdLock.Unlock()
      conn, err = muxerClient.OpenConn(strconv.Itoa(currConnid))
      log.Printf("open kitty proxy conn %s for host: %s", conn, host)
    } else {
      err = fmt.Errorf("kitty proxy muxerClient is null, %v", proxy)
    }
  } else {
    proxyAddr := proxy.Host + ":" + strconv.Itoa(int(proxy.Port))
    conn, err = net.Dial("tcp", proxyAddr)
  }
  return
}

func connectToSock5Proxy(host string, port uint16, proxy *Proxy, proxyMgr *ProxyMgr) (conn io.ReadWriteCloser, err error) {
  proxyConn, er := getProxyConn(host, port, proxy, proxyMgr)
  if er != nil {
    err = fmt.Errorf("connect to %s error! %v", proxy, er)
    return
  }

  proxyConn.Write([]byte{sock5Version, 0x01, 0x00})
  log.Printf("conn %v write handshake, waiting hanshake over", proxyConn)
  shakeBuf := make([]byte, 2)
  _, er = io.ReadAtLeast(proxyConn, shakeBuf, 2)

  if er != nil {
    err = fmt.Errorf("read from %s error! %v", proxy, er)
    return
  }

  if shakeBuf[0] != sock5Version {
    err = ErrNotSock5Version
    return
  }

  // version, cmd ,reserve, attr type
  proxyConn.Write([]byte{sock5Version, 0x01, 0x00, 0x03})

  hostByteArray := []byte(host)
  bytesLen := []byte{uint8(len(hostByteArray))}
  proxyConn.Write(bytesLen)
  proxyConn.Write(hostByteArray)
  portBytes := []byte{uint8(port>> 8), uint8(port)}
  proxyConn.Write(portBytes)

  log.Printf("conn %v write conn cmd, waiting return", proxyConn)

  requestRetBuf := make([]byte, 10)
  if  _, er = io.ReadAtLeast(proxyConn, requestRetBuf, 10); er != nil {
    err = fmt.Errorf("read from %s error! %v", proxy, er)
  }

  log.Printf("conn %v conn cmd result return", proxyConn)

  return proxyConn, nil
}
