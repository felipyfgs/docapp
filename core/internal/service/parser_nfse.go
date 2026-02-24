package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type NFSeDocument struct {
	ChaveAcesso      string
	NumeroNFSe       string
	DataEmissao      *time.Time
	Competencia      string
	PrestadorCNPJ    string
	PrestadorNome    string
	TomadorCNPJ      string
	TomadorNome      string
	ValorServico     float64
	ValorLiquido     float64
	StatusDocumento  string
	XML              string
}

var (
	reNFSeChave      = regexp.MustCompile(`<chNFSe>([^<]+)</chNFSe>`)
	reNFSeNumero     = regexp.MustCompile(`<nNFSe>([^<]+)</nNFSe>`)
	reNFSeCompetencia = regexp.MustCompile(`<competencia>([^<]+)</competencia>`)
	reNFSeSituacao   = regexp.MustCompile(`<sit>([^<]+)</sit>`)

	reNFSeDhEmi      = regexp.MustCompile(`<dhEmi>([^<]+)</dhEmi>`)
	reNFSeDhCompetencia = regexp.MustCompile(`<dhCompetencia>([^<]+)</dhCompetencia>`)

	reNFSeVServico   = regexp.MustCompile(`<vServPrest>\s*<vReceb>([^<]+)</vReceb>`)
	reNFSeVLiquid    = regexp.MustCompile(`<vLiq>([^<]+)</vLiq>`)
	reNFSeVServ      = regexp.MustCompile(`<vServ>([^<]+)</vServ>`)

	reNFSePrestCNPJ  = regexp.MustCompile(`<prest>.*?<CNPJ>([^<]+)</CNPJ>`)
	reNFSePrestNome  = regexp.MustCompile(`<prest>.*?<xNome>([^<]+)</xNome>`)
	reNFSeTomaCNPJ   = regexp.MustCompile(`<toma>.*?<CNPJ>([^<]+)</CNPJ>`)
	reNFSeTomaNome   = regexp.MustCompile(`<toma>.*?<xNome>([^<]+)</xNome>`)
)

func ParseNFSeXML(xmlContent string) NFSeDocument {
	doc := NFSeDocument{
		XML:             xmlContent,
		StatusDocumento: "autorizada",
	}

	doc.ChaveAcesso = extractMatch(reNFSeChave, xmlContent)
	doc.NumeroNFSe = extractMatch(reNFSeNumero, xmlContent)

	if sit := extractMatch(reNFSeSituacao, xmlContent); sit != "" {
		switch sit {
		case "1":
			doc.StatusDocumento = "autorizada"
		case "2":
			doc.StatusDocumento = "cancelada"
		default:
			doc.StatusDocumento = "autorizada"
		}
	}

	doc.PrestadorCNPJ = extractNFSePartyField(xmlContent, "prest", "CNPJ")
	doc.PrestadorNome = extractNFSePartyField(xmlContent, "prest", "xNome")
	doc.TomadorCNPJ = extractNFSePartyField(xmlContent, "toma", "CNPJ")
	doc.TomadorNome = extractNFSePartyField(xmlContent, "toma", "xNome")

	if cpf := extractNFSePartyField(xmlContent, "prest", "CPF"); cpf != "" && doc.PrestadorCNPJ == "" {
		doc.PrestadorCNPJ = cpf
	}
	if cpf := extractNFSePartyField(xmlContent, "toma", "CPF"); cpf != "" && doc.TomadorCNPJ == "" {
		doc.TomadorCNPJ = cpf
	}

	doc.ValorServico = parseFloatSafe(extractMatch(reNFSeVServ, xmlContent))
	doc.ValorLiquido = parseFloatSafe(extractMatch(reNFSeVLiquid, xmlContent))
	if doc.ValorLiquido == 0 {
		doc.ValorLiquido = doc.ValorServico
	}

	if dhEmi := extractMatch(reNFSeDhEmi, xmlContent); dhEmi != "" {
		if t, err := time.Parse(time.RFC3339, dhEmi); err == nil {
			doc.DataEmissao = &t
		} else if t, err := time.Parse("2006-01-02T15:04:05", dhEmi); err == nil {
			doc.DataEmissao = &t
		}
	}

	if comp := extractMatch(reNFSeCompetencia, xmlContent); comp != "" {
		doc.Competencia = comp[:7] // YYYY-MM
	} else if doc.DataEmissao != nil {
		doc.Competencia = doc.DataEmissao.Format("2006-01")
	}

	return doc
}

func extractNFSePartyField(xml, section, fieldTag string) string {
	sectionPattern := regexp.MustCompile(`<` + section + `>([\s\S]*?)</` + section + `>`)
	sectionMatch := sectionPattern.FindStringSubmatch(xml)
	if len(sectionMatch) < 2 {
		return ""
	}
	fieldPattern := regexp.MustCompile(`<` + fieldTag + `>([^<]+)</` + fieldTag + `>`)
	fieldMatch := fieldPattern.FindStringSubmatch(sectionMatch[1])
	if len(fieldMatch) < 2 {
		return ""
	}
	return strings.TrimSpace(fieldMatch[1])
}

func extractMatch(re *regexp.Regexp, content string) string {
	m := re.FindStringSubmatch(content)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func parseFloatSafe(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
