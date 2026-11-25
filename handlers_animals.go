package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func handleAnimals(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, _ := db.Query("SELECT a.id, a.species, a.age, a.sex, a.nickname, a.inventory_no, o.full_name as owner_name FROM animals a JOIN owners o ON a.owner_id = o.id ORDER BY a.nickname")
		defer rows.Close()
		var animals []map[string]string
		for rows.Next() {
			var id, species, age, sex, nick, inv, owner string
			rows.Scan(&id, &species, &age, &sex, &nick, &inv, &owner)
			animals = append(animals, map[string]string{"id": id, "species": species, "age": age, "sex": sex, "nickname": nick, "inventory_no": inv, "owner_name": owner})
		}
		json.NewEncoder(w).Encode(animals)
	} else if r.Method == http.MethodPost {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		id := uuid.New().String()
		db.Exec("INSERT INTO animals (id, owner_id, species, age, sex, nickname, inventory_no) VALUES (?, ?, ?, ?, ?, ?, ?)", id, req["owner_id"], req["species"], req["age"], req["sex"], req["nickname"], req["inventory_no"])
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	} else if r.Method == http.MethodPut {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		db.Exec("UPDATE animals SET species = ?, age = ?, sex = ?, nickname = ?, inventory_no = ?, owner_id = ? WHERE id = ?", req["species"], req["age"], req["sex"], req["nickname"], req["inventory_no"], req["owner_id"], req["id"])
		w.WriteHeader(http.StatusOK)
	} else if r.Method == http.MethodDelete {
		id := strings.TrimPrefix(r.URL.Path, "/api/animals/")
		if id == "" {
			http.Error(w, "ID не указан", http.StatusBadRequest)
			return
		}
		db.Exec("DELETE FROM animals WHERE id = ?", id)
		w.WriteHeader(http.StatusOK)
	}
}