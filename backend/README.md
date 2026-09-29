# Football Team Management Backend API

> **Technical Test: PT Ayo Indonesia Utama (2026)**  
> Production-grade RESTful API yang dibangun menggunakan **Go (Golang)**, **Gin Framework**, **PostgreSQL (pgx v5)**, dan **Cloudinary**.

---

## 📋 Panduan Setup & Menjalankan Aplikasi

Berikut adalah langkah-langkah lengkap untuk mengonfigurasi, menginisialisasi database, dan menjalankan server backend:

### 1. Konfigurasi Environment Variables (`.env`)
Salin file `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```
> File `.env` sudah dilengkapi dengan konfigurasi default yang siap langsung dipakai (Port: `8080`, Database: `football_db`).

#### ☁️ Cara Mendapatkan Kredensial Cloudinary API (Untuk Upload Logo Tim):

![alt text](image.png)

1. Login ke [Cloudinary Console](https://console.cloudinary.com).
2. Buka menu **Settings** -> **Product environment settings** -> **API Keys**.
3. Ambil nilai **Cloud name**, **API Key**, dan **API Secret** Anda, lalu masukkan ke file `.env`:
   ```env
   CLOUDINARY_CLOUD_NAME=lsmxmhrm
   CLOUDINARY_API_KEY=391269266273676
   CLOUDINARY_API_SECRET=<your_api_secret_here>
   ```

---

### 2. Menyiapkan & Menyalakan Database PostgreSQL

* **Menggunakan Docker Compose (Sangat Direkomendasikan):**
  ```bash
  docker-compose up -d
  ```
  *Penjelasan:* Container PostgreSQL akan berjalan secara otomatis pada port `5432` dan membuat database bernama `football_db`.

* **Menggunakan PostgreSQL Lokal (Tanpa Docker):**
  Buat database manual melalui psql / pgAdmin:
  ```sql
  CREATE DATABASE football_db;
  ```
  Lalu pastikan `DB_USER` dan `DB_PASSWORD` di `.env` disesuaikan dengan kredensial PostgreSQL lokal Anda.

---

### 3. Menjalankan Database Migrations (Membuat Tabel)

Jalankan perintah berikut untuk membuat seluruh skema tabel di database:

```bash
make migrate-up
```

*Jika tidak menggunakan `make`, gunakan CLI `golang-migrate`:*
```bash
migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/football_db?sslmode=disable" up
```

---

### 4. Menjalankan Server Backend Go

Jalankan server aplikasi:

```bash
make run
```
*Atau jalankan langsung dengan perintah Go:*
```bash
go run cmd/api/main.go
```

Jika server berhasil berjalan, log berikut akan muncul di terminal:
```json
{"level":"INFO","msg":"connected to database successfully"}
{"level":"INFO","msg":"token cleanup worker started"}
{"level":"INFO","msg":"server starting","addr":":8080"}
```

---

### 🔑 5. Kredensial & Cara Login Admin Initial (Auto-Seed)

Pada saat server pertama kali dinyalakan, sistem akan secara otomatis membuat akun **Admin Utama** di database jika tabel `admins` masih kosong.

* **Email Admin:** `admin@ayo.test`
* **Password Admin:** `ayoindonesiamaju123`

#### Tembak Endpoint Login via Postman / cURL:
* **Method:** `POST`
* **URL:** `http://localhost:8080/api/v1/auth/login`
* **Body (JSON):**
  ```json
  {
    "email": "admin@ayo.test",
    "password": "ayoindonesiamaju123"
  }
  ```

#### Penggunaan Token JWT:
Salin nilai `token` dari response login, lalu masukkan ke dalam Header setiap request selanjutnya:
```http
Authorization: Bearer <TOKEN_JWT_HASIL_LOGIN>
```

---

## 🏛️ Arsitektur & Prinsip Utama Kode

```
cmd/api/main.go ──> internal/app ──> middleware (JWT Auth & Blacklist)
                                        │
┌───────────────────────────────────────┴───────────────────────────────────────┐
│                                                                               │
▼                                                                               ▼
Handlers (Gin) ──> Services (Business Logic) ──> Repositories (Raw SQL pgxpool) ──> PostgreSQL
                                              └─> External Storage Provider       ──> Cloudinary
```

1. **Arsitektur 4-Layer Ketat:**
   - `handler`: Melakukan parsing request HTTP, validasi UUID & payload DTO, memanggil layer service, dan mengembalikan format JSON standar.
   - `service`: Berisi seluruh logika bisnis & aturan PRD (BR-1 s.d. BR-15). **Bebas dari impor Gin & statement SQL.**
   - `repository`: Mengeksekusi query raw SQL terparameterisasi dengan `pgxpool`.
   - `provider`: Layanan pihak ketiga (Cloudinary storage).
2. **Stateless JWT + Revocation Blacklist:**
   - Setiap token dilengkapi `jti` unik (UUID).
   - Saat logout, `jti` disimpan di tabel `revoked_tokens`.
   - Middleware auth mengecek keabsahan token & memastikan `jti` tidak masuk daftar revocation list.
   - Background worker membersihkan token expired secara berkala.
3. **Transaksi Database Atomik:**
   - Pelaporan hasil pertandingan dijalankan di dalam transaksi PostgreSQL (`tx.Begin` & `defer tx.Rollback`) untuk menjamin konsistensi skor dan statistik gol.

---

## 🧪 Menjalankan Unit Testing

Uji seluruh package dengan race detector aktif:

```bash
go test ./... -race
```

---

## 📋 Ringkasan API Endpoints (19 Endpoints)

### 🔑 Autentikasi
| Method | Endpoint | Deskripsi | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | Login & dapatkan token JWT | Tidak |
| `POST` | `/api/v1/auth/logout` | Logout & me-revoke token JWT | Ya |
| `GET` | `/api/v1/auth/me` | Profil admin yang sedang login | Ya |

### 🛡️ Tim (Teams)
| Method | Endpoint | Deskripsi | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/teams` | Buat tim baru (JSON / Upload Logo) | Ya |
| `GET` | `/api/v1/teams` | Daftar tim aktif (Paginated) | Ya |
| `GET` | `/api/v1/teams/:id` | Detail tim | Ya |
| `PUT` | `/api/v1/teams/:id` | Update data tim | Ya |
| `DELETE` | `/api/v1/teams/:id` | Soft delete tim | Ya |

### 🏃 Pemain (Players)
| Method | Endpoint | Deskripsi | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/players` | Tambah pemain (Validasi nomor punggung unik per tim) | Ya |
| `GET` | `/api/v1/players` | Daftar pemain (Bisa difilter `team_id`) | Ya |
| `GET` | `/api/v1/players/:id` | Detail pemain | Ya |
| `PUT` | `/api/v1/players/:id` | Update data pemain | Ya |
| `DELETE` | `/api/v1/players/:id` | Soft delete pemain | Ya |

### ⚽ Pertandingan (Matches & Reports)
| Method | Endpoint | Deskripsi | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/matches` | Jadwalkan pertandingan 2 tim | Ya |
| `GET` | `/api/v1/matches` | Daftar pertandingan (Paginated) | Ya |
| `GET` | `/api/v1/matches/:id` | Detail pertandingan | Ya |
| `POST` | `/api/v1/matches/:id/result` | Laporkan hasil tanding (Skor + Pencetak Gol) | Ya |
| `PUT` | `/api/v1/matches/:id/result` | Update hasil tanding (Atomik) | Ya |
| `GET` | `/api/v1/matches/:id/report` | Laporan pertandingan (Top scorer & total kemenangan) | Ya |

---

## 📬 Postman Collection
Impor file Postman Collection dari path berikut:  
[`docs/postman_collection.json`](file:///Users/williamchandra/Documents/VSCode/SD-TECHNICAL-TEST-WILLIAM-CHANDRA/backend/docs/postman_collection.json)
