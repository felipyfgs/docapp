package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"docapp/core/internal/model"
	"docapp/core/internal/repository"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

type EmpresaService struct {
	repo *repository.EmpresaRepository
}

func NewEmpresaService(repo *repository.EmpresaRepository) *EmpresaService {
	return &EmpresaService{repo: repo}
}

func (s *EmpresaService) List() ([]model.Empresa, error) {
	return s.repo.List(context.Background())
}

func (s *EmpresaService) Create(e *model.Empresa) error {
	return s.repo.Create(context.Background(), e)
}

func (s *EmpresaService) GetByID(id uint) (*model.Empresa, error) {
	return s.repo.GetByID(context.Background(), id)
}

func (s *EmpresaService) Update(id uint, updates *model.Empresa) (*model.Empresa, error) {
	return s.repo.Update(context.Background(), id, updates)
}

func (s *EmpresaService) Delete(id uint) error {
	return s.repo.Delete(context.Background(), id)
}

func (s *EmpresaService) ListAtivas() ([]model.Empresa, error) {
	return s.repo.ListAtivas(context.Background())
}

func (s *EmpresaService) ListAtivasComCertificado() ([]model.Empresa, error) {
	return s.repo.ListAtivasComCertificado(context.Background())
}

func (s *EmpresaService) UpdateUltNSU(id uint, ultNSU string) error {
	trimmed := strings.TrimSpace(ultNSU)
	if trimmed == "" {
		return errors.New("ult_nsu inválido")
	}

	return s.repo.UpdateSyncState(context.Background(), id, repository.SyncStatePatch{UltNSU: &trimmed})
}

func (s *EmpresaService) UpdateCertificadoPFX(id uint, pfx []byte, senha, siglaUF string, tpAmb int) error {
	if len(pfx) == 0 {
		return errors.New("certificado pfx vazio")
	}
	if strings.TrimSpace(senha) == "" {
		return errors.New("senha do certificado obrigatória")
	}
	uf := normalizeSiglaUF(siglaUF)
	if uf == "" {
		return errors.New("sigla_uf inválida")
	}
	if tpAmb != 1 && tpAmb != 2 {
		tpAmb = 1
	}

	validoAte, _ := ParseCertificadoValidade(pfx, senha)
	return s.repo.UpsertCertificado(context.Background(), id, pfx, senha, uf, tpAmb, validoAte)
}

func ParseCertificadoValidade(pfxData []byte, senha string) (*time.Time, error) {
	_, cert, _, err := pkcs12.DecodeChain(pfxData, senha)
	if err != nil {
		return nil, err
	}
	return &cert.NotAfter, nil
}

func (s *EmpresaService) AtualizarValidadeCertificados() error {
	certificados, err := s.repo.ListCertificadosSemValidade(context.Background())
	if err != nil {
		return err
	}

	for _, cert := range certificados {
		validoAte, err := ParseCertificadoValidade(cert.CertificadoPFX, cert.CertificadoSenha)
		if err != nil {
			continue
		}

		if err := s.repo.UpdateCertificadoValidade(context.Background(), cert.EmpresaID, validoAte); err != nil {
			return fmt.Errorf("updating certificate validity for empresa %d: %w", cert.EmpresaID, err)
		}
	}
	return nil
}
