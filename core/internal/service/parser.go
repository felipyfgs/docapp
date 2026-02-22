package service

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
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
	NSU          string     `json:"nsu"`
	Schema       string     `json:"schema"`
	DocumentType string     `json:"document_type"`
	XML          string     `json:"xml"`
	ChaveAcesso  string     `json:"chave_acesso"`
	DataEmissao  *time.Time `json:"data_emissao,omitempty"`
}

type retDistDFeInt struct {
	XMLName       xml.Name `xml:"retDistDFeInt"`
	CStat         string   `xml:"cStat"`
	XMotivo       string   `xml:"xMotivo"`
	TpAmb         string   `xml:"tpAmb"`
	VerAplic      string   `xml:"verAplic"`
	DHResp        string   `xml:"dhResp"`
	UltNSU        string   `xml:"ultNSU"`
	MaxNSU        string   `xml:"maxNSU"`
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
	var ret retDistDFeInt
	if err := xml.Unmarshal([]byte(rawXML), &ret); err != nil {
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

			doc := Document{
				NSU:          dz.NSU,
				Schema:       dz.Schema,
				DocumentType: documentTypeFromSchema(dz.Schema),
				XML:          docXML,
				ChaveAcesso:  extractChaveAcesso(docXML),
				DataEmissao:  extractDataEmissao(docXML),
			}

			resp.Documents = append(resp.Documents, doc)
		}
	}

	return resp, nil
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

func documentTypeFromSchema(schema string) string {
	if schema == "" {
		return "unknown"
	}
	parts := strings.Split(strings.ToLower(schema), "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func extractChaveAcesso(xmlContent string) string {
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

		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05-07:00", "2006-01-02"} {
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
