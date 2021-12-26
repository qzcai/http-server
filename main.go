package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qzcai/http-server/metrics"
	"io"
	"log"
	"math/rand"
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
	metrics.Register()

	mux := http.NewServeMux()
	mux.Handle("/", WithLogging(rootHandler))
	mux.Handle("/tracing", WithLogging(tracing))
	mux.Handle("/notfound", WithLogging(http.NotFound))
	mux.Handle("/healthz", WithLogging(healthz))
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    ":8080",
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	timer := metrics.NewExecutionTimer()
	defer timer.ObserveTotal()

	delay := randInt(10, 2000)
	time.Sleep(time.Millisecond * time.Duration(delay))

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

func tracing(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", "http://service2", nil)
	if err != nil {
		fmt.Printf("%s", err)
	}
	lowerCaseHeader := make(http.Header)
	for key, value := range r.Header {
		lowerCaseHeader[strings.ToLower(key)] = value
	}
	req.Header = lowerCaseHeader
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Println("HTTP get failed with error: ", "error", err)
	} else {
		log.Println("HTTP get succeeded")
	}
	if resp != nil {
		resp.Write(w)
	}
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}

	return r.RemoteAddr
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}
