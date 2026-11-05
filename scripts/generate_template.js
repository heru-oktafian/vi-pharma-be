const fs = require('fs');
const path = require('path');

// 1. Columns Definition
const columns = [
  { key: 'kode_obat', label: 'KODE OBAT (*)', width: 130, example: 'OBT-00005', desc: 'Wajib, unik (contoh: OBT-00005, ALK-00001)' },
  { key: 'barcode', label: 'BARCODE', width: 130, example: '8991234560055', desc: 'Opsional / barcode kemasan' },
  { key: 'nama_obat', label: 'NAMA OBAT / PRODUK (*)', width: 250, example: 'Cefixime 200 mg Kapsul', desc: 'Wajib, nama dagang lengkap' },
  { key: 'nama_generik', label: 'NAMA GENERIK / KANDUNGAN', width: 200, example: 'Cefixime Trihydrate', desc: 'Zat aktif farmasi' },
  { key: 'nie_bpom', label: 'NIE BPOM / KEMENKES', width: 150, example: 'GKL1505047801A1', desc: 'Nomor Izin Edar resmi' },
  { key: 'pabrik_produsen', label: 'PRODUSEN / PABRIK', width: 170, example: 'PT Dexa Medica', desc: 'Industri farmasi / manufaktur' },
  { key: 'bentuk_sediaan', label: 'BENTUK SEDIAAN', width: 130, example: 'Kapsul', desc: 'Tablet, Kapsul, Sirup, Salep, Cairan, dsb.' },
  { key: 'golongan_bpom', label: 'GOLONGAN BPOM (*)', width: 140, example: 'keras', desc: 'bebas, bebas_terbatas, keras, narkotika, psikotropika, alkes, herbal, suplemen' },
  { key: 'kategori', label: 'KATEGORI TERAPI (*)', width: 180, example: 'Antibiotik', desc: 'Nama kategori atau ID (CAT000000000001 s/d 08)' },
  { key: 'satuan_dasar', label: 'SATUAN DASAR / TERKECIL (*)', width: 160, example: 'Kapsul', desc: 'Nama satuan atau ID (UNT000000000001 s/d 10)' },
  { key: 'perlu_resep', label: 'PERLU RESEP (Y/N)', width: 130, example: 'Y', desc: 'Y atau N (otomatis Y jika keras/narkotik)' },
  { key: 'min_stok', label: 'MIN STOK', width: 90, example: '20', desc: 'Batas peringatan stok menipis (default: 10)' },
  { key: 'max_stok', label: 'MAX STOK', width: 90, example: '500', desc: 'Batas kapasitas maksimal stok (default: 1000)' },
  { key: 'lokasi_rak', label: 'LOKASI RAK / GUDANG', width: 140, example: 'Rak A-02 (Antibiotik)', desc: 'Lokasi penyimpanan rak apotek' },
  { key: 'jenis_etiket', label: 'JENIS ETIKET (*)', width: 120, example: 'dalam', desc: 'dalam (Putih) atau luar (Biru)' },
  { key: 'aturan_pakai_signa', label: 'ATURAN PAKAI / SIGNA', width: 220, example: '2 x Sehari 1 Kapsul (Tiap 12 Jam)', desc: 'Template signa etiket cetak' },
  { key: 'waktu_minum', label: 'WAKTU MINUM', width: 160, example: 'Sesudah Makan', desc: 'Sebelum/Sesudah Makan, Pagi/Malam, dsb.' },
  { key: 'dosis_pagi', label: 'DOSIS PAGI', width: 90, example: '1', desc: 'Dosis pagi (default: 1)' },
  { key: 'dosis_siang', label: 'DOSIS SIANG', width: 90, example: '0', desc: 'Dosis siang (default: 1)' },
  { key: 'dosis_malam', label: 'DOSIS MALAM', width: 90, example: '1', desc: 'Dosis malam (default: 1)' },
  { key: 'peringatan_khusus', label: 'PERINGATAN KHUSUS', width: 200, example: 'Harus Dihabiskan', desc: 'Label peringatan tambahan (pisahkan koma)' },
  { key: 'petunjuk_khusus', label: 'PETUNJUK KHUSUS', width: 220, example: 'Minum pada jam yang sama setiap hari', desc: 'Instruksi edukasi pasien' },
  { key: 'bud_hari', label: 'BUD (HARI)', width: 90, example: '0', desc: 'Beyond Use Date dalam hari (0 = ikut ED batch)' },
  { key: 'harga_jual_umum', label: 'HARGA JUAL UMUM (*)', width: 140, example: '4500', desc: 'Tarif jual kasir reguler OTC' },
  { key: 'harga_jual_resep', label: 'HARGA JUAL RESEP', width: 140, example: '5000', desc: 'Tarif jual resep dokter (opsional)' },
  { key: 'harga_jual_member', label: 'HARGA JUAL MEMBER', width: 140, example: '4200', desc: 'Tarif khusus member/pelanggan setia' }
];

// Sample rows
const sampleRows = [
  {
    kode_obat: 'OBT-00005',
    barcode: '8991234560055',
    nama_obat: 'Cefixime 200 mg Kapsul',
    nama_generik: 'Cefixime Trihydrate',
    nie_bpom: 'GKL1505047801A1',
    pabrik_produsen: 'PT Dexa Medica',
    bentuk_sediaan: 'Kapsul',
    golongan_bpom: 'keras',
    kategori: 'Antibiotik',
    satuan_dasar: 'Kapsul',
    perlu_resep: 'Y',
    min_stok: 20,
    max_stok: 500,
    lokasi_rak: 'Rak A-02 (Antibiotik)',
    jenis_etiket: 'dalam',
    aturan_pakai_signa: '2 x Sehari 1 Kapsul (Tiap 12 Jam)',
    waktu_minum: 'Sesudah Makan',
    dosis_pagi: '1',
    dosis_siang: '0',
    dosis_malam: '1',
    peringatan_khusus: 'Harus Dihabiskan (Antibiotik)',
    petunjuk_khusus: 'Minum pada jam yang sama tiap 12 jam',
    bud_hari: 0,
    harga_jual_umum: 4500,
    harga_jual_resep: 5000,
    harga_jual_member: 4200
  },
  {
    kode_obat: 'OBT-00006',
    barcode: '8992779110012',
    nama_obat: 'Sanmol Sirup 120 mg/5 ml (60 ml)',
    nama_generik: 'Paracetamol',
    nie_bpom: 'DBL7622204137A1',
    pabrik_produsen: 'PT Sanbe Farma',
    bentuk_sediaan: 'Sirup',
    golongan_bpom: 'bebas',
    kategori: 'Analgesik & Antipiretik',
    satuan_dasar: 'Botol',
    perlu_resep: 'N',
    min_stok: 15,
    max_stok: 200,
    lokasi_rak: 'Rak B-01 (Sirup Anak)',
    jenis_etiket: 'dalam',
    aturan_pakai_signa: '3 x Sehari 1 Sendok Takar (5 ml)',
    waktu_minum: 'Sesudah Makan (Bila Demam)',
    dosis_pagi: '1 Cth',
    dosis_siang: '1 Cth',
    dosis_malam: '1 Cth',
    peringatan_khusus: 'Kocok Dahulu, Hentikan jika demam reda',
    petunjuk_khusus: 'Gunakan sendok takar yang tersedia',
    bud_hari: 30,
    harga_jual_umum: 22000,
    harga_jual_resep: 23000,
    harga_jual_member: 21000
  },
  {
    kode_obat: 'OBT-00007',
    barcode: '8993214560077',
    nama_obat: 'Salep Kulit Pi Kang Shuang 5 gr',
    nama_generik: 'Triamcinolone Acetonide + Miconazole',
    nie_bpom: 'DTL9123405629A1',
    pabrik_produsen: 'PT Bina San Prima',
    bentuk_sediaan: 'Salep / Krim',
    golongan_bpom: 'bebas_terbatas',
    kategori: 'Obat Luar & Kulit',
    satuan_dasar: 'Tube',
    perlu_resep: 'N',
    min_stok: 10,
    max_stok: 100,
    lokasi_rak: 'Rak C-03 (Salep & Krim)',
    jenis_etiket: 'luar',
    aturan_pakai_signa: 'Oles tipis 2-3 x sehari pada kulit gatal/sakit',
    waktu_minum: 'Pagi dan Malam (Setelah Mandi)',
    dosis_pagi: 'Secukupnya',
    dosis_siang: 'Secukupnya',
    dosis_malam: 'Secukupnya',
    peringatan_khusus: 'Hanya Untuk Obat Luar, Hindari Area Mata',
    petunjuk_khusus: 'Bersihkan area kulit terlebih dahulu sebelum dioles',
    bud_hari: 30,
    harga_jual_umum: 14500,
    harga_jual_resep: 15500,
    harga_jual_member: 13500
  },
  {
    kode_obat: 'ALK-00001',
    barcode: '8994567890123',
    nama_obat: 'Kassa Steril Husada 16x16 cm (Isi 10)',
    nama_generik: 'Kassa Hidrofil Steril',
    nie_bpom: 'AKD 11603510245',
    pabrik_produsen: 'PT Husada Farma',
    bentuk_sediaan: 'Pembalut Luka',
    golongan_bpom: 'alkes',
    kategori: 'Obat Luar & Kulit',
    satuan_dasar: 'Box',
    perlu_resep: 'N',
    min_stok: 10,
    max_stok: 150,
    lokasi_rak: 'Rak Alkes D-01',
    jenis_etiket: 'luar',
    aturan_pakai_signa: 'Gunakan untuk menutup luka yang telah dibersihkan',
    waktu_minum: 'Saat Ganti Perban',
    dosis_pagi: '1',
    dosis_siang: '0',
    dosis_malam: '1',
    peringatan_khusus: 'Sekali Pakai, Jangan Digunakan Jika Kemasan Rusak',
    petunjuk_khusus: 'Simpan di tempat bersih dan kering',
    bud_hari: 0,
    harga_jual_umum: 12000,
    harga_jual_resep: 12000,
    harga_jual_member: 11000
  },
  {
    kode_obat: 'ALK-00002',
    barcode: '8995678901234',
    nama_obat: 'Spuit / Jarum Suntik 3 cc OneMed',
    nama_generik: 'Disposable Syringe 3 ml with Needle',
    nie_bpom: 'AKD 20902900254',
    pabrik_produsen: 'PT Jayamas Medica Industri (OneMed)',
    bentuk_sediaan: 'Alat Suntik',
    golongan_bpom: 'alkes',
    kategori: 'Obat Luar & Kulit',
    satuan_dasar: 'Pcs',
    perlu_resep: 'N',
    min_stok: 50,
    max_stok: 500,
    lokasi_rak: 'Rak Alkes D-02',
    jenis_etiket: 'luar',
    aturan_pakai_signa: 'Digunakan oleh tenaga medis berwenang',
    waktu_minum: 'Sesuai Tindakan Medis',
    dosis_pagi: '1',
    dosis_siang: '0',
    dosis_malam: '0',
    peringatan_khusus: 'Single Use Only (Sekali Pakai), Steril',
    petunjuk_khusus: 'Buang ke dalam safety box setelah digunakan',
    bud_hari: 0,
    harga_jual_umum: 2500,
    harga_jual_resep: 2500,
    harga_jual_member: 2200
  }
];

// Helper to escape XML
function escapeXml(unsafe) {
  if (unsafe === undefined || unsafe === null) return '';
  return String(unsafe)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

// Helper to escape CSV
function escapeCsv(val) {
  if (val === undefined || val === null) return '';
  const str = String(val);
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return '"' + str.replace(/"/g, '""') + '"';
  }
  return str;
}

// 2. Generate CSV
const csvHeaders = columns.map(c => c.key).join(',');
const csvLines = [csvHeaders];
for (const row of sampleRows) {
  const line = columns.map(c => escapeCsv(row[c.key])).join(',');
  csvLines.push(line);
}
const csvContent = csvLines.join('\r\n');
const csvPath = path.join(__dirname, 'template_master_obat.csv');
fs.writeFileSync(csvPath, csvContent, 'utf8');
console.log('CSV template created:', csvPath);

// 3. Generate SpreadsheetML .xls
const xmlHeader = `<?xml version="1.0"?>
<?mso-application progid="Excel.Sheet"?>
<Workbook xmlns="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:o="urn:schemas-microsoft-com:office:office"
 xmlns:x="urn:schemas-microsoft-com:office:excel"
 xmlns:ss="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:html="http://www.w3.org/TR/REC-html40">
 <DocumentProperties xmlns="urn:schemas-microsoft-com:office:office">
  <Author>Vi-Pharma System</Author>
  <LastAuthor>Vi-Pharma System</LastAuthor>
  <Created>2026-11-05T21:00:00Z</Created>
  <Company>Vi-Pharma Medika</Company>
  <Version>16.00</Version>
 </DocumentProperties>
 <Styles>
  <Style ss:ID="Default" ss:Name="Normal">
   <Alignment ss:Vertical="Center"/>
   <Borders/>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="11" ss:Color="#1E293B"/>
   <Interior/>
   <NumberFormat/>
   <Protection/>
  </Style>
  <Style ss:ID="Title">
   <Alignment ss:Horizontal="Left" ss:Vertical="Center"/>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="16" ss:Bold="1" ss:Color="#0F172A"/>
  </Style>
  <Style ss:ID="Subtitle">
   <Alignment ss:Horizontal="Left" ss:Vertical="Center"/>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="10" ss:Italic="1" ss:Color="#64748B"/>
  </Style>
  <Style ss:ID="HeaderIdent">
   <Alignment ss:Horizontal="Center" ss:Vertical="Center" ss:WrapText="1"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#0369A1"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#0369A1"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#0369A1"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#0369A1"/>
   </Borders>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="10" ss:Color="#FFFFFF" ss:Bold="1"/>
   <Interior ss:Color="#0284C7" ss:Pattern="Solid"/>
  </Style>
  <Style ss:ID="HeaderClass">
   <Alignment ss:Horizontal="Center" ss:Vertical="Center" ss:WrapText="1"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#334155"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#334155"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#334155"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#334155"/>
   </Borders>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="10" ss:Color="#FFFFFF" ss:Bold="1"/>
   <Interior ss:Color="#1E293B" ss:Pattern="Solid"/>
  </Style>
  <Style ss:ID="HeaderEtiket">
   <Alignment ss:Horizontal="Center" ss:Vertical="Center" ss:WrapText="1"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#6D28D9"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#6D28D9"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#6D28D9"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#6D28D9"/>
   </Borders>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="10" ss:Color="#FFFFFF" ss:Bold="1"/>
   <Interior ss:Color="#7C3AED" ss:Pattern="Solid"/>
  </Style>
  <Style ss:ID="HeaderPrice">
   <Alignment ss:Horizontal="Center" ss:Vertical="Center" ss:WrapText="1"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#047857"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#047857"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#047857"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#047857"/>
   </Borders>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="10" ss:Color="#FFFFFF" ss:Bold="1"/>
   <Interior ss:Color="#059669" ss:Pattern="Solid"/>
  </Style>
  <Style ss:ID="HeaderHint">
   <Alignment ss:Horizontal="Left" ss:Vertical="Center" ss:WrapText="1"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#CBD5E1"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#CBD5E1"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#CBD5E1"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#CBD5E1"/>
   </Borders>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="9" ss:Italic="1" ss:Color="#64748B"/>
   <Interior ss:Color="#F8FAFC" ss:Pattern="Solid"/>
  </Style>
  <Style ss:ID="CellText">
   <Alignment ss:Vertical="Center"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
   </Borders>
   <NumberFormat ss:Format="@"/>
  </Style>
  <Style ss:ID="CellCenter">
   <Alignment ss:Horizontal="Center" ss:Vertical="Center"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
   </Borders>
   <NumberFormat ss:Format="@"/>
  </Style>
  <Style ss:ID="CellNumber">
   <Alignment ss:Horizontal="Right" ss:Vertical="Center"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
   </Borders>
   <NumberFormat ss:Format="#,##0"/>
  </Style>
  <Style ss:ID="CellMoney">
   <Alignment ss:Horizontal="Right" ss:Vertical="Center"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#E2E8F0"/>
   </Borders>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="11" ss:Color="#0F172A" ss:Bold="1"/>
   <NumberFormat ss:Format="&quot;Rp &quot;#,##0"/>
  </Style>
  <Style ss:ID="RefHeader">
   <Alignment ss:Horizontal="Center" ss:Vertical="Center"/>
   <Borders>
    <Border ss:Position="Bottom" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#047857"/>
    <Border ss:Position="Left" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#047857"/>
    <Border ss:Position="Right" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#047857"/>
    <Border ss:Position="Top" ss:LineStyle="Continuous" ss:Weight="1" ss:Color="#047857"/>
   </Borders>
   <Font ss:FontName="Calibri" x:Family="Swiss" ss:Size="11" ss:Color="#FFFFFF" ss:Bold="1"/>
   <Interior ss:Color="#047857" ss:Pattern="Solid"/>
  </Style>
 </Styles>
`;

let xmlBody = ` <Worksheet ss:Name="Master Farmasi &amp; Alkes">
  <Table ss:DefaultRowHeight="20">
`;

// Column widths
columns.forEach(col => {
  xmlBody += `   <Column ss:Width="${col.width}"/>\n`;
});

// Row 1: Title
xmlBody += `   <Row ss:Height="26">
    <Cell ss:MergeAcross="${columns.length - 1}" ss:StyleID="Title"><Data ss:Type="String">TEMPLATE IMPORT MASTER DATA FARMASI &amp; ALKES - VI-PHARMA APOTEK</Data></Cell>
   </Row>
   <Row ss:Height="18">
    <Cell ss:MergeAcross="${columns.length - 1}" ss:StyleID="Subtitle"><Data ss:Type="String">Petunjuk: Kolom bertanda (*) WAJIB diisi. Simpan file ini sebagai .CSV atau gunakan query SQL pendamping untuk eksekusi kolektif.</Data></Cell>
   </Row>
`;

// Row 3: Header Labels
xmlBody += `   <Row ss:Height="28">\n`;
columns.forEach((col, idx) => {
  let style = 'HeaderClass';
  if (idx < 6) style = 'HeaderIdent';
  else if (idx >= 6 && idx <= 13) style = 'HeaderClass';
  else if (idx >= 14 && idx <= 22) style = 'HeaderEtiket';
  else style = 'HeaderPrice';
  xmlBody += `    <Cell ss:StyleID="${style}"><Data ss:Type="String">${escapeXml(col.label)}</Data></Cell>\n`;
});
xmlBody += `   </Row>\n`;

// Row 4: Column Descriptions / Hints
xmlBody += `   <Row ss:Height="24">\n`;
columns.forEach(col => {
  xmlBody += `    <Cell ss:StyleID="HeaderHint"><Data ss:Type="String">${escapeXml(col.desc)}</Data></Cell>\n`;
});
xmlBody += `   </Row>\n`;

// Rows 5+: Sample Rows
sampleRows.forEach(row => {
  xmlBody += `   <Row ss:Height="22">\n`;
  columns.forEach(col => {
    const val = row[col.key];
    const isMoney = col.key.startsWith('harga_');
    const isNum = ['min_stok', 'max_stok', 'bud_hari'].includes(col.key);
    const isCenter = ['kode_obat', 'barcode', 'golongan_bpom', 'perlu_resep', 'jenis_etiket', 'dosis_pagi', 'dosis_siang', 'dosis_malam'].includes(col.key);

    if (isMoney) {
      xmlBody += `    <Cell ss:StyleID="CellMoney"><Data ss:Type="Number">${val}</Data></Cell>\n`;
    } else if (isNum) {
      xmlBody += `    <Cell ss:StyleID="CellNumber"><Data ss:Type="Number">${val}</Data></Cell>\n`;
    } else if (isCenter) {
      xmlBody += `    <Cell ss:StyleID="CellCenter"><Data ss:Type="String">${escapeXml(val)}</Data></Cell>\n`;
    } else {
      xmlBody += `    <Cell ss:StyleID="CellText"><Data ss:Type="String">${escapeXml(val)}</Data></Cell>\n`;
    }
  });
  xmlBody += `   </Row>\n`;
});

xmlBody += `  </Table>
  <WorksheetOptions xmlns="urn:schemas-microsoft-com:office:excel">
   <FreezePanes/>
   <FrozenNoSplit/>
   <SplitPane>
    <PaneNumber>3</PaneNumber>
    <RowNumber>4</RowNumber>
   </SplitPane>
   <ActivePane>2</ActivePane>
  </WorksheetOptions>
 </Worksheet>
`;

// Reference Sheet
xmlBody += ` <Worksheet ss:Name="Daftar Referensi Sistem">
  <Table ss:DefaultRowHeight="20">
   <Column ss:Width="140"/>
   <Column ss:Width="200"/>
   <Column ss:Width="260"/>
   <Row ss:Height="24">
    <Cell ss:MergeAcross="2" ss:StyleID="Title"><Data ss:Type="String">DAFTAR KODE REFERENSI MASTER (DATABASE VI-PHARMA)</Data></Cell>
   </Row>
   <Row ss:Height="10"></Row>

   <!-- Golongan BPOM -->
   <Row ss:Height="24">
    <Cell ss:MergeAcross="2" ss:StyleID="RefHeader"><Data ss:Type="String">1. GOLONGAN BPOM (Pilihan Valid di Kolom golongan_bpom)</Data></Cell>
   </Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">bebas</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Obat Bebas</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Lingkaran Hijau Garis Tepi Hitam (OTC)</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">bebas_terbatas</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Obat Bebas Terbatas</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Lingkaran Biru Garis Tepi Hitam (P1-P6)</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">keras</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Obat Keras (Daftar G)</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Lingkaran Merah Huruf K (Resep Dokter)</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">narkotika</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Narkotika (Daftar O)</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Palang Merah Medali (Regulasi Ketat)</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">psikotropika</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Psikotropika</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Obat Penenang / SSP (Resep Dokter)</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">alkes</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Alat Kesehatan &amp; BMHP</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Kassa, Spuit, Plester, Masker, Handscoon</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">herbal</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Obat Tradisional / Jamu / OHT</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Minyak Telon, Minyak Gosok, Tolak Angin, dll</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">suplemen</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Suplemen Makanan / Vitamin</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Ester C, Blackmores, Enervon C, dll</Data></Cell></Row>
   <Row ss:Height="10"></Row>

   <!-- Kategori Terapi -->
   <Row ss:Height="24">
    <Cell ss:MergeAcross="2" ss:StyleID="RefHeader"><Data ss:Type="String">2. KATEGORI TERAPI (Bisa ketik Nama atau ID)</Data></Cell>
   </Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">CAT000000000001</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Antibiotik</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Obat pembunuh atau penghambat bakteri</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">CAT000000000002</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Analgesik &amp; Antipiretik</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Pereda nyeri dan penurun demam</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">CAT000000000003</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Vitamin &amp; Suplemen</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Multivitamin dan daya tahan tubuh</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">CAT000000000004</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Saluran Cerna</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Antasida, anti-maag, dan pencernaan</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">CAT000000000005</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Antihipertensi &amp; Jantung</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Pengontrol tekanan darah dan jantung</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">CAT000000000006</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Antidiabetes</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Pengontrol gula darah</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">CAT000000000007</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Antihistamin &amp; Alergi</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Pereda alergi dan gatal</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">CAT000000000008</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Obat Luar &amp; Kulit</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Salep, krim antijamur, antiseptik, alkes luar</Data></Cell></Row>
   <Row ss:Height="10"></Row>

   <!-- Satuan Dasar -->
   <Row ss:Height="24">
    <Cell ss:MergeAcross="2" ss:StyleID="RefHeader"><Data ss:Type="String">3. SATUAN DASAR / KEMASAN TERKECIL (Bisa ketik Nama atau ID)</Data></Cell>
   </Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000001</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Tablet</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Satuan sediaan padat tablet</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000002</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Strip</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Kemasan blister / strip</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000003</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Box</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Kemasan kotak / dus</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000004</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Kapsul</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Satuan sediaan kapsul</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000005</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Botol</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Kemasan botol sirup / cairan</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000006</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Tube</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Kemasan tube salep / gel / krim</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000007</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Sachet</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Kemasan sachet serbuk / granul</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000008</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Ampul</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Sediaan injeksi ampul</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000009</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Vial</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Sediaan injeksi vial</Data></Cell></Row>
   <Row><Cell ss:StyleID="CellCenter"><Data ss:Type="String">UNT000000000010</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Pcs</Data></Cell><Cell ss:StyleID="CellText"><Data ss:Type="String">Satuan pieces alat kesehatan / BMHP</Data></Cell></Row>
  </Table>
 </Worksheet>
`;

const fullXml = xmlHeader + xmlBody + '</Workbook>\n';
const xlsPath = path.join(__dirname, 'template_master_obat.xls');
fs.writeFileSync(xlsPath, fullXml, 'utf8');
console.log('XLS template created:', xlsPath);
