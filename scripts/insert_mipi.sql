-- SQL Script: Insert Item MIPI Minyak Gosok 60 ml & Batch Fisik (ED 2029)
-- Database: vipharma_db

BEGIN;

-- 1. Insert Katalog Obat (jika belum ada)
INSERT INTO medicines (
    id, item_code, barcode, name, generic_name, nie, manufacturer,
    dosage_form, drug_classification, category_id, base_unit_id,
    requires_prescription, min_stock, max_stock, storage_location,
    default_supplier_id, is_active, label_type, default_signa,
    consumption_time, dose_morning, dose_noon, dose_night,
    warning_tags, special_instructions, bud_days, created_at, updated_at
) VALUES (
    'MED000000000004',
    'OBT-00004',
    '8993430174989',
    'MIPI Minyak Gosok Pereda Nyeri Otot 60 ml',
    'Methyl Salicylate, Menthol, Camphor, Capsicum Oleoresin, Clove Oil',
    'POM QD 061 606 731',
    'PT Etercon Pharma (Novell)',
    'Cairan Obat Luar',
    'herbal',
    'CAT000000000008',
    'UNT000000000005',
    false,
    5,
    100,
    'Rak Minyak Gosok A-01',
    'SUP000000000001',
    true,
    'luar',
    'Dioleskan & dipijat 10-15 menit pada area nyeri/pegal',
    'Bila Perlu / Saat Nyeri',
    'Secukupnya',
    'Secukupnya',
    'Secukupnya',
    'Hanya Untuk Obat Luar,Hindari Area Mata dan Luka Terbuka,Tidak Untuk Anak < 2 Tahun',
    'Simpan pada suhu di bawah 30°C. Jangan digunakan langsung di bawah lubang hidung.',
    90,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 2. Insert Konversi Satuan (1 Botol = 1 Base Unit)
INSERT INTO medicine_units (
    id, medicine_id, unit_id, conversion_factor, is_base_unit, barcode, created_at, updated_at
) VALUES (
    'MNU000000000010',
    'MED000000000004',
    'UNT000000000005',
    1,
    true,
    '8993430174989',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 3. Insert Konfigurasi Harga Jual (Umum, Resep, Member)
INSERT INTO medicine_prices (
    id, medicine_id, unit_id, price_type, price, created_at, updated_at
) VALUES 
    ('MPR000000000010', 'MED000000000004', 'UNT000000000005', 'umum', 24500, NOW(), NOW()),
    ('MPR000000000011', 'MED000000000004', 'UNT000000000005', 'resep', 25000, NOW(), NOW()),
    ('MPR000000000012', 'MED000000000004', 'UNT000000000005', 'member', 23000, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 4. Insert Batch Fisik Aktif (No. Batch EFG146, Expired Date: 2029-07-31)
INSERT INTO medicine_batches (
    id, medicine_id, branch_id, batch_number, expired_date,
    qty_warehouse, qty_shelf, total_stock, buy_price, is_active, created_at, updated_at
) VALUES (
    'BTC000000000004',
    'MED000000000004',
    'CBG000000000001',
    'EFG146',
    '2029-07-31',
    5.00,
    10.00,
    15.00,
    19500.00,
    true,
    NOW(),
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    expired_date = '2029-07-31',
    qty_shelf = 10.00,
    total_stock = 15.00;

-- 5. Insert Initial Stock Card Audit Trail (Mutations)
INSERT INTO stock_mutations (
    id, branch_id, medicine_id, batch_id, mutation_type, location, reference_no,
    qty_in, qty_out, balance_after, total_after, notes, user_id, created_at
) VALUES 
    ('MUT000000000001', 'CBG000000000001', 'MED000000000004', 'BTC000000000004', 'PEMBELIAN', 'GUDANG', 'SALDO-AWAL-001', 5.00, 0.00, 5.00, 5.00, 'Stok Awal Fisik Gudang Belakang (Batch EFG146)', 'USR000000000001', NOW()),
    ('MUT000000000002', 'CBG000000000001', 'MED000000000004', 'BTC000000000004', 'PEMBELIAN', 'RAK', 'SALDO-AWAL-001', 10.00, 0.00, 10.00, 15.00, 'Stok Awal Fisik Rak Penjualan (Batch EFG146)', 'USR000000000001', NOW())
ON CONFLICT (id) DO NOTHING;

COMMIT;
