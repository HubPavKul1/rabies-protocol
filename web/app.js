"use strict";

let currentProtocolId = null;
document.addEventListener("DOMContentLoaded", async () => {
  try {
    await loadSettings();
    await loadVets();
    await loadOwners();
    await loadAnimals();
    setupFlagListeners();
    updateFormView();
  } catch (e) {
    console.error("Ошибка инициализации:", e);
    alert("Ошибка при загрузке: " + e);
  }
});

function showTab(tabId, evt) {
  document
    .querySelectorAll(".content > div")
    .forEach((d) => d.classList.add("hidden"));
  document.getElementById("tab-" + tabId).classList.remove("hidden");
  document
    .querySelectorAll(".nav button")
    .forEach((b) => b.classList.remove("active"));
  if (evt && evt.target) evt.target.classList.add("active");

  // Загружаем историю при переключении на вкладку
  if (tabId === "history") {
    loadProtocols();
  }
}

let editingProtocolId = null;

async function loadProtocols() {
  const protocols = await api("/api/protocols/list");
  const tbody = document.querySelector("#protocols-table tbody");
  tbody.innerHTML = "";

  (Array.isArray(protocols) ? protocols : []).forEach((p) => {
    const date = new Date(p.created_at).toLocaleDateString("ru-RU");
    const decType = p.decision_type === "send" ? "Ф-20" : "Отказ";
    const decColor = p.decision_type === "send" ? "#c53030" : "#2c5282";

    tbody.innerHTML += `
            <tr>
              <td>${date}</td>
              <td>${p.owner_name}</td>
              <td>${p.species} "${p.nickname || "—"}"</td>
              <td><span style="color: ${decColor}; font-weight: bold;">${decType}</span></td>
              <td>
                <button class="btn btn-sm btn-primary" onclick="viewProtocol('${p.id}')">Просмотр</button>
                <button class="btn btn-sm btn-secondary" onclick="editProtocol('${p.id}')">Редактировать</button>
                <button class="btn btn-sm btn-danger" onclick="deleteProtocol('${p.id}')">Удалить</button>
              </td>
            </tr>
          `;
  });
}

document
  .getElementById("search-protocols")
  .addEventListener("input", async (e) => {
    const query = e.target.value.toLowerCase();
    const protocols = await api("/api/protocols/list");
    const tbody = document.querySelector("#protocols-table tbody");
    tbody.innerHTML = "";

    const filtered = (Array.isArray(protocols) ? protocols : []).filter(
      (p) =>
        p.owner_name.toLowerCase().includes(query) ||
        p.species.toLowerCase().includes(query) ||
        (p.nickname && p.nickname.toLowerCase().includes(query)),
    );

    filtered.forEach((p) => {
      const date = new Date(p.created_at).toLocaleDateString("ru-RU");
      const decType = p.decision_type === "send" ? "Ф-20" : "Отказ";
      const decColor = p.decision_type === "send" ? "#c53030" : "#2c5282";

      tbody.innerHTML += `
            <tr>
              <td>${date}</td>
              <td>${p.owner_name}</td>
              <td>${p.species} "${p.nickname || "—"}"</td>
              <td><span style="color: ${decColor}; font-weight: bold;">${decType}</span></td>
              <td>
                <button class="btn btn-sm btn-primary" onclick="viewProtocol('${p.id}')">Просмотр</button>
                <button class="btn btn-sm btn-secondary" onclick="editProtocol('${p.id}')">Редактировать</button>
                <button class="btn btn-sm btn-danger" onclick="deleteProtocol('${p.id}')">Удалить</button>
              </td>
            </tr>
          `;
    });
  });

async function viewProtocol(id) {
  showPrintView(id);
}

async function editProtocol(id) {
  editingProtocolId = id;
  const data = await api("/api/protocol/edit/" + id);

  // Заполняем форму данными из протокола
  document.getElementById("d-vet").value = data.vet_id;
  document.getElementById("d-owner").value = data.owner_id;
  document.getElementById("d-animal").value = data.animal_id;

  // Восстанавливаем галочки
  if (data.survey && data.survey.flags) {
    document.getElementById("flag-wild").checked =
      data.survey.flags.wild || false;
    document.getElementById("flag-bite").checked =
      data.survey.flags.bite || false;
    document.getElementById("flag-saliva").checked =
      data.survey.flags.saliva || false;
    document.getElementById("flag-signs").checked =
      data.survey.flags.signs || false;
    document.getElementById("flag-corpse").checked =
      data.survey.flags.corpse || false;
  }

  // Заполняем детали опроса
  if (data.survey) {
    document.getElementById("s-housing").value = data.survey.housing || "";
    document.getElementById("s-contacts").value = data.survey.contacts || "";
    document.getElementById("s-signs").value = data.survey.signs || "";
  }

  // Заполняем поля Ф-20
  document.getElementById("d-disease").value = data.disease || "";
  document.getElementById("d-material-name").value = data.material_name || "";
  document.getElementById("d-material-qty").value =
    data.material_quantity || "";
  document.getElementById("d-illness-date").value = data.illness_date || "";
  document.getElementById("d-death-date").value = data.death_date || "";
  document.getElementById("d-clinical").value = data.clinical || "";
  document.getElementById("d-patho").value = data.patho || "";
  document.getElementById("d-presumptive-diag").value =
    data.presumptive_diagnosis || "";
  document.getElementById("d-sample-place").value = data.sample_place || "";
  document.getElementById("d-sample-time").value = data.sample_time || "";
  document.getElementById("d-act-no").value = data.act_no || "";
  document.getElementById("d-send-date").value = data.send_date || "";
  document.getElementById("d-pack-type").value = data.pack_type || "";

  // Параметры исследования
  document.getElementById("d-acc-scope").value =
    data.accreditation_scope || "Области аккредитации";
  document.getElementById("d-method-choice").value =
    data.method_choice || "Оставляю за лабораторией";
  document.getElementById("d-uncertainty").value =
    data.uncertainty_info || "нет";
  document.getElementById("d-result-delivery").value =
    data.result_delivery || "выдать на руки";
  document.getElementById("d-result-email").value = data.result_email || "";
  document.getElementById("d-result-copies").value = data.result_copies || "1";
  document.getElementById("d-owner-rep-pos").value =
    data.owner_rep_position || "";
  document.getElementById("d-sample-receiver-pos").value =
    data.sample_receiver_position || "Ветеринарный врач";

  // Для отказа
  if (data.decision_type === "refuse") {
    document.getElementById("d-refuse-reason").value =
      data.decision_reason || "";
  }

  updateFormView();
  showTab("new-doc");
}

async function deleteProtocol(id) {
  if (!confirm("Вы уверены, что хотите удалить этот протокол?")) return;
  await api("/api/protocol/delete/" + id, "DELETE");
  loadProtocols();
}

function showDir(dir) {
  ["vets", "owners", "animals"].forEach((d) =>
    document.getElementById("dir-" + d).classList.add("hidden"),
  );
  document.getElementById("dir-" + dir).classList.remove("hidden");
}
function setupFlagListeners() {
  const flags = [
    "flag-wild",
    "flag-bite",
    "flag-saliva",
    "flag-signs",
    "flag-corpse",
    "flag-wild-carcass",
  ];
  flags.forEach((id) => {
    document.getElementById(id).addEventListener("change", updateFormView);
  });

  // Добавляем слушатели для радио-кнопок уточнения
  document
    .getElementById("indoor-yes")
    .addEventListener("change", updateFormView);
  document
    .getElementById("indoor-no")
    .addEventListener("change", updateFormView);
}
function updateFormView() {
  const flagWild = document.getElementById("flag-wild").checked;
  const flagBite = document.getElementById("flag-bite").checked;
  const flagSaliva = document.getElementById("flag-saliva").checked;
  const flagSigns = document.getElementById("flag-signs").checked;
  const flagCorpse = document.getElementById("flag-corpse").checked;
  const flagWildCarcass = document.getElementById("flag-wild-carcass").checked;

  const hardTriggers = flagWild || flagSigns || flagCorpse || flagWildCarcass;
  const softTriggers = flagBite || flagSaliva;

  const indoorCheckBlock = document.getElementById("block-indoor-check");
  const blockSend = document.getElementById("block-send");
  const blockRefuse = document.getElementById("block-refuse");

  // Сбрасываем радио-кнопки при изменении чекбоксов
  if (!softTriggers || hardTriggers) {
    document.getElementById("indoor-yes").checked = false;
    document.getElementById("indoor-no").checked = false;
  }

  if (hardTriggers) {
    // Жесткие триггеры: всегда Ф-20
    indoorCheckBlock.classList.add("hidden");
    blockSend.classList.remove("hidden");
    blockRefuse.classList.add("hidden");
  } else if (softTriggers) {
    // Мягкие триггеры: показываем уточнение
    indoorCheckBlock.classList.remove("hidden");

    // Показываем форму в зависимости от выбранного радио
    const indoorYes = document.getElementById("indoor-yes").checked;
    const indoorNo = document.getElementById("indoor-no").checked;

    if (indoorNo) {
      blockSend.classList.remove("hidden");
      blockRefuse.classList.add("hidden");
    } else if (indoorYes) {
      blockSend.classList.add("hidden");
      blockRefuse.classList.remove("hidden");
    } else {
      // Если радио еще не выбрано, скрываем обе формы, ждем выбора
      blockSend.classList.add("hidden");
      blockRefuse.classList.add("hidden");
    }
  } else {
    // Нет никаких флагов: только отказ
    indoorCheckBlock.classList.add("hidden");
    blockSend.classList.add("hidden");
    blockRefuse.classList.remove("hidden");
  }
}

async function api(url, method = "GET", data = null) {
  const opts = {
    method,
    headers: { "Content-Type": "application/json" },
  };
  if (data) opts.body = JSON.stringify(data);
  const res = await fetch(url, opts);
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${await res.text()}`);
  return method === "GET" ? await res.json() : res;
}

async function loadSettings() {
  const cfg = await api("/api/config");
  document.getElementById("cfg-clinic-name").value = cfg.clinic_name || "";
  document.getElementById("cfg-clinic-address").value =
    cfg.clinic_address || "";
  document.getElementById("cfg-clinic-phone").value = cfg.clinic_phone || "";
  document.getElementById("cfg-clinic-email").value = cfg.clinic_email || "";
  document.getElementById("cfg-lab-name").value = cfg.lab_name || "";
  document.getElementById("cfg-lab-address").value = cfg.lab_address || "";
  document.getElementById("cfg-lab-phone").value = cfg.lab_phone || "";
  document.getElementById("cfg-lab-email").value = cfg.lab_email || "";
}
async function saveSettings() {
  await api("/api/config", "POST", {
    clinic_name: document.getElementById("cfg-clinic-name").value,
    clinic_address: document.getElementById("cfg-clinic-address").value,
    clinic_phone: document.getElementById("cfg-clinic-phone").value,
    clinic_email: document.getElementById("cfg-clinic-email").value,
    lab_name: document.getElementById("cfg-lab-name").value,
    lab_address: document.getElementById("cfg-lab-address").value,
    lab_phone: document.getElementById("cfg-lab-phone").value,
    lab_email: document.getElementById("cfg-lab-email").value,
  });
  alert("✅ Настройки сохранены!");
}

async function loadVets() {
  const vets = await api("/api/vets");
  const select = document.getElementById("d-vet");
  select.innerHTML = '<option value="">Выберите...</option>';
  const tbody = document.querySelector("#table-vets tbody");
  tbody.innerHTML = "";
  (Array.isArray(vets) ? vets : []).forEach((v) => {
    select.innerHTML += `<option value="${v.id}">${v.full_name} (${v.position})</option>`;
    tbody.innerHTML += `<tr>
            <td>${v.full_name}</td>
            <td>${v.position}</td>
            <td>
              <button class="btn btn-sm btn-secondary" onclick="editVet('${v.id}')">Изменить</button>
              <button class="btn btn-sm btn-danger" onclick="deleteVet('${v.id}')">Удалить</button>
            </td>
          </tr>`;
  });
}

async function addVet() {
  const name = document.getElementById("new-vet-name").value.trim();
  const pos = document.getElementById("new-vet-pos").value.trim();

  if (!name) return alert("⚠️ Введите ФИО врача");

  try {
    await api("/api/vets", "POST", {
      full_name: name,
      position: pos || "Ветеринарный врач",
    });
    // Очищаем поля после успешного добавления
    document.getElementById("new-vet-name").value = "";
    document.getElementById("new-vet-pos").value = "";
    await loadVets(); // Обновляем список
  } catch (err) {
    alert("Ошибка добавления: " + err.message);
  }
}

async function editVet(id) {
  const vet = (await api("/api/vets")).find((v) => v.id === id);
  const name = prompt("ФИО:", vet.full_name);
  if (name === null) return;
  const pos = prompt("Должность:", vet.position);
  if (pos === null) return;
  await api("/api/vets", "PUT", { id, full_name: name, position: pos });
  loadVets();
}

async function deleteVet(id) {
  if (
    !confirm(
      "Вы уверены, что хотите удалить этого врача? (Он исчезнет из списков, но история его протоколов сохранится)",
    )
  )
    return;
  await api(`/api/vets/${id}`, "DELETE");
  loadVets();
}

async function loadOwners() {
  const owners = await api("/api/owners");
  const select = document.getElementById("d-owner");
  const animalOwnerSelect = document.getElementById("new-animal-owner");
  select.innerHTML = '<option value="">Выберите...</option>';
  animalOwnerSelect.innerHTML = "";
  const tbody = document.querySelector("#table-owners tbody");
  tbody.innerHTML = "";
  (Array.isArray(owners) ? owners : []).forEach((o) => {
    select.innerHTML += `<option value="${o.id}">${o.full_name}</option>`;
    animalOwnerSelect.innerHTML += `<option value="${o.id}">${o.full_name}</option>`;
    tbody.innerHTML += `<tr>
            <td>${o.full_name}</td>
            <td>${o.doc_info}</td>
            <td>${o.address}</td>
            <td>${o.phone}</td>
            <td>
              <button class="btn btn-sm btn-secondary" onclick="editOwner('${o.id}')">Изменить</button>
              <button class="btn btn-sm btn-danger" onclick="deleteOwner('${o.id}')">Удалить</button>
            </td>
          </tr>`;
  });
}

async function addOwner() {
  const name = document.getElementById("new-owner-name").value.trim();
  if (!name) return alert("⚠️ Введите ФИО или название организации");

  try {
    await api("/api/owners", "POST", {
      full_name: name,
      doc_info: document.getElementById("new-owner-doc").value.trim(),
      address: document.getElementById("new-owner-addr").value.trim(),
      phone: document.getElementById("new-owner-phone").value.trim(),
    });

    // Очищаем поля
    [
      "new-owner-name",
      "new-owner-doc",
      "new-owner-addr",
      "new-owner-phone",
    ].forEach((id) => {
      document.getElementById(id).value = "";
    });
    await loadOwners();
  } catch (err) {
    alert("Ошибка добавления: " + err.message);
  }
}

async function editOwner(id) {
  const owner = (await api("/api/owners")).find((o) => o.id === id);

  // Последовательные запросы для редактирования всех полей
  const name = prompt("ФИО / Название организации:", owner.full_name);
  if (name === null) return;

  const doc = prompt("Документ (паспорт и т.д.):", owner.doc_info);
  if (doc === null) return;

  const address = prompt("Адрес:", owner.address);
  if (address === null) return;

  const phone = prompt("Телефон:", owner.phone);
  if (phone === null) return;

  await api("/api/owners", "PUT", {
    id,
    full_name: name,
    doc_info: doc,
    address: address,
    phone: phone,
  });
  loadOwners();
}

async function deleteOwner(id) {
  if (!confirm("Вы уверены, что хотите удалить этого владельца?")) return;
  await api(`/api/owners/${id}`, "DELETE");
  loadOwners();
}

async function loadAnimals() {
  const animals = await api("/api/animals");
  const select = document.getElementById("d-animal");
  select.innerHTML = '<option value="">Выберите...</option>';
  const tbody = document.querySelector("#table-animals tbody");
  tbody.innerHTML = "";
  (Array.isArray(animals) ? animals : []).forEach((a) => {
    select.innerHTML += `<option value="${a.id}">${a.species} "${a.nickname || a.inventory_no}" (${a.owner_name})</option>`;
    tbody.innerHTML += `<tr>
            <td>${a.owner_name}</td>
            <td>${a.species}</td>
            <td>${a.nickname || a.inventory_no}</td>
            <td>${a.age}, ${a.sex}</td>
            <td>
              <button class="btn btn-sm btn-secondary" onclick="editAnimal('${a.id}')">Изменить</button>
              <button class="btn btn-sm btn-danger" onclick="deleteAnimal('${a.id}')">Удалить</button>
            </td>
          </tr>`;
  });
}

async function deleteAnimal(id) {
  if (!confirm("Вы уверены, что хотите удалить это животное из справочника?"))
    return;
  await api(`/api/animals/${id}`, "DELETE");
  loadAnimals();
}
async function addAnimal() {
  await api("/api/animals", "POST", {
    owner_id: document.getElementById("new-animal-owner").value,
    species: document.getElementById("new-animal-species").value,
    nickname: document.getElementById("new-animal-nick").value,
    age: document.getElementById("new-animal-age").value,
    sex: document.getElementById("new-animal-sex").value,
    inventory_no: "",
  });
  ["new-animal-species", "new-animal-nick", "new-animal-age"].forEach(
    (id) => (document.getElementById(id).value = ""),
  );
  loadAnimals();
}
async function editAnimal(id) {
  const animal = (await api("/api/animals")).find((a) => a.id === id);
  const nick = prompt("Кличка/Инв.№:", animal.nickname || animal.inventory_no);
  if (nick) {
    await api("/api/animals", "PUT", { ...animal, nickname: nick });
    loadAnimals();
  }
}

async function submitDoc(e) {
  e.preventDefault();
  if (!document.getElementById("d-vet").value)
    return alert("⚠️ Выберите врача");
  if (!document.getElementById("d-owner").value)
    return alert("⚠️ Выберите владельца");
  if (!document.getElementById("d-animal").value)
    return alert("⚠️ Выберите животное");

  const flagWild = document.getElementById("flag-wild").checked;
  const flagBite = document.getElementById("flag-bite").checked;
  const flagSaliva = document.getElementById("flag-saliva").checked;
  const flagSigns = document.getElementById("flag-signs").checked;
  const flagCorpse = document.getElementById("flag-corpse").checked;
  const flagWildCarcass = document.getElementById("flag-wild-carcass").checked;
  const indoorYes = document.getElementById("indoor-yes").checked;

  const hardTriggers = flagWild || flagSigns || flagCorpse || flagWildCarcass;
  const softTriggers = flagBite || flagSaliva;

  // Определяем тип решения
  let decisionType = "refuse";
  if (hardTriggers) {
    decisionType = "send";
  } else if (softTriggers) {
    if (!indoorYes) {
      return alert(
        "⚠️ При наличии укуса/ослюнения необходимо подтвердить строгое домашнее содержание без контактов для оформления отказа. Иначе требуется направление на исследование (Ф-20).",
      );
    }
    decisionType = "refuse"; // Врач подтвердил домашнее содержание, идем в отказ
  }

  // Валидация полей в зависимости от решения
  if (decisionType === "send") {
    if (!document.getElementById("d-material-name").value.trim())
      return alert("⚠️ Укажите наименование материала");
    if (!document.getElementById("d-material-qty").value.trim())
      return alert("⚠️ Укажите количество материала");
  } else {
    if (!document.getElementById("d-refuse-reason").value.trim()) {
      return alert(
        "⚠️ Для обоснованного отказа обязательно укажите диагноз и обстоятельства (например: 'Укус на фоне стоматита, строгое домашнее содержание, контакт с другими животными исключен').",
      );
    }
  }

  const survey = {
    housing: document.getElementById("s-housing").value,
    contacts: document.getElementById("s-contacts").value,
    signs: document.getElementById("s-signs").value,
    flags: {
      wild: flagWild,
      bite: flagBite,
      saliva: flagSaliva,
      signs: flagSigns,
      corpse: flagCorpse,
      wild_carcass: flagWildCarcass,
      indoor_confirmed: indoorYes, // Сохраняем подтверждение врача в БД для юридической защиты
    },
  };

  const payload = {
    vet_id: document.getElementById("d-vet").value,
    owner_id: document.getElementById("d-owner").value,
    animal_id: document.getElementById("d-animal").value,
    decision_type: decisionType,
    decision_reason:
      decisionType === "send"
        ? "Наличие оснований для подозрения на бешенство (п. 10 Правил)"
        : document.getElementById("d-refuse-reason").value,
    survey: survey,
    disease_to_exclude: document.getElementById("d-disease")?.value || "",
    material_name: document.getElementById("d-material-name")?.value || "",
    material_quantity: document.getElementById("d-material-qty")?.value || "",
    illness_date: document.getElementById("d-illness-date")?.value || "",
    death_date: document.getElementById("d-death-date")?.value || "",
    clinical_picture: document.getElementById("d-clinical")?.value || "",
    patho_data: document.getElementById("d-patho")?.value || "",
    presumptive_diagnosis:
      document.getElementById("d-presumptive-diag")?.value || "",
    sampling_place: document.getElementById("d-sample-place")?.value || "",
    sampling_datetime: document.getElementById("d-sample-time")?.value || "",
    act_no: document.getElementById("d-act-no")?.value || "",
    send_date:
      document.getElementById("d-send-date")?.value ||
      new Date().toISOString().split("T")[0],
    packaging_type: document.getElementById("d-pack-type")?.value || "",
    accreditation_scope: document.getElementById("d-acc-scope")?.value || "",
    method_choice: document.getElementById("d-method-choice")?.value || "",
    uncertainty_info: document.getElementById("d-uncertainty")?.value || "",
    result_delivery: document.getElementById("d-result-delivery")?.value || "",
    result_email: document.getElementById("d-result-email")?.value || "",
    result_copies: document.getElementById("d-result-copies")?.value || "",
    owner_rep_position: document.getElementById("d-owner-rep-pos")?.value || "",
    sample_receiver_position:
      document.getElementById("d-sample-receiver-pos")?.value || "",
  };

  try {
    let res;
    if (editingProtocolId) {
      res = await api(
        "/api/protocol/update/" + editingProtocolId,
        "PUT",
        payload,
      );
      editingProtocolId = null;
      alert("✅ Протокол обновлен!");
      showTab("history");
      loadProtocols();
    } else {
      res = await api("/api/protocols", "POST", payload);
      const data = await res.json();
      showPrintView(data.id);
    }
  } catch (err) {
    alert("Ошибка: " + err.message);
  }
}

function formatDateRu(dateStr) {
  if (!dateStr) return "«___» ________ 20__ г.";
  const d = new Date(dateStr);
  const months = [
    "января",
    "февраля",
    "марта",
    "апреля",
    "мая",
    "июня",
    "июля",
    "августа",
    "сентября",
    "октября",
    "ноября",
    "декабря",
  ];
  return `«${d.getDate()}» ${months[d.getMonth()]} ${d.getFullYear()} г.`;
}

function flagText(survey) {
  const items = [];
  if (!survey || !survey.flags) return "Не выявлено";
  if (survey.flags.wild)
    items.push("контакты с дикими/безнадзорными животными");
  if (survey.flags.bite) items.push("факт укуса людей");
  if (survey.flags.saliva) items.push("ослюнение людей");
  if (survey.flags.signs) items.push("клинические признаки бешенства");
  if (survey.flags.corpse) items.push("следы укусов/ослюнения на трупе");
  return items.length ? items.join("; ") : "Не выявлено";
}

async function showPrintView(id) {
  const printView = document.getElementById("print-view");
  document
    .querySelectorAll(".content > div")
    .forEach((d) => d.classList.add("hidden"));
  document.querySelector(".nav").classList.add("hidden");
  printView.classList.remove("hidden");
  printView.style.display = "block";

  // 1. Сохраняем ID протокола для кнопки скачивания
  currentProtocolId = id;

  try {
    const data = await api("/api/protocol/" + id);
    const container = document.getElementById("print-document");
    const survey = data.survey || {};
    const today = formatDateRu(new Date().toISOString().split("T")[0]);

    // 2. Показываем кнопку "Скачать Word" ТОЛЬКО для Ф-20 (decision_type === "send")
    document.getElementById("btn-download-word").style.display =
      data.decision_type === "send" ? "inline-block" : "none";

    if (data.decision_type === "refuse") {
      // ===== ОТКАЗ =====
      container.innerHTML = `
              <div class="print-page">
                <div class="refusal-doc">
                  <div class="refusal-header-clinic">${data.clinic_name}</div>
                  <div class="refusal-header-clinic-contacts">
                    ${data.clinic_address}<br>
                    ${data.clinic_phone}${data.clinic_email ? ", " + data.clinic_email : ""}
                  </div>
                  <div style="text-align:right; font-size:11pt; margin-bottom:30px;">
                    Исх. № ______<br>от ${today}
                  </div>
                  <div class="refusal-title">ОБОСНОВАННЫЙ ОТКАЗ<br>в выдаче направления на исследование<br>патологического материала на бешенство</div>
                  <div class="refusal-section">
                    Уважаемая(ый) <b>${data.owner_name}</b>!<br><br>
                    Ваше обращение о выдаче направления на исследование трупа принадлежащего Вам животного 
                    <b>${data.species}</b> (возраст: ${data.age}, пол: ${data.sex}), кличка/инв.№: <b>${data.nickname || data.inv_no}</b> 
                    в аккредитованную ветеринарную лабораторию на бешенство рассмотрено.
                  </div>
                  <div class="refusal-section">
                    <b>1. Фактические обстоятельства:</b><br>
                    ${data.decision_reason}<br>
                    Эпидемиологические факторы, требующие исследования: <b>${flagText(survey)}</b>.
                  </div>
                  <div class="refusal-section">
                    <b>2. Правовое обоснование:</b><br>
                    В соответствии с п. 10 «Ветеринарных правил осуществления профилактических, диагностических, ограничительных и иных мероприятий, установления и отмены карантина и иных ограничений, направленных на предотвращение распространения и ликвидацию очагов бешенства», утвержденных приказом Минсельхоза России от 25.11.2020 № 705, исчерпывающий перечень оснований для подозрения на бешенство включает: наличие клинических признаков, контакт с больными животными, обнаружение трупа со следами укусов/ослюнения.<br>
                    В данном случае указанные основания <b>отсутствуют</b>.
                  </div>
                  <div class="refusal-section">
                    <b>3. Разъяснения:</b>
                    <ul>
                      <li>Труп животного подлежит утилизации в порядке, установленном Ветеринарными правилами сбора, хранения, перемещения, утилизации и уничтожения биологических отходов.</li>
                      <li>При наличии факта укуса или ослюнения животным Вам рекомендовано немедленно обратиться в медицинскую организацию (травмпункт) для решения вопроса о проведении курса антирабической вакцинации.</li>
                      <li>Вы вправе обратиться в стороннюю аккредитованную лабораторию для проведения патоморфологического исследования в частном порядке.</li>
                    </ul>
                  </div>
                  <div class="refusal-signatures">
                    <div>
                      ${data.vet_pos}:<br><br>
                      _________________ / ${data.vet_name} /
                    </div>
                    <div>
                      С отказом ознакомлен(а):<br><br>
                      _________________ / ${data.owner_name} /<br>
                      ${today}
                    </div>
                  </div>
                </div>
              </div>
            `;
    } else {
      // ===== Ф-20 ПО УТВЕРЖДЕННОЙ ФОРМЕ =====
      const sampleTimeStr = data.sample_time
        ? new Date(data.sample_time).toLocaleString("ru-RU", {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
          })
        : "";
      const illnessDateStr = data.illness_date
        ? formatDateRu(data.illness_date)
        : "";
      const deathDateStr = data.death_date ? formatDateRu(data.death_date) : "";
      const sendDateStr = data.send_date ? formatDateRu(data.send_date) : "";

      const labNameFormatted = (data.lab_name || "")
        .replace(
          /(Бюджетное государственное учреждение Ивановской области)/,
          "$1<br>",
        )
        .replace(/(«[^»]+»)/, "$1<br>");

      container.innerHTML = `
              <div class="print-page f20-page">
                <table class="f20-header-table" style="width:100%; border-collapse:collapse; font-size:9pt; margin-bottom:0;">
                  <tr>
                    <td style="border:1px solid #000; padding:6px; width:50%; vertical-align:top;">
                      <b>Заказчик:</b><br>
                      ${data.owner_name}<br>
                      ${data.owner_doc || "—"}<br>
                      ${data.clinic_address}<br>
                      Тел.: ${data.clinic_phone}
                    </td>
                    <td style="border:1px solid #000; padding:6px; width:50%; vertical-align:top; font-size:9pt;">
                      ${labNameFormatted}<br>
                      ${data.lab_address}<br>
                      ${data.lab_phone}<br>
                      mail: ${data.lab_email}
                    </td>
                  </tr>
                </table>

                <table class="f20-reg-table" style="width:100%; border-collapse:collapse; font-size:9pt; margin-bottom:8px;">
                  <tr>
                    <td style="padding:4px; width:25%; text-align:center;">Регистрационный номер заявки</td>
                    <td style="padding:4px; width:25%; text-align:center;">Шифр образца</td>
                    <td style="padding:4px; width:25%; text-align:center;">Дата поступления</td>
                    <td style="padding:4px; width:25%; text-align:center;">Время поступления</td>
                  </tr>
                </table>

                <div class="f20-title" style="text-align:center; font-weight:bold; font-size:11pt; margin:10px 0; text-transform:uppercase;">
                  СОПРОВОДИТЕЛЬНЫЙ ДОКУМЕНТ-ЗАЯВКА<br>К ПАТОЛОГИЧЕСКОМУ (БИОЛОГИЧЕСКОМУ) МАТЕРИАЛУ
                </div>

                <table class="f20-main-table" style="width:100%; border-collapse:collapse; font-size:9pt;">
                  <tr>
                    <td style="padding:4px 6px; width:30%; font-weight:bold;">Направляется для исключения</td>
                    <td style="padding:4px 6px;">(наименования заболевания)<br><b>${data.disease || ""}</b></td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">Патологический (биологический) материал<br><span style="font-weight:normal; font-size:8pt;">(указать наименование материала, количество материала):</span></td>
                    <td style="padding:4px 6px;"><b>${data.material_name || ""}</b>, количество: <b>${data.material_quantity || ""}</b></td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">от</td>
                    <td style="padding:4px 6px;">(вид и возраст животного)<br><b>${data.species || ""}</b>, возраст: <b>${data.age || ""}</b></td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">принадлежащего</td>
                    <td style="padding:4px 6px;">(наименование организации, хозяйства, предприятия, ФИО владельца животного)<br><b>${data.owner_name || ""}</b></td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">(фактический/юридический адрес)</td>
                    <td style="padding:4px 6px;"><b>${data.owner_addr || ""}</b></td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">Дата заболевания животного</td>
                    <td style="padding:4px 6px;">${illnessDateStr} &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Дата падежа: ${deathDateStr}</td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">Клиническая картина</td>
                    <td style="padding:4px 6px;">${data.clinical || ""}</td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">Данные патологоанатомического вскрытия</td>
                    <td style="padding:4px 6px;">${data.patho || ""}</td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">Предположительный диагноз</td>
                    <td style="padding:4px 6px;">${data.presumptive_diagnosis || ""}</td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">Место отбора</td>
                    <td style="padding:4px 6px;">${data.sample_place || ""}</td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">Дата, время отбора</td>
                    <td style="padding:4px 6px;">${sampleTimeStr} &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; № акта: ${data.act_no || ""}</td>
                  </tr>
                  <tr>
                    <td style="padding:4px 6px; font-weight:bold;">Дата отправки материала</td>
                    <td style="padding:4px 6px;">${sendDateStr} &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Вид упаковки: ${data.pack_type || ""}</td>
                  </tr>
                </table>

                <div style="margin-top:12px; font-size:9pt;">
                  <b>Материал отобран, упакован и отправлен:</b><br><br>
                  <b>${data.vet_pos || ""}</b> &nbsp; ____________ &nbsp; <b>${data.vet_name || ""}</b><br>
                  <span style="font-size:8pt;">(должность) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(подпись) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(ФИО)</span>
                </div>

                <div style="margin-top:15px; font-size:9pt; line-height:1.6;">
                  <div><b>Испытания провести согласно:</b> &nbsp; ☐ Области аккредитации &nbsp;&nbsp; ☐ Вне области аккредитации &nbsp;&nbsp; <i>отметить нужное</i></div>
                  <div style="margin-top:6px;"><b>Право выбора оптимального метода и методики испытания проб(ы) оставляю за лабораторией:</b> &nbsp; ☐ да &nbsp;&nbsp; ☐ нет</div>
                  <div style="margin-top:6px;"><b>ПРЕДОСТАВЛЕНИЕ СВЕДЕНИЙ ОБ ОЦЕНЕННОЙ НЕОПРЕДЕЛЕННОСТИ В ПРОТОКОЛЕ ЛАБОРАТОРНЫХ ИСПЫТАНИЙ:</b> &nbsp; ☐ да &nbsp;&nbsp; ☐ нет</div>
                  <div style="margin-top:6px;"><b>Прошу оригинал/копию (ненужное зачеркнуть) результата исследования:</b></div>
                  <div style="margin-left:20px;">☐ выдать на руки</div>
                  <div style="margin-left:20px;">☐ отправить по электронной почте &nbsp; <u>${data.result_email || "(указать адрес электронной почты)"}</u></div>
                  <div style="margin-top:6px;"><b>Количество необходимых экземпляров результатов исследований:</b> <u>${data.result_copies || "1"}</u></div>
                </div>

                <div style="margin-top:15px; font-size:7.5pt; line-height:1.25; text-align:justify;">
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

                <table style="width:100%; border-collapse:collapse; font-size:9pt; margin-top:15px;">
                  <tr>
                    <td style="padding:6px; width:40%;"><b>Представитель заказчика ознакомлен и согласен</b></td>
                    <td style="padding:6px; width:20%;"><div style="font-size:8pt; color:#666;">(должность)</div><div>${data.owner_rep_position || ""}</div></td>
                    <td style="padding:6px; width:20%;"><div style="font-size:8pt; color:#666;">(подпись)</div></td>
                    <td style="padding:6px; width:20%;"><div style="font-size:8pt; color:#666;">(расшифровка подписи)</div><div><b>${data.owner_name || ""}</b></div></td>
                  </tr>
                  <tr>
                    <td style="padding:6px;"><b>Подпись лица, сдавшего пробы, дата</b></td>
                    <td style="padding:6px;"><div style="font-size:8pt; color:#666;">(подпись)</div></td>
                    <td style="padding:6px;"><div style="font-size:8pt; color:#666;">(расшифровка подписи)</div><div><b>${data.vet_name || ""}</b></div></td>
                    <td style="padding:6px;"><div style="font-size:8pt; color:#666;">(дата)</div><div>${sendDateStr}</div></td>
                  </tr>
                  <tr>
                    <td style="padding:6px;"><b>Сотрудник, принявший пробы, дата</b></td>
                    <td style="padding:6px;"><div style="font-size:8pt; color:#666;">(подпись)</div></td>
                    <td style="padding:6px;"><div style="font-size:8pt; color:#666;">(расшифровка подписи)</div></td>
                    <td style="padding:6px;"><div style="font-size:8pt; color:#666;">(дата)</div></td>
                  </tr>
                </table>

                <div style="margin-top:15px; border:1px solid #000; padding:8px; font-size:9pt;">
                  <b>Отметка лаборатории:</b><br>
                  Доставлено проб (кол-во): ____________ &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Забраковано (кол-во): ____________<br>
                  Причина браковки проб: ____________________________________________________________________________
                </div>

                <div style="margin-top:12px; font-size:9pt;">
                  <b>Анализ заявки произведен и согласован:</b><br>
                  <div style="margin-top:6px;">Представитель заказчика: &nbsp; _____________________________ &nbsp; (ФИО) &nbsp; _____________________ &nbsp; (подпись)</div>
                  <div style="margin-top:6px;">Представитель ${data.lab_name || "лаборатории"}: &nbsp; ________________ &nbsp; (должность) &nbsp; ____________ &nbsp; (подпись) &nbsp; _______________________ &nbsp; (расшифровка подписи)</div>
                </div>

                <div style="margin-top:15px; font-size:7.5pt; font-style:italic; color:#333;">Примечание:</div>
              </div>

              <div class="print-page">
                <div class="f20-inventory-title">Опись проб к сопроводительному документу-заявке:</div>
                <div style="font-size:9pt; margin-bottom:10px;">
                  (регистрационный номер заявки) ____________ &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; (дата поступления) ____________
                </div>
                <table class="f20-inventory-table">
                  <thead>
                    <tr>
                      <th style="width:7%;">№ п/п</th>
                      <th style="width:25%;">Инвентарный номер/кличка</th>
                      <th style="width:18%;">Возраст, пол</th>
                      <th style="width:18%;">Шифр пробы</th>
                      <th style="width:32%;">Результат исследования</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>1</td>
                      <td>${data.nickname || data.inv_no || ""}</td>
                      <td>${data.age}, ${data.sex}</td>
                      <td></td>
                      <td></td>
                    </tr>
                    <tr><td>2</td><td></td><td></td><td></td><td></td></tr>
                    <tr><td>3</td><td></td><td></td><td></td><td></td></tr>
                    <tr><td>4</td><td></td><td></td><td></td><td></td></tr>
                    <tr><td>5</td><td></td><td></td><td></td><td></td></tr>
                  </tbody>
                </table>
                <div class="f20-footer">
                  <div>
                    <b>Заказчик (владелец):</b><br><br>
                    ____________ &nbsp; <b>${data.owner_name}</b><br>
                    <span style="font-size:8pt;">(подпись) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(расшифровка подписи)</span>
                  </div>
                  <div style="text-align:right;">стр. 2 из 2</div>
                </div>
              </div>
            `;
    }

    // ===== ПРОТОКОЛ ОПРОСА =====
    const surveyPage = `
            <div class="print-page">
              <div class="survey-doc">
                <div class="survey-header">ПРОТОКОЛ<br>эпидемиологического опроса владельца животного</div>
                <div class="survey-meta">
                  Дата составления: ${today}<br>
                  Место составления: ${data.clinic_name}, ${data.clinic_address}
                </div>
                <table class="survey-table">
                  <tr><td class="survey-label">Опрос проведен ветеринарным врачом</td><td>${data.vet_name}, ${data.vet_pos}</td></tr>
                  <tr><td class="survey-label">Владелец (ФИО, документ, адрес, телефон)</td><td>${data.owner_name}<br>${data.owner_doc || "—"}<br>${data.owner_addr}<br>Тел.: ${data.owner_phone}</td></tr>
                  <tr><td class="survey-label">Вид, кличка/инв.№, возраст, пол животного</td><td>${data.species}, ${data.nickname || data.inv_no || "—"}, возраст: ${data.age}, пол: ${data.sex}</td></tr>
                  <tr><td class="survey-label">Условия содержания</td><td>${survey.housing || "—"}</td></tr>
                  <tr><td class="survey-label">Контакты с дикими/безнадзорными животными</td><td>${survey.flags && survey.flags.wild ? "ДА. " + (survey.contacts || "") : "Не выявлено"}</td></tr>
                  <tr><td class="survey-label">Факты укусов людей за последние 60 дней</td><td>${survey.flags && survey.flags.bite ? "ДА. " + (survey.contacts || "") : "Не выявлено"}</td></tr>
                  <tr><td class="survey-label">Факты ослюнения людей</td><td>${survey.flags && survey.flags.saliva ? "ДА. " + (survey.contacts || "") : "Не выявлено"}</td></tr>
                  <tr><td class="survey-label">Клинические признаки бешенства</td><td>${survey.flags && survey.flags.signs ? "ДА. " + (survey.signs || "") : "Не выявлено"}</td></tr>
                  <tr><td class="survey-label">Следы укусов/ослюнения на трупе</td><td>${survey.flags && survey.flags.corpse ? "ДА. " + (survey.signs || "") : "Не выявлено"}</td></tr>
                  <tr><td class="survey-label">Обстоятельства падежа / диагноз</td><td>${data.decision_reason || "—"}</td></tr>
                  <tr><td class="survey-label">Заключение врача</td><td>${data.decision_type === "send" ? "Имеются основания для подозрения на бешенство (п. 10 Правил, Приказ Минсельхоза РФ № 705). Оформлена заявка Ф-20 для направления материала в лабораторию." : "Основания для подозрения на бешенство отсутствуют. Оформлен обоснованный отказ в выдаче направления."}</td></tr>
                </table>
                <div class="survey-signatures">
                  <div>Ветеринарный врач:<br><br>____________ / ${data.vet_name} /</div>
                  <div>Владелец (с протоколом ознакомлен):<br><br>____________ / ${data.owner_name} /<br>${today}</div>
                </div>
              </div>
            </div>
          `;
    container.innerHTML += surveyPage;
  } catch (err) {
    alert("Ошибка загрузки документа: " + err.message);
  }
}

function downloadWord() {
  if (!currentProtocolId) return alert("Нет активного протокола");
  window.location.href = "/api/protocol-docx/" + currentProtocolId;
}

function closePrint() {
  const printView = document.getElementById("print-view");
  printView.classList.add("hidden");
  printView.style.display = "";
  document.querySelector(".nav").classList.remove("hidden");
  showTab("new-doc");
  document.getElementById("doc-form").reset();
  updateFormView();
}
