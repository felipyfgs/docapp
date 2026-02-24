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
	reNFSeNumero     = regexp.MustCompile(`<nNFSe>([^<]+)</nNFSe>`)
	reNFSeDhEmi      = regexp.MustCompile(`<dhEmi>([^<]+)</dhEmi>`)
	reNFSeDCompet    = regexp.MustCompile(`<dCompet>([^<]+)</dCompet>`)
	reNFSeVLiquid    = regexp.MustCompile(`<vLiq>([^<]+)</vLiq>`)
	reNFSeVServ      = regexp.MustCompile(`<vServ>([^<]+)</vServ>`)
	reNFSeCStat      = regexp.MustCompile(`<cStat>([^<]+)</cStat>`)
)

func ParseNFSeXML(xmlContent string) NFSeDocument {
	doc := NFSeDocument{
		XML:             xmlContent,
		StatusDocumento: "autorizada",
	}

	doc.NumeroNFSe = extractMatch(reNFSeNumero, xmlContent)

	// cStat: 100=autorizada, 101=cancelada
	if cStat := extractMatch(reNFSeCStat, xmlContent); cStat == "101" {
		doc.StatusDocumento = "cancelada"
	}

	// Prestador: try <emit> first (infNFSe level), fallback to <prest> (DPS level)
	doc.PrestadorCNPJ = extractNFSePartyField(xmlContent, "emit", "CNPJ")
	doc.PrestadorNome = extractNFSePartyField(xmlContent, "emit", "xNome")
	if doc.PrestadorCNPJ == "" {
		doc.PrestadorCNPJ = extractNFSePartyField(xmlContent, "prest", "CNPJ")
	}
	if doc.PrestadorNome == "" {
		doc.PrestadorNome = extractNFSePartyField(xmlContent, "prest", "xNome")
	}

	// Tomador
	doc.TomadorCNPJ = extractNFSePartyField(xmlContent, "toma", "CNPJ")
	doc.TomadorNome = extractNFSePartyField(xmlContent, "toma", "xNome")

	// CPF fallback
	if doc.PrestadorCNPJ == "" {
		doc.PrestadorCNPJ = extractNFSePartyField(xmlContent, "emit", "CPF")
	}
	if doc.PrestadorCNPJ == "" {
		doc.PrestadorCNPJ = extractNFSePartyField(xmlContent, "prest", "CPF")
	}
	if doc.TomadorCNPJ == "" {
		doc.TomadorCNPJ = extractNFSePartyField(xmlContent, "toma", "CPF")
	}

	// Valores
	doc.ValorServico = parseFloatSafe(extractMatch(reNFSeVServ, xmlContent))
	doc.ValorLiquido = parseFloatSafe(extractMatch(reNFSeVLiquid, xmlContent))
	if doc.ValorLiquido == 0 {
		doc.ValorLiquido = doc.ValorServico
	}

	// Data emissao
	if dhEmi := extractMatch(reNFSeDhEmi, xmlContent); dhEmi != "" {
		if t, err := time.Parse(time.RFC3339, dhEmi); err == nil {
			doc.DataEmissao = &t
		} else if t, err := time.Parse("2006-01-02T15:04:05", dhEmi); err == nil {
			doc.DataEmissao = &t
		}
	}

	// Competencia: <dCompet> first, fallback to data emissao
	if comp := extractMatch(reNFSeDCompet, xmlContent); comp != "" {
		if len(comp) >= 7 {
			doc.Competencia = comp[:7]
		} else {
			doc.Competencia = comp
		}
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
