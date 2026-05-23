package maps

// Cidade representa uma cidade vinculada a uma UF.
type Cidade struct {
	Nome string
	UF   string
}

// CidadesMT contem as cidades de Mato Grosso (UF de operacao da empresa).
// Para outras UFs, consulte fontes externas (IBGE). Esta lista foi extraida
// de database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabCidade.csv.
var CidadesMT = []Cidade{
	{"Acorizal", "MT"}, {"Agua Boa", "MT"}, {"Alta Floresta", "MT"},
	{"Alto Araguaia", "MT"}, {"Alto da Boa Vista", "MT"}, {"Alto Garcas", "MT"},
	{"Alto Paraguai", "MT"}, {"Alto Taquari", "MT"}, {"Apiacas", "MT"},
	{"Araguainha", "MT"}, {"Araputanga", "MT"}, {"Arenapolis", "MT"},
	{"Aripuana", "MT"}, {"Barao de Melgaco", "MT"}, {"Barra do Bugres", "MT"},
	{"Barra do Garcas", "MT"}, {"Bom Jesus do Araguaia", "MT"}, {"Brasnorte", "MT"},
	{"Caceres", "MT"}, {"Campinapolis", "MT"}, {"Campo Novo do Parecis", "MT"},
	{"Campo Verde", "MT"}, {"Campos de Julio", "MT"}, {"Canabrava do Norte", "MT"},
	{"Canarana", "MT"}, {"Carlinda", "MT"}, {"Castanheira", "MT"},
	{"Chapada dos Guimaraes", "MT"}, {"Claudia", "MT"}, {"Cocalinho", "MT"},
	{"Colider", "MT"}, {"Colniza", "MT"}, {"Comodoro", "MT"},
	{"Confresa", "MT"}, {"Conquista d'Oeste", "MT"}, {"Cuiaba", "MT"},
	{"Curvelandia", "MT"}, {"Denise", "MT"}, {"Diamantino", "MT"},
	{"Dom Aquino", "MT"}, {"Feliz Natal", "MT"}, {"Figueiropolis d'Oeste", "MT"},
	{"Gaucha do Norte", "MT"}, {"General Carneiro", "MT"}, {"Gloria d'Oeste", "MT"},
	{"Guaranta do Norte", "MT"}, {"Guiratinga", "MT"}, {"Indiavai", "MT"},
	{"Ipiranga do Norte", "MT"}, {"Itanhanga", "MT"}, {"Itauba", "MT"},
	{"Itiquira", "MT"}, {"Jaciara", "MT"}, {"Jangada", "MT"},
	{"Jauru", "MT"}, {"Juara", "MT"}, {"Juina", "MT"},
	{"Juruena", "MT"}, {"Juscimeira", "MT"}, {"Lambari d'Oeste", "MT"},
	{"Lucas do Rio Verde", "MT"}, {"Luciara", "MT"}, {"Marcelandia", "MT"},
	{"Matupa", "MT"}, {"Mirassol d'Oeste", "MT"}, {"Nobres", "MT"},
	{"Nortelandia", "MT"}, {"Nossa Senhora do Livramento", "MT"},
	{"Nova Bandeirantes", "MT"}, {"Nova Brasilandia", "MT"},
	{"Nova Canaa do Norte", "MT"}, {"Nova Guarita", "MT"}, {"Nova Lacerda", "MT"},
	{"Nova Marilandia", "MT"}, {"Nova Maringa", "MT"}, {"Nova Monte Verde", "MT"},
	{"Nova Mutum", "MT"}, {"Nova Nazare", "MT"}, {"Nova Olimpia", "MT"},
	{"Nova Santa Helena", "MT"}, {"Nova Ubirata", "MT"}, {"Nova Xavantina", "MT"},
	{"Novo Horizonte do Norte", "MT"}, {"Novo Mundo", "MT"},
	{"Novo Santo Antonio", "MT"}, {"Novo Sao Joaquim", "MT"}, {"Paranaita", "MT"},
	{"Paranatinga", "MT"}, {"Pedra Preta", "MT"}, {"Peixoto de Azevedo", "MT"},
	{"Planalto da Serra", "MT"}, {"Pocone", "MT"}, {"Pontal do Araguaia", "MT"},
	{"Ponte Branca", "MT"}, {"Pontes e Lacerda", "MT"},
	{"Porto Alegre do Norte", "MT"}, {"Porto dos Gauchos", "MT"},
	{"Porto Esperidiao", "MT"}, {"Porto Estrela", "MT"}, {"Poxoreu", "MT"},
	{"Primavera do Leste", "MT"}, {"Querencia", "MT"}, {"Reserva do Cabacal", "MT"},
	{"Ribeirao Cascalheira", "MT"}, {"Ribeiraozinho", "MT"}, {"Rio Branco", "MT"},
	{"Rondolandia", "MT"}, {"Rondonopolis", "MT"}, {"Rosario Oeste", "MT"},
	{"Salto do Ceu", "MT"}, {"Santa Carmem", "MT"}, {"Santa Cruz do Xingu", "MT"},
	{"Santa Rita do Trivelato", "MT"}, {"Santa Terezinha", "MT"},
	{"Santo Afonso", "MT"}, {"Santo Antonio do Leste", "MT"},
	{"Santo Antonio do Leverger", "MT"}, {"Sao Felix do Araguaia", "MT"},
	{"Sao Jose do Povo", "MT"}, {"Sao Jose do Rio Claro", "MT"},
	{"Sao Jose do Xingu", "MT"}, {"Sao Jose dos Quatro Marcos", "MT"},
	{"Sao Pedro da Cipa", "MT"}, {"Sapezal", "MT"}, {"Serra Nova Dourada", "MT"},
	{"Sinop", "MT"}, {"Sorriso", "MT"}, {"Tabapora", "MT"},
	{"Tangara da Serra", "MT"}, {"Tapurah", "MT"}, {"Terra Nova do Norte", "MT"},
	{"Tesouro", "MT"}, {"Torixoreu", "MT"}, {"Uniao do Sul", "MT"},
	{"Vale de Sao Domingos", "MT"}, {"Varzea Grande", "MT"}, {"Vera", "MT"},
	{"Vila Bela da Santissima Trindade", "MT"}, {"Vila Rica", "MT"},
}

// CidadesPorUF agrupa as cidades disponiveis por UF. Hoje apenas MT esta
// populado; adicione novas UFs conforme a operacao se expande.
var CidadesPorUF = map[string][]Cidade{
	"MT": CidadesMT,
}
