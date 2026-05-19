CREATE TABLE IF NOT EXISTS service_orders (
	id BIGSERIAL PRIMARY KEY,
	code TEXT NOT NULL UNIQUE,
	customer_name TEXT NOT NULL,
	event_name TEXT NOT NULL,
	event_city TEXT NOT NULL,
	event_date DATE NOT NULL,
	status TEXT NOT NULL CHECK (
		status IN (
			'orcamento',
			'agendado',
			'em_execucao',
			'aguardando_retorno',
			'finalizado'
		)
	),
	crew_size INTEGER NOT NULL DEFAULT 1 CHECK (crew_size > 0),
	vehicle_label TEXT NOT NULL,
	inventory_summary TEXT NOT NULL,
	total_amount NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
	balance_due NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (balance_due >= 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
