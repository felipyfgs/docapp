package service

import (
	"regexp"
	"strconv"
	"strings"

	"docapp/core/internal/model"
)

var detRegex = regexp.MustCompile(`(?s)<det\s+nItem="(\d+)">(.*?)</det>`)

func ExtractItens(xmlContent string) []model.DocumentoItem {
	matches := detRegex.FindAllStringSubmatch(xmlContent, -1)
	if len(matches) == 0 {
		return nil
	}

	itens := make([]model.DocumentoItem, 0, len(matches))
	for _, m := range matches {
		nItem, _ := strconv.Atoi(m[1])
		detBody := m[2]

		item := model.DocumentoItem{
			NItem: nItem,
		}

		prodBody := extractSectionBody(detBody, "prod")
		if prodBody == "" {
			continue
		}

		item.CProd = extractTagValue(prodBody, "cProd")
		item.CEAN = extractTagValue(prodBody, "cEAN")
		item.XProd = extractTagValue(prodBody, "xProd")
		if item.XProd == "" {
			continue
		}
		item.NCM = extractTagValue(prodBody, "NCM")
		item.CEST = extractTagValue(prodBody, "CEST")
		item.CFOP = extractTagValue(prodBody, "CFOP")
		item.UCom = extractTagValue(prodBody, "uCom")
		item.QCom = parseDecimal(extractTagValue(prodBody, "qCom"))
		item.VUnCom = parseDecimal(extractTagValue(prodBody, "vUnCom"))
		item.VProd = parseDecimal(extractTagValue(prodBody, "vProd"))
		item.VDesc = parseDecimal(extractTagValue(prodBody, "vDesc"))
		item.VFrete = parseDecimal(extractTagValue(prodBody, "vFrete"))
		item.VSeg = parseDecimal(extractTagValue(prodBody, "vSeg"))
		item.VOutro = parseDecimal(extractTagValue(prodBody, "vOutro"))
		item.XPed = extractTagValue(prodBody, "xPed")
		item.NItemPed = extractTagValue(prodBody, "nItemPed")

		item.InfAdProd = extractTagValue(detBody, "infAdProd")

		impostoBody := extractSectionBody(detBody, "imposto")
		if impostoBody != "" {
			item.VTotTrib = parseDecimalPtr(extractTagValue(impostoBody, "vTotTrib"))
			extractICMS(&item, impostoBody)
			extractIPI(&item, impostoBody)
			extractPIS(&item, impostoBody)
			extractCOFINS(&item, impostoBody)
			extractIBSCBS(&item, impostoBody)
		}

		itens = append(itens, item)
	}

	return itens
}

func extractICMS(item *model.DocumentoItem, impostoBody string) {
	icmsBody := extractSectionBody(impostoBody, "ICMS")
	if icmsBody == "" {
		return
	}

	// Try all ICMS variants - regime normal
	for _, cst := range []string{"ICMS00", "ICMS10", "ICMS20", "ICMS30", "ICMS40", "ICMS51", "ICMS60", "ICMS70", "ICMS90"} {
		body := extractSectionBody(icmsBody, cst)
		if body == "" {
			continue
		}
		item.ICMSOrig = extractTagValue(body, "orig")
		item.ICMSCST = extractTagValue(body, "CST")
		item.ICMSModBC = extractTagValue(body, "modBC")
		item.ICMSPRedBC = parseDecimalPtr(extractTagValue(body, "pRedBC"))
		item.ICMSVBC = parseDecimalPtr(extractTagValue(body, "vBC"))
		item.ICMSPICMS = parseDecimalPtr(extractTagValue(body, "pICMS"))
		item.ICMSVICMS = parseDecimal(extractTagValue(body, "vICMS"))
		item.ICMSVICMSDeson = parseDecimalPtr(extractTagValue(body, "vICMSDeson"))

		// ST fields (ICMS10, ICMS30, ICMS70, ICMS90 have vBCST/vICMSST; ICMS60 has vBCSTRet/vICMSSTRet)
		if v := extractTagValue(body, "vBCST"); v != "" {
			item.ICMSVBCST = parseDecimalPtr(v)
		} else {
			item.ICMSVBCST = parseDecimalPtr(extractTagValue(body, "vBCSTRet"))
		}
		item.ICMSPST = parseDecimalPtr(extractTagValue(body, "pICMSST", "pST"))
		if v := extractTagValue(body, "vICMSST"); v != "" {
			item.ICMSVICMSST = parseDecimalPtr(v)
		} else {
			item.ICMSVICMSST = parseDecimalPtr(extractTagValue(body, "vICMSSTRet"))
		}
		return
	}

	// Simples Nacional variants
	for _, csosn := range []string{"ICMSSN101", "ICMSSN102", "ICMSSN201", "ICMSSN202", "ICMSSN500", "ICMSSN900"} {
		body := extractSectionBody(icmsBody, csosn)
		if body == "" {
			continue
		}
		item.ICMSOrig = extractTagValue(body, "orig")
		item.ICMSCST = extractTagValue(body, "CSOSN")
		item.ICMSVBC = parseDecimalPtr(extractTagValue(body, "vBC"))
		item.ICMSPICMS = parseDecimalPtr(extractTagValue(body, "pCredSN"))
		item.ICMSVICMS = parseDecimal(extractTagValue(body, "vCredICMSSN"))

		// ST fields for ICMSSN201, ICMSSN202, ICMSSN500
		if v := extractTagValue(body, "vBCST"); v != "" {
			item.ICMSVBCST = parseDecimalPtr(v)
		} else {
			item.ICMSVBCST = parseDecimalPtr(extractTagValue(body, "vBCSTRet"))
		}
		item.ICMSPST = parseDecimalPtr(extractTagValue(body, "pICMSST", "pST"))
		if v := extractTagValue(body, "vICMSST"); v != "" {
			item.ICMSVICMSST = parseDecimalPtr(v)
		} else {
			item.ICMSVICMSST = parseDecimalPtr(extractTagValue(body, "vICMSSTRet"))
		}
		return
	}

	// CT-e ICMSSN
	body := extractSectionBody(icmsBody, "ICMSSN")
	if body != "" {
		item.ICMSCST = extractTagValue(body, "CST")
		return
	}
}

func extractIPI(item *model.DocumentoItem, impostoBody string) {
	ipiBody := extractSectionBody(impostoBody, "IPI")
	if ipiBody == "" {
		return
	}

	if body := extractSectionBody(ipiBody, "IPITrib"); body != "" {
		item.IPICST = extractTagValue(body, "CST")
		item.IPIVBC = parseDecimalPtr(extractTagValue(body, "vBC"))
		item.IPIPIPI = parseDecimalPtr(extractTagValue(body, "pIPI"))
		item.IPIVIPI = parseDecimal(extractTagValue(body, "vIPI"))
		return
	}

	if body := extractSectionBody(ipiBody, "IPINT"); body != "" {
		item.IPICST = extractTagValue(body, "CST")
		return
	}
}

func extractPIS(item *model.DocumentoItem, impostoBody string) {
	pisBody := extractSectionBody(impostoBody, "PIS")
	if pisBody == "" {
		return
	}

	for _, variant := range []string{"PISAliq", "PISQtde", "PISOutr", "PISST", "PISNT"} {
		body := extractSectionBody(pisBody, variant)
		if body == "" {
			continue
		}
		item.PISCST = extractTagValue(body, "CST")
		item.PISVBC = parseDecimalPtr(extractTagValue(body, "vBC"))
		item.PISPPIS = parseDecimalPtr(extractTagValue(body, "pPIS"))
		item.PISVPIS = parseDecimal(extractTagValue(body, "vPIS"))
		return
	}
}

func extractCOFINS(item *model.DocumentoItem, impostoBody string) {
	cofinsBody := extractSectionBody(impostoBody, "COFINS")
	if cofinsBody == "" {
		return
	}

	for _, variant := range []string{"COFINSAliq", "COFINSQtde", "COFINSOutr", "COFINSST", "COFINSNT"} {
		body := extractSectionBody(cofinsBody, variant)
		if body == "" {
			continue
		}
		item.COFINSCST = extractTagValue(body, "CST")
		item.COFINSVBC = parseDecimalPtr(extractTagValue(body, "vBC"))
		item.COFINSPCOFINS = parseDecimalPtr(extractTagValue(body, "pCOFINS"))
		item.COFINSVCOFINS = parseDecimal(extractTagValue(body, "vCOFINS"))
		return
	}
}

func extractIBSCBS(item *model.DocumentoItem, impostoBody string) {
	ibscbsBody := extractSectionBody(impostoBody, "IBSCBS")
	if ibscbsBody == "" {
		return
	}

	item.IBSCBSCST = extractTagValue(ibscbsBody, "CST")
	item.IBSCBSCClassTrib = extractTagValue(ibscbsBody, "cClassTrib")

	gibscbs := extractSectionBody(ibscbsBody, "gIBSCBS")
	if gibscbs == "" {
		return
	}

	item.IBSCBSVBC = parseDecimalPtr(extractTagValue(gibscbs, "vBC"))

	gibsuf := extractSectionBody(gibscbs, "gIBSUF")
	if gibsuf != "" {
		item.IBSCBSPIBSUF = parseDecimalPtr(extractTagValue(gibsuf, "pIBSUF"))
		item.IBSCBSVIBSUF = parseDecimalPtr(extractTagValue(gibsuf, "vIBSUF"))
	}

	gibsmun := extractSectionBody(gibscbs, "gIBSMun")
	if gibsmun != "" {
		item.IBSCBSPIBSMun = parseDecimalPtr(extractTagValue(gibsmun, "pIBSMun"))
		item.IBSCBSVIBSMun = parseDecimalPtr(extractTagValue(gibsmun, "vIBSMun"))
	}

	gcbs := extractSectionBody(gibscbs, "gCBS")
	if gcbs != "" {
		item.IBSCBSPCBS = parseDecimalPtr(extractTagValue(gcbs, "pCBS"))
		item.IBSCBSVCBS = parseDecimalPtr(extractTagValue(gcbs, "vCBS"))
	}
}

func parseDecimalPtr(raw string) *float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &v
}

func ExtractItensDescricoes(itens []model.DocumentoItem) string {
	if len(itens) == 0 {
		return ""
	}
	var b strings.Builder
	for _, item := range itens {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strings.ToLower(item.XProd))
	}
	return b.String()
}
