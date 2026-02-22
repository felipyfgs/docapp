package model

import (
	"time"

	"gorm.io/gorm"
)

type Empresa struct {
	gorm.Model
	CNPJ                string `gorm:"uniqueIndex;size:14;not null"`
	RazaoSocial         string `gorm:"not null"`
	NomeFantasia        string
	SituacaoCadastral   string
	Logradouro          string
	Numero              string
	Complemento         string
	Bairro              string
	CEP                 string
	Cidade              string
	Estado              string
	Telefone            string
	Email               string
	CNAE                string
	Porte               string
	NaturezaJuridica    string
	DataInicioAtividade string
	LookbackDays        int    `gorm:"default:90"`
	UltNSU              string `gorm:"default:'000000000000000'"`
	Ativo               bool   `gorm:"default:true"`
}

type DocumentoFiscal struct {
	gorm.Model
	EmpresaID   uint       `gorm:"index;not null"`
	NSU         string     `gorm:"index;size:15"`
	ChaveAcesso string     `gorm:"index;size:44"`
	Tipo        string
	Schema      string
	XML         string     `gorm:"type:text"`
	DataEmissao *time.Time
}
