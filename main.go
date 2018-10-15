package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"

	"jrubin.io/bl/handler"
	"jrubin.io/bl/selfcert"
)

var (
	addr     = flag.String("addr", ":443", "address:port to listen for requests on")
	certFile = flag.String("cert", "", "tls certificate file")
	keyFile  = flag.String("key", "", "tls key file")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("%#v", err)
	}
}

func getCert() (tls.Certificate, error) {
	if *certFile != "" && *keyFile != "" {
		return tls.LoadX509KeyPair(*certFile, *keyFile)
	}

	log.Print("generating self signed certificate")
	return selfcert.NewCert("::1", "127.0.0.1", "localhost")
}

func listen(cert tls.Certificate) (ln net.Listener, err error) {
	ln, err = net.Listen("tcp", *addr)
	if err != nil {
		return
	}

	log.Printf("listening at %s", *addr)

	tc := tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2"},
	}

	ln = tls.NewListener(ln, &tc)

	return
}

func run() (err error) {
	cert, err := getCert()
	if err != nil {
		return
	}

	ln, err := listen(cert)
	if err != nil {
		return
	}
	defer ln.Close()

	mux := http.NewServeMux()
	mux.Handle("/v1/clicks/country", handler.Handler())

	server := http.Server{Handler: mux}
	return server.Serve(ln)
}
