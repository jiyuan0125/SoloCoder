package main

import (
	"encoding/json"
	"log"
	"net/http"

	"skip-list/common"
	"skip-list/skiplist"

	"github.com/gorilla/mux"
)

type server struct {
	sl *skiplist.SkipList
}

func (s *server) putHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var req common.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sl.Put(key, []byte(req.Value))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"key":    key,
	})
}

func (s *server) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, found := s.sl.Get(key)

	w.Header().Set("Content-Type", "application/json")
	resp := common.GetResponse{
		Found: found,
	}
	if found {
		resp.Key = key
		resp.Value = string(value)
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	deleted := s.sl.Delete(key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.DeleteResponse{
		Deleted: deleted,
	})
}

func (s *server) rangeHandler(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	result := s.sl.Range(from, to)

	kvPairs := make([]*common.KVPair, len(result))
	for i, kv := range result {
		kvPairs[i] = &common.KVPair{
			Key:   kv.Key,
			Value: string(kv.Value),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.RangeResponse{
		Data: kvPairs,
	})
}

func (s *server) rankHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	rank, found := s.sl.Rank(key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.RankResponse{
		Found: found,
		Rank:  rank,
	})
}

func main() {
	s := &server{
		sl: skiplist.New(),
	}

	r := mux.NewRouter()

	api := r.PathPrefix("/skiplist").Subrouter()

	api.HandleFunc("/range", s.rangeHandler).Methods("GET")
	api.HandleFunc("/rank/{key}", s.rankHandler).Methods("GET")
	api.HandleFunc("/{key}", s.putHandler).Methods("PUT")
	api.HandleFunc("/{key}", s.getHandler).Methods("GET")
	api.HandleFunc("/{key}", s.deleteHandler).Methods("DELETE")

	log.Printf("Server starting on :8103")
	if err := http.ListenAndServe(":8103", r); err != nil {
		log.Fatal(err)
	}
}
