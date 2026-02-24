package service

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"hash/fnv"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type DistDFeResponse struct {
	CStat     string     `json:"cstat"`
	XMotivo   string     `json:"xmotivo"`
	TpAmb     string     `json:"tp_amb"`
	VerAplic  string     `json:"ver_aplic"`
	DHResp    string     `json:"dh_resp"`
	UltNSU    string     `json:"ult_nsu"`
	MaxNSU    string     `json:"max_nsu"`
	Documents []Document `json:"documents"`
}

type Document struct {
	NSU              string     `json:"nsu"`
	Schema           string     `json:"schema"`
	DocumentType     string     `json:"document_type"`
	StatusDocumento  string     `json:"status_documento"`
	NumeroDocumento  string     `json:"numero_documento"`
	EmitenteNome     string     `json:"emitente_nome"`
	EmitenteCNPJ     string     `json:"emitente_cnpj"`
	DestinatarioNome string     `json:"destinatario_nome"`
	DestinatarioCNPJ string     `json:"destinatario_cnpj"`
	Competencia      string     `json:"competencia"`
	XMLResumo        bool       `json:"xml_resumo"`
	XML              string     `json:"xml"`
	ChaveAcesso      string     `json:"chave_acesso"`
	DataEmissao      *time.Time `json:"data_emissao,omitempty"`
}

type retDistDFeInt struct {
	XMLName        xml.Name        `xml:"retDistDFeInt"`
	CStat          string          `xml:"cStat"`
	XMotivo        string          `xml:"xMotivo"`
	TpAmb          string          `xml:"tpAmb"`
	VerAplic       string          `xml:"verAplic"`
	DHResp         string          `xml:"dhResp"`
	UltNSU         string          `xml:"ultNSU"`
	MaxNSU         string          `xml:"maxNSU"`
	LoteDistDFeInt *loteDistDFeInt `xml:"loteDistDFeInt"`
}

type loteDistDFeInt struct {
	DocZips []docZip `xml:"docZip"`
}

type docZip struct {
	NSU    string `xml:"NSU,attr"`
	Schema string `xml:"schema,attr"`
	Value  string `xml:",chardata"`
}

func ParseDistDFeResponse(rawXML string) (*DistDFeResponse, error) {
	ret, err := extractRetDistDFeInt(rawXML)
	if err != nil {
		return nil, fmt.Errorf("parsing retDistDFeInt: %w", err)
	}

	resp := &DistDFeResponse{
		CStat:    ret.CStat,
		XMotivo:  ret.XMotivo,
		TpAmb:    ret.TpAmb,
		VerAplic: ret.VerAplic,
		DHResp:   ret.DHResp,
		UltNSU:   ret.UltNSU,
		MaxNSU:   ret.MaxNSU,
	}

	if ret.LoteDistDFeInt != nil {
		for _, dz := range ret.LoteDistDFeInt.DocZips {
			docXML, err := decodeDocZip(dz.Value)
			if err != nil {
				continue
			}

			dataEmissao := extractDataEmissao(docXML)
			docType := documentTypeFromSchemaAndXML(dz.Schema, docXML)

			doc := Document{
				NSU:              dz.NSU,
				Schema:           dz.Schema,
				DocumentType:     docType,
				StatusDocumento:  extractStatusDocumento(docXML),
				NumeroDocumento:  extractNumeroDocumento(docXML),
				EmitenteNome:     extractEmitenteNome(docXML),
				EmitenteCNPJ:     extractEmitenteCNPJ(docXML),
				DestinatarioNome: extractDestinatarioNome(docXML),
				DestinatarioCNPJ: extractDestinatarioCNPJ(docXML),
				Competencia:      extractCompetencia(dataEmissao),
				XMLResumo:        isResumoDocument(dz.Schema),
				XML:              docXML,
				ChaveAcesso:      extractChaveAcesso(docXML),
				DataEmissao:      dataEmissao,
			}

			resp.Documents = append(resp.Documents, doc)
		}
	}

	return resp, nil
}

func extractRetDistDFeInt(rawXML string) (*retDistDFeInt, error) {
	var direct retDistDFeInt
	if err := xml.Unmarshal([]byte(rawXML), &direct); err == nil {
		if direct.XMLName.Local == "retDistDFeInt" {
			return &direct, nil
		}
	}

	decoder := xml.NewDecoder(strings.NewReader(rawXML))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		if start.Name.Local != "retDistDFeInt" {
			continue
		}

		var nested retDistDFeInt
		if err := decoder.DecodeElement(&nested, &start); err != nil {
			return nil, err
		}

		return &nested, nil
	}

	return nil, fmt.Errorf("retDistDFeInt not found")
}

func decodeDocZip(encoded string) (string, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return "", fmt.Errorf("empty docZip content")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	reader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("gzip read: %w", err)
	}

	return string(content), nil
}
func documentTypeFromSchemaAndXML(schema, xmlContent string) string {
	schemaLower := strings.ToLower(schema)
	if strings.Contains(schemaLower, "nfse") {
		return "nfs-e"
	}
	if strings.Contains(schemaLower, "cte") {
		return "ct-e"
	}
	if strings.Contains(schemaLower, "nfe") {
		mod := extractTagValue(xmlContent, "mod")
		if mod == "" {
			chave := extractChaveAcesso(xmlContent)
			if len(chave) == 44 {
				mod = chave[20:22]
			}
		}
		if mod == "65" {
			return "nfc-e"
		}
		return "nf-e"
	}

	mod := extractTagValue(xmlContent, "mod")
	if mod == "" {
		chave := extractChaveAcesso(xmlContent)
		if len(chave) == 44 {
			mod = chave[20:22]
		}
	}

	switch mod {
	case "57":
		return "ct-e"
	case "65":
		return "nfc-e"
	case "55":
		return "nf-e"
	}

	return "desconhecido"
}

func extractChaveAcesso(xmlContent string) string {
	if id := extractAttributeValue(xmlContent, "Id"); id != "" {
		re := regexp.MustCompile(`(\d{44})`)
		if match := re.FindStringSubmatch(id); len(match) == 2 {
			return match[1]
		}
	}

	for _, tag := range []string{"<chNFe>", "<chCTe>"} {
		start := strings.Index(xmlContent, tag)
		if start == -1 {
			continue
		}
		start += len(tag)
		endTag := "</" + tag[1:]
		end := strings.Index(xmlContent[start:], endTag)
		if end == 44 {
			return xmlContent[start : start+end]
		}
	}
	return ""
}

func extractEmitenteNome(xmlContent string) string {
	if name := extractPartyName(xmlContent, "emit"); name != "" {
		return name
	}
	return extractTagValue(xmlContent, "xNome")
}

func extractEmitenteCNPJ(xmlContent string) string {
	if cnpj := extractPartyCNPJ(xmlContent, "emit"); cnpj != "" {
		return cnpj
	}
	v := digitsOnly(extractTagValue(xmlContent, "CNPJ"))
	if len(v) == 14 {
		return v
	}
	return ""
}

func extractDestinatarioNome(xmlContent string) string {
	for _, section := range []string{"dest", "rem", "toma"} {
		if value := extractPartyName(xmlContent, section); value != "" {
			return value
		}
	}
	return ""
}

func extractDestinatarioCNPJ(xmlContent string) string {
	for _, section := range []string{"dest", "rem", "toma"} {
		if value := extractPartyCNPJ(xmlContent, section); value != "" {
			return value
		}
	}
	return ""
}

func extractNumeroDocumento(xmlContent string) string {
	if n := extractTagValue(xmlContent, "nNF", "nCT", "nNFS", "nDoc"); n != "" {
		return n
	}
	chave := extractChaveAcesso(xmlContent)
	if len(chave) == 44 {
		return strings.TrimLeft(chave[25:34], "0")
	}
	return ""
}

func extractStatusDocumento(xmlContent string) string {
	sit := extractTagValue(xmlContent, "cSitNFe", "cSitCTe", "cSit")
	switch sit {
	case "1":
		return "autorizada"
	case "2":
		return "cancelada"
	case "3":
		return "denegada"
	}

	statusValues := extractTagValues(xmlContent, "cStat")

	for _, code := range statusValues {
		if isCancelledStatusCode(code) {
			return "cancelada"
		}
	}

	for _, code := range statusValues {
		if isAuthorizedStatusCode(code) {
			return "autorizada"
		}
	}

	for _, code := range statusValues {
		if isDeniedStatusCode(code) {
			return "denegada"
		}
	}

	return "desconhecido"
}

func extractCompetencia(dataEmissao *time.Time) string {
	if dataEmissao == nil {
		return ""
	}
	return dataEmissao.Format("2006/01")
}

func isResumoDocument(schema string) bool {
	s := strings.ToLower(strings.TrimSpace(schema))
	if s == "" {
		return false
	}
	if strings.Contains(s, "proc") {
		return false
	}
	return strings.Contains(s, "res")
}

func extractPartyName(xmlContent, section string) string {
	body := extractSectionBody(xmlContent, section)
	if body == "" {
		return ""
	}

	return extractTagValue(body, "xNome", "xNomeDest", "xRazao")
}

func extractPartyCNPJ(xmlContent, section string) string {
	body := extractSectionBody(xmlContent, section)
	if body == "" {
		return ""
	}

	v := digitsOnly(extractTagValue(body, "CNPJ"))
	if len(v) == 14 {
		return v
	}
	return ""
}

func extractPartyAddressField(xmlContent, section, fieldTag string) string {
	body := extractSectionBody(xmlContent, section)
	if body == "" {
		return ""
	}
	return extractTagValue(body, fieldTag)
}

func extractSectionBody(xmlContent, section string) string {
	re := regexp.MustCompile(`(?is)<(?:[a-z0-9_]+:)?` + regexp.QuoteMeta(section) + `\b[^>]*>(.*?)</(?:[a-z0-9_]+:)?` + regexp.QuoteMeta(section) + `>`)
	match := re.FindStringSubmatch(xmlContent)
	if len(match) != 2 {
		return ""
	}

	return match[1]
}

func extractTagValue(xmlContent string, tags ...string) string {
	for _, tag := range tags {
		re := regexp.MustCompile(`(?is)<(?:[a-z0-9_]+:)?` + regexp.QuoteMeta(tag) + `\b[^>]*>\s*([^<]+?)\s*</(?:[a-z0-9_]+:)?` + regexp.QuoteMeta(tag) + `>`)
		match := re.FindStringSubmatch(xmlContent)
		if len(match) == 2 {
			return strings.TrimSpace(match[1])
		}
	}

	return ""
}

func extractTagValues(xmlContent string, tag string) []string {
	re := regexp.MustCompile(`(?is)<(?:[a-z0-9_]+:)?` + regexp.QuoteMeta(tag) + `\b[^>]*>\s*([^<]+?)\s*</(?:[a-z0-9_]+:)?` + regexp.QuoteMeta(tag) + `>`)
	matches := re.FindAllStringSubmatch(xmlContent, -1)
	if len(matches) == 0 {
		return nil
	}

	values := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) == 2 {
			values = append(values, strings.TrimSpace(match[1]))
		}
	}

	return values
}

func extractAttributeValue(xmlContent, attr string) string {
	re := regexp.MustCompile(`(?is)\b` + regexp.QuoteMeta(attr) + `\s*=\s*"([^"]+)"`)
	match := re.FindStringSubmatch(xmlContent)
	if len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func digitsOnly(value string) string {
	if value == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}

	return b.String()
}

func isAuthorizedStatusCode(code string) bool {
	switch code {
	case "100", "150":
		return true
	default:
		return false
	}
}

func isCancelledStatusCode(code string) bool {
	switch code {
	case "101", "135", "136", "151", "155":
		return true
	default:
		return false
	}
}

func isDeniedStatusCode(code string) bool {
	switch code {
	case "110", "301", "302":
		return true
	default:
		return false
	}
}

func extractDataEmissao(xmlContent string) *time.Time {
	for _, tag := range []string{"<dhEmi>", "<dEmi>"} {
		start := strings.Index(xmlContent, tag)
		if start == -1 {
			continue
		}

		start += len(tag)
		endTag := "</" + tag[1:]
		end := strings.Index(xmlContent[start:], endTag)
		if end == -1 {
			continue
		}

		raw := strings.TrimSpace(xmlContent[start : start+end])

		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05-07:00", "2006-01-02", "2006-01-02T15:04:05"} {
			if t, err := time.Parse(layout, raw); err == nil {
				return &t
			}
		}
	}

	return nil
}

type ConsultaChaveResponse struct {
	CStat     string `json:"cstat"`
	XMotivo   string `json:"xmotivo"`
	Chave     string `json:"chave"`
	TpAmb     string `json:"tp_amb"`
	Protocolo string `json:"protocolo,omitempty"`
	DHRecbto  string `json:"dh_recbto,omitempty"`
}

type retConsSitNFe struct {
	XMLName xml.Name `xml:"retConsSitNFe"`
	CStat   string   `xml:"cStat"`
	XMotivo string   `xml:"xMotivo"`
	ChNFe   string   `xml:"chNFe"`
	TpAmb   string   `xml:"tpAmb"`
	ProtNFe *protNFe `xml:"protNFe"`
}

type protNFe struct {
	InfProt infProt `xml:"infProt"`
}

type infProt struct {
	NProt    string `xml:"nProt"`
	DHRecbto string `xml:"dhRecbto"`
}

func ParseConsultaChaveResponse(rawXML string) (*ConsultaChaveResponse, error) {
	var ret retConsSitNFe
	if err := xml.Unmarshal([]byte(rawXML), &ret); err != nil {
		return nil, fmt.Errorf("parsing retConsSitNFe: %w", err)
	}

	resp := &ConsultaChaveResponse{
		CStat:   ret.CStat,
		XMotivo: ret.XMotivo,
		Chave:   ret.ChNFe,
		TpAmb:   ret.TpAmb,
	}

	if ret.ProtNFe != nil {
		resp.Protocolo = ret.ProtNFe.InfProt.NProt
		resp.DHRecbto = ret.ProtNFe.InfProt.DHRecbto
	}

	return resp, nil
}

// ParseNFeProcXML parses a raw nfeProc/procNFe XML (without the retDistDFeInt wrapper)
// and returns a Document. nsu can be extracted from the filename via NSUFromFilename.
func ParseNFeProcXML(xmlContent string, nsu string) Document {
	dataEmissao := extractDataEmissao(xmlContent)
	schema := schemaFromRawXML(xmlContent)
	return Document{
		NSU:              nsu,
		Schema:           schema,
		DocumentType:     documentTypeFromSchemaAndXML(schema, xmlContent),
		StatusDocumento:  extractStatusDocumento(xmlContent),
		NumeroDocumento:  extractNumeroDocumento(xmlContent),
		EmitenteNome:     extractEmitenteNome(xmlContent),
		EmitenteCNPJ:     extractEmitenteCNPJ(xmlContent),
		DestinatarioNome: extractDestinatarioNome(xmlContent),
		DestinatarioCNPJ: extractDestinatarioCNPJ(xmlContent),
		Competencia:      extractCompetencia(dataEmissao),
		XMLResumo:        false,
		XML:              xmlContent,
		ChaveAcesso:      extractChaveAcesso(xmlContent),
		DataEmissao:      dataEmissao,
	}
}

// NSUFromFilename extracts the 15-digit NSU from filenames like WS_<NSU>_<chave>.xml.
func NSUFromFilename(filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	parts := strings.Split(base, "_")
	if len(parts) >= 3 {
		candidate := parts[1]
		matched, _ := regexp.MatchString(`^\d{15}$`, candidate)
		if matched {
			return candidate
		}
	}
	return ""
}

// CNPJFromChave extracts the 14-digit emitente CNPJ from a 44-digit chave_acesso.
// chave layout: cUF(2) + AAMM(4) + CNPJ(14) + mod(2) + serie(3) + nNF(9) + tpEmis(1) + cNF(8) + cDV(1)
func CNPJFromChave(chave string) string {
	chave = strings.TrimSpace(chave)
	if len(chave) != 44 {
		return ""
	}
	return chave[6:20]
}

// nsuFromChave derives a deterministic 15-digit NSU from a chave_acesso.
func nsuFromChave(chave string) string {
	h := fnv.New64a()
	h.Write([]byte(chave))
	return fmt.Sprintf("%015d", h.Sum64()%1_000_000_000_000_000)
}

// schemaFromRawXML detects the schema name from the root XML tag.
func schemaFromRawXML(xmlContent string) string {
	lower := strings.ToLower(xmlContent)
	switch {
	case strings.Contains(lower, "<nfeproc") || strings.Contains(lower, "<nfproc"):
		return "procNFe_v4.00.xsd"
	case strings.Contains(lower, "<cteproc"):
		return "procCTe_v4.00.xsd"
	case strings.Contains(lower, "<mdfeproc"):
		return "procMDFe_v3.00.xsd"
	default:
		return "procNFe_v4.00.xsd"
	}
}
