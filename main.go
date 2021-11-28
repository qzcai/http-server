package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func WithLogging(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &StatusRecorder{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}
		handler(recorder, r)
		log.Printf("Handling request for %s from %s, status: %d", r.URL.Path, GetIP(r), recorder.Status)
	})
}

func main() {
	log.Println("Starting http server...")

	mux := http.NewServeMux()
	mux.Handle("/", WithLogging(rootHandler))
	mux.Handle("/notfound", WithLogging(http.NotFound))
	mux.Handle("/healthz", WithLogging(healthz))

	srv := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	// initialize the server in goroutine so that
	// it won't block the graceful shutdown handling logic
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "ok\n")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// copy request header to response header
	for k, v := range r.Header {
		w.Header().Add(k, strings.Join(v, ","))
	}

	// read VERSION environment variable and add into response header
	key := "VERSION"
	w.Header().Add(key, os.Getenv(key))

	w.Header().Add("Content-Type", "application/json")
	body, _ := json.Marshal(r.Header)
	_, _ = w.Write(body)
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}

	return r.RemoteAddr
}
