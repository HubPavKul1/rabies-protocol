package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

//go:embed index.html
var indexHTML embed.FS

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
		decision_type TEXT, decision_reason TEXT, 
		disease_to_exclude TEXT, material_desc TEXT, illness_date TEXT, death_date TEXT,
		clinical_picture TEXT, patho_data TEXT, sampling_place TEXT, sampling_datetime TEXT, 
		act_no TEXT, send_date TEXT, packaging_type TEXT, status TEXT
	);`
	if _, err := db.Exec(schema); err != nil {
		log.Fatal("Ошибка создания таблиц:", err)
	}

	// Настройки по умолчанию
	var clinicName string
	db.QueryRow("SELECT value FROM config WHERE key='clinic_name'").Scan(&clinicName)
	if clinicName == "" {
		db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES ('clinic_name', 'ГБУ Ивановской области «Ивановская областная СББЖ»')")
	}

	var labName string
	db.QueryRow("SELECT value FROM config WHERE key='lab_name'").Scan(&labName)
	if labName == "" {
		db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES ('lab_name', 'Бюджетное государственное учреждение Ивановской области «Ивановская областная ветеринарная лаборатория»')")
		db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES ('lab_address', '153022, г. Иваново, ул. Ташкентская, д. 66')")
		db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES ('lab_phone', 'тел/факс (4932) 23-46-75, 23-47-64')")
		db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES ('lab_email', 'ivoblvetlab-info@mail.ru')")
	}
}

func main() {
	initDB()
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := indexHTML.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки интерфейса: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("/api/vets", func(w http.ResponseWriter, r *http.Request) {
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
		}
	})

	mux.HandleFunc("/api/owners", func(w http.ResponseWriter, r *http.Request) {
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
		}
	})

	mux.HandleFunc("/api/animals", func(w http.ResponseWriter, r *http.Request) {
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
		}
	})

	mux.HandleFunc("/api/protocols", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			VetID            string `json:"vet_id"`
			OwnerID          string `json:"owner_id"`
			AnimalID         string `json:"animal_id"`
			DecisionType     string `json:"decision_type"`
			DecisionReason   string `json:"decision_reason"`
			DiseaseToExclude string `json:"disease_to_exclude"`
			MaterialDesc     string `json:"material_desc"`
			IllnessDate      string `json:"illness_date"`
			DeathDate        string `json:"death_date"`
			ClinicalPicture  string `json:"clinical_picture"`
			PathoData        string `json:"patho_data"`
			SamplingPlace    string `json:"sampling_place"`
			SamplingDatetime string `json:"sampling_datetime"`
			ActNo            string `json:"act_no"`
			SendDate         string `json:"send_date"`
			PackagingType    string `json:"packaging_type"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Ошибка чтения данных: "+err.Error(), http.StatusBadRequest)
			return
		}

		id := uuid.New().String()
		_, err := db.Exec(`INSERT INTO protocols (id, created_at, vet_id, owner_id, animal_id, decision_type, decision_reason, disease_to_exclude, material_desc, illness_date, death_date, clinical_picture, patho_data, sampling_place, sampling_datetime, act_no, send_date, packaging_type, status) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, time.Now().Format(time.RFC3339), req.VetID, req.OwnerID, req.AnimalID,
			req.DecisionType, req.DecisionReason, req.DiseaseToExclude, req.MaterialDesc, req.IllnessDate, req.DeathDate,
			req.ClinicalPicture, req.PathoData, req.SamplingPlace, req.SamplingDatetime,
			req.ActNo, req.SendDate, req.PackagingType, "completed")

		if err != nil {
			http.Error(w, "Ошибка сохранения в БД: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	})

	mux.HandleFunc("/api/protocol/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/protocol/"):]
		row := db.QueryRow(`
			SELECT p.created_at, v.full_name, v.position, o.full_name, o.doc_info, o.address, o.phone, 
				   a.species, a.age, a.sex, a.nickname, a.inventory_no,
				   p.decision_type, p.decision_reason, p.disease_to_exclude, p.material_desc, p.illness_date, p.death_date,
				   p.clinical_picture, p.patho_data, p.sampling_place, p.sampling_datetime, p.act_no, p.send_date, p.packaging_type
			FROM protocols p
			JOIN vets v ON p.vet_id = v.id
			JOIN owners o ON p.owner_id = o.id
			JOIN animals a ON p.animal_id = a.id
			WHERE p.id = ?`, id)
		
		var createdAt, vetName, vetPos, ownerName, ownerDoc, ownerAddr, ownerPhone string
		var species, age, sex, nick, invNo string
		var decType, decReason, disease, material, illnessDate, deathDate, clinical, patho, samplePlace, sampleTime, actNo, sendDate, packType string
		
		err := row.Scan(&createdAt, &vetName, &vetPos, &ownerName, &ownerDoc, &ownerAddr, &ownerPhone,
			&species, &age, &sex, &nick, &invNo, &decType, &decReason, &disease, &material, &illnessDate, &deathDate,
			&clinical, &patho, &samplePlace, &sampleTime, &actNo, &sendDate, &packType)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		
		var clinicName, labName, labAddr, labPhone, labEmail string
		db.QueryRow("SELECT value FROM config WHERE key='clinic_name'").Scan(&clinicName)
		db.QueryRow("SELECT value FROM config WHERE key='lab_name'").Scan(&labName)
		db.QueryRow("SELECT value FROM config WHERE key='lab_address'").Scan(&labAddr)
		db.QueryRow("SELECT value FROM config WHERE key='lab_phone'").Scan(&labPhone)
		db.QueryRow("SELECT value FROM config WHERE key='lab_email'").Scan(&labEmail)

		response := map[string]interface{}{
			"clinic_name":    clinicName,
			"lab_name":       labName, "lab_address": labAddr, "lab_phone": labPhone, "lab_email": labEmail,
			"created_at":     createdAt, "vet_name": vetName, "vet_pos": vetPos,
			"owner_name":     ownerName, "owner_doc": ownerDoc, "owner_addr": ownerAddr, "owner_phone": ownerPhone,
			"species":        species, "age": age, "sex": sex, "nickname": nick, "inv_no": invNo,
			"decision_type":  decType, "decision_reason": decReason,
			"disease":        disease, "material": material, "illness_date": illnessDate, "death_date": deathDate,
			"clinical":       clinical, "patho": patho, "sample_place": samplePlace, "sample_time": sampleTime,
			"act_no":         actNo, "send_date": sendDate, "pack_type": packType,
		}
		json.NewEncoder(w).Encode(response)
	})

	port := "8080"
	url := fmt.Sprintf("http://localhost:%s", port)
	fmt.Printf("🚀 Сервер запущен: %s\n", url)
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	log.Fatal(http.ListenAndServe(":"+port, mux))
}