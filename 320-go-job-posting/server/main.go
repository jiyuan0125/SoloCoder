package main

import (
	"log"
	"net/http"

	"jobposting/server/handler"
	"jobposting/server/store"
)

func main() {
	s := store.NewStore("data.json")
	h := handler.NewHandler(s)

	http.HandleFunc("/company/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateJob(w, r)
		} else {
			h.ListJobsForCompany(w, r)
		}
	})

	http.HandleFunc("/company/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.OfflineJob(w, r)
		} else {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/company/applications", h.ListApplications)

	http.HandleFunc("/company/applications/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.UpdateApplicationStatus(w, r)
		} else {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/jobs", h.ListJobsForJobseeker)

	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.ApplyJob(w, r)
		} else {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/my/applications", h.ListMyApplications)

	log.Println("服务端启动，监听端口 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
