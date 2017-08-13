package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/sevketarisu/GoDashPlayer/utils"
	"golang.org/x/net/http2"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	verbose := flag.Bool("v", false, "verbose")
	quic := flag.Bool("quic", false, "quic")
	h2 := flag.Bool("h2", false, "h2")
	flag.Parse()
	args := flag.Args()

	if *verbose {
		utils.SetLogLevel(utils.LogLevelDebug)
	} else {
		utils.SetLogLevel(utils.LogLevelInfo)
	}
	utils.SetLogTimeFormat("")

	segment_limit, _ := strconv.Atoi(args[0])
	url_base := args[1]
	//url_base = "https://caddy.quic/media/BigBuckBunny/4sec/bunny_3936261bps/BigBuckBunny_4s$SEGMENT$.m4s"
	var hclient *http.Client

	urls := make([]string, segment_limit)

	for i := 0; i < len(urls); i++ {
		urls[i] = strings.Replace(url_base, "$SEGMENT$", strconv.Itoa(i+1), -1)
	}

	if *quic {
		utils.Infof("QUIC CLIENT")
		utils.Infof(url_base)

		hclient = &http.Client{
			Transport: &h2quic.RoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
	} else if *h2 {
		utils.Infof("HTTP2 CLIENT")
		utils.Infof(url_base)

		tr := &http2.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		hclient = &http.Client{Transport: tr}

	} else {
		utils.Infof("HTTP1.1 CLIENT")
		utils.Infof(url_base)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		hclient = &http.Client{Transport: tr}

	}
	var wg sync.WaitGroup
	wg.Add(len(urls))
	startTime := GetNow()
	for _, addr := range urls {
		utils.Infof("GET %s", addr)
		go func(addr string) {
			utils.Infof("Downloading %s", addr)
			rsp, err := hclient.Get(addr)
			if err != nil {
				panic(err)
			}
			defer rsp.Body.Close()

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				panic(err)
			}
			//utils.Infof("Request Body:")
			utils.Infof("%s", body.Bytes())
			utils.Infof("Finished %s:", addr)
			wg.Done()
		}(addr)
		//	time.Sleep(100 * time.Millisecond)
	}
	wg.Wait()
	utils.Infof("TOTAL DURATION: %s", FloatToString((GetNow() - startTime)))
}

func FloatToString(input_num float64) string {

	// to convert a float number to a string, precision 2 digits
	return strconv.FormatFloat(input_num, 'f', 2, 64)
}

func GetNow() float64 {

	now := time.Now()
	secs := now.Unix()
	nanos := now.UnixNano()

	// Note that there is no `UnixMillis`, so to get the
	// milliseconds since epoch you'll need to manually
	// divide from nanoseconds.
	millis := nanos / 10000000
	str := strconv.FormatInt(secs, 10) + "." + strconv.FormatInt(millis-secs*100, 10)
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return f
}
