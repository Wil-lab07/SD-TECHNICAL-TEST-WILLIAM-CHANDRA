# Technical Test Documentation — Software Developer (AYO Indonesia 2026)

Dokumentasi utama untuk hasil pengerjaan **Technical Test Software Developer PT Ayo Indonesia Utama**.

---

## 🗺️ Lokasi Jawaban & Panduan Soal

Proyek ini terbagi menjadi beberapa bagian sesuai dengan nomor soal:

### 1. SOAL 1: Backend API (Football Team Management)
* **Lokasi Kode Program:** Folder [`backend/`](file:///Users/williamchandra/Documents/VSCode/SD-TECHNICAL-TEST-WILLIAM-CHANDRA/backend)
* **Panduan Step-by-Step Menjalankan Server & Database:**  
  Seluruh langkah-langkah rinci (pembuatan `.env`, setup database `football_db`, migrasi tabel, cara run server, serta cara login admin & penggunaan JWT token) **dapat dibaca langsung pada file README backend:**  
  👉 **[Lihat Dokumentasi & Panduan Backend di `backend/README.md`](file:///Users/williamchandra/Documents/VSCode/SD-TECHNICAL-TEST-WILLIAM-CHANDRA/backend/README.md)**

---

### 2. SOAL 2: Analisis Teknikal & Peningkatan Sistem AYO Indonesia
* **Dokumentasi Jawaban:**  
  👉 [`docs/2/1.md`](file:///Users/williamchandra/Documents/VSCode/SD-TECHNICAL-TEST-WILLIAM-CHANDRA/docs/2/1.md) dan [`JAWABAN_SOAL_2.md`](file:///Users/williamchandra/Documents/VSCode/SD-TECHNICAL-TEST-WILLIAM-CHANDRA/JAWABAN_SOAL_2.md)
* **Konteks:** Analisis mendalam mengenai *input validation bug*, *rate limiting*, *unbounded pagination*, *caching*, trade-off eksposur API Key Google Maps, serta rekomendasi arsitektur mobile (offline state, real-time WebSocket, BFF over-fetching).

---

### 3. SOAL 3: Analisis Tambahan & Studi Kasus Arsitektur
* **Dokumentasi Jawaban:**  
  👉 Folder [`docs/3/`](file:///Users/williamchandra/Documents/VSCode/SD-TECHNICAL-TEST-WILLIAM-CHANDRA/docs/3)

---

## 📂 Struktur Utama Repository

```
.
├── README.md                    # Dokumentasi utama proyek
├── JAWABAN_SOAL_2.md            # Dokumentasi Jawaban Soal 2
├── backend/                     # [SOAL 1] Implementation Code (Go API)
│   ├── README.md                # 👈 Panduan lengkap cara jalankan server & DB
│   ├── cmd/api/main.go          # Entrypoint server Go
│   ├── db/migrations/           # File SQL Migration
│   ├── docs/                    # PRD & Postman Collection JSON
│   ├── internal/                # Logic modul (auth, teams, players, matches)
│   ├── .env.example             # Template environment variables
│   └── docker-compose.yml       # PostgreSQL Docker Container
└── docs/                        # [SOAL 2 & 3] Dokumentasi Jawaban Soal
    ├── 2/
    │   └── 1.md                 # Analisis Teknikal Sistem AYO (Soal 2)
    └── 3/                       # Dokumentasi Jawaban Soal 3
```
