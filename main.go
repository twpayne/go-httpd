package main

import (
	"crypto/subtle"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

func addACAOHeader(value string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", value)
		h.ServeHTTP(w, r)
	})
}

func basicAuth(realm, username, password string, n int, h http.Handler) http.Handler {
	// Extend n if needed.
	if len(username) > n {
		n = len(username)
	}
	if len(password) > n {
		n = len(password)
	}

	// Prepare extended username and password.
	usernameBytes := extendStringBytes(username, n)
	passwordBytes := extendStringBytes(password, n)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		// We can safely extended u and p to n characters without revealing
		// information in timing side-channel attacks as the duration is a
		// function of the input u and p, not our secret username and password.
		uOK := uint8(subtle.ConstantTimeCompare(extendStringBytes(u, n), usernameBytes))
		pOK := uint8(subtle.ConstantTimeCompare(extendStringBytes(p, n), passwordBytes))
		// Equally, ok is a function of the input, not our secrets, so using
		// lazy boolean evaluation does not reveal anything that an attacker
		// does not already know.
		if subtle.ConstantTimeByteEq(uOK&pOK, 1) == 0 || !ok {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q", realm))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// extendStringBytes returns s extended to at least n bytes by adding null
// characters if necessary.
func extendStringBytes(s string, n int) []byte {
	m := n - len(s)
	if m <= 0 {
		return []byte(s)
	}
	return []byte(s + strings.Repeat("\x00", m))
}

func main() {
	var (
		acao     = flag.String("acao", "*", "Access-Control-Allow-Origin")
		addr     = flag.String("addr", ":8080", "addr")
		realm    = flag.String("realm", "go-httpd", "realm")
		root     = flag.String("root", ".", "root")
		password = flag.String("password", "", "password")
		prefix   = flag.String("prefix", "/", "prefix")
		username = flag.String("username", "", "username")
	)
	flag.Parse()

	h := http.StripPrefix(*prefix, http.FileServer(http.Dir(*root)))
	if *acao != "" {
		h = addACAOHeader(*acao, h)
	}
	if *realm != "" || *username != "" || *password != "" {
		h = basicAuth(*realm, *username, *password, 1024, h)
	}
	h = handlers.LoggingHandler(os.Stdout, h)

	http.Handle(*prefix, h)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
