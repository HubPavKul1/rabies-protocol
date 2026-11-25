package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func handleVets(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, _ := db.Query("SELECT id, full_name, position FROM vets WHERE is_active = 1 ORDER BY full_name")
		defer rows.Close()
		var vets []map[string]string
		for rows.Next() {
			var id, name, pos string
			rows.Scan(&id, &name, &pos)
			vets = append(vets, map[string]string{"id": id, "full_name": name, "position": pos})
		}
		json.NewEncoder(w).Encode(vets)
	} else if r.Method == http.MethodPost {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		id := uuid.New().String()
		db.Exec("INSERT INTO vets (id, full_name, position, is_active) VALUES (?, ?, ?, 1)", id, req["full_name"], req["position"])
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	} else if r.Method == http.MethodPut {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		db.Exec("UPDATE vets SET full_name = ?, position = ? WHERE id = ?", req["full_name"], req["position"], req["id"])
		w.WriteHeader(http.StatusOK)
	} else if r.Method == http.MethodDelete {
		id := strings.TrimPrefix(r.URL.Path, "/api/vets/")
		if id == "" {
			http.Error(w, "ID не указан", http.StatusBadRequest)
			return
		}
		db.Exec("UPDATE vets SET is_active = 0 WHERE id = ?", id)
		w.WriteHeader(http.StatusOK)
	}
}