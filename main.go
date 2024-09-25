package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	apiCfg := &apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.HandlerFunc(fileserverHandler)))
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/metrics", apiCfg.getMetrics)
	mux.HandleFunc("/reset", apiCfg.resetMetrics)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func fileserverHandler(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/app/", http.FileServer(http.Dir("./"))).ServeHTTP(w, r)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getMetrics(w http.ResponseWriter, r *http.Request) {
	metricsCount := strconv.Itoa(int(cfg.fileserverHits.Load()))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Hits: " + metricsCount))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Metrics reset"))
}
