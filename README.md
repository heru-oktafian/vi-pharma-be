# Vi-Pharma Backend Service

Backend service untuk sistem informasi manajemen jaringan apotek multi-cabang **Vi-Pharma**. Dibangun menggunakan **Golang** terbaru dengan arsitektur **Clean Architecture**, **Fiber Web Framework**, **Fiber WebSocket**, dan **GORM** terhubung ke database **PostgreSQL**.

Mendukung alur otentikasi terintegrasi dengan mobile authenticator **`vi-token`** (Flutter) melalui mekanisme **QR Code Pairing** dan pemilihan **Cabang Apotek**.

---

## 🏛️ Struktur Clean Architecture

```text
pharma-be/
├── cmd/
│   └── api/
│       └── main.go                         # Entry point aplikasi, dependency injection & graceful shutdown
├── pkg/
│   ├── config/
│   │   └── config.go                       # Konfigurasi app & database dari environment (.env)
│   └── utils/
│       ├── password.go                     # Bcrypt hashing & password validation
│       ├── jwt.go                          # JWT generator (mendukung claim branch_id & division)
│       └── idgen.go                        # Generator ID 15 karakter (prefix: USR, CBG, UBR)
├── internal/
│   ├── domain/                             # Enterprise Business Rules & Entities
│   │   ├── user.go                         # Entity User, Enum role/status, DTOs
│   │   ├── branch.go                       # Entity Branch (profil apotek lengkap), UserBranch (pivot)
│   │   ├── qr_session.go                   # Model state sesi QR, event WebSocket & DTO
│   │   └── errors.go                       # Custom domain errors
│   ├── repository/                         # Data Access Layer
│   │   ├── memory/
│   │   │   ├── qr_session_store.go         # In-memory store untuk tiket QR (Mutex safe + auto-cleanup)
│   │   │   └── qr_session_store_test.go    # Unit test session store
│   │   └── postgres/
│   │       ├── db.go                       # Inisialisasi DB, connection pool, auto-migrasi ENUM & seeder
│   │       ├── user_repository.go          # Implementasi UserRepository
│   │       └── branch_repository.go        # Implementasi BranchRepository & query cabang per user
│   ├── usecase/                            # Application Business Logic
│   │   ├── auth_usecase.go                 # Login dasar mobile & profil
│   │   ├── qr_auth_usecase.go              # Orkestrator alur QR (Generate, Authorize via WS, Select Branch)
│   │   ├── user_usecase.go                 # Manajemen User
│   │   └── branch_usecase.go               # Manajemen Cabang & penugasan user
│   └── delivery/                           # Interface Adapters Layer (Fiber HTTP & WS)
│       └── http/
│           ├── handler/
│           │   ├── auth_handler.go         # Handler login mobile & user profile
│           │   ├── qr_auth_handler.go      # Handler generate QR, scan authorize, select branch, WS
│           │   ├── user_handler.go         # Handler User CRUD
│           │   ├── branch_handler.go       # Handler Branch CRUD & Assign
│           │   └── response.go             # Standard JSON API response helper
│           ├── middleware/
│           │   ├── auth_middleware.go      # JWT auth, role validation, branch guard (RequireBranch)
│           │   └── logger.go               # Request logging middleware
│           ├── websocket/
│           │   └── hub.go                  # WebSocket Connection Manager
│           └── route/
│               └── router.go               # Registrasi endpoint REST API & WebSocket
├── migrations/
│   ├── 000001_create_user_types_and_table.sql        # Migrasi tabel users & enum
│   └── 000002_create_branches_and_user_branches.sql  # Migrasi tabel branches, user_branches & enum
├── .env.example
├── .env
├── go.mod
├── go.sum
└── README.md
```

---

## 🗄️ Spesifikasi Skema Database

### 1. Tabel `users`
| Kolom | Tipe Data | Nullable | PK | Keterangan |
| :--- | :--- | :--- | :--- | :--- |
| `id` | `varchar(15)` | False | True | Unique 15 karakter (`USRxxxxxxxxxxxx`) |
| `username` | `varchar(255)` | False | False | Unique username |
| `password` | `text` | False | False | Bcrypt hashed password |
| `name` | `varchar(255)` | False | False | Nama lengkap pengguna |
| `user_role` | `user_role` (Type) | False | False | `superadmin`, `admin`, `pharmacist`, `cashier`, `inventory` |
| `user_status` | `user_status` (Type) | False | False | `active`, `inactive`, `suspended` |
| `created_at` | `timestamptz` | False | False | Timestamp pembuatan |
| `updated_at` | `timestamptz` | False | False | Timestamp pembaruan |

### 2. Tabel `branches` (Profil Lengkap Apotek)
| Kolom | Tipe Data | Keterangan |
| :--- | :--- | :--- |
| `id` | `varchar(15)` | Primary Key (`CBGxxxxxxxxxxxx`) |
| `branch_code` | `varchar(50)` | Kode Unik Cabang (contoh: `CBG-PST-01`) |
| `branch_name` | `text` | Nama Cabang Apotek |
| `address` | `text` | Alamat fisik lengkap |
| `phone` | `varchar(100)` | Nomor kontak telepon |
| `email` | `varchar(100)` | Email cabang |
| `sia_id` | `varchar(100)` | Nomor SIA (Surat Izin Apotek) |
| `sia_name` | `varchar(255)` | Nama pada SIA |
| `psa_id` | `varchar(100)` | Nomor PSA (Pemilik Sarana Apotek) |
| `psa_name` | `varchar(255)` | Nama Pemilik Sarana Apotek |
| `sipa` | `varchar(100)` | Nomor SIPA (Surat Izin Praktik Apoteker) |
| `sipa_name` | `varchar(255)` | Nama Apoteker Pemegang SIPA |
| `aping_id` | `varchar(100)` | Nomor SIPA Apoteker Pendamping |
| `aping_name` | `varchar(255)` | Nama Apoteker Pendamping |
| `bank_name` | `varchar(255)` | Nama Bank Rekening Apotek |
| `account_name` | `varchar(255)` | Nama Pemilik Rekening |
| `account_number` | `varchar(100)` | Nomor Rekening Bank |
| `tax_percentage` | `int8` | Persentase Pajak / PPN |
| `journal_method` | `journal_method` (Type) | `perpetual`, `periodic` |
| `branch_status` | `branch_status` (Type) | `active`, `inactive`, `suspended` |
| `license_date` | `timestamptz` | Tanggal masa berlaku izin apotek |
| `default_member` | `varchar(15)` | ID Member default walk-in |
| `subscription_type` | `subscription_type` (Type) | `trial`, `basic`, `pro`, `enterprise` |
| `quota` | `int4` | Kuota data/transaksi |
| `real_asset` | `real_asset` (Type) | `owned`, `rented`, `partnership` |
| `created_at`, `updated_at` | `timestamptz` | Timestamps |

### 3. Tabel `user_branches` (Pivot Relasi Many-to-Many)
| Kolom | Tipe Data | Keterangan |
| :--- | :--- | :--- |
| `id` | `varchar(15)` | Primary Key (`UBRxxxxxxxxxxxx`) |
| `user_id` | `varchar(15)` | Foreign Key ke `users.id` |
| `branch_id` | `varchar(15)` | Foreign Key ke `branches.id` |
| `is_default` | `boolean` | Penanda cabang utama (*home base*) |
| `created_at`, `updated_at` | `timestamptz` | Timestamps |

---

## 🔄 Alur Integrasi: QR Pairing `vi-pharma` (Web) & `vi-token` (Mobile)

```text
[vi-pharma (ReactJS)]         [Backend (Golang)]            [vi-token (Flutter)]
        |                             |                              |
        | 1. Buka Halaman Login       |                              |
        |---> POST /auth/qr/generate -|                              |
        |<--- (session_id, QR text)---|                              |
        |     (Tampilkan QR Code)     |                              |
        |                             |                              |
        | 2. Konek WebSocket          |                              |
        |===> GET /auth/qr/ws/:id ===>|                              |
        |<=== {"event":"CONNECTED"} ==|                              |
        |                             |     3. Login Username/Pass   |
        |                             |<--- POST /auth/login --------|
        |                             |---> (Simpan Mobile JWT) ---->|
        |                             |                              |
        |                             |     4. Scan QR & Otorisasi   |
        |                             |<--- POST /auth/qr/authorize -|
        |                             |     (session_id + Mobile JWT)|
        | 5. Notifikasi Realtime      |     (Cek Cabang Terafiliasi) |
        |<=== {"event":"QR_AUTHORIZED"|                              |
        |     "branches":[...]} ======|                              |
        |                             |                              |
        | 6. Tampilkan List Cabang    |                              |
        |    (User Klik Cabang)       |                              |
        |---> POST /select-branch --->|                              |
        |<--- (Web Token + Cabang) ---|                              |
        | 7. Redirect ke Dashboard!   |                              |
```

---

## 📡 Daftar Endpoint API

### 1. Autentikasi Mobile & QR Pairing
- **Mobile Login (`vi-token`):**
  - `POST /api/v1/auth/login`
  - Body: `{"username": "admin", "password": "admin123"}`
  - Response: Mobile JWT Token.

- **Generate QR (`vi-pharma` Web):**
  - `POST /api/v1/auth/qr/generate`
  - Response: `{"session_id": "<uuid>", "qr_string": "vipharma:qr:<uuid>", "expires_in_seconds": 120}`

- **WebSocket Stream (`vi-pharma` Web):**
  - `GET /api/v1/auth/qr/ws/:sessionId`
  - Realtime Events:
    - `CONNECTED` (Saat browser terhubung)
    - `QR_AUTHORIZED` (Saat mobile berhasil scan, membawa list cabang)
    - `QR_CONSUMED` (Saat user telah memilih cabang)

- **Mobile Authorize (`vi-token`):**
  - `POST /api/v1/auth/qr/authorize`
  - Headers: `Authorization: Bearer <mobile_jwt_token>`
  - Body: `{"session_id": "<uuid>", "device_id": "hp-android-budi"}`

- **Pilih Cabang (`vi-pharma` Web):**
  - `POST /api/v1/auth/qr/select-branch`
  - Body: `{"session_id": "<uuid>", "branch_id": "CBG000000000001"}`
  - Response: Web Session JWT (membawa claim `branch_id`).

### 2. Sesi & Workstation Aktif
- **Get Profile Saat Ini:**
  - `GET /api/v1/auth/me`
  - Headers: `Authorization: Bearer <token>`

- **Verifikasi Workstation Cabang Aktif:**
  - `GET /api/v1/workstation/current`
  - Headers: `Authorization: Bearer <web_token_dengan_branch_id>`
  - Middleware Guard: `RequireBranch()` memastikan token terikat ke cabang aktif.

### 3. Master Data Cabang & User (SuperAdmin / Admin)
- `GET /api/v1/branches` - Daftar cabang (paginated)
- `GET /api/v1/branches/:id` - Detail profil lengkap cabang
- `POST /api/v1/branches` - Tambah cabang apotek baru
- `POST /api/v1/branches/assign` - Menautkan user ke cabang (`user_id`, `branch_id`, `is_default`)
- `GET /api/v1/users` - Daftar user
- `POST /api/v1/users` - Tambah user baru

---

## ⚡ Cara Menjalankan

1. Pastikan PostgreSQL lokal aktif di port `5432` dan database `vipharma_db` sudah dibuat:
   ```sql
   CREATE DATABASE vipharma_db;
   ```
2. Jalankan backend:
   ```bash
   cd pharma-be
   go run ./cmd/api
   ```
3. Seeder otomatis menyiapkan:
   - **Admin:** `username: admin`, `password: admin123` (Role: `superadmin`)
   - **Cabang Default:**
     - `CBG-PST-01` (Apotek Vi-Pharma Pusat)
     - `CBG-SDR-02` (Apotek Vi-Pharma Sudirman)
   - Akun admin otomatis ditautkan ke kedua cabang tersebut.
