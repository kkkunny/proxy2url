package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

func main() {
	svr := chi.NewRouter()
	svr.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		paths := strings.Split(r.URL.Path, "/")
		paths = paths[1:]

		var proxy func(*http.Request) (*url.URL, error)
		switch proxyType := paths[0]; proxyType {
		case "socks5":
			if len(paths) < 2 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			proxyPath, err := url.Parse(fmt.Sprintf("%s://%s", proxyType, paths[1]))
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			proxy = http.ProxyURL(proxyPath)
			paths = paths[2:]
		default:
			http.Error(w, "unknown proxy type", http.StatusBadRequest)
			return
		}

		toUrlStr := fmt.Sprintf("http://%s", strings.Join(paths, "/"))
		toUrl, err := url.Parse(toUrlStr)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		r.URL, _ = url.Parse(toUrl.String())
		toUrl.Path = ""
		reverseProxy := httputil.NewSingleHostReverseProxy(toUrl)
		reverseProxy.Transport = &http.Transport{Proxy: proxy}
		reverseProxy.ServeHTTP(w, r)
	})
	if err := http.ListenAndServe(":80", svr); err != nil {
		panic(err)
	}
}
