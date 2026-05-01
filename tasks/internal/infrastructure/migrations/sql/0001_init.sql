CREATE TABLE IF NOT EXISTS tasks (
	id BIGINT PRIMARY KEY,
	title VARCHAR(255) NOT NULL,
	description TEXT,
	performer_id BIGINT NOT NULL,
	creator_id BIGINT NOT NULL,
	status VARCHAR(50) NOT NULL DEFAULT 'pending',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tasks_performer_id ON tasks (performer_id);
CREATE INDEX IF NOT EXISTS idx_tasks_creator_id ON tasks (creator_id);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks (deleted_at);

CREATE TABLE IF NOT EXISTS observers (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL,
	task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_observers_user_id ON observers (user_id);
CREATE INDEX IF NOT EXISTS idx_observers_task_id ON observers (task_id);

CREATE TABLE IF NOT EXISTS id_allocator (
	id INT PRIMARY KEY,
	next_id BIGINT NOT NULL
);
