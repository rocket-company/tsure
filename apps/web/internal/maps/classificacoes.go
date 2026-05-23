package maps

// Classificacao agrupa um codigo curto (usado como id interno) com a
// descricao apresentada ao usuario.
type Classificacao struct {
	Codigo    string
	Descricao string
}

// ClassificacoesServico espelha CadClass do banco legado: categorias
// macro dos servicos prestados. A mesma lista esta presente no seed do
// banco; mantemos aqui para uso em dropdowns sem round-trip.
var ClassificacoesServico = []Classificacao{
	{"palco", "Palco"},
	{"tenda", "Tenda"},
	{"sonorizacao", "Sonorizacao"},
	{"iluminacao", "Iluminacao"},
	{"grupo_gerador", "Grupo Gerador"},
	{"mesas_cadeiras", "Mesas e Cadeiras"},
	{"banheiros", "Banheiros"},
	{"climatizador", "Climatizador"},
	{"caixa_termica", "Caixa Termica"},
	{"estrutura", "Estrutura"},
	{"extras", "Extras"},
}

var classByCodigo = func() map[string]Classificacao {
	out := make(map[string]Classificacao, len(ClassificacoesServico))
	for _, c := range ClassificacoesServico {
		out[c.Codigo] = c
	}
	return out
}()

// ClassificacaoByCodigo busca uma classificacao pelo codigo curto.
func ClassificacaoByCodigo(codigo string) (Classificacao, bool) {
	c, ok := classByCodigo[codigo]
	return c, ok
}
