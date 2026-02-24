package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/config"
	"docapp/core/internal/db"
	"docapp/core/internal/repository"
	"docapp/core/internal/service"
)

func main() {
	cnpj := flag.String("cnpj", "", "CNPJ da empresa (opcional)")
	flag.Parse()

	cfg := config.Load()
	log := config.NewLogger(cfg.Env)

	bunDB, err := db.Connect(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	c := client.New(cfg.SpedServiceURL, cfg.SpedTimeoutSeconds)

	empresaRepo := repository.NewEmpresaRepository(bunDB)
	documentoRepo := repository.NewDocumentoRepository(bunDB)
	empresaService := service.NewEmpresaService(empresaRepo)

	empresas, err := empresaService.ListAtivasComCertificado()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to list empresas")
	}

	fmt.Println("Iniciando rotina de Ciência da Operação...")

	for _, emp := range empresas {
		if *cnpj != "" && emp.CNPJ != *cnpj {
			continue
		}

		fmt.Printf("Processando empresa: %s - %s\n", emp.CNPJ, emp.RazaoSocial)
		siglaUF := emp.SiglaUF
		if siglaUF == "" {
			siglaUF = emp.Estado
		}

		if siglaUF == "" || len(siglaUF) < 2 {
			log.Warn().Str("cnpj", emp.CNPJ).Msg("UF da empresa não configurada")
			continue
		}

		if len(siglaUF) > 2 {
			siglaUF = siglaUF[:2]
		}

		pendentes, err := documentoRepo.ListPendingCiencia(context.Background(), emp.ID)
		if err != nil {
			log.Warn().Err(err).Str("cnpj", emp.CNPJ).Msg("failed to list pending ciencia")
			continue
		}

		if len(pendentes) == 0 {
			fmt.Println("  -> Nenhuma nota pendente de ciência.")
			continue
		}

		fmt.Printf("  -> Forçando envio de %d ciências...\n", len(pendentes))
		for _, doc := range pendentes {
			fmt.Printf("     - Chave: %s\n", doc.ChaveAcesso)
			manifResp, err := c.Manifesta(
				context.Background(),
				emp.CertificadoPFX,
				emp.CertificadoSenha,
				emp.CNPJ,
				emp.RazaoSocial,
				siglaUF,
				emp.TpAmb,
				doc.ChaveAcesso,
				"210210", // Ciência da Operação
				"",
			)
			if err != nil {
				log.Warn().Err(err).Str("chave", doc.ChaveAcesso).Msg("ciencia request failed")
				continue
			}

			cStat := ""
			if len(manifResp.CStat) > 0 {
				cStat = manifResp.CStat
			}
			
			if cStat != "128" && cStat != "135" && cStat != "573" {
				log.Warn().
					Str("chave", doc.ChaveAcesso).
					Str("cstat", cStat).
					Str("xmotivo", manifResp.XMotivo).
					Msg("ciencia rejected by SEFAZ")
				continue
			}

			now := time.Now()
			_ = documentoRepo.UpdateManifestacaoStatus(context.Background(), doc.ID, "ciencia", now)
		}
	}

	fmt.Println("Concluído!")
	os.Exit(0)
}
