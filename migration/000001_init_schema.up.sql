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
    role            VARCHAR(20) NOT NULL DEFAULT 'member',
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
    status          VARCHAR(20) NOT NULL DEFAULT 'todo',
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