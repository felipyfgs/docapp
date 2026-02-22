<?php

namespace App\Services\Sped;

use App\Services\Sped\Exceptions\TenantProfileNotFoundException;
use Illuminate\Support\Str;
use JsonException;
use RuntimeException;

class ConfigProfileLoader
{
    public function load(string $tenant): array
    {
        $tenant = $this->sanitizeTenant($tenant);
        $profilePath = $this->profilePath($tenant);

        if (!is_file($profilePath)) {
            throw new TenantProfileNotFoundException("Perfil fiscal do tenant '{$tenant}' não encontrado.");
        }

        $content = file_get_contents($profilePath);

        if ($content === false) {
            throw new RuntimeException("Falha ao ler o perfil fiscal do tenant '{$tenant}'.");
        }

        try {
            $decoded = json_decode($content, true, 512, JSON_THROW_ON_ERROR);
        } catch (JsonException $exception) {
            throw new RuntimeException("Perfil fiscal do tenant '{$tenant}' contém JSON inválido.", 0, $exception);
        }

        if (!is_array($decoded)) {
            throw new RuntimeException("Perfil fiscal do tenant '{$tenant}' inválido.");
        }

        return $decoded;
    }

    public function sanitizeTenant(string $tenant): string
    {
        $sanitized = Str::of($tenant)
            ->lower()
            ->trim()
            ->replaceMatches('/[^a-z0-9_-]/', '')
            ->value();

        if ($sanitized === '') {
            throw new RuntimeException('Tenant inválido.');
        }

        return $sanitized;
    }

    private function profilePath(string $tenant): string
    {
        return rtrim((string) config('sped.certs_path'), '/')."/{$tenant}.json";
    }
}
