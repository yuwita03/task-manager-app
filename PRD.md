# PRD — Task Manager App (Backend)

## Tech Stack

- Go + Gin (web framework)
- PostgreSQL + pgx (raw SQL, tanpa ORM)
- golang-migrate — migration dijalankan via kode di `main.go`, bukan CLI manual
- godotenv — load `.env`
- JWT (golang-jwt/jwt/v5) + bcrypt — auth
- go-playground/validator — validasi request
- Panic-based error handling — repository/service panic, ditangkap `exception.ErrorHandlerMiddleware()` sebagai Gin middleware
- Arsitektur: **Layered — Handler → Service → Repository** (bukan Clean Architecture literal; repository berbentuk interface, testable via mock)
- Testing: `httptest` + testify, integration test dengan database real (`task_manager_test`, di-truncate tiap test)
- Docker + Docker Compose, GitHub Actions CI
- Deferred v2: gorilla/websocket (real-time sync)

## Struktur Folder

```
task-manager-app/
├── cmd/main.go                    — entry: load env, migration, start server
├── internal/
│   ├── app/router.go              — semua wiring repo→service→handler + route, dipakai main.go & test
│   ├── config/config.go
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/domain/              — struct tabel DB
│   ├── model/web/                 — request/response API
│   ├── middleware/                — AuthMiddleware (JWT check)
│   └── exception/                 — typed errors + panic recovery middleware
├── migration/
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── tests/
├── docker-compose.yml / Dockerfile
├── .github/workflows/ci.yml
└── .env / .env.example
```

## Schema

```sql
CREATE TABLE users (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE boards (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    owner_id        INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_boards_owner_id ON boards(owner_id);

CREATE TABLE board_members (
    board_id        INT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    user_id         INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL DEFAULT 'member', -- 'owner' | 'member'
    joined_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (board_id, user_id)
);

CREATE TABLE lists (
    id              SERIAL PRIMARY KEY,
    board_id        INT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    position        INT NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_lists_board_id ON lists(board_id);

CREATE TABLE tasks (
    id              SERIAL PRIMARY KEY,
    list_id         INT NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    title           VARCHAR(200) NOT NULL,
    description     TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'todo', -- kolom independen, bukan turunan list_id
    position        INT NOT NULL,
    due_date        TIMESTAMP,
    created_by      INT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tasks_list_id ON tasks(list_id);

CREATE TABLE task_assignees (
    task_id         INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id         INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, user_id)
);
CREATE INDEX idx_task_assignees_user_id ON task_assignees(user_id);

CREATE TABLE labels (
    id              SERIAL PRIMARY KEY,
    board_id        INT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name            VARCHAR(50) NOT NULL,
    color           VARCHAR(7) NOT NULL
);

CREATE TABLE task_labels (
    task_id         INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label_id        INT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, label_id)
);
```

Design notes: board multi-user via `board_members` (owner ikut tercatat sebagai row dengan role='owner'); task multi-assignee via `task_assignees`; `tasks.status` independen dari `list_id`. Primary key & UNIQUE constraint otomatis ter-index; foreign key tidak otomatis ter-index di PostgreSQL, jadi ditambah manual di atas.

## API Endpoints

### Auth
| Method | Endpoint |
|---|---|
| POST | /api/auth/register |
| POST | /api/auth/login |
| GET | /api/auth/me *(protected)* |

### Boards
| Method | Endpoint |
|---|---|
| POST | /api/boards |
| GET | /api/boards *(list milik user)* |
| GET | /api/boards/:id |
| PUT | /api/boards/:id |
| DELETE | /api/boards/:id |

### Board Members
| Method | Endpoint |
|---|---|
| GET | /api/boards/:id/members |
| POST | /api/boards/:id/members *(invite by email)* |
| DELETE | /api/boards/:id/members/:userId |

### Lists
| Method | Endpoint |
|---|---|
| POST | /api/boards/:id/lists |
| GET | /api/boards/:id/lists |
| PUT | /api/lists/:id |
| PUT | /api/lists/:id/reorder |
| DELETE | /api/lists/:id |

**Catatan routing:** param board pakai `:id` (bukan `:boardId`) — Gin tidak mengizinkan 2 nama wildcard berbeda di segment path yang sama level dengan route Board (`/boards/:id`).

### Tasks
| Method | Endpoint |
|---|---|
| POST | /api/lists/:id/tasks |
| GET | /api/lists/:id/tasks |
| GET | /api/tasks/:id |
| PUT | /api/tasks/:id |
| DELETE | /api/tasks/:id |
| PUT | /api/tasks/:id/move *(pindah list + posisi, drag-drop)* |

### Task Assignees
| Method | Endpoint |
|---|---|
| POST | /api/tasks/:id/assignees |
| DELETE | /api/tasks/:id/assignees/:userId |

### Labels
| Method | Endpoint |
|---|---|
| GET | /api/boards/:id/labels |
| POST | /api/boards/:id/labels |
| DELETE | /api/labels/:id |
| POST | /api/tasks/:id/labels |
| DELETE | /api/tasks/:id/labels/:labelId |

Semua endpoint kecuali auth wajib JWT middleware.

## Authorization Rules

Opsi yang perlu diputuskan:
- Semua member bebas CRUD list/task/label, cuma owner yang boleh hapus board & invite/remove member
- Semua member cuma boleh create/edit, delete dibatasi owner
- Full sama rata (member = owner) kecuali hapus board

`BoardMemberRepository.IsMember(ctx, boardId, userId)` sudah tersedia untuk dipakai di service layer begitu rule-nya diputuskan.


## Infra

- `Dockerfile` (multi-stage build, base image versi Go harus disamakan dengan `go.mod`) + `docker-compose.yml` (app + PostgreSQL, healthcheck)
- `.github/workflows/ci.yml` — spin up service Postgres, apply migration via `psql -f`, run test, build

## Testing

Pattern: `httptest.NewRecorder()` + real DB (`task_manager_test`, truncate tiap test run) — integration test, bukan mock. `internal/app/router.go` di-extract supaya `main.go` dan test pakai wiring yang sama.