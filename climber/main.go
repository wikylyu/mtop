package main

import (
	"crypto/tls"
	"log"

	"github.com/wikylyu/mtop/tunnel/protocol"
)

func main() {
	log.SetFlags(log.Lshortfile)

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", "127.0.0.1:4433", conf)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	reqMessage := protocol.NewMTopAuthenticationMessage(protocol.MTopVersion1, protocol.MTopMethodConnect, "hello", "world", protocol.NewMTopAddress(protocol.MTopAddressTypeDomain, nil, "www.baidu.com", 443))

	n, err := conn.Write(reqMessage.Bytes())
	if err != nil {
		log.Println(n, err)
		return
	}

}
