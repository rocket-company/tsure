package budgets

const schema = `
CREATE TABLE IF NOT EXISTS budgets (
	id BIGSERIAL PRIMARY KEY,
	code TEXT NOT NULL UNIQUE,
	customer_name TEXT NOT NULL,
	event_name TEXT NOT NULL,
	event_city TEXT NOT NULL,
	event_date DATE NOT NULL,
	install_date DATE,
	return_date DATE,
	status TEXT NOT NULL DEFAULT 'rascunho' CHECK (status IN ('rascunho','em_analise','aprovado','recusado','expirado','convertido')),
	crew_size INTEGER NOT NULL DEFAULT 1 CHECK (crew_size > 0),
	vehicle_label TEXT NOT NULL DEFAULT '',
	total_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
	valid_until DATE,
	notes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS budget_items (
	id BIGSERIAL PRIMARY KEY,
	budget_id BIGINT NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
	equipment_id BIGINT,
	kit_id BIGINT,
	quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
	unit_price NUMERIC(12, 2) NOT NULL DEFAULT 0,
	total_price NUMERIC(12, 2) NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_budget_items_budget ON budget_items(budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_status ON budgets(status);
`
