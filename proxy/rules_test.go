package proxy

import (
  "testing"
  "fmt"
)

func TestInitProxyMgr(t *testing.T) {

  proxyMgr, err := LoadProxyMgrFromFile("../proxy.conf")

  if err != nil {
    fmt.Println(err)
  } else {
    fmt.Printf("%+v\n", proxyMgr)
    fmt.Println(proxyMgr.Match("google.com"))
    fmt.Println(proxyMgr.Match("www.google.com"))
    fmt.Println(proxyMgr.Match("www.google.com.hk"))
    fmt.Println(proxyMgr.Match("github.com"))
    fmt.Println(proxyMgr.Match("baidu.com"))
    fmt.Println(proxyMgr.GetDirectForwardAddr("example.com:80"))
  }
}
