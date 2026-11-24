package main

import (
	"bytes"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// shortenName преобразует "Иванов Иван Владимирович" в "Иванов И. В."
func shortenName(fullName string) string {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return ""
	}

	// strings.Fields разбивает строку по любым пробельным символам
	parts := strings.Fields(fullName)

	if len(parts) >= 3 {
		// Преобразуем в []rune, чтобы корректно получить первый СИМВОЛ, а не байт
		firstName := []rune(parts[1])
		patronymic := []rune(parts[2])
		
		return fmt.Sprintf("%s %c. %c.", parts[0], firstName[0], patronymic[0])
	} else if len(parts) == 2 {
		firstName := []rune(parts[1])
		return fmt.Sprintf("%s %c.", parts[0], firstName[0])
	}

	// Если указано только одно слово (например, просто "Иванов")
	return parts[0]
}

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
			VetID            string          `json:"vet_id"`
			OwnerID          string          `json:"owner_id"`
			AnimalID         string          `json:"animal_id"`
			DecisionType     string          `json:"decision_type"`
			DecisionReason   string          `json:"decision_reason"`
			Survey           json.RawMessage `json:"survey"`
			DiseaseToExclude string          `json:"disease_to_exclude"`
			MaterialName     string          `json:"material_name"`
			MaterialQuantity string          `json:"material_quantity"`
			IllnessDate      string          `json:"illness_date"`
			DeathDate        string          `json:"death_date"`
			ClinicalPicture  string          `json:"clinical_picture"`
			PathoData        string          `json:"patho_data"`
			PresumptiveDiag  string          `json:"presumptive_diagnosis"`
			SamplingPlace    string          `json:"sampling_place"`
			SamplingDatetime string          `json:"sampling_datetime"`
			ActNo            string          `json:"act_no"`
			SendDate         string          `json:"send_date"`
			PackagingType    string          `json:"packaging_type"`
			AccreditationScope string        `json:"accreditation_scope"`
			MethodChoice     string          `json:"method_choice"`
			UncertaintyInfo  string          `json:"uncertainty_info"`
			ResultDelivery   string          `json:"result_delivery"`
			ResultEmail      string          `json:"result_email"`
			ResultCopies     string          `json:"result_copies"`
			OwnerRepPosition string          `json:"owner_rep_position"`
			SampleReceiverPos string         `json:"sample_receiver_position"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Ошибка чтения данных: "+err.Error(), http.StatusBadRequest)
			return
		}

		id := uuid.New().String()
		surveyStr := "{}"
		if len(req.Survey) > 0 {
			surveyStr = string(req.Survey)
		}

		_, err := db.Exec(`INSERT INTO protocols (id, created_at, vet_id, owner_id, animal_id, decision_type, decision_reason, survey_json,
			disease_to_exclude, material_name, material_quantity, illness_date, death_date, clinical_picture, patho_data, presumptive_diagnosis,
			sampling_place, sampling_datetime, act_no, send_date, packaging_type,
			accreditation_scope, method_choice, uncertainty_info, result_delivery, result_email, result_copies,
			owner_rep_position, sample_receiver_position, status) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, time.Now().Format(time.RFC3339), req.VetID, req.OwnerID, req.AnimalID,
			req.DecisionType, req.DecisionReason, surveyStr,
			req.DiseaseToExclude, req.MaterialName, req.MaterialQuantity, req.IllnessDate, req.DeathDate,
			req.ClinicalPicture, req.PathoData, req.PresumptiveDiag,
			req.SamplingPlace, req.SamplingDatetime, req.ActNo, req.SendDate, req.PackagingType,
			req.AccreditationScope, req.MethodChoice, req.UncertaintyInfo,
			req.ResultDelivery, req.ResultEmail, req.ResultCopies,
			req.OwnerRepPosition, req.SampleReceiverPos, "completed")

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
				   p.decision_type, p.decision_reason, p.survey_json,
				   p.disease_to_exclude, p.material_name, p.material_quantity, 
				   p.illness_date, p.death_date, p.clinical_picture, p.patho_data, p.presumptive_diagnosis,
				   p.sampling_place, p.sampling_datetime, p.act_no, p.send_date, p.packaging_type,
				   p.accreditation_scope, p.method_choice, p.uncertainty_info,
				   p.result_delivery, p.result_email, p.result_copies,
				   p.owner_rep_position, p.sample_receiver_position
			FROM protocols p
			JOIN vets v ON p.vet_id = v.id
			JOIN owners o ON p.owner_id = o.id
			JOIN animals a ON p.animal_id = a.id
			WHERE p.id = ?`, id)

		var createdAt, vetName, vetPos, ownerName, ownerDoc, ownerAddr, ownerPhone string
		var species, age, sex, nick, invNo string
		var decType, decReason, surveyJSON string
		var disease, materialName, materialQty, illnessDate, deathDate, clinical, patho, presumptiveDiag string
		var samplePlace, sampleTime, actNo, sendDate, packType string
		var accScope, methodChoice, uncertaintyInfo string
		var resultDelivery, resultEmail, resultCopies string
		var ownerRepPos, sampleReceiverPos string

		err := row.Scan(&createdAt, &vetName, &vetPos, &ownerName, &ownerDoc, &ownerAddr, &ownerPhone,
			&species, &age, &sex, &nick, &invNo,
			&decType, &decReason, &surveyJSON,
			&disease, &materialName, &materialQty, &illnessDate, &deathDate,
			&clinical, &patho, &presumptiveDiag,
			&samplePlace, &sampleTime, &actNo, &sendDate, &packType,
			&accScope, &methodChoice, &uncertaintyInfo,
			&resultDelivery, &resultEmail, &resultCopies,
			&ownerRepPos, &sampleReceiverPos)
		if err != nil {
			http.Error(w, "Not found: "+err.Error(), http.StatusNotFound)
			return
		}

		rows, _ := db.Query("SELECT key, value FROM config")
		defer rows.Close()
		cfg := make(map[string]string)
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			cfg[k] = v
		}

		var survey interface{}
		json.Unmarshal([]byte(surveyJSON), &survey)

		response := map[string]interface{}{
			"clinic_name": cfg["clinic_name"], "clinic_address": cfg["clinic_address"],
			"clinic_phone": cfg["clinic_phone"], "clinic_email": cfg["clinic_email"],
			"lab_name": cfg["lab_name"], "lab_address": cfg["lab_address"],
			"lab_phone": cfg["lab_phone"], "lab_email": cfg["lab_email"],
			"created_at": createdAt, "vet_name": vetName, "vet_pos": vetPos,
			"owner_name": ownerName, "owner_doc": ownerDoc, "owner_addr": ownerAddr, "owner_phone": ownerPhone,
			"species": species, "age": age, "sex": sex, "nickname": nick, "inv_no": invNo,
			"decision_type": decType, "decision_reason": decReason, "survey": survey,
			"disease": disease, "material_name": materialName, "material_quantity": materialQty,
			"illness_date": illnessDate, "death_date": deathDate,
			"clinical": clinical, "patho": patho, "presumptive_diagnosis": presumptiveDiag,
			"sample_place": samplePlace, "sample_time": sampleTime,
			"act_no": actNo, "send_date": sendDate, "pack_type": packType,
			"accreditation_scope": accScope, "method_choice": methodChoice,
			"uncertainty_info": uncertaintyInfo,
			"result_delivery": resultDelivery, "result_email": resultEmail, "result_copies": resultCopies,
			"owner_rep_position": ownerRepPos, "sample_receiver_position": sampleReceiverPos,
		}
		json.NewEncoder(w).Encode(response)
	})

		// --- API: Скачивание Ф-20 как Word-документа ---
	mux.HandleFunc("/api/protocol-word/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/protocol-word/"):]
		row := db.QueryRow(`
			SELECT p.created_at, v.full_name, v.position, o.full_name, o.doc_info, o.address, o.phone, 
				   a.species, a.age, a.sex, a.nickname, a.inventory_no,
				   p.decision_type, p.decision_reason, p.survey_json,
				   p.disease_to_exclude, p.material_name, p.material_quantity, 
				   p.illness_date, p.death_date, p.clinical_picture, p.patho_data, p.presumptive_diagnosis,
				   p.sampling_place, p.sampling_datetime, p.act_no, p.send_date, p.packaging_type,
				   p.accreditation_scope, p.method_choice, p.uncertainty_info,
				   p.result_delivery, p.result_email, p.result_copies,
				   p.owner_rep_position, p.sample_receiver_position
			FROM protocols p
			JOIN vets v ON p.vet_id = v.id
			JOIN owners o ON p.owner_id = o.id
			JOIN animals a ON p.animal_id = a.id
			WHERE p.id = ?`, id)

		var createdAt, vetName, vetPos, ownerName, ownerDoc, ownerAddr, ownerPhone string
		var species, age, sex, nick, invNo string
		var decType, decReason, surveyJSON string
		var disease, materialName, materialQty, illnessDate, deathDate, clinical, patho, presumptiveDiag string
		var samplePlace, sampleTime, actNo, sendDate, packType string
		var accScope, methodChoice, uncertaintyInfo string
		var resultDelivery, resultEmail, resultCopies string
		var ownerRepPos, sampleReceiverPos string

		err := row.Scan(&createdAt, &vetName, &vetPos, &ownerName, &ownerDoc, &ownerAddr, &ownerPhone,
			&species, &age, &sex, &nick, &invNo,
			&decType, &decReason, &surveyJSON,
			&disease, &materialName, &materialQty, &illnessDate, &deathDate,
			&clinical, &patho, &presumptiveDiag,
			&samplePlace, &sampleTime, &actNo, &sendDate, &packType,
			&accScope, &methodChoice, &uncertaintyInfo,
			&resultDelivery, &resultEmail, &resultCopies,
			&ownerRepPos, &sampleReceiverPos)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		rows, _ := db.Query("SELECT key, value FROM config")
		defer rows.Close()
		cfg := make(map[string]string)
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			cfg[k] = v
		}

		// Форматирование дат
		formatDateRu := func(s string) string {
			if s == "" { return "____________________" }
			t, err := time.Parse("2006-01-02", s)
			if err != nil { return s }
			months := []string{"января","февраля","марта","апреля","мая","июня","июля","августа","сентября","октября","ноября","декабря"}
			return fmt.Sprintf("«%d» %s %d г.", t.Day(), months[t.Month()-1], t.Year())
		}
		formatDateTimeRu := func(s string) string {
			if s == "" { return "____________________" }
			t, err := time.Parse(time.RFC3339, s)
			if err != nil { return s }
			return t.Format("02.01.2006 15:04")
		}

		// Опции с галочками
		accMark := "☐"
		accOutsideMark := "☐"
		if accScope == "Области аккредитации" { accMark = "☑" } else { accOutsideMark = "☑" }
		methodYes := "☐"
		methodNo := "☐"
		if methodChoice == "Оставляю за лабораторией" { methodYes = "☑" } else { methodNo = "☑" }
		uncYes := "☐"
		uncNo := "☐"
		if uncertaintyInfo == "да" { uncYes = "☑" } else { uncNo = "☑" }
		resultHand := "☐"
		resultEmailMark := "☐"
		if resultDelivery == "выдать на руки" { resultHand = "☑" } else { resultEmailMark = "☑" }

		// Название лаборатории для Word (с переносами)
		labName := cfg["lab_name"]
		labAddr := cfg["lab_address"]
		labPhone := cfg["lab_phone"]
		labEmail := cfg["lab_email"]
		clinicName := cfg["clinic_name"]

		html := fmt.Sprintf(`<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:w="urn:schemas-microsoft-com:office:word" xmlns="http://www.w3.org/TR/REC-html40">
<head>
<meta charset="UTF-8">
<title>Ф-20</title>
<!--[if gte mso 9]>
<xml>
<w:WordDocument>
<w:View>Print</w:View>
<w:Zoom>100</w:Zoom>
</w:WordDocument>
</xml>
<![endif]-->
<style>
@page { size: A4; margin: 1.5cm 1.5cm 2cm 2cm; }
body { font-family: 'Times New Roman', serif; font-size: 11pt; line-height: 1.3; }
.watermark {
    position: fixed;
    top: 45%%;
    left: 50%%;
    transform: translate(-50%%, -50%%) rotate(-45deg);
    font-size: 110pt;
    color: rgba(200, 0, 0, 0.12);
    font-weight: bold;
    font-family: 'Arial Black', sans-serif;
    letter-spacing: 8px;
    white-space: nowrap;
    z-index: -1;
    pointer-events: none;
    text-transform: uppercase;
}
.header-table { width: 100%%; border-collapse: collapse; margin-bottom: 10px; }
.header-table td { vertical-align: top; padding: 5px; width: 50%%; font-size: 10pt; }
.reg-row { width: 100%%; border-collapse: collapse; margin-bottom: 8px; font-size: 9pt; }
.reg-row td { border-top: 1px solid #000; padding: 2px 4px; text-align: center; }
.title { text-align: center; font-weight: bold; font-size: 12pt; margin: 12px 0; text-transform: uppercase; }
.field { margin: 6px 0; }
.field-label { font-weight: bold; }
.field-value { border-bottom: 1px solid #000; display: inline-block; min-width: 200px; padding: 0 4px; vertical-align: bottom; }
.field-value-wide { border-bottom: 1px solid #000; display: block; min-height: 18px; padding: 2px 4px; margin-top: 2px; }
.field-value-full { border-bottom: 1px solid #000; display: block; min-height: 40px; padding: 2px 4px; margin-top: 2px; }
.signature-block { margin-top: 20px; font-size: 10pt; }
.sig-row { display: flex; justify-content: space-between; margin-top: 15px; }
.sig-cell { width: 32%%; }
.sig-line { border-bottom: 1px solid #000; display: inline-block; min-width: 150px; }
.sig-hint { font-size: 8pt; color: #555; text-align: center; margin-top: 2px; }
.options { margin: 10px 0; font-size: 10pt; line-height: 1.6; }
.disclaimer { font-size: 8pt; line-height: 1.25; text-align: justify; margin-top: 15px; }
.lab-mark { border: 1px solid #000; padding: 8px; margin-top: 15px; font-size: 9pt; }
.analysis { margin-top: 12px; font-size: 10pt; }
</style>
</head>
<body>
<div class="watermark">ГОСЗАДАНИЕ</div>

<!-- ШАПКА: слева лаборатория, справа заказчик (СББЖ) -->
<table class="header-table">
<tr>
<td style="width:50%%;">
%s<br>
%s<br>
%s<br>
mail: %s
</td>
<td style="width:50%%;">
<b>Заказчик:</b><br>
%s<br>
%s<br>
%s<br>
Тел.: %s
</td>
</tr>
</table>

<!-- РЕГИСТРАЦИОННАЯ СТРОКА -->
<table class="reg-row">
<tr>
<td>Регистрационный номер заявки</td>
<td>Шифр образца</td>
<td>Дата поступления</td>
<td>Время поступления</td>
</tr>
</table>

<div class="title">СОПРОВОДИТЕЛЬНЫЙ ДОКУМЕНТ-ЗАЯВКА<br>К ПАТОЛОГИЧЕСКОМУ (БИОЛОГИЧЕСКОМУ) МАТЕРИАЛУ</div>

<!-- ТЕЛО ДОКУМЕНТА В ВИДЕ СТРОК С ПОДЧЕРКИВАНИЕМ -->
<div class="field">
<span class="field-label">Направляется для исключения</span> <span class="field-value">%s</span> <span style="font-size:9pt;">(наименования заболевания)</span>
</div>

<div class="field">
<span class="field-label">Патологический (биологический) материал</span> <span style="font-size:9pt;">(указать наименование материала, количество материала):</span>
<div class="field-value-wide">%s, количество: %s</div>
</div>

<div class="field">
<span class="field-label">от</span> <span style="font-size:9pt;">(вид и возраст животного)</span>
<div class="field-value-wide">%s, возраст: %s</div>
</div>

<div class="field">
<span class="field-label">принадлежащего</span> <span style="font-size:9pt;">(наименование организации, хозяйства, предприятия, ФИО владельца животного)</span>
<div class="field-value-wide">%s</div>
</div>

<div class="field">
<span style="font-size:9pt;">(фактический/юридический адрес)</span>
<div class="field-value-wide">%s</div>
</div>

<div class="field">
<span class="field-label">Дата заболевания животного</span> <span class="field-value">%s</span>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<span class="field-label">Дата падежа</span> <span class="field-value">%s</span>
</div>

<div class="field">
<span class="field-label">Клиническая картина</span>
<div class="field-value-full">%s</div>
</div>

<div class="field">
<span class="field-label">Данные патологоанатомического вскрытия</span>
<div class="field-value-full">%s</div>
</div>

<div class="field">
<span class="field-label">Предположительный диагноз</span>
<div class="field-value-wide">%s</div>
</div>

<div class="field">
<span class="field-label">Место отбора</span> <span class="field-value-wide">%s</span>
</div>

<div class="field">
<span class="field-label">Дата, время отбора</span> <span class="field-value">%s</span>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<span class="field-label">№ акта</span> <span class="field-value">%s</span>
</div>

<div class="field">
<span class="field-label">Дата отправки материала</span> <span class="field-value">%s</span>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<span class="field-label">Вид упаковки</span> <span class="field-value">%s</span>
</div>

<div class="signature-block">
<b>Материал отобран, упакован и отправлен:</b><br><br>
%s &nbsp;&nbsp; <span class="sig-line"></span> &nbsp;&nbsp; %s<br>
<div class="sig-hint">(должность) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(подпись) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(ФИО)</div>
</div>

<div class="options">
<b>Испытания провести согласно:</b> &nbsp; %s Области аккредитации &nbsp;&nbsp; %s Вне области аккредитации &nbsp;&nbsp; <i>отметить нужное</i>
</div>
<div class="options">
<b>Право выбора оптимального метода и методики испытания проб(ы) оставляю за лабораторией:</b> &nbsp; %s да &nbsp;&nbsp; %s нет
</div>
<div class="options">
<b>ПРЕДОСТАВЛЕНИЕ СВЕДЕНИЙ ОБ ОЦЕНЕННОЙ НЕОПРЕДЕЛЕННОСТИ В ПРОТОКОЛЕ ЛАБОРАТОРНЫХ ИСПЫТАНИЙ:</b> &nbsp; %s да &nbsp;&nbsp; %s нет
</div>
<div class="options">
<b>Прошу оригинал/копию (ненужное зачеркнуть) результата исследования:</b><br>
&nbsp;&nbsp;&nbsp;&nbsp;%s выдать на руки<br>
&nbsp;&nbsp;&nbsp;&nbsp;%s отправить по электронной почте &nbsp; <span class="field-value">%s</span>
</div>
<div class="options">
<b>Количество необходимых экземпляров результатов исследований:</b> <span class="field-value">%s</span>
</div>

<div class="disclaimer">
<b>Заказчик уведомлен и дает свое согласие на:</b><br>
Обработку персональных данных<br>
Сроки проведения испытаний<br>
Оплату всех расходов за проведение испытаний<br>
Не возврат и списание образцов после проведения испытаний без представителя заказчика<br>
Лаборатория не несет ответственность за стадию отбора, условия доставки образца(-ов), представленного(-ых) заказчиком для исследований<br>
Лаборатория не дает оценку соответствия продукции, не выдает заключений о выданных результатах исследований.<br>
Информация, указанная в данной заявке, будет отражена в протоколе испытаний. За информацию, предоставленную заказчиком, лаборатория ответственности не несет.<br>
В случае получения результата, не соответствующего требованиям НД, будет сформирован срочный отчет по форме 4-ВЕТ-Б, согласно ПРИКАЗУ МИНСЕЛЬХОЗА РФ от 21.02.2022 № 89.<br>
В соответствии со статьей 4.1 Закона РФ «О Ветеринарии» от 14.05.1993 №4979-1, Постановлением правительства РФ от 07.11.2016 №1140 «О порядке создания, развития и эксплуатации федеральной государственной информационной системы в области ветеринарии» и порядком предоставления информации в федеральную государственную информационную систему в области ветеринарии и получения информации из нее, утвержденного Приказом Минсельхоза России от 30.06.2017 №318, испытательная лаборатория осуществляет работу с использованием компонента ФГИС «ВЕТИС»;<br>
В соответствии с Приказом Минэкономразвития РФ от 24.10.2020 г. № 704 «Об утверждении положения о составе сведений о результатах деятельности аккредитованных лиц, об изменениях состава их работников и о компетенции этих работников, об изменениях технической оснащенности, представляемых аккредитованными лицами в Федеральную службу по аккредитации, порядке и сроках представления аккредитованными лицами таких сведений в Федеральную службу по аккредитации» сведения о выданных протоколах испытаний, в том числе о Заказчике, представляются в Федеральную службу по аккредитации;
</div>

<table style="width:100%%; border-collapse:collapse; margin-top:15px; font-size:10pt;">
<tr>
<td style="border:1px solid #000; padding:6px; width:33%%;">
<b>Представитель заказчика ознакомлен и согласен</b>
</td>
<td style="border:1px solid #000; padding:6px; width:22%%;">
<div class="sig-hint">(должность)</div>
<div>%s</div>
</td>
<td style="border:1px solid #000; padding:6px; width:22%%;">
<div class="sig-hint">(подпись)</div>
</td>
<td style="border:1px solid #000; padding:6px; width:23%%;">
<div class="sig-hint">(расшифровка подписи)</div>
<div><b>%s</b></div>
</td>
</tr>
<tr>
<td style="border:1px solid #000; padding:6px;">
<b>Подпись лица, сдавшего пробы, дата</b>
</td>
<td style="border:1px solid #000; padding:6px;">
<div class="sig-hint">(подпись)</div>
</td>
<td style="border:1px solid #000; padding:6px;">
<div class="sig-hint">(расшифровка подписи)</div>
<div><b>%s</b></div>
</td>
<td style="border:1px solid #000; padding:6px;">
<div class="sig-hint">(дата)</div>
<div>%s</div>
</td>
</tr>
<tr>
<td style="border:1px solid #000; padding:6px;">
<b>Сотрудник, принявший пробы, дата</b>
</td>
<td style="border:1px solid #000; padding:6px;">
<div class="sig-hint">(подпись)</div>
</td>
<td style="border:1px solid #000; padding:6px;">
<div class="sig-hint">(расшифровка подписи)</div>
</td>
<td style="border:1px solid #000; padding:6px;">
<div class="sig-hint">(дата)</div>
</td>
</tr>
</table>

<div class="lab-mark">
<b>Отметка лаборатории:</b><br>
Доставлено проб (кол-во): ____________ &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Забраковано (кол-во): ____________<br>
Причина браковки проб: ____________________________________________________________________________
</div>

<div class="analysis">
<b>Анализ заявки произведен и согласован:</b><br>
Представитель заказчика: &nbsp; _____________________________ &nbsp; (ФИО) &nbsp; _____________________ &nbsp; (подпись)<br>
Представитель %s: &nbsp; ________________ &nbsp; (должность) &nbsp; ____________ &nbsp; (подпись) &nbsp; _______________________ &nbsp; (расшифровка подписи)
</div>

<p style="margin-top:15px; font-size:9pt; font-style:italic;">Примечание:</p>

<div style="page-break-before: always;"></div>

<h3 style="text-align:center; text-transform:uppercase;">Опись проб к сопроводительному документу-заявке</h3>
<div style="font-size:9pt; margin-bottom:10px;">
(регистрационный номер заявки) ____________ &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; (дата поступления) ____________
</div>
<table style="width:100%%; border-collapse:collapse; font-size:10pt;">
<thead>
<tr>
<th style="border:1px solid #000; padding:5px; width:7%%;">№ п/п</th>
<th style="border:1px solid #000; padding:5px; width:25%%;">Инвентарный номер/кличка</th>
<th style="border:1px solid #000; padding:5px; width:18%%;">Возраст, пол</th>
<th style="border:1px solid #000; padding:5px; width:18%%;">Шифр пробы</th>
<th style="border:1px solid #000; padding:5px; width:32%%;">Результат исследования</th>
</tr>
</thead>
<tbody>
<tr>
<td style="border:1px solid #000; padding:5px; text-align:center;">1</td>
<td style="border:1px solid #000; padding:5px;">%s</td>
<td style="border:1px solid #000; padding:5px;">%s, %s</td>
<td style="border:1px solid #000; padding:5px;"></td>
<td style="border:1px solid #000; padding:5px;"></td>
</tr>
<tr><td style="border:1px solid #000; padding:5px; text-align:center;">2</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr>
<tr><td style="border:1px solid #000; padding:5px; text-align:center;">3</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr>
<tr><td style="border:1px solid #000; padding:5px; text-align:center;">4</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr>
<tr><td style="border:1px solid #000; padding:5px; text-align:center;">5</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr>
</tbody>
</table>

<div style="margin-top:30px; display:flex; justify-content:space-between; font-size:10pt;">
<div>
<b>Заказчик (владелец):</b><br><br>
____________ &nbsp; <b>%s</b><br>
<span style="font-size:8pt;">(подпись) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(расшифровка подписи)</span>
</div>
<div style="text-align:right;">
стр. 2 из 2
</div>
</div>

</body>
</html>`,
			labName, labAddr, labPhone, labEmail,
			clinicName, ownerDoc, ownerAddr, ownerPhone,
			disease,
			materialName, materialQty,
			species, age,
			ownerName,
			ownerAddr,
			formatDateRu(illnessDate), formatDateRu(deathDate),
			clinical,
			patho,
			presumptiveDiag,
			samplePlace,
			formatDateTimeRu(sampleTime), actNo,
			formatDateRu(sendDate), packType,
			vetPos, vetName,
			accMark, accOutsideMark,
			methodYes, methodNo,
			uncYes, uncNo,
			resultHand, resultEmailMark, resultEmail,
			resultCopies,
			ownerRepPos, ownerName,
			vetName, formatDateRu(sendDate),
			labName,
			nick, age, sex,
			ownerName,
		)

		w.Header().Set("Content-Type", "application/msword")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="F20_%s.doc"`, id[:8]))
		w.Write([]byte(html))
	})

		// --- API: Генерация Ф-20 через docxtemplater (Word) ---
	mux.HandleFunc("/api/protocol-docx/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/protocol-docx/"):]
		
		// 1. Получаем данные из БД (используем ту же логику, что и в /api/protocol/)
		row := db.QueryRow(`
			SELECT p.created_at, v.full_name, v.position, o.full_name, o.doc_info, o.address, o.phone, 
				   a.species, a.age, a.sex, a.nickname, a.inventory_no,
				   p.disease_to_exclude, p.material_name, p.material_quantity, 
				   p.illness_date, p.death_date, p.clinical_picture, p.patho_data, p.presumptive_diagnosis,
				   p.sampling_place, p.sampling_datetime, p.act_no, p.send_date, p.packaging_type,
				   p.accreditation_scope, p.method_choice, p.uncertainty_info,
				   p.result_delivery, p.result_email, p.result_copies,
				   p.owner_rep_position
			FROM protocols p
			JOIN vets v ON p.vet_id = v.id
			JOIN owners o ON p.owner_id = o.id
			JOIN animals a ON p.animal_id = a.id
			WHERE p.id = ?`, id)

		var createdAt, vetName, vetPos, ownerName, ownerDoc, ownerAddr, ownerPhone string
		var species, age, sex, nick, invNo string
		var disease, materialName, materialQty, illnessDate, deathDate, clinical, patho, presumptiveDiag string
		var samplePlace, sampleTime, actNo, sendDate, packType string
		var accScope, methodChoice, uncertaintyInfo string
		var resultDelivery, resultEmail, resultCopies, ownerRepPos string

		err := row.Scan(&createdAt, &vetName, &vetPos, &ownerName, &ownerDoc, &ownerAddr, &ownerPhone,
			&species, &age, &sex, &nick, &invNo,
			&disease, &materialName, &materialQty, &illnessDate, &deathDate,
			&clinical, &patho, &presumptiveDiag,
			&samplePlace, &sampleTime, &actNo, &sendDate, &packType,
			&accScope, &methodChoice, &uncertaintyInfo,
			&resultDelivery, &resultEmail, &resultCopies, &ownerRepPos)
		if err != nil {
			http.Error(w, "Протокол не найден", http.StatusNotFound)
			return
		}

		// Вспомогательная функция для форматирования даты
		formatDateRu := func(s string) string {
			if s == "" { return "«___» ________ 20__ г." }
			t, err := time.Parse("2006-01-02", s)
			if err != nil { return s }
			months := []string{"января","февраля","марта","апреля","мая","июня","июля","августа","сентября","октября","ноября","декабря"}
			return fmt.Sprintf("«%d» %s %d г.", t.Day(), months[t.Month()-1], t.Year())
		}

		// 2. Формируем словарь данных для шаблона Word
		templateData := map[string]interface{}{
			"id": id,
			"vetOrganizationName": "БГУ Ивановской области «Ивановская городская СББЖ»", // Можно брать из config
			"customerName": ownerName,
			"customerDoc": ownerDoc,
			"customerAddress": ownerAddr,
			"customerPhone": ownerPhone,
			"diseases": disease,
			"biomaterialName": materialName,
			"biomaterialQuantity": materialQty,
			"animalSpecies": species,
			"animalAge": age,
			"animalNickname": nick,
			"illnessDate": formatDateRu(illnessDate),
			"deathDate": formatDateRu(deathDate),
			"clinicalPicture": clinical,
			"pathoData": patho,
			"presumptiveDiagnosis": presumptiveDiag,
			"samplingPlace": samplePlace,
			"samplingDate": sampleTime, // Можно отформатировать при необходимости
			"actNo": actNo,
			"materialSamplingDate": formatDateRu(sendDate),
			"packagingTypeName": packType,

			// === ЗДЕСЬ ПРИМЕНЯЕМ ФУНКЦИЮ СОКРАЩЕНИЯ ===
			"vetDoctorPosition": vetPos,
			"vetDoctorShortName": shortenName(vetName),       // "Куличков П. В."
			"ownerRepShortName":  shortenName(ownerName),     // "Петров С. В." (или используйте ownerRepPos, если там указано ФИО представителя)
			
			"resultEmail": resultEmail,
			"resultCopies": resultCopies,
			// Для цикла в описи (если понадобится в будущем)
			"animalsList": []map[string]interface{}{
				{"index": 1, "speciesAndNickname": fmt.Sprintf("%s, %s", species, nick), "ageAndGender": fmt.Sprintf("%s, %s", age, sex)},
			},
		}

		dataJSON, _ := json.Marshal(templateData)

		// 3. Вызываем Node.js скрипт
		cmd := exec.Command("node", "generate.js", "template-f20.docx")
		cmd.Stdin = bytes.NewReader(dataJSON)
		
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			// Если Node.js упал, вернем ошибку в браузер для отладки
			http.Error(w, "Ошибка генерации документа: "+stderr.String(), http.StatusInternalServerError)
			return
		}

		// 4. Отдаем сгенерированный файл браузеру
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="F20_%s.docx"`, id[:8]))
		w.Write(out.Bytes())
	})

	port := "8080"
	url := fmt.Sprintf("http://localhost:%s", port)
	fmt.Printf("🚀 Сервер запущен: %s\n", url)
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	log.Fatal(http.ListenAndServe(":"+port, mux))
}