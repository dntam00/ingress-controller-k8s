package test_web_server

import (
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"time"
)

func main() {
	headersOk := handlers.AllowedHeaders([]string{
		"sec-ch-ua", "x-owner-content", "sec-ch-ua-mobile", "User-Agent",
		"Accept", "Referer", "device", "sec-ch-ua-platform",
	})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	r := mux.NewRouter()
	r.HandleFunc("/v2/endpoint", func(writer http.ResponseWriter, request *http.Request) {
		hostname, err := os.Hostname()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			errMsg := fmt.Sprintf("Error getting hostname: %v", err)
			http.Error(writer, errMsg, http.StatusInternalServerError)
			return
		}
		_, err = writer.Write([]byte(fmt.Sprintf("hostname: %s\n", hostname)))
		if err != nil {
			fmt.Println("write error: ", err)
		}
	}).Methods("GET")

	srv := &http.Server{
		Addr:        ":" + "7888",
		Handler:     handlers.CORS(originsOk, headersOk, methodsOk)(r),
		IdleTimeout: 10 * time.Second,
	}

	fmt.Println("start server")

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("failed to start server")
		return
	}
}
