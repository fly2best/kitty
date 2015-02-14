package httpmgr

import (
  "testing"
  "kitty/proxy"
)

func TestInitProxyMgr(t *testing.T) {

  proxyMgr := new(proxy.ProxyMgr)
  proxyMgr.Init("../proxy.conf")
  StartHttpServer(proxyMgr, ":6001")
}
