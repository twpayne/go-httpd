package main

import (
	"flag"
	"github.com/gorilla/handlers"
	"github.com/twpayne/gombtiles/mbtiles"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func stripExt(path string) string {
	return path[:len(path)-len(filepath.Ext(path))]
}

func main() {

	var (
		acao     = flag.String("acao", "*", "Access-Control-Allow-Origin")
		addr     = flag.String("addr", "127.0.0.1:8080", "addr")
		root     = flag.String("root", ".", "root")
		mbtiles_ = flag.String("mbtiles", "", "MBTiles")
		prefix   = flag.String("prefix", "/", "prefix")
	)
	flag.Parse()

	http.Handle(*prefix,
		handlers.LoggingHandler(os.Stdout,
			addACAOHeader(*acao,
				http.StripPrefix(*prefix,
					http.FileServer(http.Dir(*root))))))

	if *mbtiles_ != "" {
		if mbtilesBase := filepath.Base(*mbtiles_); mbtilesBase != "." {
			mbtilesPrefix := "/" + stripExt(mbtilesBase) + "/"
			ts, err := mbtiles.NewTileServer(*mbtiles_)
			if err != nil {
				log.Fatal(err)
			}
			defer ts.Close()
			http.Handle(mbtilesPrefix,
				handlers.LoggingHandler(os.Stdout,
					addACAOHeader(*acao,
						http.StripPrefix(mbtilesPrefix,
							ts))))
		}
	}

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}

}
