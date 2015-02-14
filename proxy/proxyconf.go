package proxy
import (
  "regexp"
  "bufio"
  "strings"
  "strconv"
  "bytes"
  "fmt"
  "log"
)

type Rules struct {
  Name string
  rawRules []string
  regexps []*regexp.Regexp
}

func (rules *Rules) Init(rulesName string, scanner *bufio.Scanner) (err error) {

  log.Printf("%s started to init\n", rulesName)

  //1. read the rules
  rules.Name = rulesName
  rawRules := make([]string, 0)
  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())
    if strings.HasPrefix(line, "#") {
      continue
    }
    if line != "}" {
      line = strings.TrimSpace(scanner.Text())
      rawRules = append(rawRules, line)
    } else {
      break
    }
  }
  rules.rawRules = rawRules

  //2. init the regex
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

func (rules *Rules) String() (str string) {
  var buffer bytes.Buffer

  fmt.Fprintf(&buffer, "rules %s {\n", rules.Name)
  for _, rawRule := range rules.rawRules {
    fmt.Fprintf(&buffer, "  %s\n", rawRule)
  }
  fmt.Fprintf(&buffer, "}")

  return buffer.String()
}

type DirectForward map[string]string

func (df *DirectForward) String() (str string) {
  var buffer bytes.Buffer
  fmt.Fprintf(&buffer, "directforward {\n")
  for key, value := range *df {
    fmt.Fprintf(&buffer, "  %s -> %s\n", key, value)
  }
  fmt.Fprintf(&buffer, "}")
  return buffer.String()
}

type Proxy struct {
  Name string
  ProxyType string
  Host string
  Port uint16
}

func (proxy *Proxy) Init(proxyName string, scanner *bufio.Scanner) (err error) {

  log.Printf("%s started to init\n", proxyName)
  proxy.Name = proxyName

  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())
    if strings.HasPrefix(line, "#") {
      continue
    }

    if line != "}" {
      fileds := strings.Fields(line)
      if fileds[0] == "type" {
	proxy.ProxyType = fileds[2]
      } else if fileds[0] == "host" {
	proxy.Host = fileds[2]
      } else if fileds[0] == "port" {
	if port, err := strconv.Atoi(fileds[2]); err == nil {
	  proxy.Port = uint16(port)
	}
      }
    } else {
      break
    }
  }
  return nil
}

func (proxy *Proxy) String() (str string) {
  var buffer bytes.Buffer
  fmt.Fprintf(&buffer, "proxy %s {\n", proxy.Name)
  fmt.Fprintf(&buffer, "  type = %s\n", proxy.ProxyType)
  fmt.Fprintf(&buffer, "  host = %s\n", proxy.Host)
  fmt.Fprintf(&buffer, "  port = %d\n", proxy.Port)
  fmt.Fprintf(&buffer, "}")
  return buffer.String()
}

type ProxyConfig map[string]string

func (config *ProxyConfig ) String() (str string) {
  var buffer bytes.Buffer
  fmt.Fprintf(&buffer, "config {\n")
  for key, value := range *config {
    fmt.Fprintf(&buffer, "  %s -> %s\n", key, value)
  }
  fmt.Fprintf(&buffer, "}")
  return buffer.String()
}
