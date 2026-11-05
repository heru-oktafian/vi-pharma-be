-- ==============================================================================
-- VI-PHARMA APOTEK: BATCH IMPORT MASTER DATA FARMASI & ALKES
-- File: pharma-be/scripts/import_master_medicines.sql
-- Database: vipharma_db (PostgreSQL 14+)
-- Deskripsi: Script query SQL untuk import kolektif katalog obat & alkes,
--            sekaligus otomatis mendaftarkan satuan dasar (medicine_units)
--            dan konfigurasi harga jual multi-tier (medicine_prices: umum, resep, member).
-- ==============================================================================

BEGIN;

-- ------------------------------------------------------------------------------
-- 1. Buat Tabel Staging Sementara (Temp Table)
-- ------------------------------------------------------------------------------
CREATE TEMP TABLE IF NOT EXISTS temp_master_obat_staging (
    kode_obat           VARCHAR(50),
    barcode             VARCHAR(50),
    nama_obat           VARCHAR(255),
    nama_generik        VARCHAR(255),
    nie_bpom            VARCHAR(100),
    pabrik_produsen     VARCHAR(150),
    bentuk_sediaan      VARCHAR(100),
    golongan_bpom       VARCHAR(50),
    kategori            VARCHAR(150),  -- Bisa berupa Nama Kategori atau ID (misal: CAT000000000001)
    satuan_dasar        VARCHAR(100),  -- Bisa berupa Nama Satuan atau ID (misal: UNT000000000001)
    perlu_resep         VARCHAR(10),   -- Y/N/True/False
    min_stok            INT,
    max_stok            INT,
    lokasi_rak          VARCHAR(100),
    jenis_etiket        VARCHAR(20),   -- 'dalam' (Putih) atau 'luar' (Biru)
    aturan_pakai_signa  VARCHAR(150),
    waktu_minum         VARCHAR(100),
    dosis_pagi          VARCHAR(50),
    dosis_siang         VARCHAR(50),
    dosis_malam         VARCHAR(50),
    peringatan_khusus   TEXT,
    petunjuk_khusus     TEXT,
    bud_hari            INT,
    harga_jual_umum     NUMERIC(15,2),
    harga_jual_resep    NUMERIC(15,2),
    harga_jual_member   NUMERIC(15,2)
) ON COMMIT DROP;

-- ------------------------------------------------------------------------------
-- 2. Input Data Kolektif ke Tabel Staging
--    (Silakan tambahkan / ganti baris data di bawah ini sesuai kebutuhan apotek)
--    Atau jika ingin import dari CSV di terminal VPS:
--    \copy temp_master_obat_staging FROM 'template_master_obat.csv' WITH (FORMAT csv, HEADER true, DELIMITER ',');
-- ------------------------------------------------------------------------------
INSERT INTO temp_master_obat_staging (
    kode_obat, barcode, nama_obat, nama_generik, nie_bpom, pabrik_produsen,
    bentuk_sediaan, golongan_bpom, kategori, satuan_dasar, perlu_resep,
    min_stok, max_stok, lokasi_rak, jenis_etiket, aturan_pakai_signa, waktu_minum,
    dosis_pagi, dosis_siang, dosis_malam, peringatan_khusus, petunjuk_khusus, bud_hari,
    harga_jual_umum, harga_jual_resep, harga_jual_member
) VALUES 
(
    'OBT-00005', '8991234560055', 'Cefixime 200 mg Kapsul', 'Cefixime Trihydrate',
    'GKL1505047801A1', 'PT Dexa Medica', 'Kapsul', 'keras',
    'Antibiotik', 'Kapsul', 'Y',
    20, 500, 'Rak A-02 (Antibiotik)', 'dalam',
    '2 x Sehari 1 Kapsul (Tiap 12 Jam)', 'Sesudah Makan',
    '1', '0', '1', 'Harus Dihabiskan (Antibiotik)',
    'Minum pada jam yang sama tiap 12 jam', 0,
    4500, 5000, 4200
),
(
    'OBT-00006', '8992779110012', 'Sanmol Sirup 120 mg/5 ml (60 ml)', 'Paracetamol',
    'DBL7622204137A1', 'PT Sanbe Farma', 'Sirup', 'bebas',
    'Analgesik & Antipiretik', 'Botol', 'N',
    15, 200, 'Rak B-01 (Sirup Anak)', 'dalam',
    '3 x Sehari 1 Sendok Takar (5 ml)', 'Sesudah Makan (Bila Demam)',
    '1 Cth', '1 Cth', '1 Cth', 'Kocok Dahulu, Hentikan jika demam reda',
    'Gunakan sendok takar yang tersedia', 30,
    22000, 23000, 21000
),
(
    'OBT-00007', '8993214560077', 'Salep Kulit Pi Kang Shuang 5 gr', 'Triamcinolone Acetonide + Miconazole',
    'DTL9123405629A1', 'PT Bina San Prima', 'Salep', 'bebas_terbatas',
    'Obat Luar & Kulit', 'Tube', 'N',
    10, 100, 'Rak C-03 (Salep & Krim)', 'luar',
    'Oles tipis 2-3 x sehari pada kulit gatal/sakit', 'Pagi dan Malam (Setelah Mandi)',
    'Secukupnya', 'Secukupnya', 'Secukupnya', 'Hanya Untuk Obat Luar, Hindari Area Mata',
    'Bersihkan area kulit terlebih dahulu sebelum dioles', 30,
    14500, 15500, 13500
),
(
    'ALK-00001', '8994567890123', 'Kassa Steril Husada 16x16 cm (Isi 10)', 'Kassa Hidrofil Steril',
    'AKD 11603510245', 'PT Husada Farma', 'Pembalut Luka', 'alkes',
    'Obat Luar & Kulit', 'Box', 'N',
    10, 150, 'Rak Alkes D-01', 'luar',
    'Gunakan untuk menutup luka yang telah dibersihkan', 'Saat Ganti Perban',
    '1', '0', '1', 'Sekali Pakai, Jangan Digunakan Jika Kemasan Rusak',
    'Simpan di tempat bersih dan kering', 0,
    12000, 12000, 11000
),
(
    'ALK-00002', '8995678901234', 'Spuit / Jarum Suntik 3 cc OneMed', 'Disposable Syringe 3 ml with Needle',
    'AKD 20902900254', 'PT Jayamas Medica Industri (OneMed)', 'Alat Suntik', 'alkes',
    'Obat Luar & Kulit', 'Pcs', 'N',
    50, 500, 'Rak Alkes D-02', 'luar',
    'Digunakan oleh tenaga medis berwenang', 'Sesuai Tindakan Medis',
    '1', '0', '0', 'Single Use Only (Sekali Pakai), Steril',
    'Buang ke dalam safety box setelah digunakan', 0,
    2500, 2500, 2200
);

-- ------------------------------------------------------------------------------
-- 3. Eksekusi Transformasi & Batch Insert ke Tabel Utama
-- ------------------------------------------------------------------------------
DO $$
DECLARE
    rec RECORD;
    v_med_id VARCHAR(15);
    v_cat_id VARCHAR(15);
    v_unit_id VARCHAR(15);
    v_requires_rx BOOLEAN;
    v_label_type VARCHAR(20);
    v_drug_class drug_classification;
    v_unit_row_id VARCHAR(15);
    v_price_umum_id VARCHAR(15);
    v_price_resep_id VARCHAR(15);
    v_price_member_id VARCHAR(15);
    v_price_resep NUMERIC(15,2);
    v_price_member NUMERIC(15,2);
    v_count_inserted INT := 0;
    v_count_updated INT := 0;
BEGIN
    FOR rec IN SELECT * FROM temp_master_obat_staging LOOP
        -- A. Validasi & Mapping Kategori (Cari berdasarkan ID atau Nama)
        SELECT id INTO v_cat_id 
        FROM categories 
        WHERE id = TRIM(rec.kategori) 
           OR LOWER(TRIM(name)) = LOWER(TRIM(rec.kategori))
        LIMIT 1;

        IF v_cat_id IS NULL THEN
            -- Fallback jika kategori tidak ditemukan: gunakan kategori umum pertama
            SELECT id INTO v_cat_id FROM categories ORDER BY id ASC LIMIT 1;
        END IF;

        -- B. Validasi & Mapping Satuan Dasar (Cari berdasarkan ID, Simbol, atau Nama)
        SELECT id INTO v_unit_id 
        FROM units 
        WHERE id = TRIM(rec.satuan_dasar) 
           OR LOWER(TRIM(name)) = LOWER(TRIM(rec.satuan_dasar))
           OR LOWER(TRIM(symbol)) = LOWER(TRIM(rec.satuan_dasar))
        LIMIT 1;

        IF v_unit_id IS NULL THEN
            -- Fallback jika satuan tidak ditemukan: gunakan Tablet (UNT000000000001)
            SELECT id INTO v_unit_id FROM units ORDER BY id ASC LIMIT 1;
        END IF;

        -- C. Parsing Golongan BPOM
        BEGIN
            v_drug_class := LOWER(TRIM(rec.golongan_bpom))::drug_classification;
        EXCEPTION WHEN OTHERS THEN
            v_drug_class := 'bebas'::drug_classification;
        END;

        -- D. Parsing Perlu Resep
        IF UPPER(TRIM(rec.perlu_resep)) IN ('Y', 'YA', 'TRUE', '1', 'T') 
           OR v_drug_class IN ('keras', 'narkotika', 'psikotropika') THEN
            v_requires_rx := true;
        ELSE
            v_requires_rx := false;
        END IF;

        -- E. Parsing Jenis Etiket
        IF LOWER(TRIM(rec.jenis_etiket)) IN ('luar', 'biru') THEN
            v_label_type := 'luar';
        ELSE
            v_label_type := 'dalam';
        END IF;

        -- F. Cek apakah Medicine dengan ItemCode ini sudah ada
        SELECT id INTO v_med_id FROM medicines WHERE item_code = TRIM(rec.kode_obat);

        IF v_med_id IS NOT NULL THEN
            -- UPDATE data medicine yang sudah ada
            UPDATE medicines SET
                barcode              = COALESCE(NULLIF(TRIM(rec.barcode), ''), barcode),
                name                 = TRIM(rec.nama_obat),
                generic_name         = COALESCE(TRIM(rec.nama_generik), generic_name),
                nie                  = COALESCE(TRIM(rec.nie_bpom), nie),
                manufacturer         = COALESCE(TRIM(rec.pabrik_produsen), manufacturer),
                dosage_form          = COALESCE(TRIM(rec.bentuk_sediaan), dosage_form),
                drug_classification  = v_drug_class,
                category_id          = v_cat_id,
                base_unit_id         = v_unit_id,
                requires_prescription= v_requires_rx,
                min_stock            = COALESCE(rec.min_stok, min_stock, 10),
                max_stock            = COALESCE(rec.max_stok, max_stock, 1000),
                storage_location     = COALESCE(TRIM(rec.lokasi_rak), storage_location),
                label_type           = v_label_type,
                default_signa        = COALESCE(TRIM(rec.aturan_pakai_signa), default_signa),
                consumption_time     = COALESCE(TRIM(rec.waktu_minum), consumption_time),
                dose_morning         = COALESCE(NULLIF(TRIM(rec.dosis_pagi), ''), dose_morning, '1'),
                dose_noon            = COALESCE(NULLIF(TRIM(rec.dosis_siang), ''), dose_noon, '1'),
                dose_night           = COALESCE(NULLIF(TRIM(rec.dosis_malam), ''), dose_night, '1'),
                warning_tags         = COALESCE(TRIM(rec.peringatan_khusus), warning_tags),
                special_instructions = COALESCE(TRIM(rec.petunjuk_khusus), special_instructions),
                bud_days             = COALESCE(rec.bud_hari, bud_days, 0),
                updated_at           = NOW()
            WHERE id = v_med_id;

            v_count_updated := v_count_updated + 1;
        ELSE
            -- INSERT medicine baru dengan ID sequential 15 karakter
            v_med_id := 'MED' || LPAD((
                COALESCE((SELECT MAX(SUBSTRING(id, 4)::bigint) FROM medicines WHERE id ~ '^MED[0-9]+$'), 0) + 1
            )::TEXT, 12, '0');

            INSERT INTO medicines (
                id, item_code, barcode, name, generic_name, nie, manufacturer,
                dosage_form, drug_classification, category_id, base_unit_id,
                requires_prescription, min_stock, max_stock, storage_location,
                is_active, label_type, default_signa, consumption_time,
                dose_morning, dose_noon, dose_night, warning_tags,
                special_instructions, bud_days, created_at, updated_at
            ) VALUES (
                v_med_id,
                TRIM(rec.kode_obat),
                COALESCE(NULLIF(TRIM(rec.barcode), ''), TRIM(rec.kode_obat)),
                TRIM(rec.nama_obat),
                COALESCE(TRIM(rec.nama_generik), '-'),
                COALESCE(TRIM(rec.nie_bpom), '-'),
                COALESCE(TRIM(rec.pabrik_produsen), '-'),
                COALESCE(TRIM(rec.bentuk_sediaan), 'Tablet'),
                v_drug_class,
                v_cat_id,
                v_unit_id,
                v_requires_rx,
                COALESCE(rec.min_stok, 10),
                COALESCE(rec.max_stok, 1000),
                COALESCE(TRIM(rec.lokasi_rak), 'Gudang Utama'),
                true,
                v_label_type,
                COALESCE(TRIM(rec.aturan_pakai_signa), '3 x Sehari 1 Tablet'),
                COALESCE(TRIM(rec.waktu_minum), 'Sesudah Makan'),
                COALESCE(NULLIF(TRIM(rec.dosis_pagi), ''), '1'),
                COALESCE(NULLIF(TRIM(rec.dosis_siang), ''), '1'),
                COALESCE(NULLIF(TRIM(rec.dosis_malam), ''), '1'),
                COALESCE(TRIM(rec.peringatan_khusus), ''),
                COALESCE(TRIM(rec.petunjuk_khusus), ''),
                COALESCE(rec.bud_hari, 0),
                NOW(),
                NOW()
            );

            v_count_inserted := v_count_inserted + 1;
        END IF;

        -- G. Pastikan Satuan Dasar (Base Unit) terdaftar di medicine_units
        IF NOT EXISTS (
            SELECT 1 FROM medicine_units 
            WHERE medicine_id = v_med_id AND unit_id = v_unit_id
        ) THEN
            v_unit_row_id := 'MNU' || LPAD((
                COALESCE((SELECT MAX(SUBSTRING(id, 4)::bigint) FROM medicine_units WHERE id ~ '^MNU[0-9]+$'), 0) + 1
            )::TEXT, 12, '0');

            INSERT INTO medicine_units (
                id, medicine_id, unit_id, conversion_factor, is_base_unit, barcode, created_at, updated_at
            ) VALUES (
                v_unit_row_id,
                v_med_id,
                v_unit_id,
                1,
                true,
                COALESCE(NULLIF(TRIM(rec.barcode), ''), TRIM(rec.kode_obat)),
                NOW(),
                NOW()
            );
        END IF;

        -- H. Setup Konfigurasi Harga Jual (Umum, Resep, Member)
        v_price_resep  := COALESCE(NULLIF(rec.harga_jual_resep, 0), rec.harga_jual_umum);
        v_price_member := COALESCE(NULLIF(rec.harga_jual_member, 0), rec.harga_jual_umum);

        -- H1. Tier Harga UMUM
        IF EXISTS (SELECT 1 FROM medicine_prices WHERE medicine_id = v_med_id AND unit_id = v_unit_id AND price_type = 'umum') THEN
            UPDATE medicine_prices SET price = rec.harga_jual_umum, updated_at = NOW()
            WHERE medicine_id = v_med_id AND unit_id = v_unit_id AND price_type = 'umum';
        ELSE
            v_price_umum_id := 'MPR' || LPAD((
                COALESCE((SELECT MAX(SUBSTRING(id, 4)::bigint) FROM medicine_prices WHERE id ~ '^MPR[0-9]+$'), 0) + 1
            )::TEXT, 12, '0');

            INSERT INTO medicine_prices (id, medicine_id, unit_id, price_type, price, created_at, updated_at)
            VALUES (v_price_umum_id, v_med_id, v_unit_id, 'umum', rec.harga_jual_umum, NOW(), NOW());
        END IF;

        -- H2. Tier Harga RESEP
        IF EXISTS (SELECT 1 FROM medicine_prices WHERE medicine_id = v_med_id AND unit_id = v_unit_id AND price_type = 'resep') THEN
            UPDATE medicine_prices SET price = v_price_resep, updated_at = NOW()
            WHERE medicine_id = v_med_id AND unit_id = v_unit_id AND price_type = 'resep';
        ELSE
            v_price_resep_id := 'MPR' || LPAD((
                COALESCE((SELECT MAX(SUBSTRING(id, 4)::bigint) FROM medicine_prices WHERE id ~ '^MPR[0-9]+$'), 0) + 1
            )::TEXT, 12, '0');

            INSERT INTO medicine_prices (id, medicine_id, unit_id, price_type, price, created_at, updated_at)
            VALUES (v_price_resep_id, v_med_id, v_unit_id, 'resep', v_price_resep, NOW(), NOW());
        END IF;

        -- H3. Tier Harga MEMBER
        IF EXISTS (SELECT 1 FROM medicine_prices WHERE medicine_id = v_med_id AND unit_id = v_unit_id AND price_type = 'member') THEN
            UPDATE medicine_prices SET price = v_price_member, updated_at = NOW()
            WHERE medicine_id = v_med_id AND unit_id = v_unit_id AND price_type = 'member';
        ELSE
            v_price_member_id := 'MPR' || LPAD((
                COALESCE((SELECT MAX(SUBSTRING(id, 4)::bigint) FROM medicine_prices WHERE id ~ '^MPR[0-9]+$'), 0) + 1
            )::TEXT, 12, '0');

            INSERT INTO medicine_prices (id, medicine_id, unit_id, price_type, price, created_at, updated_at)
            VALUES (v_price_member_id, v_med_id, v_unit_id, 'member', v_price_member, NOW(), NOW());
        END IF;

    END LOOP;

    RAISE NOTICE '=== IMPORT MASTER SELESAI ===';
    RAISE NOTICE 'Obat Baru Ditambahkan : % item', v_count_inserted;
    RAISE NOTICE 'Obat Diperbarui       : % item', v_count_updated;
END $$;

COMMIT;

-- ------------------------------------------------------------------------------
-- 4. Verifikasi Hasil Import
-- ------------------------------------------------------------------------------
SELECT 
    m.item_code,
    m.name AS nama_obat,
    m.drug_classification AS golongan,
    c.name AS kategori,
    u.name AS satuan_dasar,
    m.label_type AS etiket,
    pu.price AS harga_umum,
    pr.price AS harga_resep,
    pm.price AS harga_member
FROM medicines m
JOIN categories c ON c.id = m.category_id
JOIN units u ON u.id = m.base_unit_id
LEFT JOIN medicine_prices pu ON pu.medicine_id = m.id AND pu.unit_id = m.base_unit_id AND pu.price_type = 'umum'
LEFT JOIN medicine_prices pr ON pr.medicine_id = m.id AND pr.unit_id = m.base_unit_id AND pr.price_type = 'resep'
LEFT JOIN medicine_prices pm ON pm.medicine_id = m.id AND pm.unit_id = m.base_unit_id AND pm.price_type = 'member'
WHERE m.item_code IN ('OBT-00005', 'OBT-00006', 'OBT-00007', 'ALK-00001', 'ALK-00002')
ORDER BY m.item_code ASC;
