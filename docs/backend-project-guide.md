# Backend Project Guide

Dokumen ini menjelaskan arsitektur, alur kerja, dan struktur backend dari aplikasi Task Manager.

## 1. Gambaran umum

Backend dibuat dengan:
- Go
- Gin Web Framework
- PostgreSQL
- pgx (driver PostgreSQL untuk Go)
- JWT untuk autentikasi
- golang-migrate untuk migrasi database

Tujuan backend adalah menyediakan API yang dipakai oleh frontend untuk:
- autentikasi user
- CRUD board
- CRUD list
- CRUD task
- assign assignee ke task
- label dan task label relation
- membership board
- validasi request dan error handling

---

## 2. Arsitektur aplikasi

Backend mengikuti pola layered architecture:

1. Handler
   - menangani HTTP request
   - menerima payload dari client
   - memanggil service
   - mengembalikan response JSON

2. Service
   - berisi business logic
   - menjalankan validasi dan aturan domain
   - mengintegrasikan repository dan dependency lain

3. Repository
   - komunikasi langsung ke database PostgreSQL
   - query SQL / pgx query
   - akses data berdasarkan kebutuhan domain

4. Model
   - domain model: struktur data domain seperti User, Board, List, Task
   - web model: request/response DTO untuk API

5. Middleware
   - auth middleware
   - security headers
   - rate limiter
   - exception/error handler

6. Migration
   - schema database dikelola via file SQL di folder migration

---

## 3. Struktur folder

```text
task-manager-app/
├── cmd/
│   └── main.go                 # entry point aplikasi
├── docs/
│   └── backend-project-guide.md
├── internal/
│   ├── app/
│   │   └── router.go           # routing semua endpoint
│   ├── config/
│   │   └── config.go           # membaca env dan konfigurasi
│   ├── exception/
│   │   ├── error_handler.go    # global error handling
│   │   ├── conflict_error.go
│   │   ├── not_fround_error.go
│   │   ├── unauthorized_error.go
│   │   └── validation_error.go
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── board_handler.go
│   │   ├── label_handler.go
│   │   ├── list_handler.go
│   │   └── task_handler.go
│   ├── helper/
│   │   └── panic_if_error.go
│   ├── middleware/
│   │   ├── auth_middleware.go
│   │   ├── rate_limiter.go
│   │   └── security_header.go
│   ├── model/
│   │   ├── domain/
│   │   │   ├── board.go
│   │   │   ├── label.go
│   │   │   ├── list.go
│   │   │   ├── task.go
│   │   │   └── user.go
│   │   └── web/
│   │       ├── auth_request.go
│   │       ├── auth_response.go
│   │       ├── board_request.go
│   │       ├── label_request.go
│   │       ├── list_request.go
│   │       ├── task_request.go
│   │       └── web_response.go
│   ├── repository/
│   │   ├── board_member_repository.go
│   │   ├── board_repository.go
│   │   ├── label_repository.go
│   │   ├── list_repository.go
│   │   ├── task_assignee_repository.go
│   │   ├── task_label_repository.go
│   │   ├── task_repository.go
│   │   └── user_repository.go
│   └── service/
│       ├── auth_service.go
│       ├── board_service.go
│       ├── label_service.go
│       ├── list_service.go
│       └── task_service.go
├── migration/
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── tests/
│   ├── auth_test.go
│   ├── label_test.go
│   ├── list_test.go
│   ├── setup_test.go
│   └── task_test.go
├── go.mod
├── docker-compose.yml
├── Dockerfile
└── PRD.md
```

---

## 4. Alur startup aplikasi

Saat aplikasi dijalankan:

1. `cmd/main.go` dipanggil
2. `.env` dibaca (jika ada)
3. konfigurasi dimuat dari environment variable
4. database URL dicek
5. migrasi dijalankan lewat `migrate.Up()`
6. koneksi PostgreSQL dibuat dengan `pgxpool.New()`
7. router dibuat melalui `app.NewRouter(db, cfg)`
8. HTTP server mulai berjalan di port yang ditentukan (`PORT`, default `8080`)

### Contoh konfigurasi environment

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/task_manager
JWT_SECRET=your-secret-key
PORT=8080
TOKEN_EXPIRY_HOURS=24h
FRONTEND_URL=http://localhost:5173
```

---

## 5. Routing dan API group

Semua endpoint backend dipanggil dengan prefix `/api`.

### Public endpoint
Endpoint yang dapat diakses tanpa token:
- `POST /api/auth/register`
- `POST /api/auth/login`

### Protected endpoint
Endpoint yang butuh JWT di header Authorization:

```http
Authorization: Bearer <token>
```

Contoh protected route:
- `GET /api/auth/me`
- `POST /api/boards`
- `GET /api/boards`
- `GET /api/boards/:id`
- `POST /api/boards/:id/lists`
- `POST /api/lists/:id/tasks`
- `POST /api/tasks/:id/labels`

---

## 6. Middleware yang dipakai

### AuthMiddleware
Bertanggung jawab untuk:
- membaca header `Authorization`
- memvalidasi format `Bearer <token>`
- parse JWT
- mengambil `user_id` dari claim token
- menaruh user_id di context request

Jika token invalid atau expired, response akan `401 Unauthorized`.

### SecurityHeaders
Menambahkan header keamanan untuk mencegah beberapa ancaman umum seperti:
- X-Frame-Options
- X-Content-Type-Options
- Referrer-Policy
- lain-lain

### ErrorHandlerMiddleware
Menangkap panic/error yang terjadi di service/handler dan mengubahnya jadi response JSON yang konsisten.

Contoh response umum:

```json
{
  "code": 400,
  "status": "BAD REQUEST",
  "data": "invalid input"
}
```

---

## 7. Fungsi utama backend

### Auth
Bertanggung jawab untuk login, register, dan ambil current user.

Proses:
- user register/login via handler
- service cek email, password, dan validasi
- password di-hash menggunakan bcrypt
- JWT dibuat dengan info `user_id`
- token dikirim ke frontend untuk dipakai di request berikutnya

### Board
Bertanggung jawab untuk:
- membuat board baru
- menampilkan board yang dimiliki / diikuti user
- update board
- delete board
- invite/remove member board

Model board sering melibatkan relasi `board_members` supaya member bisa berkolaborasi.

### List
Bertanggung jawab untuk:
- membuat list dalam board
- menampilkan semua list di board
- reorder list
- update nama list
- delete list

### Task
Bertanggung jawab untuk:
- membuat task di list tertentu
- update task, termasuk judul, deskripsi, status, due date
- move task ke list lain
- delete task
- assign user ke task
- ambil assignee task

### Label
Bertanggung jawab untuk:
- membuat label di board
- menampilkan label per board
- attach label ke task
- detach label dari task
- delete label

---

## 8. Flow request API

Contoh alur saat user login:

1. Frontend kirim `POST /api/auth/login` dengan email + password
2. `AuthHandler.Login` menerima request
3. `AuthService.Login` validasi credential
4. `UserRepository.FindByEmail` mengecek user di database
5. password diverifikasi via bcrypt
6. JWT dibuat
7. response dikirim: `{ token, user }`
8. frontend simpan token dan user ke client state/localStorage

Contoh alur saat user membuat task:

1. Frontend kirim `POST /api/lists/:id/tasks`
2. JWT dicek di `AuthMiddleware`
3. `TaskHandler.Create` mengekstrak user_id dari context
4. `TaskService.Create` validasi input dan access permission
5. `TaskRepository.Save` menulis data ke database
6. response task baru dikirim kembali ke frontend

---

## 9. Repository pattern

Repository dipisah per entitas utama:
- `user_repository.go`
- `board_repository.go`
- `board_member_repository.go`
- `list_repository.go`
- `task_repository.go`
- `task_assignee_repository.go`
- `label_repository.go`
- `task_label_repository.go`

Tujuannya:
- memisahkan query database dari business logic
- memudahkan testing
- membuat kode lebih mudah dipelihara

---

## 10. Model dan response shape

### Web response standar
Semua response API biasanya mengikuti format:

```json
{
  "code": 200,
  "status": "OK",
  "data": { ... }
}
```

Bentuk ini diatur melalui model `web` dan `error handler`.

### Contoh auth response

```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "token": "jwt-token",
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2026-09-21T10:00:00Z"
    }
  }
}
```

---

## 11. Database dan migrasi

Database schema dibuat via migration SQL di folder `migration`.

Fungsi migrasi:
- membuat tabel user
- membuat tabel board
- membuat tabel board_members
- membuat tabel list
- membuat tabel task
- membuat tabel label
- membuat tabel task_labels
- membuat tabel task_assignees

Setiap perubahan schema harus melalui migration agar konsisten di semua environment.

---

## 12. Error handling strategy

Backend menggunakan custom exception types:
- ValidationError
- NotFoundError
- ConflictError
- UnauthorizedError

Semua error tersebut ditangkap di middleware `ErrorHandlerMiddleware` lalu dikonversi ke response JSON yang konsisten.

Contoh:
- duplicate email -> `409 CONFLICT`
- token salah -> `401 UNAUTHORIZED`
- resource tidak ada -> `404 NOT FOUND`
- input invalid -> `400 BAD REQUEST`

---

## 13. Run backend

### Cara jalan normal
```bash
go run ./cmd/main.go
```

### Dengan Docker
```bash
docker-compose up --build
```

### Jalankan test
```bash
go test ./...
```

---

## 14. Kesimpulan

Backend project ini dibuat dengan pendekatan modular dan clean layered architecture. Setiap fitur dipisah berdasarkan tanggung jawab:
- handler untuk HTTP
- service untuk logic bisnis
- repository untuk database
- middleware untuk keamanan
- migration untuk schema

Dengan pola ini, aplikasi lebih mudah dikembangkan, diuji, dan dipelihara ketika fitur baru ditambahkan.
