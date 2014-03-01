package main

import (
	"flag"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
)

func addACAOHeader(value string, h http.Handler) http.Handler {
	if value == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", value)
		h.ServeHTTP(w, r)
	})
}

func main() {

	var (
		acao   = flag.String("acao", "*", "Access-Control-Allow-Origin")
		addr   = flag.String("addr", "127.0.0.1:8080", "addr")
		root   = flag.String("root", ".", "root")
		prefix = flag.String("prefix", "/", "prefix")
	)
	flag.Parse()

	http.Handle(*prefix,
		handlers.LoggingHandler(os.Stdout,
			addACAOHeader(*acao,
				http.StripPrefix(*prefix,
					http.FileServer(http.Dir(*root))))))

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}

}
