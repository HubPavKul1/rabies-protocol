package main

import (
	"encoding/json"
	"net/http"
)

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, _ := db.Query("SELECT key, value FROM config")
		defer rows.Close()
		config := make(map[string]string)
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			config[k] = v
		}
		json.NewEncoder(w).Encode(config)
	} else if r.Method == http.MethodPost {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		for k, v := range req {
			db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", k, v)
		}
		w.WriteHeader(http.StatusOK)
	}
}