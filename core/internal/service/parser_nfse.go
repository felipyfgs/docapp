package service

import (
	"time"
)

type NFSeDocument struct {
	ChaveAcesso     string
	NumeroNFSe      string
	DataEmissao     *time.Time
	Competencia     string
	PrestadorCNPJ   string
	PrestadorNome   string
	TomadorCNPJ     string
	TomadorNome     string
	ValorServico    float64
	ValorLiquido    float64
	StatusDocumento string
	XML             string
}

func ParseNFSeXML(xmlContent string) NFSeDocument {
	doc := NFSeDocument{
		XML:             xmlContent,
		StatusDocumento: "autorizada",
	}

	doc.NumeroNFSe = extractTagValue(xmlContent, "nNFSe")

	if extractTagValue(xmlContent, "cStat") == "101" {
		doc.StatusDocumento = "cancelada"
	}

	// Prestador: <emit> (infNFSe level) first, fallback <prest> (DPS level)
	doc.PrestadorCNPJ = nfsePartyField(xmlContent, "emit", "CNPJ")
	doc.PrestadorNome = nfsePartyField(xmlContent, "emit", "xNome")
	if doc.PrestadorCNPJ == "" {
		doc.PrestadorCNPJ = nfsePartyField(xmlContent, "prest", "CNPJ")
	}
	if doc.PrestadorNome == "" {
		doc.PrestadorNome = nfsePartyField(xmlContent, "prest", "xNome")
	}

	// Tomador
	doc.TomadorCNPJ = nfsePartyField(xmlContent, "toma", "CNPJ")
	doc.TomadorNome = nfsePartyField(xmlContent, "toma", "xNome")

	// CPF fallback
	if doc.PrestadorCNPJ == "" {
		doc.PrestadorCNPJ = nfsePartyField(xmlContent, "emit", "CPF")
	}
	if doc.PrestadorCNPJ == "" {
		doc.PrestadorCNPJ = nfsePartyField(xmlContent, "prest", "CPF")
	}
	if doc.TomadorCNPJ == "" {
		doc.TomadorCNPJ = nfsePartyField(xmlContent, "toma", "CPF")
	}

	// Valores - reuse extractValorDecimal from parser.go
	doc.ValorServico = extractValorDecimal(xmlContent, "vServ")
	doc.ValorLiquido = extractValorDecimal(xmlContent, "vLiq")
	if doc.ValorLiquido == 0 {
		doc.ValorLiquido = doc.ValorServico
	}

	// Data emissao
	if dhEmi := extractTagValue(xmlContent, "dhEmi"); dhEmi != "" {
		if t, err := time.Parse(time.RFC3339, dhEmi); err == nil {
			doc.DataEmissao = &t
		} else if t, err := time.Parse("2006-01-02T15:04:05", dhEmi); err == nil {
			doc.DataEmissao = &t
		}
	}

	// Competencia: <dCompet> first, fallback to data emissao
	if comp := extractTagValue(xmlContent, "dCompet"); comp != "" {
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

// nfsePartyField extracts a field from a section using extractSectionBody + extractTagValue from parser.go.
func nfsePartyField(xml, section, fieldTag string) string {
	body := extractSectionBody(xml, section)
	if body == "" {
		return ""
	}
	return extractTagValue(body, fieldTag)
}
