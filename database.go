package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB() {
	os.MkdirAll("data", 0755)
	var err error
	db, err = sql.Open("sqlite", "data/rabies.db")
	if err != nil {
		log.Fatal("Ошибка открытия БД:", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS config (key TEXT PRIMARY KEY, value TEXT);
	CREATE TABLE IF NOT EXISTS vets (id TEXT PRIMARY KEY, full_name TEXT, position TEXT, is_active INTEGER DEFAULT 1);
	CREATE TABLE IF NOT EXISTS owners (id TEXT PRIMARY KEY, full_name TEXT, doc_info TEXT, address TEXT, phone TEXT);
	CREATE TABLE IF NOT EXISTS animals (id TEXT PRIMARY KEY, owner_id TEXT, species TEXT, age TEXT, sex TEXT, nickname TEXT, inventory_no TEXT);
	CREATE TABLE IF NOT EXISTS protocols (
		id TEXT PRIMARY KEY, created_at TEXT, vet_id TEXT, owner_id TEXT, animal_id TEXT,
		decision_type TEXT, decision_reason TEXT, survey_json TEXT,
		disease_to_exclude TEXT, material_name TEXT, material_quantity TEXT, 
		illness_date TEXT, death_date TEXT,
		clinical_picture TEXT, patho_data TEXT, presumptive_diagnosis TEXT,
		sampling_place TEXT, sampling_datetime TEXT, act_no TEXT, 
		send_date TEXT, packaging_type TEXT,
		accreditation_scope TEXT, method_choice TEXT, uncertainty_info TEXT,
		result_delivery TEXT, result_email TEXT, result_copies TEXT,
		owner_rep_position TEXT, sample_receiver_position TEXT,
		status TEXT
	);`
	if _, err := db.Exec(schema); err != nil {
		log.Fatal("Ошибка создания таблиц:", err)
	}

	defaults := map[string]string{
		"clinic_name":    "БГУ Ивановской области «Ивановская городская СББЖ»",
		"clinic_address": "153022, г. Иваново, ул. Танкиста Белороссова, д. 30А",
		"clinic_phone":   "тел. (4932) 235257",
		"clinic_email":   "ivgorvet@mail.ru",
		"lab_name":       "Бюджетное государственное учреждение Ивановской области «Ивановская областная ветеринарная лаборатория»",
		"lab_address":    "153022, г. Иваново, ул. Ташкентская, д. 66",
		"lab_phone":      "тел/факс (4932) 23-46-75, 23-47-64",
		"lab_email":      "ivoblvetlab-info@mail.ru",
	}
	for k, v := range defaults {
		var exists string
		db.QueryRow("SELECT value FROM config WHERE key=?", k).Scan(&exists)
		if exists == "" {
			db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", k, v)
		}
	}
}