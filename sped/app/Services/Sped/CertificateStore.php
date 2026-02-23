<?php

namespace App\Services\Sped;

use App\Services\Sped\Exceptions\TenantCredentialException;
use RuntimeException;

class CertificateStore
{
    public function loadPfx(string $tenant): string
    {
        $path = rtrim((string) config('sped.certs_path'), '/')."/{$tenant}.pfx";

        if (! is_file($path)) {
            throw new TenantCredentialException("Certificado PFX do tenant '{$tenant}' não encontrado.");
        }

        $content = file_get_contents($path);

        if ($content === false || $content === '') {
            throw new RuntimeException("Falha ao ler certificado PFX do tenant '{$tenant}'.");
        }

        return $content;
    }
}
