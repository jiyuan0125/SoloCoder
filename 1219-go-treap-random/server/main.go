package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"

	"treap-demo/api"
	"treap-demo/treap"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port (e.g. 8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("TREAP_PORT")
	}
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	store := treap.New(nil)

	http.HandleFunc("/put", makePutHandler(store))
	http.HandleFunc("/get", makeGetHandler(store))
	http.HandleFunc("/delete", makeDeleteHandler(store))
	http.HandleFunc("/prev", makePrevHandler(store))
	http.HandleFunc("/next", makeNextHandler(store))
	http.HandleFunc("/list", makeListHandler(store))

	_ = http.ListenAndServe(port, nil)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func makePutHandler(store *treap.Treap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.PutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.PutResponse{
				Success: false,
				Error:   "invalid request body",
			})
			return
		}

		err := store.Insert(req.Key, req.Value)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, api.PutResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, api.PutResponse{
			Success: true,
		})
	}
}

func makeGetHandler(store *treap.Treap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyStr := r.URL.Query().Get("key")
		key, err := strconv.ParseInt(keyStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.GetResponse{
				Success: false,
				Error:   "invalid key parameter",
			})
			return
		}

		value, err := store.Search(key)
		if err != nil {
			if errors.Is(err, treap.ErrKeyNotFound) {
				writeJSON(w, http.StatusNotFound, api.GetResponse{
					Success: false,
					Error:   err.Error(),
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, api.GetResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, api.GetResponse{
			Success: true,
			Key:     key,
			Value:   value,
		})
	}
}

func makeDeleteHandler(store *treap.Treap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyStr := r.URL.Query().Get("key")
		key, err := strconv.ParseInt(keyStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.DeleteResponse{
				Success: false,
				Error:   "invalid key parameter",
			})
			return
		}

		err = store.Delete(key)
		if err != nil {
			if errors.Is(err, treap.ErrEmptyTree) || errors.Is(err, treap.ErrKeyNotFound) {
				writeJSON(w, http.StatusNotFound, api.DeleteResponse{
					Success: false,
					Error:   err.Error(),
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, api.DeleteResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, api.DeleteResponse{
			Success: true,
		})
	}
}

func makePrevHandler(store *treap.Treap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyStr := r.URL.Query().Get("key")
		key, err := strconv.ParseInt(keyStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.PrevNextResponse{
				Success: false,
				Error:   "invalid key parameter",
			})
			return
		}

		node, err := store.Predecessor(key)
		if err != nil {
			if errors.Is(err, treap.ErrEmptyTree) {
				writeJSON(w, http.StatusNotFound, api.PrevNextResponse{
					Success: false,
					Exists:  false,
					Error:   err.Error(),
				})
				return
			}
			if errors.Is(err, treap.ErrNoPrev) {
				writeJSON(w, http.StatusOK, api.PrevNextResponse{
					Success: true,
					Exists:  false,
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, api.PrevNextResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, api.PrevNextResponse{
			Success: true,
			Key:     node.Key,
			Value:   node.Value,
			Exists:  true,
		})
	}
}

func makeNextHandler(store *treap.Treap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyStr := r.URL.Query().Get("key")
		key, err := strconv.ParseInt(keyStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.PrevNextResponse{
				Success: false,
				Error:   "invalid key parameter",
			})
			return
		}

		node, err := store.Successor(key)
		if err != nil {
			if errors.Is(err, treap.ErrEmptyTree) {
				writeJSON(w, http.StatusNotFound, api.PrevNextResponse{
					Success: false,
					Exists:  false,
					Error:   err.Error(),
				})
				return
			}
			if errors.Is(err, treap.ErrNoNext) {
				writeJSON(w, http.StatusOK, api.PrevNextResponse{
					Success: true,
					Exists:  false,
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, api.PrevNextResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, api.PrevNextResponse{
			Success: true,
			Key:     node.Key,
			Value:   node.Value,
			Exists:  true,
		})
	}
}

func makeListHandler(store *treap.Treap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodes := store.InOrder()
		items := make([]api.KeyValuePair, 0, len(nodes))
		for _, n := range nodes {
			items = append(items, api.KeyValuePair{Key: n.Key, Value: n.Value})
		}

		writeJSON(w, http.StatusOK, api.ListResponse{
			Success: true,
			Items:   items,
		})
	}
}
