package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

// --- Создание протокола ---
func handleCreateProtocol(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		VetID, OwnerID, AnimalID, DecisionType, DecisionReason string
		Survey           json.RawMessage
		DiseaseToExclude, MaterialName, MaterialQuantity       string
		IllnessDate, DeathDate, ClinicalPicture, PathoData     string
		PresumptiveDiag, SamplingPlace, SamplingDatetime       string
		ActNo, SendDate, PackagingType                         string
		AccreditationScope, MethodChoice, UncertaintyInfo      string
		ResultDelivery, ResultEmail, ResultCopies              string
		OwnerRepPosition, SampleReceiverPos                    string
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
		sampling_place, sampling_datetime, act_no, send_date, packaging_type, accreditation_scope, method_choice, uncertainty_info, 
		result_delivery, result_email, result_copies, owner_rep_position, sample_receiver_position, status) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, time.Now().Format(time.RFC3339), req.VetID, req.OwnerID, req.AnimalID, req.DecisionType, req.DecisionReason, surveyStr,
		req.DiseaseToExclude, req.MaterialName, req.MaterialQuantity, req.IllnessDate, req.DeathDate, req.ClinicalPicture, req.PathoData, req.PresumptiveDiag,
		req.SamplingPlace, req.SamplingDatetime, req.ActNo, req.SendDate, req.PackagingType, req.AccreditationScope, req.MethodChoice, req.UncertaintyInfo,
		req.ResultDelivery, req.ResultEmail, req.ResultCopies, req.OwnerRepPosition, req.SampleReceiverPos, "completed")

	if err != nil {
		http.Error(w, "Ошибка сохранения в БД: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// --- Получение протокола для печати (HTML) ---
func handleGetProtocol(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/protocol/"):]
	row := db.QueryRow(`SELECT p.created_at, v.full_name, v.position, o.full_name, o.doc_info, o.address, o.phone, 
		a.species, a.age, a.sex, a.nickname, a.inventory_no, p.decision_type, p.decision_reason, p.survey_json,
		p.disease_to_exclude, p.material_name, p.material_quantity, p.illness_date, p.death_date, p.clinical_picture, 
		p.patho_data, p.presumptive_diagnosis, p.sampling_place, p.sampling_datetime, p.act_no, p.send_date, p.packaging_type,
		p.accreditation_scope, p.method_choice, p.uncertainty_info, p.result_delivery, p.result_email, p.result_copies,
		p.owner_rep_position, p.sample_receiver_position
		FROM protocols p JOIN vets v ON p.vet_id = v.id JOIN owners o ON p.owner_id = o.id JOIN animals a ON p.animal_id = a.id WHERE p.id = ?`, id)

	var createdAt, vetName, vetPos, ownerName, ownerDoc, ownerAddr, ownerPhone string
	var species, age, sex, nick, invNo string
	var decType, decReason, surveyJSON string
	var disease, materialName, materialQty, illnessDate, deathDate, clinical, patho, presumptiveDiag string
	var samplePlace, sampleTime, actNo, sendDate, packType string
	var accScope, methodChoice, uncertaintyInfo, resultDelivery, resultEmail, resultCopies, ownerRepPos, sampleReceiverPos string

	err := row.Scan(&createdAt, &vetName, &vetPos, &ownerName, &ownerDoc, &ownerAddr, &ownerPhone,
		&species, &age, &sex, &nick, &invNo, &decType, &decReason, &surveyJSON,
		&disease, &materialName, &materialQty, &illnessDate, &deathDate, &clinical, &patho, &presumptiveDiag,
		&samplePlace, &sampleTime, &actNo, &sendDate, &packType, &accScope, &methodChoice, &uncertaintyInfo,
		&resultDelivery, &resultEmail, &resultCopies, &ownerRepPos, &sampleReceiverPos)
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"clinic_name": cfg["clinic_name"], "clinic_address": cfg["clinic_address"], "clinic_phone": cfg["clinic_phone"], "clinic_email": cfg["clinic_email"],
		"lab_name": cfg["lab_name"], "lab_address": cfg["lab_address"], "lab_phone": cfg["lab_phone"], "lab_email": cfg["lab_email"],
		"created_at": createdAt, "vet_name": vetName, "vet_pos": vetPos, "owner_name": ownerName, "owner_doc": ownerDoc, "owner_addr": ownerAddr, "owner_phone": ownerPhone,
		"species": species, "age": age, "sex": sex, "nickname": nick, "inv_no": invNo, "decision_type": decType, "decision_reason": decReason, "survey": survey,
		"disease": disease, "material_name": materialName, "material_quantity": materialQty, "illness_date": illnessDate, "death_date": deathDate,
		"clinical": clinical, "patho": patho, "presumptive_diagnosis": presumptiveDiag, "sample_place": samplePlace, "sample_time": sampleTime,
		"act_no": actNo, "send_date": sendDate, "pack_type": packType, "accreditation_scope": accScope, "method_choice": methodChoice,
		"uncertainty_info": uncertaintyInfo, "result_delivery": resultDelivery, "result_email": resultEmail, "result_copies": resultCopies,
		"owner_rep_position": ownerRepPos, "sample_receiver_position": sampleReceiverPos,
	})
}

// --- Генерация HTML Word ---
func handleGetProtocolWord(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/protocol-word/"):]
	// (Здесь используется та же логика запроса к БД, что и выше, для краткости я использую упрощенный вариант, 
	// но в реальном коде лучше вынести запрос в отдельную функцию getProlocolData(id))
	// Для экономии места я оставлю структуру как в оригинале, но с вызовом formatDateRu из utils.go
	
	row := db.QueryRow(`SELECT p.created_at, v.full_name, v.position, o.full_name, o.doc_info, o.address, o.phone, 
		a.species, a.age, a.sex, a.nickname, a.inventory_no, p.disease_to_exclude, p.material_name, p.material_quantity, 
		p.illness_date, p.death_date, p.clinical_picture, p.patho_data, p.presumptive_diagnosis, p.sampling_place, 
		p.sampling_datetime, p.act_no, p.send_date, p.packaging_type, p.accreditation_scope, p.method_choice, 
		p.uncertainty_info, p.result_delivery, p.result_email, p.result_copies, p.owner_rep_position, p.sample_receiver_position
		FROM protocols p JOIN vets v ON p.vet_id = v.id JOIN owners o ON p.owner_id = o.id JOIN animals a ON p.animal_id = a.id WHERE p.id = ?`, id)

	var createdAt, vetName, vetPos, ownerName, ownerDoc, ownerAddr, ownerPhone string
	var species, age, sex, nick, invNo, disease, materialName, materialQty, illnessDate, deathDate, clinical, patho, presumptiveDiag string
	var samplePlace, sampleTime, actNo, sendDate, packType, accScope, methodChoice, uncertaintyInfo, resultDelivery, resultEmail, resultCopies, ownerRepPos, sampleReceiverPos string

	err := row.Scan(&createdAt, &vetName, &vetPos, &ownerName, &ownerDoc, &ownerAddr, &ownerPhone, &species, &age, &sex, &nick, &invNo,
		&disease, &materialName, &materialQty, &illnessDate, &deathDate, &clinical, &patho, &presumptiveDiag, &samplePlace, &sampleTime, &actNo, &sendDate, &packType,
		&accScope, &methodChoice, &uncertaintyInfo, &resultDelivery, &resultEmail, &resultCopies, &ownerRepPos, &sampleReceiverPos)
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

	accMark, accOutsideMark := "☐", "☐"
	if accScope == "Области аккредитации" {
		accMark = "☑"
	} else {
		accOutsideMark = "☑"
	}
	methodYes, methodNo := "☐", "☐"
	if methodChoice == "Оставляю за лабораторией" {
		methodYes = "☑"
	} else {
		methodNo = "☑"
	}
	uncYes, uncNo := "☐", "☐"
	if uncertaintyInfo == "да" {
		uncYes = "☑"
	} else {
		uncNo = "☑"
	}
	resultHand, resultEmailMark := "☐", "☐"
	if resultDelivery == "выдать на руки" {
		resultHand = "☑"
	} else {
		resultEmailMark = "☑"
	}

	html := fmt.Sprintf(`<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:w="urn:schemas-microsoft-com:office:word" xmlns="http://www.w3.org/TR/REC-html40">
<head><meta charset="UTF-8"><title>Ф-20</title>
<!--[if gte mso 9]><xml><w:WordDocument><w:View>Print</w:View><w:Zoom>100</w:Zoom></w:WordDocument></xml><![endif]-->
<style>
@page { size: A4; margin: 1.5cm 1.5cm 2cm 2cm; }
body { font-family: 'Times New Roman', serif; font-size: 11pt; line-height: 1.3; }
.watermark { position: fixed; top: 45%%; left: 50%%; transform: translate(-50%%, -50%%) rotate(-45deg); font-size: 110pt; color: rgba(200, 0, 0, 0.12); font-weight: bold; font-family: 'Arial Black', sans-serif; letter-spacing: 8px; white-space: nowrap; z-index: -1; pointer-events: none; text-transform: uppercase; }
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
.sig-line { border-bottom: 1px solid #000; display: inline-block; min-width: 150px; }
.sig-hint { font-size: 8pt; color: #555; text-align: center; margin-top: 2px; }
.options { margin: 10px 0; font-size: 10pt; line-height: 1.6; }
.disclaimer { font-size: 8pt; line-height: 1.25; text-align: justify; margin-top: 15px; }
.lab-mark { border: 1px solid #000; padding: 8px; margin-top: 15px; font-size: 9pt; }
.analysis { margin-top: 12px; font-size: 10pt; }
</style></head><body>
<div class="watermark">ГОСЗАДАНИЕ</div>
<table class="header-table"><tr><td style="width:50%%;">%s<br>%s<br>%s<br>mail: %s</td><td style="width:50%%;"><b>Заказчик:</b><br>%s<br>%s<br>%s<br>Тел.: %s</td></tr></table>
<table class="reg-row"><tr><td>Регистрационный номер заявки</td><td>Шифр образца</td><td>Дата поступления</td><td>Время поступления</td></tr></table>
<div class="title">СОПРОВОДИТЕЛЬНЫЙ ДОКУМЕНТ-ЗАЯВКА<br>К ПАТОЛОГИЧЕСКОМУ (БИОЛОГИЧЕСКОМУ) МАТЕРИАЛУ</div>
<div class="field"><span class="field-label">Направляется для исключения</span> <span class="field-value">%s</span> <span style="font-size:9pt;">(наименования заболевания)</span></div>
<div class="field"><span class="field-label">Патологический (биологический) материал</span> <span style="font-size:9pt;">(указать наименование материала, количество материала):</span><div class="field-value-wide">%s, количество: %s</div></div>
<div class="field"><span class="field-label">от</span> <span style="font-size:9pt;">(вид и возраст животного)</span><div class="field-value-wide">%s, возраст: %s</div></div>
<div class="field"><span class="field-label">принадлежащего</span> <span style="font-size:9pt;">(наименование организации, хозяйства, предприятия, ФИО владельца животного)</span><div class="field-value-wide">%s</div></div>
<div class="field"><span style="font-size:9pt;">(фактический/юридический адрес)</span><div class="field-value-wide">%s</div></div>
<div class="field"><span class="field-label">Дата заболевания животного</span> <span class="field-value">%s</span>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<span class="field-label">Дата падежа</span> <span class="field-value">%s</span></div>
<div class="field"><span class="field-label">Клиническая картина</span><div class="field-value-full">%s</div></div>
<div class="field"><span class="field-label">Данные патологоанатомического вскрытия</span><div class="field-value-full">%s</div></div>
<div class="field"><span class="field-label">Предположительный диагноз</span><div class="field-value-wide">%s</div></div>
<div class="field"><span class="field-label">Место отбора</span> <span class="field-value-wide">%s</span></div>
<div class="field"><span class="field-label">Дата, время отбора</span> <span class="field-value">%s</span>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<span class="field-label">№ акта</span> <span class="field-value">%s</span></div>
<div class="field"><span class="field-label">Дата отправки материала</span> <span class="field-value">%s</span>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<span class="field-label">Вид упаковки</span> <span class="field-value">%s</span></div>
<div class="signature-block"><b>Материал отобран, упакован и отправлен:</b><br><br>%s &nbsp;&nbsp; <span class="sig-line"></span> &nbsp;&nbsp; %s<br><div class="sig-hint">(должность) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(подпись) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(ФИО)</div></div>
<div class="options"><b>Испытания провести согласно:</b> &nbsp; %s Области аккредитации &nbsp;&nbsp; %s Вне области аккредитации &nbsp;&nbsp; <i>отметить нужное</i></div>
<div class="options"><b>Право выбора оптимального метода и методики испытания проб(ы) оставляю за лабораторией:</b> &nbsp; %s да &nbsp;&nbsp; %s нет</div>
<div class="options"><b>ПРЕДОСТАВЛЕНИЕ СВЕДЕНИЙ ОБ ОЦЕНЕННОЙ НЕОПРЕДЕЛЕННОСТИ В ПРОТОКОЛЕ ЛАБОРАТОРНЫХ ИСПЫТАНИЙ:</b> &nbsp; %s да &nbsp;&nbsp; %s нет</div>
<div class="options"><b>Прошу оригинал/копию (ненужное зачеркнуть) результата исследования:</b><br>&nbsp;&nbsp;&nbsp;&nbsp;%s выдать на руки<br>&nbsp;&nbsp;&nbsp;&nbsp;%s отправить по электронной почте &nbsp; <span class="field-value">%s</span></div>
<div class="options"><b>Количество необходимых экземпляров результатов исследований:</b> <span class="field-value">%s</span></div>
<div class="disclaimer"><b>Заказчик уведомлен и дает свое согласие на:</b><br>Обработку персональных данных<br>Сроки проведения испытаний<br>Оплату всех расходов за проведение испытаний<br>Не возврат и списание образцов после проведения испытаний без представителя заказчика<br>Лаборатория не несет ответственность за стадию отбора, условия доставки образца(-ов), представленного(-ых) заказчиком для исследований<br>Лаборатория не дает оценку соответствия продукции, не выдает заключений о выданных результатах исследований.<br>Информация, указанная в данной заявке, будет отражена в протоколе испытаний. За информацию, предоставленную заказчиком, лаборатория ответственности не несет.<br>В случае получения результата, не соответствующего требованиям НД, будет сформирован срочный отчет по форме 4-ВЕТ-Б, согласно ПРИКАЗУ МИНСЕЛЬХОЗА РФ от 21.02.2022 № 89.<br>В соответствии со статьей 4.1 Закона РФ «О Ветеринарии» от 14.05.1993 №4979-1, Постановлением правительства РФ от 07.11.2016 №1140 «О порядке создания, развития и эксплуатации федеральной государственной информационной системы в области ветеринарии» и порядком предоставления информации в федеральную государственную информационную систему в области ветеринарии и получения информации из нее, утвержденного Приказом Минсельхоза России от 30.06.2017 №318, испытательная лаборатория осуществляет работу с использованием компонента ФГИС «ВЕТИС»;<br>В соответствии с Приказом Минэкономразвития РФ от 24.10.2020 г. № 704 «Об утверждении положения о составе сведений о результатах деятельности аккредитованных лиц, об изменениях состава их работников и о компетенции этих работников, об изменениях технической оснащенности, представляемых аккредитованными лицами в Федеральную службу по аккредитации, порядке и сроках представления аккредитованными лицами таких сведений в Федеральную службу по аккредитации» сведения о выданных протоколах испытаний, в том числе о Заказчике, представляются в Федеральную службу по аккредитации;</div>
<table style="width:100%%; border-collapse:collapse; margin-top:15px; font-size:10pt;"><tr><td style="border:1px solid #000; padding:6px; width:33%%;"><b>Представитель заказчика ознакомлен и согласен</b></td><td style="border:1px solid #000; padding:6px; width:22%%;"><div class="sig-hint">(должность)</div><div>%s</div></td><td style="border:1px solid #000; padding:6px; width:22%%;"><div class="sig-hint">(подпись)</div></td><td style="border:1px solid #000; padding:6px; width:23%%;"><div class="sig-hint">(расшифровка подписи)</div><div><b>%s</b></div></td></tr><tr><td style="border:1px solid #000; padding:6px;"><b>Подпись лица, сдавшего пробы, дата</b></td><td style="border:1px solid #000; padding:6px;"><div class="sig-hint">(подпись)</div></td><td style="border:1px solid #000; padding:6px;"><div class="sig-hint">(расшифровка подписи)</div><div><b>%s</b></div></td><td style="border:1px solid #000; padding:6px;"><div class="sig-hint">(дата)</div><div>%s</div></td></tr><tr><td style="border:1px solid #000; padding:6px;"><b>Сотрудник, принявший пробы, дата</b></td><td style="border:1px solid #000; padding:6px;"><div class="sig-hint">(подпись)</div></td><td style="border:1px solid #000; padding:6px;"><div class="sig-hint">(расшифровка подписи)</div></td><td style="border:1px solid #000; padding:6px;"><div class="sig-hint">(дата)</div></td></tr></table>
<div class="lab-mark"><b>Отметка лаборатории:</b><br>Доставлено проб (кол-во): ____________ &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Забраковано (кол-во): ____________<br>Причина браковки проб: ____________________________________________________________________________</div>
<div class="analysis"><b>Анализ заявки произведен и согласован:</b><br>Представитель заказчика: &nbsp; _____________________________ &nbsp; (ФИО) &nbsp; _____________________ &nbsp; (подпись)<br>Представитель %s: &nbsp; ________________ &nbsp; (должность) &nbsp; ____________ &nbsp; (подпись) &nbsp; _______________________ &nbsp; (расшифровка подписи)</div>
<p style="margin-top:15px; font-size:9pt; font-style:italic;">Примечание:</p>
<div style="page-break-before: always;"></div>
<h3 style="text-align:center; text-transform:uppercase;">Опись проб к сопроводительному документу-заявке</h3>
<div style="font-size:9pt; margin-bottom:10px;">(регистрационный номер заявки) ____________ &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; (дата поступления) ____________</div>
<table style="width:100%%; border-collapse:collapse; font-size:10pt;"><thead><tr><th style="border:1px solid #000; padding:5px; width:7%%;">№ п/п</th><th style="border:1px solid #000; padding:5px; width:25%%;">Инвентарный номер/кличка</th><th style="border:1px solid #000; padding:5px; width:18%%;">Возраст, пол</th><th style="border:1px solid #000; padding:5px; width:18%%;">Шифр пробы</th><th style="border:1px solid #000; padding:5px; width:32%%;">Результат исследования</th></tr></thead><tbody><tr><td style="border:1px solid #000; padding:5px; text-align:center;">1</td><td style="border:1px solid #000; padding:5px;">%s</td><td style="border:1px solid #000; padding:5px;">%s, %s</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr><tr><td style="border:1px solid #000; padding:5px; text-align:center;">2</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr><tr><td style="border:1px solid #000; padding:5px; text-align:center;">3</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr><tr><td style="border:1px solid #000; padding:5px; text-align:center;">4</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr><tr><td style="border:1px solid #000; padding:5px; text-align:center;">5</td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td><td style="border:1px solid #000; padding:5px;"></td></tr></tbody></table>
<div style="margin-top:30px; display:flex; justify-content:space-between; font-size:10pt;"><div><b>Заказчик (владелец):</b><br><br>____________ &nbsp; <b>%s</b><br><span style="font-size:8pt;">(подпись) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(расшифровка подписи)</span></div><div style="text-align:right;">стр. 2 из 2</div></div>
</body></html>`,
		cfg["lab_name"], cfg["lab_address"], cfg["lab_phone"], cfg["lab_email"],
		cfg["clinic_name"], ownerDoc, ownerAddr, ownerPhone,
		disease, materialName, materialQty, species, age, ownerName, ownerAddr,
		formatDateRu(illnessDate), formatDateRu(deathDate), clinical, patho, presumptiveDiag, samplePlace,
		formatDateTimeRu(sampleTime), actNo, formatDateRu(sendDate), packType,
		vetPos, vetName, accMark, accOutsideMark, methodYes, methodNo, uncYes, uncNo,
		resultHand, resultEmailMark, resultEmail, resultCopies,
		ownerRepPos, ownerName, vetName, formatDateRu(sendDate), cfg["lab_name"], nick, age, sex, ownerName)

	w.Header().Set("Content-Type", "application/msword")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="F20_%s.doc"`, id[:8]))
	w.Write([]byte(html))
}

// --- Генерация DOCX через Node.js ---
func handleGetProtocolDocx(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/protocol-docx/"):]
	row := db.QueryRow(`SELECT p.created_at, v.full_name, v.position, o.full_name, o.doc_info, o.address, o.phone, 
		a.species, a.age, a.sex, a.nickname, a.inventory_no, p.disease_to_exclude, p.material_name, p.material_quantity, 
		p.illness_date, p.death_date, p.clinical_picture, p.patho_data, p.presumptive_diagnosis, p.sampling_place, 
		p.sampling_datetime, p.act_no, p.send_date, p.packaging_type, p.accreditation_scope, p.method_choice, 
		p.uncertainty_info, p.result_delivery, p.result_email, p.result_copies, p.owner_rep_position
		FROM protocols p JOIN vets v ON p.vet_id = v.id JOIN owners o ON p.owner_id = o.id JOIN animals a ON p.animal_id = a.id WHERE p.id = ?`, id)

	var createdAt, vetName, vetPos, ownerName, ownerDoc, ownerAddr, ownerPhone string
	var species, age, sex, nick, invNo, disease, materialName, materialQty, illnessDate, deathDate, clinical, patho, presumptiveDiag string
	var samplePlace, sampleTime, actNo, sendDate, packType, accScope, methodChoice, uncertaintyInfo, resultDelivery, resultEmail, resultCopies, ownerRepPos string

	err := row.Scan(&createdAt, &vetName, &vetPos, &ownerName, &ownerDoc, &ownerAddr, &ownerPhone, &species, &age, &sex, &nick, &invNo,
		&disease, &materialName, &materialQty, &illnessDate, &deathDate, &clinical, &patho, &presumptiveDiag, &samplePlace, &sampleTime, &actNo, &sendDate, &packType,
		&accScope, &methodChoice, &uncertaintyInfo, &resultDelivery, &resultEmail, &resultCopies, &ownerRepPos)
	if err != nil {
		http.Error(w, "Протокол не найден", http.StatusNotFound)
		return
	}

	templateData := map[string]interface{}{
		"id": id, "vetOrganizationName": "БГУ Ивановской области «Ивановская городская СББЖ»",
		"customerName": ownerName, "customerDoc": ownerDoc, "customerAddress": ownerAddr, "customerPhone": ownerPhone,
		"diseases": disease, "biomaterialName": materialName, "biomaterialQuantity": materialQty,
		"animalSpecies": species, "animalAge": age, "animalNickname": nick,
		"illnessDate": formatDateRu(illnessDate), "deathDate": formatDateRu(deathDate),
		"clinicalPicture": clinical, "pathoData": patho, "presumptiveDiagnosis": presumptiveDiag,
		"samplingPlace": samplePlace, "samplingDate": sampleTime, "actNo": actNo,
		"materialSamplingDate": formatDateRu(sendDate), "packagingTypeName": packType,
		"vetDoctorPosition": vetPos, "vetDoctorShortName": shortenName(vetName),
		"ownerRepShortName": shortenName(ownerName), "resultEmail": resultEmail, "resultCopies": resultCopies,
		"animalsList": []map[string]interface{}{
			{"index": 1, "speciesAndNickname": fmt.Sprintf("%s, %s", species, nick), "ageAndGender": fmt.Sprintf("%s, %s", age, sex)},
		},
	}

	dataJSON, _ := json.Marshal(templateData)
	cmd := exec.Command("node", "generate.js", "template-f20.docx")
	cmd.Stdin = bytes.NewReader(dataJSON)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		http.Error(w, "Ошибка генерации документа: "+stderr.String(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="F20_%s.docx"`, id[:8]))
	w.Write(out.Bytes())
}

// --- Список протоколов (История) ---
func handleListProtocols(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT p.id, p.created_at, p.decision_type, v.full_name as vet_name, o.full_name as owner_name, a.species, a.nickname, p.disease_to_exclude, p.material_name FROM protocols p JOIN vets v ON p.vet_id = v.id JOIN owners o ON p.owner_id = o.id JOIN animals a ON p.animal_id = a.id ORDER BY p.created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var protocols []map[string]interface{}
	for rows.Next() {
		var id, createdAt, decType, vetName, ownerName, species, nickname, disease, materialName string
		rows.Scan(&id, &createdAt, &decType, &vetName, &ownerName, &species, &nickname, &disease, &materialName)
		protocols = append(protocols, map[string]interface{}{
			"id": id, "created_at": createdAt, "decision_type": decType, "vet_name": vetName,
			"owner_name": ownerName, "species": species, "nickname": nickname, "disease": disease, "material_name": materialName,
		})
	}
	json.NewEncoder(w).Encode(protocols)
}

// --- Получение протокола для редактирования ---
func handleEditProtocol(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/protocol/edit/"):]
	row := db.QueryRow(`SELECT p.created_at, p.vet_id, p.owner_id, p.animal_id, p.decision_type, p.decision_reason, p.survey_json,
		p.disease_to_exclude, p.material_name, p.material_quantity, p.illness_date, p.death_date, p.clinical_picture, p.patho_data, p.presumptive_diagnosis,
		p.sampling_place, p.sampling_datetime, p.act_no, p.send_date, p.packaging_type, p.accreditation_scope, p.method_choice, p.uncertainty_info,
		p.result_delivery, p.result_email, p.result_copies, p.owner_rep_position, p.sample_receiver_position FROM protocols p WHERE p.id = ?`, id)

	var createdAt, vetID, ownerID, animalID, decType, decReason, surveyJSON string
	var disease, materialName, materialQty, illnessDate, deathDate, clinical, patho, presumptiveDiag string
	var samplePlace, sampleTime, actNo, sendDate, packType, accScope, methodChoice, uncertaintyInfo, resultDelivery, resultEmail, resultCopies, ownerRepPos, sampleReceiverPos string

	err := row.Scan(&createdAt, &vetID, &ownerID, &animalID, &decType, &decReason, &surveyJSON,
		&disease, &materialName, &materialQty, &illnessDate, &deathDate, &clinical, &patho, &presumptiveDiag,
		&samplePlace, &sampleTime, &actNo, &sendDate, &packType, &accScope, &methodChoice, &uncertaintyInfo,
		&resultDelivery, &resultEmail, &resultCopies, &ownerRepPos, &sampleReceiverPos)
	if err != nil {
		http.Error(w, "Протокол не найден", http.StatusNotFound)
		return
	}

	var survey interface{}
	json.Unmarshal([]byte(surveyJSON), &survey)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id, "created_at": createdAt, "vet_id": vetID, "owner_id": ownerID, "animal_id": animalID,
		"decision_type": decType, "decision_reason": decReason, "survey": survey,
		"disease": disease, "material_name": materialName, "material_quantity": materialQty,
		"illness_date": illnessDate, "death_date": deathDate, "clinical": clinical, "patho": patho, "presumptive_diagnosis": presumptiveDiag,
		"sample_place": samplePlace, "sample_time": sampleTime, "act_no": actNo, "send_date": sendDate, "pack_type": packType,
		"accreditation_scope": accScope, "method_choice": methodChoice, "uncertainty_info": uncertaintyInfo,
		"result_delivery": resultDelivery, "result_email": resultEmail, "result_copies": resultCopies,
		"owner_rep_position": ownerRepPos, "sample_receiver_position": sampleReceiverPos,
	})
}

// --- Обновление протокола ---
func handleUpdateProtocol(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Path[len("/api/protocol/update/"):]
	var req struct {
		VetID, OwnerID, AnimalID, DecisionType, DecisionReason string
		Survey           json.RawMessage
		DiseaseToExclude, MaterialName, MaterialQuantity       string
		IllnessDate, DeathDate, ClinicalPicture, PathoData     string
		PresumptiveDiag, SamplingPlace, SamplingDatetime       string
		ActNo, SendDate, PackagingType                         string
		AccreditationScope, MethodChoice, UncertaintyInfo      string
		ResultDelivery, ResultEmail, ResultCopies              string
		OwnerRepPosition, SampleReceiverPos                    string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ошибка чтения данных", http.StatusBadRequest)
		return
	}

	surveyStr := "{}"
	if len(req.Survey) > 0 {
		surveyStr = string(req.Survey)
	}

	_, err := db.Exec(`UPDATE protocols SET vet_id=?, owner_id=?, animal_id=?, decision_type=?, decision_reason=?, survey_json=?,
		disease_to_exclude=?, material_name=?, material_quantity=?, illness_date=?, death_date=?, clinical_picture=?, patho_data=?, presumptive_diagnosis=?,
		sampling_place=?, sampling_datetime=?, act_no=?, send_date=?, packaging_type=?, accreditation_scope=?, method_choice=?, uncertainty_info=?,
		result_delivery=?, result_email=?, result_copies=?, owner_rep_position=?, sample_receiver_position=? WHERE id=?`,
		req.VetID, req.OwnerID, req.AnimalID, req.DecisionType, req.DecisionReason, surveyStr,
		req.DiseaseToExclude, req.MaterialName, req.MaterialQuantity, req.IllnessDate, req.DeathDate,
		req.ClinicalPicture, req.PathoData, req.PresumptiveDiag, req.SamplingPlace, req.SamplingDatetime, req.ActNo, req.SendDate, req.PackagingType,
		req.AccreditationScope, req.MethodChoice, req.UncertaintyInfo, req.ResultDelivery, req.ResultEmail, req.ResultCopies,
		req.OwnerRepPosition, req.SampleReceiverPos, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// --- Удаление протокола ---
func handleDeleteProtocol(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Path[len("/api/protocol/delete/"):]
	_, err := db.Exec("DELETE FROM protocols WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}