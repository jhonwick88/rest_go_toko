# TokoPintar REST API (rest_go_toko)

Backend REST API berperforma tinggi dan berbobot ringan untuk ekosistem aplikasi kasir **TokoPintar (POS)**. Dibangun menggunakan bahasa **Go (Golang)** dengan framework **Gin** dan engine database **SQLite murni (Pure Go / CGO-Free)**.

Proyek ini menyediakan layanan terpusat untuk katalog inventaris, transaksi penjualan (POS), pembukuan kas kasir, manajemen pengguna, audit log, serta proteksi lisensi enterprise berbasis perangkat keras (Hardware Fingerprint).

---

## 🚀 Fitur Utama (Features)

### 1. 🔐 Sistem Lisensi Enterprise & Anti-Tamper (Hardware-Bound Licensing)
* **Hardware ID Fingerprinting**: Lisensi diikat secara permanen dengan identitas perangkat keras (`machineid.ProtectedID`) untuk mencegah duplikasi atau pembajakan aplikasi ke mesin lain.
* **Offline Cryptographic Token Verification**: Memvalidasi token lisensi (`license.token`) bertanda tangan kriptografis secara instan dan aman bahkan tanpa koneksi internet.
* **Aktivasi Otomatis & Trial License**: Mendukung aktivasi lisensi resmi maupun aktivasi masa percobaan (*14-day trial*) langsung via API (`/api/license/trial`) untuk produk `TP-POS`.
* **Feature Tiering / Gate Middleware**: Membatasi akses endpoint secara granular berdasarkan paket lisensi aktif (misalnya fitur **PRO** seperti *Supplier Management* dan *Stock Ledger*).
* **Global License Shield**: Middleware proteksi global yang otomatis mengunci akses operasional jika lisensi tidak sah, kadaluarsa, atau diubah secara ilegal.

---

### 2. 📦 Manajemen Produk, Barcode, & Multi-Satuan
* **Katalog & Kategori Produk**: Pengelolaan kategori (`ITEM_CATEGORY`) dan master produk (`ITEM`) secara lengkap (CRUD).
* **Pencarian Cepat & Barcode Scanner**: Endpoint khusus pencarian teks instan (*case-insensitive*) dan lookup barcode untuk integrasi scanner kasir.
* **Manajemen Multi-Satuan (Units)**: Fleksibilitas satuan penjualan (Pcs, Dus, Pack, Lusin, dll).
* **Penyesuaian Stok Cepat (Stock Adjustment)**: Update stok langsung via endpoint `PATCH /api/items/:itemno/stock`.
* **Item Cepat / Quick Access Items**: Konfigurasi produk favorit untuk mempercepat transaksi kasir di layar sentuh POS.

---

### 3. 💳 Transaksi Penjualan & Kasir (POS Sales)
* **Penerbitan Faktur & Penjualan**: Pencatatan transaksi penjualan secara atomik dengan kalkulasi subtotal, diskon, dan total tagihan.
* **Pelacakan Status Faktur**: Pembaruan status pembayaran transaksi (lunas, pending, void).
* **Riwayat Transaksi**: Pengambilan data riwayat transaksi penjualan dengan filter dan pagination efisien.

---

### 4. 💼 Rekonsiliasi Kas & Shift Kasir (Cash Reconciliation)
* **Buka & Tutup Shift**: Pencatatan saldo kas awal kasir (*opening cash*).
* **Hitung Kas Fisik & Selisih**: Perhitungan otomatis antara penerimaan sistem dengan uang fisik kasir di laci uang (*cash drawer*).
* **Pelacakan Discrepancy**: Mendeteksi selisih lebih (*overage*) atau kurang (*shortage*) per shift.

---

### 5. ⭐ Fitur Eksklusif PRO (Tiered Features)
* **Supplier Management**: Kelola data pemasok barang dagangan (memerlukan fitur `supplier_management`).
* **Buku Besar Pergerakan Stok (Stock Ledger)**: Lacak histori alur masuk dan keluar barang secara detail per nomor item (memerlukan fitur `stock_movement`).

---

### 6. 🛡️ Keamanan, Audit Trail, & Pemeliharaan Sistem
* **Manajemen Pengguna Kasir**: Kelola akun kasir dan hak akses.
* **Audit Logging**: Jejak audit komprehensif untuk setiap aktivitas penting pada sistem.
* **Profil Perusahaan / Toko**: Kelola data nama toko, alamat, telepon, dan pesan struk belanja.
* **Backup & Restore Database**: Endpoint instan untuk mencadangkan dan memulihkan file database secara langsung.

---

### 7. ⚡ Performa & Arsitektur Backend
* **Pure Go SQLite (CGO-Free)**: Menggunakan driver modern `modernc.org/sqlite` tanpa ketergantungan DLL C atau kompilator CGO, sangat portabel untuk deployment Windows & Linux.
* **SQLite WAL & Busy Timeout**: Konfigurasi `PRAGMA journal_mode=WAL` dan `PRAGMA busy_timeout=5000` untuk performa konkurensi baca/tulis tinggi tanpa deadlock.
* **Koneksi Pooling Otomatis**: Manajemen *idle connection* dan *connection lifetime* terkelola optimal.
* **Paginasi Server-Side (Limit & Offset)**: Paginasi data berkecepatan tinggi untuk dataset produk yang besar.
* **CORS Middleware**: Siap dikonsumsi langsung oleh client multi-platform (Flutter Desktop TokoPintar, Mobile, atau Web).
* **Graceful Shutdown**: Menangkap sinyal OS (`SIGINT`, `SIGTERM`) untuk menyelesaikan request berjalan sebelum server mati dengan aman.
* **Format Respons JSON Terstandar**:
  ```json
  {
    "success": true,
    "message": "Operasi berhasil",
    "data": [...]
  }
  ```

---

## 📂 Struktur Proyek

```text
rest_go_toko/
├── main.go               # Entrypoint aplikasi & konfigurasi Graceful Shutdown
├── go.mod / go.sum       # Manajemen dependensi Golang
├── .env                  # Variabel lingkungan runtime (konfigurasi lokal)
├── .env.example          # Template konfigurasi environment
├── config/
│   └── config.go         # Loader konfigurasi sistem (godotenv)
├── database/
│   └── database.go       # Inisialisasi pool SQLite (WAL, Foreign Keys, Pool limits)
├── middleware/
│   ├── cors.go           # Penanganan CORS request
│   └── license.go        # Middleware validasi lisensi & fitur (RequireLicense, RequireFeature)
├── models/
│   ├── category.go       # Model data kategori
│   ├── item.go           # Model master barang & stok
│   ├── sale.go           # Model transaksi penjualan & faktur
│   ├── reconciliation.go # Model rekonsiliasi kas kasir
│   └── response.go       # Struktur standar respons JSON & pagination
├── handlers/
│   ├── license_handler.go        # Handler aktivasi, trial, dan status lisensi
│   ├── item_handler.go           # Handler master barang, barcode, pencarian, dan stok
│   ├── category_handler.go       # Handler kategori barang
│   ├── sale_handler.go           # Handler faktur dan transaksi POS
│   ├── reconciliation_handler.go # Handler buka-tutup kas kasir
│   ├── quick_item_handler.go     # Handler item favorit / shortcut POS
│   ├── supplier_handler.go       # Handler supplier (PRO)
│   ├── stock_ledger_handler.go   # Handler buku besar stok (PRO)
│   ├── user_handler.go           # Handler pengguna & kasir
│   ├── audit_handler.go          # Handler catatan audit sistem
│   ├── company_handler.go        # Handler profil toko/perusahaan
│   └── system_handler.go         # Handler backup & restore database
├── routes/
│   └── routes.go         # Pendaftaran seluruh rute API dan proteksi middleware
└── services/
    └── license_service.go # Core engine verifikasi token lisensi, fingerprint & validasi kriptografi
```

---

## 🛠️ Panduan Penggunaan & Setup

### 1. Prasyarat
* **Go 1.22+** terinstal di sistem
* Database SQLite TokoPintar (`.db` atau `.sqlite`)

### 2. Konfigurasi Environment (`.env`)
Salin berkas `.env.example` menjadi `.env`:
```env
# Port server REST API
SERVER_PORT=8181

# Path file database SQLite lokal
DB_PATH=H:\AMAN\tokopintar.db
```

### 3. Instalasi Dependensi
```bash
go mod tidy
```

### 4. Menjalankan Server
```bash
go run main.go
```

### 5. Kompilasi Binary (Produksi / Windows `.exe`)
```bash
go build -ldflags="-s -w" -o rest_go_toko.exe main.go
```
Jalankan file biner hasil kompilasi:
```bash
./rest_go_toko.exe
```

---

## 🌐 Ringkasan Endpoint API

### Lisensi (Unprotected)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `POST` | `/api/license/activate` | Aktivasi lisensi menggunakan License Key |
| `POST` | `/api/license/trial` | Permintaan aktivasi lisensi Trial (14 Hari) |
| `GET`  | `/api/license/status` | Pengecekan status lisensi mesin saat ini |

### Master Data & Inventori (Protected)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET`    | `/api/categories` | Daftar semua kategori |
| `POST`   | `/api/categories` | Tambah kategori baru |
| `GET`    | `/api/items` | Ambil barang berpaginasi (`?page=1&limit=50`) |
| `GET`    | `/api/items/search` | Cari barang (`?q=mie&page=1&limit=50`) |
| `GET`    | `/api/items/barcode/:barcode` | Cari barang berdasarkan barcode |
| `GET`    | `/api/items/:itemno` | Ambil detail satu barang |
| `PATCH`  | `/api/items/:itemno/stock` | Penyesuaian jumlah stok barang |
| `GET`    | `/api/units` | Daftar satuan barang |
| `GET`    | `/api/quick-items` | Daftar item cepat / shortcut kasir |

### Transaksi & Kasir (Protected)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET`   | `/api/sales` | Riwayat transaksi penjualan |
| `POST`  | `/api/sales` | Simpan transaksi penjualan baru |
| `PATCH` | `/api/sales/:invoiceno/status` | Update status faktur |
| `GET`   | `/api/cash-reconciliations` | Riwayat buka/tutup kasir |
| `POST`  | `/api/cash-reconciliations` | Simpan rekonsiliasi kas kasir |

### Fitur Eksklusif PRO (Protected by Feature Tier)
| Method | Endpoint | Fitur yang Diperlukan | Deskripsi |
|---|---|---|---|
| `GET`  | `/api/suppliers` | `supplier_management` | Ambil data daftar supplier |
| `POST` | `/api/suppliers` | `supplier_management` | Tambah supplier baru |
| `GET`  | `/api/stock-ledger/:itemno` | `stock_movement` | Kartu riwayat pergerakan stok barang |

### Sistem & Pengaturan (Protected)
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET`  | `/api/company` | Ambil profil identitas toko |
| `PUT`  | `/api/company` | Update profil toko |
| `GET`  | `/api/audit-logs` | Ambil riwayat audit log |
| `GET`  | `/api/backup` | Download cadangan database |
| `POST` | `/api/restore` | Pulihkan database dari backup |
