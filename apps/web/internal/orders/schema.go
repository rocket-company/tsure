package orders

// bootstrapSchema cria as tabelas usadas pela camada legada de ordens de
// servico (service_orders + relacionadas). O modelo canonico do ERP vive
// em database/schema.sql; estas tabelas v0 coexistem para nao quebrar a
// UI atual e serao gradualmente migradas para o modelo novo.
const bootstrapSchema = `
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
	inventory_summary TEXT NOT NULL DEFAULT '',
	total_amount NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
	balance_due NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (balance_due >= 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE service_orders ADD COLUMN IF NOT EXISTS install_date DATE;
ALTER TABLE service_orders ADD COLUMN IF NOT EXISTS return_date DATE;
ALTER TABLE service_orders ALTER COLUMN inventory_summary SET DEFAULT '';

CREATE TABLE IF NOT EXISTS service_order_items (
	id BIGSERIAL PRIMARY KEY,
	service_order_id BIGINT NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
	equipment_id BIGINT NOT NULL,
	quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_order_items_so ON service_order_items(service_order_id);

CREATE TABLE IF NOT EXISTS service_order_checklist (
	id BIGSERIAL PRIMARY KEY,
	service_order_id BIGINT NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
	equipment_id BIGINT NOT NULL,
	status TEXT NOT NULL CHECK (status IN ('ok','avariado','perdido','nao_retornado')),
	notes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (service_order_id, equipment_id)
);

CREATE INDEX IF NOT EXISTS idx_service_order_checklist_so ON service_order_checklist(service_order_id);

CREATE TABLE IF NOT EXISTS service_order_charges (
	id BIGSERIAL PRIMARY KEY,
	service_order_id BIGINT NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
	equipment_id BIGINT,
	charge_type TEXT NOT NULL CHECK (charge_type IN ('avaria','perda','despesa','multa')),
	amount NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
	description TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_order_charges_so ON service_order_charges(service_order_id);
`
