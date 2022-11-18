package downright

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func SlowHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = io.WriteString(w, "only accept GET request")
		}

		log.Printf("new request from %s \n", req.RemoteAddr)

		var err error
		_, err = fmt.Fprintf(w, "Request processed")
		if err != nil {
			log.Printf("error while writing response: %s\n", err)
		} else {
			log.Printf("responded successfully.\n")
		}
	})
}
