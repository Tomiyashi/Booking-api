CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    capacity INT NOT NULL
);

CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY,
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(room_id, start_time)
);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    slot_id UUID REFERENCES slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(slot_id) WHERE (status = 'active')
);

CREATE INDEX IF NOT EXISTS idx_slots_time ON slots(room_id, start_time);