package maps

// Departamento herda CadDepto do banco legado Radelgo. Os valores
// originais vinham de outro ERP (industria de alimentos) e nao tem uso
// operacional aqui  mantemos apenas para auditoria/historico.
type Departamento struct {
	ID        int
	Descricao string
}

// DepartamentosLegado preserva a tabela CadDepto original, somente para
// referencia ao migrar dados antigos. Nao usar em fluxos novos.
var DepartamentosLegado = []Departamento{
	{1, "Controladoria"},
	{2, "RH"},
	{3, "Manutencao"},
	{4, "Gerencias Unidade"},
	{5, "DGQ"},
	{6, "Fabrica de Aves"},
	{7, "Fabrica de Industrializados"},
	{8, "Fabrica de CCC"},
	{9, "Fabrica de Bovino"},
	{10, "CPD"},
	{11, "Logistica Secundaria"},
	{12, "Logistica Primaria"},
	{13, "Vendas Sadia"},
	{14, "Terceiros"},
	{15, "Apoio Adm"},
	{16, "Food Service"},
	{17, "Rezende"},
	{18, "Pesquisa e Desenvolvimento"},
	{19, "SIF"},
	{20, "Fabrica de Empanados"},
}
