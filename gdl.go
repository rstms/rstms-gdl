package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const Version = "0.1.5"

func main() {
	var ca, cert, key, outputFilename string
	var version, verbose bool

	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	flag.Usage = func() {
		log.SetFlags(0)
		log.SetPrefix("")
		log.Printf("\ngdl v%s\n", Version)
		log.Println(`
Issue TLS GET request using designated CA and client certificate, 
writing response data to OUTPUT_FILE

Requires PEM-formatted CA, client certificate, client_key files, 
specified with flags or environment variables.

If not provided, OUTPUT_FILE is set from the final element of the URL.
Use - to write output to stdout.

usage: gdl [flags] URL [OUTPUT_FILE]
`)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.StringVar(&ca, "ca", os.Getenv("GDL_CA"), "certificate authority `file` [GDL_CA]")
	flag.StringVar(&cert, "cert", os.Getenv("GDL_CERT"), "client cert `file` [GDL_CERT]")
	flag.StringVar(&key, "key", os.Getenv("GDL_KEY"), "client cert key `file` [GDL_KEY]")
	flag.BoolVar(&version, "version", false, "output version")
	flag.Parse()

	if version {
		fmt.Printf("gdl v%s\n", Version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		flag.Usage()
	}
	url := flag.Arg(0)
	if flag.NArg() > 1 {
		outputFilename = flag.Arg(1)
	}

	if ca == "" {
		ca = "/etc/ssl/cert.pem"
	}
	if cert == "" {
		cert = "/etc/ssl/netboot.pem"
		_, err := os.Stat(cert)
		if err != nil {
			cert = ""
		}
	}
	if key == "" {
		key = "/etc/ssl/netboot.key"
		_, err := os.Stat(key)
		if err != nil {
			key = ""
		}
	}

	if verbose {
		log.SetFlags(log.Lshortfile)
		log.Printf("url=%s", url)
		log.Printf("ca=%s", ca)
		log.Printf("cert=%s", cert)
		log.Printf("key=%s", key)
		log.Printf("outputFile=%s", outputFilename)
	}

	GetFile(url, ca, cert, key, outputFilename, verbose)
}

func GetFile(url, ca, cert, key, outputFilename string, verbose bool) {

	caCert, err := ioutil.ReadFile(ca)
	if err != nil {
		log.Fatalf("Error reading CA cert file: %v", err)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Fatal("Failed to append CA cert")
	}

	tlsConfig := tls.Config{
		RootCAs: caCertPool,
	}

	if cert != "" {
		clientCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			log.Fatalf("Error reading client cert and key: %v", err)
		}

		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	transport := &http.Transport{
		TLSClientConfig: &tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
	}

	// Use the custom client to make a GET request
	response, err := client.Get(url)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		if verbose {
			log.Printf("HTTP Status: %s\n", response.Status)
		}
	} else {
		log.Printf("HTTP Error: %s\n", response.Status)
	}

	if verbose {
		log.Printf("response=%+v\n", response)
	}

	if outputFilename == "" {
		fields := strings.Split(url, "/")
		if len(fields) < 1 {
			log.Fatalf("missing / in url")
		}
		outputFilename = fields[len(fields)-1]
	}

	oFile := os.Stdout

	if outputFilename != "-" {
		out, err := os.Create(outputFilename)
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}
		defer out.Close()
		oFile = out
	}

	byteCount, err := io.Copy(oFile, response.Body)
	if err != nil {
		log.Fatalf("Error copying response body to stdout: %v", err)
	}

	if verbose {
		log.Printf("%v bytes written", byteCount)
	}
}
