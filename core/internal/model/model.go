package model

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        uint           `gorm:"primarykey"  json:"id"`
	CreatedAt time.Time      `                   json:"created_at"`
	UpdatedAt time.Time      `                   json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"       json:"-"`
}

type Empresa struct {
	Base
	CNPJ                string     `gorm:"uniqueIndex;size:14;not null"     json:"cnpj"`
	RazaoSocial         string     `gorm:"not null"                          json:"razao_social"`
	NomeFantasia        string     `                                         json:"nome_fantasia"`
	SituacaoCadastral   string     `                                         json:"situacao_cadastral"`
	Logradouro          string     `                                         json:"logradouro"`
	Numero              string     `                                         json:"numero"`
	Complemento         string     `                                         json:"complemento"`
	Bairro              string     `                                         json:"bairro"`
	CEP                 string     `                                         json:"cep"`
	Cidade              string     `                                         json:"cidade"`
	Estado              string     `                                         json:"estado"`
	Telefone            string     `                                         json:"telefone"`
	Email               string     `                                         json:"email"`
	CNAE                string     `                                         json:"cnae"`
	Porte               string     `                                         json:"porte"`
	NaturezaJuridica    string     `                                         json:"natureza_juridica"`
	DataInicioAtividade string     `                                         json:"data_inicio_atividade"`
	LookbackDays            int    `gorm:"default:90"                        json:"lookback_days"`
	UltNSU                  string `gorm:"default:'000000000000000'"         json:"ult_nsu"`
	Ativo                   bool   `gorm:"default:true"                      json:"ativo"`
	CertificadoCaminho      string `                                         json:"certificado_caminho"`
	CertificadoSenha        string `                                         json:"certificado_senha"`
}

type DocumentoFiscal struct {
	Base
	EmpresaID   uint       `gorm:"index;not null"  json:"empresa_id"`
	NSU         string     `gorm:"index;size:15"   json:"nsu"`
	ChaveAcesso string     `gorm:"index;size:44"   json:"chave_acesso"`
	Tipo        string     `                        json:"tipo"`
	Schema      string     `                        json:"schema"`
	XML         string     `gorm:"type:text"        json:"xml"`
	DataEmissao *time.Time `                        json:"data_emissao"`
}
