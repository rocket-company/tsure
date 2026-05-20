package inventory

const schema = `
CREATE TABLE IF NOT EXISTS equipment (
	id BIGSERIAL PRIMARY KEY,
	code TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	category TEXT NOT NULL DEFAULT 'geral',
	quantity_total INTEGER NOT NULL DEFAULT 1 CHECK (quantity_total > 0),
	status TEXT NOT NULL DEFAULT 'disponivel' CHECK (status IN ('disponivel','em_manutencao','descartado')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kits (
	id BIGSERIAL PRIMARY KEY,
	code TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kit_items (
	kit_id BIGINT NOT NULL REFERENCES kits(id) ON DELETE CASCADE,
	equipment_id BIGINT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
	quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
	PRIMARY KEY (kit_id, equipment_id)
);

CREATE TABLE IF NOT EXISTS equipment_reservations (
	id BIGSERIAL PRIMARY KEY,
	equipment_id BIGINT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
	service_order_id BIGINT NOT NULL,
	reserved_from DATE NOT NULL,
	reserved_until DATE NOT NULL,
	quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
	status TEXT NOT NULL DEFAULT 'reservado' CHECK (status IN ('reservado','em_campo','aguardando_retorno','retornado','cancelado')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reservations_equipment ON equipment_reservations(equipment_id);
CREATE INDEX IF NOT EXISTS idx_reservations_dates ON equipment_reservations(reserved_from, reserved_until);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON equipment_reservations(status);
CREATE INDEX IF NOT EXISTS idx_equipment_status ON equipment(status);
`
