# Panduan Import Kolektif Master Data Farmasi & Alkes (Vi-Pharma)

Dokumen ini menjelaskan cara menggunakan file template Excel `.xls` / `.csv` dan query database PostgreSQL untuk menginput atau memperbarui data katalog obat & alkes secara massal (batch import).

---

## 1. Berkas yang Disediakan

| Nama Berkas | Lokasi | Kegunaan |
| :--- | :--- | :--- |
| **`template_master_obat.xls`** | `pharma-be/scripts/template_master_obat.xls` | File spreadsheet Excel (lengkap dengan styling warna, petunjuk kolom, dan sheet referensi). |
| **`template_master_obat.csv`** | `pharma-be/scripts/template_master_obat.csv` | Format CSV murni (UTF-8, pemisah koma) untuk import cepat lewat PostgreSQL `\copy`. |
| **`import_master_medicines.sql`** | `pharma-be/scripts/import_master_medicines.sql` | Query SQL otomatis (staging, lookup kategori & satuan, upsert medicines, units, dan harga jual). |

---

## 2. Struktur Kolom Template

| No | Nama Kolom | Wajib? | Contoh Nilai | Keterangan |
|:---|:---|:---:|:---|:---|
| 1 | `kode_obat` | **Ya** | `OBT-00005`, `ALK-00001` | Kode unik item (Primary Key bisnis). |
| 2 | `barcode` | Tidak | `8991234560055` | Barcode fisik kemasan (jika ada). |
| 3 | `nama_obat` | **Ya** | `Cefixime 200 mg Kapsul` | Nama resmi obat / alkes. |
| 4 | `nama_generik` | Tidak | `Cefixime Trihydrate` | Zat aktif farmakologis. |
| 5 | `nie_bpom` | Tidak | `GKL1505047801A1` | Nomor Izin Edar resmi BPOM / Kemenkes. |
| 6 | `pabrik_produsen`| Tidak | `PT Dexa Medica` | Nama manufaktur / pabrik farmasi. |
| 7 | `bentuk_sediaan` | Tidak | `Kapsul`, `Tablet`, `Sirup`, `Salep` | Bentuk sediaan fisik. |
| 8 | `golongan_bpom` | **Ya** | `keras`, `bebas`, `bebas_terbatas`, `alkes`, `herbal` | Wajib sesuai ENUM regulasi (lihat tabel di bawah). |
| 9 | `kategori` | **Ya** | `Antibiotik` atau `CAT000000000001` | Nama kategori terapi atau ID resmi. |
| 10 | `satuan_dasar` | **Ya** | `Kapsul` atau `UNT000000000004` | Satuan kemasan terkecil (Base Unit). |
| 11 | `perlu_resep` | Tidak | `Y` atau `N` | `Y` untuk wajib resep dokter (otomatis Y jika keras/narkotik). |
| 12 | `min_stok` | Tidak | `20` | Batas peringatan stok menipis (default: 10). |
| 13 | `max_stok` | Tidak | `500` | Batas kapasitas maksimal stok (default: 1000). |
| 14 | `lokasi_rak` | Tidak | `Rak A-02 (Antibiotik)` | Lokasi rak fisik apotek. |
| 15 | `jenis_etiket` | **Ya** | `dalam` (Putih) atau `luar` (Biru) | Jenis etiket dispensing farmasi. |
| 16 | `aturan_pakai_signa` | Tidak | `2 x Sehari 1 Kapsul` | Signa default saat cetak etiket / resep. |
| 17 | `waktu_minum` | Tidak | `Sesudah Makan` | Aturan waktu konsumsi obat. |
| 18 | `dosis_pagi` | Tidak | `1` | Dosis pagi hari (default: '1'). |
| 19 | `dosis_siang` | Tidak | `0` | Dosis siang hari (default: '1'). |
| 20 | `dosis_malam` | Tidak | `1` | Dosis malam hari (default: '1'). |
| 21 | `peringatan_khusus`| Tidak | `Harus Dihabiskan` | Label peringatan tambahan etiket. |
| 22 | `petunjuk_khusus` | Tidak | `Minum tiap 12 jam` | Instruksi edukasi pasien. |
| 23 | `bud_hari` | Tidak | `30` | Beyond Use Date dalam hari (0 = ikut ED batch). |
| 24 | `harga_jual_umum` | **Ya** | `4500` | Tarif jual kasir umum (OTC). |
| 25 | `harga_jual_resep`| Tidak | `5000` | Tarif jual resep (jika kosong, ikut harga umum). |
| 26 | `harga_jual_member`| Tidak| `4200` | Tarif jual member (jika kosong, ikut harga umum). |

---

## 3. Kamus Nilai Referensi

### A. Golongan BPOM (`golongan_bpom`)
- `bebas` : Obat Bebas (Lingkaran Hijau)
- `bebas_terbatas` : Obat Bebas Terbatas (Lingkaran Biru)
- `keras` : Obat Keras / Daftar G (Lingkaran Merah K)
- `narkotika` : Narkotika / Daftar O (Palang Merah)
- `psikotropika` : Obat Psikotropika (SSP)
- `prekursor` : Prekursor Farmasi
- `oot` : Obat-Obat Tertentu
- `alkes` : Alat Kesehatan & BMHP (Kassa, Spuit, dll.)
- `herbal` : Jamu, OHT, Fitofarmaka, Minyak Gosok
- `suplemen` : Suplemen & Multivitamin

### B. Daftar Kategori Terapi Bawaan
| ID Kategori | Nama Kategori |
| :--- | :--- |
| `CAT000000000001` | Antibiotik |
| `CAT000000000002` | Analgesik & Antipiretik |
| `CAT000000000003` | Vitamin & Suplemen |
| `CAT000000000004` | Saluran Cerna |
| `CAT000000000005` | Antihipertensi & Jantung |
| `CAT000000000006` | Antidiabetes |
| `CAT000000000007` | Antihistamin & Alergi |
| `CAT000000000008` | Obat Luar & Kulit |

### C. Daftar Satuan Standar Bawaan
| ID Satuan | Nama Satuan | Simbol |
| :--- | :--- | :--- |
| `UNT000000000001` | Tablet | tab |
| `UNT000000000002` | Strip | str |
| `UNT000000000003` | Box | box |
| `UNT000000000004` | Kapsul | kap |
| `UNT000000000005` | Botol | btl |
| `UNT000000000006` | Tube | tb |
| `UNT000000000007` | Sachet | sct |
| `UNT000000000008` | Ampul | amp |
| `UNT000000000009` | Vial | vial |
| `UNT000000000010` | Pcs | pcs |

---

## 4. Cara Eksekusi Query di Database

### Cara 1: Menjalankan Langsung via Terminal VPS (Paling Cepat)
1. Buka terminal server VPS.
2. Masuk ke direktori backend:
   ```bash
   cd /var/www/vi-pharma/pharma-be
   ```
3. Eksekusi script SQL ke database PostgreSQL:
   ```bash
   sudo -u postgres psql -d vipharma_db -f scripts/import_master_medicines.sql
   ```

### Cara 2: Import dari Berkas CSV
Jika Anda memiliki ratusan baris data di `template_master_obat.csv`:
1. Upload file CSV ke server (misal di `/tmp/template_master_obat.csv`).
2. Di dalam file `scripts/import_master_medicines.sql`, aktifkan perintah `\copy`:
   ```sql
   \copy temp_master_obat_staging FROM '/tmp/template_master_obat.csv' WITH (FORMAT csv, HEADER true, DELIMITER ',');
   ```
3. Jalankan query seperti biasa.

### Cara 3: Menggunakan DBeaver / Navicat / pgAdmin
1. Buka DBeaver / Navicat / pgAdmin dan koneksikan ke database `vipharma_db`.
2. Buka berkas `pharma-be/scripts/import_master_medicines.sql`.
3. Tambahkan baris-baris data Anda di dalam klausa `INSERT INTO temp_master_obat_staging VALUES (...)`.
4. Tekan tombol **Execute Script (F5)**.
5. Periksa pesan output log dan tabel hasil verifikasi di tab bawah.
