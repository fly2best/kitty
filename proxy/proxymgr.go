package proxy

import (
  "os"
  "bufio"
  "strings"
  "fmt"
  "log"
)

type ProxyMgr struct {
  RulesList []*Rules
  ProxyList []*Proxy
  ForwardMap DirectForward
  Config map[string]string
  ProxyConfFile string
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

  if er := scanner.Err(); err != nil {
    err = fmt.Errorf("cofig file read error. %v", er)
  } else {
    proxyMgr.RulesList = rulesList
    proxyMgr.ProxyList = proxyList
    proxyMgr.ForwardMap = forwardMap
    proxyMgr.Config = config
    proxyMgr.ProxyConfFile = proxyConfFile
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
