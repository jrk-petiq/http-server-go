package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	apiCfg := &apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.HandlerFunc(fileserverHandler)))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

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
	// metricsCount := strconv.Itoa(int(cfg.fileserverHits.Load()))
	metricsCount := int(cfg.fileserverHits.Load())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	html := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
		`, metricsCount)
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Metrics reset"))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpResponse struct {
		Body string `json:"body"`
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	decoder := json.NewDecoder(r.Body)
	chirp := chirpResponse{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding chirp: %s", err)
		errMsg := map[string]string{"error": "Something went wrong"}
		j, err := json.Marshal(errMsg)
		if err != nil {
			log.Printf("Error marshalling error message: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(j)
	}
	if len(chirp.Body) > 140 {
		errMsg := map[string]string{"error": "Chirp is too long"}
		j, err := json.Marshal(errMsg)
		if err != nil {
			log.Printf("Error marshalling error message: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(j)
		return
	}
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	splitBody := strings.Split(chirp.Body, " ")
	for i, word := range splitBody {
		for _, badWord := range badWords {
			if strings.Contains(strings.ToLower(word), badWord) {
				splitBody[i] = "****"
			}
		}
	}
	replacedBody := strings.Join(splitBody, " ")

	cleanedBody := map[string]string{"cleaned_body": replacedBody}
	j, err := json.Marshal(cleanedBody)
	if err != nil {
		log.Printf("Error marshalling valid message: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(j)
}
