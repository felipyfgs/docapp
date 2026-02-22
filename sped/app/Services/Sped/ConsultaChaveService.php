<?php

namespace App\Services\Sped;

class ConsultaChaveService
{
    public function __construct(
        private readonly SefazToolsFactory $toolsFactory,
        private readonly ResponseMapper $responseMapper
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function execute(string $tenant, string $chave): array
    {
        $tools = $this->toolsFactory->make($tenant);
        $response = $tools->sefazConsultaChave($chave);

        $mapped = $this->responseMapper->mapConsultaChaveResponse($response);
        $mapped['chave'] = $chave;

        return $mapped;
    }
}
