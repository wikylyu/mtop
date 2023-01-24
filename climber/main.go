package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/wikylyu/mtop/tunnel/protocol"
)

const (
	AppName = "climber"
)

func init() {
	initLog()
}

func initLog() {
	log.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
}

func main() {
	certPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile("../keys/ca.crt")
	if err != nil {
		panic(err)
	}
	if !certPool.AppendCertsFromPEM(pem) {
		panic("failed")
	}
	conf := &tls.Config{
		// InsecureSkipVerify: true,
		RootCAs: certPool,
	}

	conn, err := tls.Dial("tcp", "127.0.0.1:4433", conf)
	if err != nil {
		log.Fatalf("connect to server error: %v", err)
	}
	defer conn.Close()

	reqMessage := protocol.NewMTopAuthenticationMessage(protocol.MTopVersion1, protocol.MTopMethodConnect, "hello", "world", protocol.NewMTopAddress(protocol.MTopAddressTypeDomain, nil, "www.baidu.com", 443))

	if _, err := conn.Write(reqMessage.Bytes()); err != nil {
		log.Fatalf("write error: %v", err)

	}

	respMessage, err := protocol.ParseMTopResponseMessage(conn)
	if err != nil {
		log.Fatalf("read error: %v", err)
	}
	log.Infof("verion: %d", respMessage.Version)
	log.Infof("Status: %d", respMessage.Status)
}
