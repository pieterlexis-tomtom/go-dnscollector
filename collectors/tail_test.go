package collectors

import (
	"bufio"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dmachard/go-dnscollector/dnsutils"
	"github.com/dmachard/go-dnscollector/loggers"
	"github.com/dmachard/go-logger"
)

func TestTailRun(t *testing.T) {
	// create a temp file
	tmpFile, err := os.CreateTemp("", "temp_tailffile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // clean up

	// config
	config := dnsutils.GetFakeConfig()
	config.Collectors.Tail.TimeLayout = "2006-01-02T15:04:05.999999999Z07:00"
	config.Collectors.Tail.FilePath = tmpFile.Name()
	config.Collectors.Tail.PatternQuery = "^(?P<timestamp>[^ ]*) (?P<identity>[^ ]*) (?P<qr>.*_QUERY) (?P<rcode>[^ ]*) (?P<queryip>[^ ]*) (?P<queryport>[^ ]*) (?P<family>[^ ]*) (?P<protocol>[^ ]*) (?P<length>[^ ]*)b (?P<domain>[^ ]*) (?P<qtype>[^ ]*) (?P<latency>[^ ]*)$"

	// init collector
	g := loggers.NewFakeLogger()
	c := NewTail([]dnsutils.Worker{g}, config, logger.New(false))
	if err := c.Follow(); err != nil {
		log.Fatal("collector tail following error: ", err)
	}
	go c.Run()

	// write fake log
	time.Sleep(5 * time.Second)
	w := bufio.NewWriter(tmpFile)
	linesToWrite := "2021-08-27T07:18:35.775473Z dnscollector CLIENT_QUERY NOERROR 192.168.1.5 45660 INET INET 43b www.google.org A 0.00000"
	if _, err := w.WriteString(linesToWrite + "\n"); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}
	w.Flush()

	// waiting message in channel
	msg := <-g.Channel()
	if msg.DNS.Qname != "www.google.org" {
		t.Errorf("want www.google.org, got %s", msg.DNS.Qname)
	}
}
