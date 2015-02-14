package httpmgr

import (
  "fmt"
  "net/http"
  "kitty/proxy"
)

type ProxyHandler struct {
  ProxyMgr * proxy.ProxyMgr
}

func StartHttpServer(proxyMgr *proxy.ProxyMgr, httpAddr string) {
  proxyHandler :=  new(ProxyHandler)
  proxyHandler.ProxyMgr = proxyMgr

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
    proxyHandler.ls(w, r)
  })

  http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request){
    proxyHandler.update(w, r)
  })

  http.ListenAndServe(httpAddr, nil)
}

func (proxyHandler *ProxyHandler) ls(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "%s", proxyHandler.ProxyMgr.String())
}

func (proxyHandler *ProxyHandler) update(w http.ResponseWriter, r *http.Request) {
  proxyHandler.ProxyMgr.ReInit()
  fmt.Fprintf(w, "updated\n")
  proxyHandler.ls(w, r)
}
