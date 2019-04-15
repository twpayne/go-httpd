package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func addACAOHeader(value string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", value)
		h.ServeHTTP(w, r)
	})
}

func main() {
	var (
		acao   = flag.String("acao", "*", "Access-Control-Allow-Origin")
		addr   = flag.String("addr", ":8080", "addr")
		root   = flag.String("root", ".", "root")
		prefix = flag.String("prefix", "/", "prefix")
	)
	flag.Parse()

	h := http.StripPrefix(*prefix, http.FileServer(http.Dir(*root)))
	if *acao != "" {
		h = addACAOHeader(*acao, h)
	}
	h = handlers.LoggingHandler(os.Stdout, h)

	http.Handle(*prefix, h)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
