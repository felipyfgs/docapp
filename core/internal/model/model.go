package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Base struct {
	ID        uint       `bun:",pk,autoincrement" json:"id"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt *time.Time `bun:"deleted_at,nullzero" json:"-"`
}

type Empresa struct {
	bun.BaseModel `bun:"table:empresas,alias:e"`

	Base
	CNPJ                string `bun:"cnpj,notnull" json:"cnpj"`
	RazaoSocial         string `bun:"razao_social,notnull" json:"razao_social"`
	NomeFantasia        string `                                      json:"nome_fantasia"`
	SituacaoCadastral   string `                                      json:"situacao_cadastral"`
	Logradouro          string `                                      json:"logradouro"`
	Numero              string `                                      json:"numero"`
	Complemento         string `                                      json:"complemento"`
	Bairro              string `                                      json:"bairro"`
	CEP                 string `                                      json:"cep"`
	Cidade              string `                                      json:"cidade"`
	Estado              string `                                      json:"estado"`
	Telefone            string `                                      json:"telefone"`
	Email               string `                                      json:"email"`
	CNAE                string `                                      json:"cnae"`
	Porte               string `                                      json:"porte"`
	NaturezaJuridica    string `                                      json:"natureza_juridica"`
	DataInicioAtividade string `                                      json:"data_inicio_atividade"`

	SyncState   *EmpresaSyncState   `bun:"rel:has-one,join:id=empresa_id" json:"-"`
	Certificado *EmpresaCertificado `bun:"rel:has-one,join:id=empresa_id" json:"-"`

	LookbackDays         int        `bun:"-" json:"lookback_days"`
	UltNSU               string     `bun:"-" json:"ult_nsu"`
	Ativo                bool       `bun:"-" json:"ativo"`
	SiglaUF              string     `bun:"-" json:"sigla_uf"`
	TpAmb                int        `bun:"-" json:"tp_amb"`
	UltimaSincronizacao  *time.Time `bun:"-" json:"ultima_sincronizacao"`
	CertificadoValidoAte *time.Time `bun:"-" json:"certificado_valido_ate"`
	CertificadoPFX       []byte     `bun:"-" json:"-"`
	CertificadoSenha     string     `bun:"-" json:"-"`
}

func (e *Empresa) CertificadoStatus() string {
	if e.CertificadoValidoAte == nil && len(e.CertificadoPFX) == 0 {
		return "sem_certificado"
	}

	if e.CertificadoValidoAte == nil {
		return "sem_certificado"
	}

	now := time.Now()
	if e.CertificadoValidoAte.Before(now) {
		return "vencido"
	}
	if e.CertificadoValidoAte.Before(now.AddDate(0, 0, 30)) {
		return "prestes_a_vencer"
	}
	return "valido"
}

func (e *Empresa) HydrateFromRelations() {
	e.LookbackDays = 90
	e.UltNSU = "000000000000000"
	e.Ativo = true
	e.TpAmb = 1

	if e.SyncState != nil {
		e.LookbackDays = e.SyncState.LookbackDays
		e.UltNSU = e.SyncState.UltNSU
		e.Ativo = e.SyncState.Ativo
		e.UltimaSincronizacao = e.SyncState.UltimaSincronizacao
	}

	if e.Certificado != nil {
		e.CertificadoPFX = e.Certificado.CertificadoPFX
		e.CertificadoSenha = e.Certificado.CertificadoSenha
		e.SiglaUF = e.Certificado.SiglaUF
		e.TpAmb = e.Certificado.TpAmb
		e.CertificadoValidoAte = e.Certificado.CertificadoValidoAte
	}
}

type EmpresaCertificado struct {
	bun.BaseModel `bun:"table:empresa_certificados,alias:ec"`

	EmpresaID            uint       `bun:"empresa_id,pk,notnull" json:"empresa_id"`
	Empresa              *Empresa   `bun:"rel:belongs-to,join:empresa_id=id" json:"-"`
	CreatedAt            time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt            *time.Time `bun:"deleted_at,nullzero" json:"-"`
	CertificadoPFX       []byte     `bun:"certificado_pfx,notnull" json:"-"`
	CertificadoSenha     string     `bun:"certificado_senha,notnull" json:"-"`
	SiglaUF              string     `bun:"sigla_uf,notnull" json:"sigla_uf"`
	TpAmb                int        `bun:"tp_amb,notnull" json:"tp_amb"`
	CertificadoValidoAte *time.Time `bun:"certificado_valido_ate,nullzero" json:"certificado_valido_ate"`
}

type EmpresaSyncState struct {
	bun.BaseModel `bun:"table:empresa_sync_states,alias:ess"`

	EmpresaID            uint       `bun:"empresa_id,pk,notnull" json:"empresa_id"`
	Empresa              *Empresa   `bun:"rel:belongs-to,join:empresa_id=id" json:"-"`
	CreatedAt            time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt            time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt            *time.Time `bun:"deleted_at,nullzero" json:"-"`
	Ativo                bool       `bun:"ativo,notnull" json:"ativo"`
	LookbackDays         int        `bun:"lookback_days,notnull" json:"lookback_days"`
	UltNSU               string     `bun:"ult_nsu,notnull" json:"ult_nsu"`
	MaxNSU               string     `bun:"max_nsu,nullzero" json:"max_nsu"`
	UltimaSincronizacao  *time.Time `bun:"ultima_sincronizacao,nullzero" json:"ultima_sincronizacao"`
	BlockedUntil         *time.Time `bun:"blocked_until,nullzero" json:"blocked_until"`
	DownloadBlockedUntil *time.Time `bun:"download_blocked_until,nullzero" json:"download_blocked_until"`
	UltimoCStat          string     `bun:"ultimo_cstat" json:"ultimo_cstat"`
	UltimoXMotivo        string     `bun:"ultimo_xmotivo" json:"ultimo_xmotivo"`
}

type DocumentoFiscal struct {
	bun.BaseModel `bun:"table:documentos_fiscais,alias:df"`

	Base
	EmpresaID          uint       `bun:"empresa_id,notnull" json:"empresa_id"`
	Empresa            *Empresa   `bun:"rel:belongs-to,join:empresa_id=id" json:"empresa,omitempty"`
	NSU                string     `bun:"nsu,notnull" json:"nsu"`
	ChaveAcesso        string     `bun:"chave_acesso" json:"chave_acesso"`
	TipoDocumento      string     `bun:"tipo_documento,notnull" json:"tipo_documento"`
	StatusDocumento    string     `bun:"status_documento,notnull" json:"status_documento"`
	NumeroDocumento    string     `bun:"numero_documento" json:"numero_documento"`
	EmitenteNome       string     `json:"emitente_nome"`
	EmitenteCNPJ       string     `bun:"emitente_cnpj,nullzero" json:"emitente_cnpj"`
	DestinatarioNome   string     `json:"destinatario_nome"`
	DestinatarioCNPJ   string     `bun:"destinatario_cnpj,nullzero" json:"destinatario_cnpj"`
	Competencia        string     `json:"competencia"`
	Schema             string     `bun:"schema_nome" json:"schema"`
	XMLObjectKey       string     `bun:"xml_object_key,notnull" json:"xml_object_key"`
	XMLSHA256          string     `bun:"xml_sha256" json:"-"`
	XMLSizeBytes       int        `bun:"xml_size_bytes" json:"-"`
	XMLResumo          bool       `bun:"xml_resumo,notnull" json:"xml_resumo"`
	DanfeObjectKey     string     `json:"danfe_object_key"`
	DanfeGeneratedAt   *time.Time `bun:"danfe_generated_at,nullzero" json:"danfe_generated_at"`
	DataEmissao        *time.Time `bun:"data_emissao,nullzero" json:"data_emissao"`
	SearchText         string     `bun:"search_text,notnull" json:"-"`
	ManifestacaoStatus string     `bun:"manifestacao_status,nullzero" json:"manifestacao_status"`
	ManifestacaoAt     *time.Time `bun:"manifestacao_at,nullzero" json:"manifestacao_at"`
}
