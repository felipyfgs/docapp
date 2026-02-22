package service

import (
	"docapp/core/internal/model"

	"gorm.io/gorm"
)

type EmpresaService struct {
	db *gorm.DB
}

func NewEmpresaService(db *gorm.DB) *EmpresaService {
	return &EmpresaService{db: db}
}

func (s *EmpresaService) List() ([]model.Empresa, error) {
	var empresas []model.Empresa
	if err := s.db.Order("created_at desc").Find(&empresas).Error; err != nil {
		return nil, err
	}

	return empresas, nil
}

func (s *EmpresaService) Create(e *model.Empresa) error {
	return s.db.Create(e).Error
}

func (s *EmpresaService) GetByID(id uint) (*model.Empresa, error) {
	var empresa model.Empresa
	if err := s.db.First(&empresa, id).Error; err != nil {
		return nil, err
	}

	return &empresa, nil
}

func (s *EmpresaService) Update(id uint, updates *model.Empresa) (*model.Empresa, error) {
	empresa, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates.ID = empresa.ID
	if err := s.db.Save(updates).Error; err != nil {
		return nil, err
	}

	return updates, nil
}

func (s *EmpresaService) Delete(id uint) error {
	return s.db.Unscoped().Delete(&model.Empresa{}, id).Error
}

func (s *EmpresaService) ListAtivas() ([]model.Empresa, error) {
	var empresas []model.Empresa
	if err := s.db.Where("ativo = true").Find(&empresas).Error; err != nil {
		return nil, err
	}

	return empresas, nil
}

func (s *EmpresaService) ListAtivasComCertificado() ([]model.Empresa, error) {
	var empresas []model.Empresa
	if err := s.db.Where("ativo = true").
		Where("certificado_pfx IS NOT NULL").
		Where("LENGTH(certificado_pfx) > 0").
		Where("certificado_senha != ''").
		Where("sigla_uf != ''").
		Find(&empresas).Error; err != nil {
		return nil, err
	}

	return empresas, nil
}

func (s *EmpresaService) UpdateUltNSU(id uint, ultNSU string) error {
	return s.db.Model(&model.Empresa{}).Where("id = ?", id).Update("ult_nsu", ultNSU).Error
}

func (s *EmpresaService) UpdateCertificadoPFX(id uint, pfx []byte, senha, siglaUF string, tpAmb int) error {
	return s.db.Model(&model.Empresa{}).Where("id = ?", id).Updates(map[string]any{
		"certificado_pfx":   pfx,
		"certificado_senha": senha,
		"sigla_uf":          siglaUF,
		"tp_amb":            tpAmb,
	}).Error
}
