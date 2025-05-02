package main

import (
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const CONTROL_PARAM = "custom_param" // header for control. if not present in the request, not forward to the backend
const VAL_CONTROL_PARAM = "Fk8pYlKMp5lCq2pzzZbo996sG6n139sx"

const PROXY_SERVER = "0.0.0.0:38768"
const REMOTE_SERVER = "https://192.168.242.129:27598" // the final remote url/ip where the backend is

func main() {
	targetURL, err := url.Parse(REMOTE_SERVER)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = httpClient.Transport

	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.TransferEncoding = nil
		resp.Header.Del("Server")
		resp.Header.Del("Via")
		resp.Header.Del("X-Powered-By")

		if location := resp.Header.Get("Location"); location != "" {
			newLocation := strings.Replace(location, REMOTE_SERVER, PROXY_SERVER, -1)
			resp.Header.Set("Location", newLocation)
		}

		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Println("Error:", err)
		http.Error(w, "Error:", http.StatusBadGateway)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var bodyBuffer bytes.Buffer
		if r.Body != nil {
			_, err := bodyBuffer.ReadFrom(r.Body)
			if err != nil {
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBuffer.Bytes()))
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error", http.StatusBadRequest)
			return
		}

		paramValue := r.FormValue(CONTROL_PARAM)

		if paramValue != VAL_CONTROL_PARAM {
			http.Error(w, "Error", http.StatusForbidden)
			return
		}
		r.Header.Del("Transfer-Encoding")
		r.Header.Del("Content-Length")
		r.Header.Del("X-Forwarded-For")

		if bodyBuffer.Len() > 0 {
			r.Body = io.NopCloser(bytes.NewReader(bodyBuffer.Bytes()))
			r.ContentLength = int64(bodyBuffer.Len())
		}

		const maxRequestSize = 10 * 1024 * 1024 // 10MB
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

		r.Host = targetURL.Host

		proxy.ServeHTTP(w, r)
	}

	server := &http.Server{
		Addr:    PROXY_SERVER,
		Handler: http.HandlerFunc(handler),
		TLSConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	}

	log.Println("HTTPS proxy running in" + PROXY_SERVER)

	err = server.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		log.Fatalf("Init error: %v", err)
	}
}
