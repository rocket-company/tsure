INSERT INTO service_orders (
	code,
	customer_name,
	event_name,
	event_city,
	event_date,
	status,
	crew_size,
	vehicle_label,
	inventory_summary,
	total_amount,
	balance_due
)
VALUES
	(
		'OS-2026-001',
		'Casa Aurora Eventos',
		'Feira de Noivas Primavera',
		'Cuiaba',
		CURRENT_DATE + INTERVAL '2 days',
		'agendado',
		6,
		'VUC 3/4 - Placa TSR-2041',
		'Palco modular, 120 cadeiras Tiffany, kit iluminacao cenario',
		18500.00,
		9250.00
	),
	(
		'OS-2026-002',
		'Grupo Pantanal Experience',
		'Convencao de Franqueados',
		'Varzea Grande',
		CURRENT_DATE,
		'em_execucao',
		10,
		'Truck BaU - Placa TSR-8830',
		'Painel LED P3, praticaveis, sonorizacao completa',
		42750.00,
		0.00
	),
	(
		'OS-2026-003',
		'Instituto Terra Viva',
		'Mutirao de Saude Corporativa',
		'Rondonopolis',
		CURRENT_DATE + INTERVAL '5 days',
		'orcamento',
		4,
		'Van Operacional - Placa TSR-1108',
		'Tendas 10x10, climatizadores, mobiliario lounge',
		13200.00,
		13200.00
	);
