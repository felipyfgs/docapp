<?php

namespace App\Services\Sped;

use InvalidArgumentException;
use NFePHP\DA\CTe\Dacte;
use NFePHP\DA\NFe\Danfce;
use NFePHP\DA\NFe\Danfe;

class DanfeService
{
    public function generate(string $tipo, string $xml): string
    {
        $normalizedType = $this->normalizeTipo($tipo);
        $xmlContent = trim($xml);

        if ($xmlContent === '') {
            throw new InvalidArgumentException('XML é obrigatório.');
        }

        return match ($normalizedType) {
            'nf-e' => (new Danfe($xmlContent))->render(),
            'nfc-e' => (new Danfce($xmlContent))->render(),
            'ct-e' => (new Dacte($xmlContent))->render(),
            default => throw new InvalidArgumentException('Tipo de documento não suportado para DANFE.'),
        };
    }

    private function normalizeTipo(string $tipo): string
    {
        return strtolower(trim($tipo));
    }
}
