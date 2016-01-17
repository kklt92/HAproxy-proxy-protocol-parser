package main

import (
  "log"
  "net"
  "fmt"
  "net/http"
  "strings"
  "bufio"
  "strconv"
)

const (
  v1Signature = "PROXY"
  v2Signature = "\x0D\x0A\x0D\x0A\x00\x0D\x0A\x51\x55\x49\x54\x0A" 
  ProxyMinLen = 12
)

type ConnInfo struct {
  Version int
  SrcIP   net.IP
  SrcPort int
  DstIP   net.IP
  DstPort int
  Inet    string
}

func (c ConnInfo) String() string {
  str := "Version: " + strconv.Itoa(c.Version) + "  Type: " + c.Inet + 
         "\nSource: " + c.SrcIP.String() + ":" + strconv.Itoa(c.SrcPort) +
         "\nDestination: " + c.DstIP.String() + ":" + strconv.Itoa(c.DstPort) //+
  //       "\nPayload:\n" + c.Payload
  
  return str
}

func parseProxy(str string) (*ConnInfo) {
  connInfo := new(ConnInfo)  
  if str[:5] == v1Signature {
    connInfo.Version = 1
    parseProxyV1(str, connInfo)
  } else if str[:12] == v2Signature {
    connInfo.Version = 2
    parseProxyV2(str, connInfo)
  } else {
  }
  return connInfo
}

func parseProxyV1(str string, connInfo *ConnInfo) {
  proxyStr := str[:len(str)-2]
  proxyTokens := strings.Split(proxyStr, " ")
  if (len(proxyTokens) <= 1 || len(proxyTokens) > 6) {
    log.Fatal("Error: packet length not valid: " + strconv.Itoa(len(proxyTokens)))
    return
  }
  if proxyTokens[1] == "UNKNOWN" {
    connInfo.Inet = "tcp" 
  } else if proxyTokens[1] == "TCP4" {
    connInfo.Inet = "tcp4"
  } else if proxyTokens[1] == "TCP6" {
    connInfo.Inet = "tcp6"
  }

  if (len(proxyTokens) == 2) {
    return
  }
  connInfo.SrcIP = net.ParseIP(proxyTokens[2])
  if (len(proxyTokens) == 3) {
    return
  }
  connInfo.DstIP = net.ParseIP(proxyTokens[3])
  if (len(proxyTokens) == 4) {
    return
  }

  port, err := strconv.Atoi(proxyTokens[4])
  if err != nil {
    log.Fatal(err)
  }
  connInfo.SrcPort = port
  
  if len(proxyTokens) == 5 {
    return
  }
  port, err = strconv.Atoi(proxyTokens[5])
  if err != nil {
    log.Fatal(err)
  }   
  connInfo.DstPort = port
}

func parseProxyV2(str string, connInfo *ConnInfo) {
  // TODO 
}

func handleDefault(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("<h1>Your IP address:</h1>\r\n" + r.Header.Get("SrcIP") +":" + r.Header.Get("SrcPort") + "\r\n"))
}

func main() {
  l, err := net.Listen("tcp", ":8000")
  if err != nil {
    log.Fatal(err)
  }
  defer l.Close()
  http.HandleFunc("/", handleDefault)
  go http.ListenAndServe("127.0.0.1:8081", nil)
  for {
    conn, err := l.Accept()
    if err != nil {
      log.Fatal(err)
    }
    go func(c net.Conn) {
      defer c.Close()
      b:= bufio.NewReader(c)
      fmt.Println("Received packet")
      proxyStr, err := b.ReadString('\n')
      connInfo := parseProxy(proxyStr)
      fmt.Println(connInfo)
      req , err := http.ReadRequest(b)
      req.Header.Add("SrcIP", connInfo.SrcIP.String())
      req.Header.Add("SrcPort", strconv.Itoa(connInfo.SrcPort))
      if err != nil {
        log.Fatal(err)
      }
      fmt.Println(req)

      be, err := net.Dial("tcp", "127.0.0.1:8081") 
      beReader := bufio.NewReader(be)
      req.Write(be)
      resp, err := http.ReadResponse(beReader, req)
      resp.Write(c)
    }(conn)
  }
}

