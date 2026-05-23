// Package maps contem tabelas de referencia estaticas (UFs, cidades,
// classificacoes, departamentos legados) que viviam como tabelas no banco
// Radelgo e agora sao mantidas em codigo. Use estas listas para popular
// dropdowns e validar entradas sem custo de I/O.
package maps

// UF representa uma unidade federativa do Brasil.
type UF struct {
	Sigla string
	Nome  string
}

// UFs lista todas as unidades federativas brasileiras, ordenadas por sigla.
var UFs = []UF{
	{"AC", "Acre"},
	{"AL", "Alagoas"},
	{"AM", "Amazonas"},
	{"AP", "Amapa"},
	{"BA", "Bahia"},
	{"CE", "Ceara"},
	{"DF", "Distrito Federal"},
	{"ES", "Espirito Santo"},
	{"GO", "Goias"},
	{"MA", "Maranhao"},
	{"MG", "Minas Gerais"},
	{"MS", "Mato Grosso do Sul"},
	{"MT", "Mato Grosso"},
	{"PA", "Para"},
	{"PB", "Paraiba"},
	{"PE", "Pernambuco"},
	{"PI", "Piaui"},
	{"PR", "Parana"},
	{"RJ", "Rio de Janeiro"},
	{"RN", "Rio Grande do Norte"},
	{"RO", "Rondonia"},
	{"RR", "Roraima"},
	{"RS", "Rio Grande do Sul"},
	{"SC", "Santa Catarina"},
	{"SE", "Sergipe"},
	{"SP", "Sao Paulo"},
	{"TO", "Tocantins"},
}

var ufBySigla = func() map[string]UF {
	out := make(map[string]UF, len(UFs))
	for _, u := range UFs {
		out[u.Sigla] = u
	}
	return out
}()

// UFBySigla busca uma UF pelo seu codigo de 2 letras. Retorna false se
// nao existir.
func UFBySigla(sigla string) (UF, bool) {
	u, ok := ufBySigla[sigla]
	return u, ok
}

// IsUF retorna true se a sigla informada corresponde a uma UF valida.
func IsUF(sigla string) bool {
	_, ok := ufBySigla[sigla]
	return ok
}
