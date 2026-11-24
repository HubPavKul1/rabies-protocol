const Docxtemplater = require("docxtemplater");
const PizZip = require("pizzip");
const fs = require("fs");

const templatePath = process.argv[2];

// Читаем шаблон в бинарном виде
const content = fs.readFileSync(templatePath, "binary");
const zip = new PizZip(content);

const doc = new Docxtemplater(zip, {
  paragraphLoop: true,
  linebreaks: true,
});

let inputData = "";
process.stdin.setEncoding("utf8");

process.stdin.on("data", (chunk) => {
  inputData += chunk;
});

process.stdin.on("end", () => {
  try {
    const data = JSON.parse(inputData);

    // Заменяем плейсхолдеры в шаблоне данными
    doc.render(data);

    // Генерируем буфер и выводим в stdout
    const buf = doc.getZip().generate({ type: "nodebuffer" });
    process.stdout.write(buf);
  } catch (error) {
    console.error("Docxtemplater error:", error);
    process.exit(1);
  }
});
