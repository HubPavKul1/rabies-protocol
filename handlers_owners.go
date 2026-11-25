package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func handleOwners(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, _ := db.Query("SELECT id, full_name, doc_info, address, phone FROM owners ORDER BY full_name")
		defer rows.Close()
		var owners []map[string]string
		for rows.Next() {
			var id, name, doc, addr, phone string
			rows.Scan(&id, &name, &doc, &addr, &phone)
			owners = append(owners, map[string]string{"id": id, "full_name": name, "doc_info": doc, "address": addr, "phone": phone})
		}
		json.NewEncoder(w).Encode(owners)
	} else if r.Method == http.MethodPost {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		id := uuid.New().String()
		db.Exec("INSERT INTO owners (id, full_name, doc_info, address, phone) VALUES (?, ?, ?, ?, ?)", id, req["full_name"], req["doc_info"], req["address"], req["phone"])
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	} else if r.Method == http.MethodPut {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		db.Exec("UPDATE owners SET full_name = ?, doc_info = ?, address = ?, phone = ? WHERE id = ?", req["full_name"], req["doc_info"], req["address"], req["phone"], req["id"])
		w.WriteHeader(http.StatusOK)
	} else if r.Method == http.MethodDelete {
		id := strings.TrimPrefix(r.URL.Path, "/api/owners/")
		if id == "" {
			http.Error(w, "ID не указан", http.StatusBadRequest)
			return
		}
		db.Exec("DELETE FROM owners WHERE id = ?", id)
		w.WriteHeader(http.StatusOK)
	}
}