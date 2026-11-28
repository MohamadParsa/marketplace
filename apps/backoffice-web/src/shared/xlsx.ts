// Minimal OOXML workbook with UTF-8 inline-string cells in a standard ZIP package.
// Cells never contain formulas, external relationships or macros.
export async function xlsxBlob(
  headers: string[],
  rows: string[][],
  signal?: AbortSignal,
): Promise<Blob> {
  const pause = async () => {
    if (signal?.aborted) throw Error("خروجی لغو شد.");
    await new Promise((resolve) => setTimeout(resolve, 0));
  };
  if (rows.length > 10000 || headers.length > 100 || headers.length === 0)
    throw Error("حداکثر ۱۰هزار ردیف و صد ستون در هر خروجی مجاز است.");
  const xml = (s: string) =>
    s
      .replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f￾￿]/g, "")
      .slice(0, 32767)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  const col = (n: number) => {
    let s = "";
    for (n++; n; n = Math.floor((n - 1) / 26))
      s = String.fromCharCode(65 + ((n - 1) % 26)) + s;
    return s;
  };
  let sheet = "",
    characters = 0;
  const dataRows = [headers, ...rows];
  for (let i = 0; i < dataRows.length; i++) {
    const row = dataRows[i];
    if (row.length !== headers.length)
      throw Error("ستون‌های خروجی ناسازگارند.");
    for (const value of row) {
      if (value.length > 32767)
        throw Error("متن یک سلول بیش از حد مجاز Excel است.");
      characters += value.length;
    }
    if (characters > 5000000)
      throw Error("حجم خروجی زیاد است؛ فیلتر را محدودتر کنید.");
    sheet += `<row r="${i + 1}">${row.map((value, j) => `<c r="${col(j)}${i + 1}" t="inlineStr"><is><t xml:space="preserve">${xml(value)}</t></is></c>`).join("")}</row>`;
    if (i % 100 === 0) await pause();
  }
  const ns = "http://schemas.openxmlformats.org/";
  const files: Record<string, string> = {
    "[Content_Types].xml": `<Types xmlns="${ns}package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
    "_rels/.rels": `<Relationships xmlns="${ns}package/2006/relationships"><Relationship Id="rId1" Type="${ns}officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
    "xl/workbook.xml": `<workbook xmlns="${ns}spreadsheetml/2006/main" xmlns:r="${ns}officeDocument/2006/relationships"><sheets><sheet name="گزارش" sheetId="1" r:id="rId1"/></sheets></workbook>`,
    "xl/_rels/workbook.xml.rels": `<Relationships xmlns="${ns}package/2006/relationships"><Relationship Id="rId1" Type="${ns}officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
    "xl/worksheets/sheet1.xml": `<worksheet xmlns="${ns}spreadsheetml/2006/main"><sheetViews><sheetView workbookViewId="0" rightToLeft="1"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews><sheetData>${sheet}</sheetData><autoFilter ref="A1:${col(headers.length - 1)}${rows.length + 1}"/></worksheet>`,
  };
  const encoder = new TextEncoder(),
    parts: Uint8Array[] = [],
    central: Uint8Array[] = [];
  let offset = 0;
  const put = (length: number) => {
    const b = new Uint8Array(length);
    return { b, v: new DataView(b.buffer) };
  };
  for (const [path, content] of Object.entries(files)) {
    const name = encoder.encode(path),
      data = encoder.encode(
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>' + content,
      );
    if (data.length > 20 * 1024 * 1024)
      throw Error(
        "حجم فایل خروجی بیش از ۲۰ مگابایت است؛ فیلتر را محدودتر کنید.",
      );
    await pause();
    let crc = 0xffffffff;
    const table = new Uint32Array(256);
    for (let i = 0; i < 256; i++) {
      let value = i;
      for (let bit = 0; bit < 8; bit++)
        value = (value >>> 1) ^ (value & 1 ? 0xedb88320 : 0);
      table[i] = value;
    }
    for (let i = 0; i < data.length; i++) {
      crc = (crc >>> 8) ^ table[(crc ^ data[i]) & 255];
      if (i > 0 && i % 500000 === 0) await pause();
    }
    crc = (crc ^ 0xffffffff) >>> 0;
    const h = put(30 + name.length);
    h.v.setUint32(0, 0x04034b50, true);
    h.v.setUint16(4, 20, true);
    h.v.setUint16(6, 0x800, true);
    h.v.setUint16(12, 0x21, true);
    h.v.setUint32(14, crc, true);
    h.v.setUint32(18, data.length, true);
    h.v.setUint32(22, data.length, true);
    h.v.setUint16(26, name.length, true);
    h.b.set(name, 30);
    parts.push(h.b, data);
    const c = put(46 + name.length);
    c.v.setUint32(0, 0x02014b50, true);
    c.v.setUint16(4, 20, true);
    c.v.setUint16(6, 20, true);
    c.v.setUint16(8, 0x800, true);
    c.v.setUint16(14, 0x21, true);
    c.v.setUint32(16, crc, true);
    c.v.setUint32(20, data.length, true);
    c.v.setUint32(24, data.length, true);
    c.v.setUint16(28, name.length, true);
    c.v.setUint32(42, offset, true);
    c.b.set(name, 46);
    central.push(c.b);
    offset += h.b.length + data.length;
  }
  const centralSize = central.reduce((n, b) => n + b.length, 0),
    end = put(22);
  end.v.setUint32(0, 0x06054b50, true);
  end.v.setUint16(8, central.length, true);
  end.v.setUint16(10, central.length, true);
  end.v.setUint32(12, centralSize, true);
  end.v.setUint32(16, offset, true);
  if (signal?.aborted) throw Error("خروجی لغو شد.");
  return new Blob(
    [...parts, ...central, end.b].map((b) => new Uint8Array(b).buffer),
    {
      type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    },
  );
}
