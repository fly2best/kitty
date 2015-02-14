package proxy

import (
  "net"
  "io"
  "fmt"
  "errors"
  "strconv"
  "log"
)

const (
  sock5Version = uint8(5)
  connCmd = uint8(1)
  atypIPV4 = uint8(1)
  atypDomainName = uint8(3)
  atypIPV6 = uint8(4)
)

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

func getRemoteConnByDestAddr(host string, port uint16, proxyMgr *ProxyMgr) (conn net.Conn, err error) {

  dstAddr := host + ":" + strconv.Itoa(int(port))
  if forwardAddr, ok := proxyMgr.GetDirectForwardAddr(dstAddr); ok {
    log.Printf("direct forward %s to %s", dstAddr, forwardAddr)
    return net.Dial("tcp", forwardAddr)
  }

  if proxy, matched := proxyMgr.Match(host); matched {
    // sock5 proxy conn
    log.Printf("connect to host %s by proxy %v", host, proxy)
    return connectToSock5Proxy(host, port, proxy)
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
    log.Printf("handleConnCmd connectingt to %s:%d", host, port)
  }

  if remoteConn, err := getRemoteConnByDestAddr(host, port, proxyMgr); err != nil {
    log.Printf("getRemoteConnByDestAddr %s:%d error: %v", host, port, err)
    return err
  } else {

    cmdRet := []byte{5, 0, 0, 1, 0, 0, 0, 0, 3, 3}
    conn.Write(cmdRet)

    go func() {
      defer conn.Close()
      defer remoteConn.Close()

      io.Copy(conn, remoteConn)
    }()

    go func() {
      defer conn.Close()
      defer remoteConn.Close()

      io.Copy(remoteConn, conn)
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

func connectToSock5Proxy(host string, port uint16, proxy *Proxy) (conn net.Conn, err error) {
  proxyAddr := proxy.Host + ":" + strconv.Itoa(int(proxy.Port))
  proxyConn, er := net.Dial("tcp", proxyAddr)
  if er != nil {
    err = fmt.Errorf("connect to %s error! %v", proxyAddr, er)
    return
  }

  proxyConn.Write([]byte{sock5Version, 0x01, 0x00})
  shakeBuf := make([]byte, 2)
  _, er = io.ReadAtLeast(proxyConn, shakeBuf, 2)

  if er != nil {
    err = fmt.Errorf("read from %s error! %v", proxyAddr, er)
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

  requestRetBuf := make([]byte, 10)
  if  _, er = io.ReadAtLeast(proxyConn, requestRetBuf, 10); er != nil {
    err = fmt.Errorf("read from %s error! %v", proxyAddr, er)
  }

  return proxyConn, nil
}
