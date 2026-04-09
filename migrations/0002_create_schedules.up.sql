CREATE TABLE IF NOT EXISTS schedules (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    schedule_type TEXT NOT NULL, --'daily' | 'monthly' | 'specific_dates' | 'even_odd'

    -- for daily
    day_interval INT,
    start_date DATE,

    -- for monthly
    day_of_month INT,

    -- for specific_dates
    dates DATE[],

    -- for even_odd
    parity TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedule_instances (
    schedule_id BIGINT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    scheduled_date DATE NOT NULL,
    PRIMARY KEY (schedule_id, scheduled_date)
);

CREATE INDEX IF NOT EXISTS idx_schedule_type ON schedules (schedule_type);
CREATE INDEX IF NOT EXISTS idx_schedule_instances_task_id ON schedule_instances (task_id);
