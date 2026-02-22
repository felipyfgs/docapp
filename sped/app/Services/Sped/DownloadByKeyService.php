<?php

namespace App\Services\Sped;

use App\Services\Sped\Exceptions\SefazThrottleException;

class DownloadByKeyService
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
        $response = $tools->sefazDownload($chave);

        $mapped = $this->responseMapper->mapDownloadResponse($response);
        $mapped['chave'] = $chave;

        if ((string) ($mapped['cstat'] ?? '') === '656') {
            throw new SefazThrottleException(
                (string) ($mapped['xmotivo'] ?? 'Consumo indevido na SEFAZ.'),
                (int) config('sped.sefaz_retry_after_seconds', 3600)
            );
        }

        return $mapped;
    }
}
