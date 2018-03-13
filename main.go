package main

import (
	"bufio"
	"flag"
	"fmt"
	"time"
	"sync"
	"os"
	"strings"
	"net/http"
	"crypto/tls"
	"io/ioutil"
	"encoding/json"
	"github.com/fatih/color"
	"regexp"
)

type Header struct {
	Key   string `json:"header"`
	Value string `json:"value"`
}

var (
	/* flags for program */
	threadsOpt  = flag.Int("t", 10, "Number of concurrent threads")
	outputOpt   = flag.String("o", "wat.out", "Output file")
	hostsOpt    = flag.String("i", "", "Newline separated hosts file")
	foundOpt    = flag.String("f", " ", "Output 'found' marking")
	missingOpt  = flag.String("m", "X", "Output 'missing' marking")
	timeoutOpt  = flag.Uint("r", 3, "Timeout for connections")
	headersOpt  = flag.String("l", "headers.json", "File containing headers")
	caseSensOpt = flag.Bool("case-sensitive", false, "Case-sensitive string matching")

	/* headers array */
	headers []Header

	/* channel for printing */
	printChan = make(chan string, 10)
	/* channel for writing to output */
	outputChan = make(chan string, 10)
)

func checkHost(host string, tcpclient http.Client) {

	resp, err := tcpclient.Head("http://" + host)
	if err != nil {
		printChan <- color.YellowString("\n[!] %s", err)
		return
	}

	/* create the array for results */
	findings := make([]string, len(headers)+1)
	findings[0] = host
	var found = false

	printLog := make([]string, len(headers)+1)
	printLog[0] = color.BlueString("\n%s", host)

	/* hacky nested loop */
	for index, header := range headers {
		found = false
		var retHead, retVal string
		for key, value := range resp.Header {
			curHeader := fmt.Sprintf("(?i)%s", header.Key)
			curValue := fmt.Sprintf("(?i)%s", header.Value)
			if *caseSensOpt {
				curHeader = header.Key
				curValue = header.Value
			}
			matchHead, _ := regexp.MatchString(curHeader, key)
			matchVal, _ := regexp.MatchString(curValue, value[0])
			if matchHead && matchVal {
				found = true
				retHead = key
				retVal = value[0]
				break
			}
		}
		if !found {
			findings[index+1] = *missingOpt
			printLog[index+1] = color.RedString("[-] %s: %s", header.Key, header.Value)
		} else {
			findings[index+1] = *foundOpt
			printLog[index+1] = color.GreenString("[+] %s: %s", retHead, retVal)
		}
	}
	outputChan <- strings.Join(findings, ",")
	printChan <- strings.Join(printLog, "\n")
}

/* launch worker threads for processing */
func launchWorker(channel chan string, workerGroup *sync.WaitGroup) {
	tcpclient := http.Client{
		Timeout: time.Duration(time.Duration(*timeoutOpt) * time.Second),
	}
	for host := range channel {
		checkHost(host, tcpclient)
	}
	workerGroup.Done()
}

/* handle output printing */
func handlePrint(printerGroup *sync.WaitGroup) {
	for result := range printChan {
		fmt.Println(result)
	}
	printerGroup.Done()
}

/* handle file output writing */
func handleFileWrite(printerGroup *sync.WaitGroup, file *os.File) {
	for result := range outputChan {
		file.WriteString(fmt.Sprintf("%s\n", result))
	}
	printerGroup.Done()
}

/* store all hosts as we read into the input channel */
func bufferHosts(scanner *bufio.Scanner, channel chan string) {
	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if len(host) > 0 {
			channel <- host
		}
	}
	/* won't be adding any more hosts */
	close(channel)
}

func main() {

	/* get current time and ensure flags are in order.. */
	start := time.Now()
	flag.Parse()

	/* ignore issues with certificates */
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	/* handle required flags */
	if *hostsOpt == "" {
		fmt.Println("[!] Please supply an input file\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	/* open output file, create if needed */
	outputHndl, err := os.Create(*outputOpt)
	if err != nil {
		fmt.Printf("[!] Failed to open output file: %s\n", *outputOpt)
		flag.PrintDefaults()
		os.Exit(1)
	}

	/* handle custom header file */
	if *headersOpt != "" {
		content, err := ioutil.ReadFile(*headersOpt)
		if err != nil {
			fmt.Printf("[!] Failed to open headers file: %s\n", *headersOpt)
			flag.PrintDefaults()
			os.Exit(1)
		}
		err = json.Unmarshal(content, &headers)
		if err != nil {
			panic(err)
		}
	}

	/* input channel for hosts */
	inputChan := make(chan string, 10)

	/* header for output */
	headerList := ""
	for _, header := range headers {
		headerList = fmt.Sprintf("%s,%s", headerList, header.Key)
	}
	outputChan <- fmt.Sprintf("Host%s", headerList)

	/* progress info */
	printChan <- "[*] Started WatHeaders"
	printChan <- fmt.Sprintf("[*] Number of Threads: %d", *threadsOpt)
	printChan <- fmt.Sprintf("[*] Hosts file: %s", *hostsOpt)
	printChan <- fmt.Sprintf("[*] Output file: %s", *outputOpt)
	printChan <- fmt.Sprintf("[*] TCP Timeout: %d sec", *timeoutOpt)

	/* create wait-group for the input threads */
	workerGroup := new(sync.WaitGroup)
	workerGroup.Add(*threadsOpt)

	/* create wait-group for the printers */
	printerGroup := new(sync.WaitGroup)
	printerGroup.Add(2)

	/* open the file list */
	hostList, err := os.Open(*hostsOpt)
	if err != nil {
		fmt.Printf("[!] Failed to open input file: %s\n", *hostsOpt)
		flag.PrintDefaults()
		os.Exit(1)
	}
	defer hostList.Close()
	scannerH := bufio.NewScanner(hostList)

	/* start reading hosts from file and buffering */
	go bufferHosts(scannerH, inputChan)

	/* launch routines to process hosts */
	for i := 0; i < *threadsOpt; i++ {
		go launchWorker(inputChan, workerGroup)
	}

	/* launch routine to handle stdout printing */
	go handlePrint(printerGroup)

	/* launch routine to handle file writing */
	go handleFileWrite(printerGroup, outputHndl)

	/* wait until all threads finished */
	workerGroup.Wait()
	close(printChan)
	close(outputChan)
	printerGroup.Wait()

	/* how long did we take? */
	elapsed := time.Since(start)
	fmt.Printf("\nCompleted in %s\n", elapsed)
}
