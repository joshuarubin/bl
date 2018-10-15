package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/pprof"

	"jrubin.io/bl/handler"
	"jrubin.io/bl/selfcert"
)

var (
	addr     = flag.String("addr", ":https", "address:port to listen for requests on")
	certFile = flag.String("cert", "", "tls certificate file (optional)")
	keyFile  = flag.String("key", "", "tls key file (optional)")
	workers  = flag.Int("workers", 16, "number of worker connections to maintain to the bitly api")
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

	// add the primary handler
	mux.Handle("/v1/clicks/country", handler.Handler(*workers))

	// set up the pprof endpoints
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := http.Server{Handler: mux}
	return server.Serve(ln)
}
