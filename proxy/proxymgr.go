package proxy

import (
  "os"
  "bufio"
  "bytes"
  "strings"
  "fmt"
  "log"
  "net"
  "kitty/mux"
  "strconv"
  "github.com/streamrail/concurrent-map"
)

type ProxyMgr struct {
  RulesList []*Rules
  ProxyList []*Proxy
  ForwardMap DirectForward
  Config ProxyConfig
  KittyMuxerClientMap map[string]*mux.MuxerClient

  ProxyConfFile string
  hostConnMap cmap.ConcurrentMap
}

func (proxyMgr *ProxyMgr) Init(proxyConfFile string) (err error) {

  file, err := os.Open(proxyConfFile)
  if err != nil {
    err = fmt.Errorf("open config file error. %v", err)
    return
  }

  defer file.Close()
  scanner := bufio.NewScanner(file)

  rulesList := make([]*Rules, 0)
  proxyList := make([]*Proxy, 0)
  config := make(map[string]string)
  forwardMap := make(map[string]string)

  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())

    if strings.HasPrefix(line, "#") {
      continue
    }

    if strings.HasPrefix(line, "rules") {
      fileds := strings.Fields(line)
      rulesName := fileds[1]
      rules := new(Rules)
      er := rules.Init(rulesName , scanner); if er == nil {
	rulesList = append(rulesList, rules)
      }else {
	log.Printf("rule init err, %v", er)
      }
    }

    if strings.HasPrefix(line, "proxy") {
      fileds := strings.Fields(line)
      proxyName := fileds[1]
      proxy := new(Proxy)
      proxy.Init(proxyName, scanner)
      proxyList = append(proxyList, proxy)
    }

    if strings.HasPrefix(line, "directforward") {
      for scanner.Scan() {
	line = strings.TrimSpace(scanner.Text())
	if strings.HasPrefix(line, "#") {
	  continue
	}
	if line != "}" {
	  fields := strings.Fields(line)
	  forwardMap[fields[0]] = fields[2]
	} else {
	  break
	}
      }
    }

    if strings.HasPrefix(line, "config") {
      for scanner.Scan() {
	line = strings.TrimSpace(scanner.Text())
	if strings.HasPrefix(line, "#") {
	  continue
	}
	if line != "}" {
	  fields := strings.Fields(line)
	  config[fields[0]] = fields[2]
	} else {
	  break
	}
      }
    }
  }

  kittyMuxerClientMap := make(map[string]*mux.MuxerClient)
  for _, proxy := range(proxyList) {
    if proxy.ProxyType == "kitty" {
      proxyAddr := proxy.Host + ":" + strconv.Itoa(int(proxy.Port))
      conn, er := net.Dial("tcp", proxyAddr)
      if er == nil {
	if muxerClient, err := mux.NewMuxerClient(conn); err == nil {
	  kittyMuxerClientMap[proxy.Name] = muxerClient
	}
      } else {
	log.Printf("init connect to kitty proxy error. %v", er)
      }
    }
  }

  if er := scanner.Err(); err != nil {
    err = fmt.Errorf("cofig file read error. %v", er)
  } else {
    proxyMgr.RulesList = rulesList
    proxyMgr.ProxyList = proxyList
    proxyMgr.ForwardMap = forwardMap
    proxyMgr.Config = config
    proxyMgr.ProxyConfFile = proxyConfFile
    proxyMgr.KittyMuxerClientMap =  kittyMuxerClientMap

    proxyMgr.hostConnMap =  cmap.New()
  }
  return
}

func (proxyMgr *ProxyMgr) Match(host string) (proxy *Proxy, matched bool) {

  matched = false
  for _, rules := range proxyMgr.RulesList {
    if rules.Match(host) {
      proxyName := proxyMgr.Config[rules.Name]
      for  _, proxy := range proxyMgr.ProxyList {
	if proxy.Name == proxyName {
	  return proxy, true
	}
      }
    }
  }
  return
}

func (proxyMgr *ProxyMgr) GetDirectForwardAddr(dstAddr string)(forwardAddr string, contained bool) {
  forwardAddr, contained = proxyMgr.ForwardMap[dstAddr]
  return
}

func (proxyMgr *ProxyMgr) ReInit() error {

  //close all conn
  for tuple := range proxyMgr.hostConnMap.Iter() {
    conn := tuple.Val.(net.Conn)
    if conn != nil {
      conn.Close()
    }
  }

  return proxyMgr.Init(proxyMgr.ProxyConfFile)
}

func (proxyMgr *ProxyMgr) RegisterConn(host string, conn net.Conn) error {
  key := fmt.Sprintf("%s%v", host, conn)
  log.Printf("register conn: %s\n", key)
  proxyMgr.hostConnMap.Set(key, conn)
  return nil
}

func (proxyMgr *ProxyMgr) RemoveConn(host string, conn net.Conn) error {
  key := fmt.Sprintf("%s%v", host, conn)
  log.Printf("remove conn: %s\n", key)
  proxyMgr.hostConnMap.Remove(key)
  return nil
}

func (proxyMgr *ProxyMgr) String() string{
  var buffer bytes.Buffer

  for _, rules := range proxyMgr.RulesList{
    fmt.Fprintf(&buffer, "%s\n", rules.String())
  }

  fmt.Fprintf(&buffer, "%s\n", proxyMgr.ForwardMap.String())

  for _, proxy := range proxyMgr.ProxyList{
    fmt.Fprintf(&buffer, "%s\n", proxy.String())
  }

  fmt.Fprintf(&buffer, "%s", proxyMgr.Config.String())
  return buffer.String()
}
