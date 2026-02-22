<?php

namespace App\Services\Sped;

use App\Services\Sped\Exceptions\TenantCredentialException;
use JsonException;
use NFePHP\Common\Certificate;
use NFePHP\NFe\Tools;
use RuntimeException;

class SefazToolsFactory
{
    public function __construct(
        private readonly ConfigProfileLoader $profileLoader,
        private readonly CertificateStore $certificateStore
    ) {
    }

    public function make(string $tenant): Tools
    {
        $tenant = $this->profileLoader->sanitizeTenant($tenant);
        $profile = $this->profileLoader->load($tenant);

        $password = (string) ($profile['pfx_password'] ?? $profile['password'] ?? '');

        if ($password === '') {
            throw new TenantCredentialException("Senha do certificado não informada para o tenant '{$tenant}'.");
        }

        $certificate = Certificate::readPfx($this->certificateStore->loadPfx($tenant), $password);

        try {
            $tools = new Tools($this->buildConfigJson($profile), $certificate);
        } catch (\Throwable $exception) {
            throw new RuntimeException('Falha ao inicializar comunicação com SEFAZ.', 0, $exception);
        }

        $tools->model('55');

        if (method_exists($tools, 'setEnvironment')) {
            $tools->setEnvironment((int) ($profile['tp_amb'] ?? $profile['tpAmb'] ?? 1));
        }

        return $tools;
    }

    /**
     * @param array<string, mixed> $profile
     */
    private function buildConfigJson(array $profile): string
    {
        $razaoSocial = (string) ($profile['razao_social'] ?? $profile['razaosocial'] ?? '');
        $cnpj = preg_replace('/\D+/', '', (string) ($profile['cnpj'] ?? ''));
        $siglaUf = strtoupper((string) ($profile['sigla_uf'] ?? $profile['siglaUF'] ?? ''));

        if ($razaoSocial === '' || $cnpj === '' || $siglaUf === '') {
            throw new RuntimeException('Perfil fiscal deve conter cnpj, razao_social e sigla_uf.');
        }

        $config = [
            'atualizacao' => date('Y-m-d H:i:s'),
            'tpAmb' => (int) ($profile['tp_amb'] ?? $profile['tpAmb'] ?? 1),
            'razaosocial' => $razaoSocial,
            'cnpj' => $cnpj,
            'siglaUF' => $siglaUf,
            'schemes' => (string) ($profile['schemes'] ?? 'PL_010_V1'),
            'versao' => (string) ($profile['versao'] ?? '4.00'),
            'tokenIBPT' => (string) ($profile['token_ibpt'] ?? $profile['tokenIBPT'] ?? ''),
            'CSC' => (string) ($profile['csc'] ?? $profile['CSC'] ?? ''),
            'CSCid' => (string) ($profile['csc_id'] ?? $profile['CSCid'] ?? ''),
            'proxyConf' => [
                'proxyIp' => (string) ($profile['proxy_ip'] ?? ''),
                'proxyPort' => (string) ($profile['proxy_port'] ?? ''),
                'proxyUser' => (string) ($profile['proxy_user'] ?? ''),
                'proxyPass' => (string) ($profile['proxy_pass'] ?? ''),
            ],
        ];

        try {
            return json_encode($config, JSON_THROW_ON_ERROR);
        } catch (JsonException $exception) {
            throw new RuntimeException('Falha ao gerar configuração da SEFAZ.', 0, $exception);
        }
    }
}
