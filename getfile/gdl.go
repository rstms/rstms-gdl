package getfile

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const Version = "1.0.1"

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
