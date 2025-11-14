package main

import (
	"fmt"
	"github.com/rstms/rstms-gdl/getfile"
	flag "github.com/spf13/pflag"
	"log"
	"os"
)

const Version = "0.0.2"

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
	flag.BoolVar(&verbose, "v", false, "verbose output")
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

	getfile.GetFile(url, ca, cert, key, outputFilename, verbose)
}
