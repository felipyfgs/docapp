<?php

namespace App\Services\Sped;

use App\Services\Sped\Exceptions\SefazThrottleException;

class DistDFeService
{
    public function __construct(
        private readonly SefazToolsFactory $toolsFactory,
        private readonly ResponseMapper $responseMapper
    ) {}

    /**
     * @return array<string, mixed>
     */
    public function execute(string $tenant, string $ultNsu = '0', int $maxLoops = 1, bool $includeXml = true): array
    {
        $tools = $this->toolsFactory->make($tenant);
        $maxLoops = max(1, min($maxLoops, (int) config('sped.distdfe_max_loops', 12)));

        $currentNsu = $this->normalizeNsu($ultNsu);
        $lastMaxNsu = $currentNsu;
        $lastCstat = null;
        $lastXMotivo = null;

        /** @var array<string, array<string, mixed>> $documentsByNsu */
        $documentsByNsu = [];

        for ($iteration = 1; $iteration <= $maxLoops; $iteration++) {
            $response = $tools->sefazDistDFe($currentNsu);
            $mapped = $this->responseMapper->mapDistDFeResponse($response, $includeXml);

            $lastCstat = (string) ($mapped['cstat'] ?? '');
            $lastXMotivo = (string) ($mapped['xmotivo'] ?? '');
            $currentNsu = $this->normalizeNsu((string) ($mapped['ult_nsu'] ?? $currentNsu));
            $lastMaxNsu = $this->normalizeNsu((string) ($mapped['max_nsu'] ?? $lastMaxNsu));

            foreach (($mapped['documents'] ?? []) as $document) {
                $nsu = (string) ($document['nsu'] ?? '');
                $key = $nsu !== '' ? $nsu : md5((string) json_encode($document));
                $documentsByNsu[$key] = $document;
            }

            if ($lastCstat === '656') {
                throw new SefazThrottleException(
                    $lastXMotivo !== '' ? $lastXMotivo : 'Consumo indevido na SEFAZ.',
                    (int) config('sped.sefaz_retry_after_seconds', 3600)
                );
            }

            if ($lastCstat === '137' || $currentNsu === $lastMaxNsu) {
                return [
                    'cstat' => $lastCstat,
                    'xmotivo' => $lastXMotivo,
                    'ult_nsu' => $currentNsu,
                    'max_nsu' => $lastMaxNsu,
                    'iterations' => $iteration,
                    'documents' => array_values($documentsByNsu),
                ];
            }

            if ($iteration < $maxLoops) {
                sleep(max(0, (int) config('sped.distdfe_sleep_seconds', 2)));
            }
        }

        return [
            'cstat' => $lastCstat,
            'xmotivo' => $lastXMotivo,
            'ult_nsu' => $currentNsu,
            'max_nsu' => $lastMaxNsu,
            'iterations' => $maxLoops,
            'documents' => array_values($documentsByNsu),
        ];
    }

    private function normalizeNsu(string $value): string
    {
        $digits = preg_replace('/\D+/', '', $value) ?? '';

        if ($digits === '') {
            return str_repeat('0', 15);
        }

        $trimmed = ltrim($digits, '0');

        if ($trimmed === '') {
            return str_repeat('0', 15);
        }

        return str_pad($trimmed, 15, '0', STR_PAD_LEFT);
    }
}
