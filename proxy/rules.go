package proxy

import (
  "regexp"
  "os"
  "bufio"
  "strings"
  "strconv"
  "fmt"
  "log"
)


type Rules struct {
  Name string
  rawRules []string
  regexps []*regexp.Regexp
}

type DirectForward map[string]string

type Proxy struct {
  Name string
  Ptype string
  Host string
  Port uint16
}

type ProxyMgr struct {
  RulesList []Rules
  ProxyList []Proxy
  ForwardMap DirectForward
  Config map[string]string
}

func LoadProxyMgrFromFile(fileName string) (mgr *ProxyMgr, err error) {

  file, err := os.Open(fileName)
  if err != nil {
    err = fmt.Errorf("open config file error. %v", err)
    return
  }

  defer file.Close()
  scanner := bufio.NewScanner(file)


  rulesList := make([]Rules, 0)
  proxyList := make([]Proxy, 0)
  config := make(map[string]string)
  forwardMap := make(map[string]string)

  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())

    if strings.HasPrefix(line, "#") {
      continue
    }

    if strings.HasPrefix(line, "rules") {
      fileds := strings.Fields(line)
      ruleName := fileds[1]
      rawRules := make([]string, 0)
      for scanner.Scan() {
	if strings.HasPrefix(line, "#") {
	  continue
	}
	if line != "}" {
	  line = strings.TrimSpace(scanner.Text())
	  rawRules = append(rawRules, line)
	} else {
	  rules := Rules{ruleName, rawRules, nil}
	  if er:= rules.init(); er != nil {
	    err = er
	    return
	  } else {
	    rulesList = append(rulesList, rules)
	  }
	  break
	}
      }
    }

    if strings.HasPrefix(line, "proxy") {
      fileds := strings.Fields(line)
      proxyName := fileds[1]
      var proxyType, proxyHost string
      var proxyPort uint16

      for scanner.Scan() {
	line = strings.TrimSpace(scanner.Text())

	if strings.HasPrefix(line, "#") {
	  continue
	}
	if line != "}" {
	  fileds := strings.Fields(line)
	  if fileds[0] == "type" {
	    proxyType = fileds[2]
	  } else if fileds[0] == "host" {
	    proxyHost = fileds[2]
	  } else if fileds[0] == "port" {
	    if port, err := strconv.Atoi(fileds[2]); err == nil {
	      proxyPort = uint16(port)
	    }
	  }
	} else {
	  proxy := Proxy{proxyName, proxyType, proxyHost, proxyPort}
	  proxyList = append(proxyList, proxy)
	  break
	}
      }
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
    mgr = &ProxyMgr{rulesList, proxyList, forwardMap, config}
  }

  return
}


func (rules *Rules) init() error {

  log.Printf("%s started to init\n", rules.Name)

  regexps := make([]*regexp.Regexp, 0)

  for _, rawRule := range rules.rawRules {
    regexStr := strings.Replace(rawRule, ".", `\.`, -1)
    regexStr = strings.Replace(regexStr, "*", ".*", -1)

    if rp, er := regexp.Compile(regexStr); er == nil {
      regexps = append(regexps, rp)
    } else {
      log.Printf("compile rule failed, %s, %s, %v\n", rawRule, regexStr, er)
    }
  }

  rules.regexps = regexps
  log.Printf("rule intid, %s\n", rules.Name)
  return nil
}


func (rules *Rules) Match(host string) (bool) {

  for _, rp:= range rules.regexps {
    if (rp.MatchString(host)) {
      log.Printf("%s matched\n", host)
      return true
    }
  }

  return false
}

func (proxyMgr *ProxyMgr) Match(host string) (proxy Proxy, matched bool) {

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
