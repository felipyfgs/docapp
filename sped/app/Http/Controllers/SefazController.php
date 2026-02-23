<?php

namespace App\Http\Controllers;

use App\Services\Sped\DanfeService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Validation\Rule;
use NFePHP\Common\Certificate;
use NFePHP\NFe\Tools;

class SefazController extends Controller
{
    public function __construct(
        private readonly DanfeService $danfeService,
    ) {}

    public function distDFe(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'certificado_pfx' => ['required', 'string'],
            'senha' => ['required', 'string'],
            'cnpj' => ['required', 'string', 'size:14'],
            'razao_social' => ['required', 'string'],
            'sigla_uf' => ['required', 'string', 'size:2'],
            'tp_amb' => ['required', 'integer', 'in:1,2'],
            'ult_nsu' => ['nullable', 'string'],
        ]);

        try {
            $tools = $this->makeTools($validated);
            $ultNsu = $this->normalizeNsu((string) ($validated['ult_nsu'] ?? '0'));

            $rawXml = $tools->sefazDistDFe($ultNsu);

            return response()->json([
                'raw_xml' => $rawXml,
                'cstat' => $this->extractCstat($rawXml),
                'xmotivo' => $this->extractXMotivo($rawXml),
            ]);
        } catch (\Throwable $e) {
            report($e);

            return response()->json([
                'error' => $e->getMessage(),
            ], 500);
        }
    }

    public function download(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'certificado_pfx' => ['required', 'string'],
            'senha' => ['required', 'string'],
            'cnpj' => ['required', 'string', 'size:14'],
            'razao_social' => ['required', 'string'],
            'sigla_uf' => ['required', 'string', 'size:2'],
            'tp_amb' => ['required', 'integer', 'in:1,2'],
            'chave' => ['required', 'string', 'size:44'],
        ]);

        try {
            $tools = $this->makeTools($validated);
            $rawXml = $tools->sefazDownload((string) $validated['chave']);

            return response()->json([
                'raw_xml' => $rawXml,
                'cstat' => $this->extractCstat($rawXml),
                'xmotivo' => $this->extractXMotivo($rawXml),
            ]);
        } catch (\Throwable $e) {
            report($e);

            return response()->json([
                'error' => $e->getMessage(),
            ], 500);
        }
    }

    public function consulta(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'certificado_pfx' => ['required', 'string'],
            'senha' => ['required', 'string'],
            'cnpj' => ['required', 'string', 'size:14'],
            'razao_social' => ['required', 'string'],
            'sigla_uf' => ['required', 'string', 'size:2'],
            'tp_amb' => ['required', 'integer', 'in:1,2'],
            'chave' => ['required', 'string', 'size:44'],
        ]);

        try {
            $tools = $this->makeTools($validated);
            $rawXml = $tools->sefazConsultaChave((string) $validated['chave']);

            return response()->json([
                'raw_xml' => $rawXml,
                'cstat' => $this->extractCstat($rawXml),
                'xmotivo' => $this->extractXMotivo($rawXml),
            ]);
        } catch (\Throwable $e) {
            report($e);

            return response()->json([
                'error' => $e->getMessage(),
            ], 500);
        }
    }

    public function manifesta(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'certificado_pfx' => ['required', 'string'],
            'senha' => ['required', 'string'],
            'cnpj' => ['required', 'string', 'size:14'],
            'razao_social' => ['required', 'string'],
            'sigla_uf' => ['required', 'string', 'size:2'],
            'tp_amb' => ['required', 'integer', 'in:1,2'],
            'chave' => ['required', 'string', 'size:44'],
            'tp_evento' => ['required', 'string', 'in:210200,210210,210220,210240'],
            'justificativa' => ['nullable', 'string', 'min:15', 'max:255'],
        ]);

        try {
            $tools = $this->makeTools($validated);
            $rawXml = $tools->sefazManifesta(
                (string) $validated['chave'],
                (string) $validated['tp_evento'],
                (string) ($validated['justificativa'] ?? ''),
                1,
            );

            return response()->json([
                'raw_xml' => $rawXml,
                'cstat' => $this->extractCstat($rawXml),
                'xmotivo' => $this->extractXMotivo($rawXml),
            ]);
        } catch (\Throwable $e) {
            report($e);

            return response()->json([
                'error' => $e->getMessage(),
            ], 500);
        }
    }

    public function danfe(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'tipo' => ['required', 'string', Rule::in(['nf-e', 'nfc-e', 'ct-e'])],
            'xml' => ['required', 'string'],
        ]);

        try {
            $pdf = $this->danfeService->generate(
                (string) $validated['tipo'],
                (string) $validated['xml'],
            );

            return response()->json([
                'pdf_base64' => base64_encode($pdf),
                'mime_type' => 'application/pdf',
            ]);
        } catch (\InvalidArgumentException $e) {
            return response()->json([
                'error' => $e->getMessage(),
            ], 422);
        } catch (\Throwable $e) {
            report($e);

            return response()->json([
                'error' => $e->getMessage(),
            ], 500);
        }
    }

    private function makeTools(array $data): Tools
    {
        $pfxContent = base64_decode((string) $data['certificado_pfx'], true);
        if ($pfxContent === false) {
            throw new \RuntimeException('Invalid base64 certificate');
        }

        $certificate = Certificate::readPfx($pfxContent, (string) $data['senha']);

        $config = json_encode([
            'atualizacao' => date('Y-m-d H:i:s'),
            'tpAmb' => (int) $data['tp_amb'],
            'razaosocial' => (string) $data['razao_social'],
            'cnpj' => (string) $data['cnpj'],
            'siglaUF' => strtoupper((string) $data['sigla_uf']),
            'schemes' => 'PL_010_V1',
            'versao' => '4.00',
        ], JSON_THROW_ON_ERROR);

        $tools = new Tools($config, $certificate);
        $tools->model('55');

        return $tools;
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

    private function extractCstat(string $xml): ?string
    {
        if (preg_match('/<cStat>(\d+)<\/cStat>/', $xml, $matches)) {
            return $matches[1];
        }

        return null;
    }

    private function extractXMotivo(string $xml): ?string
    {
        if (preg_match('/<xMotivo>([^<]+)<\/xMotivo>/', $xml, $matches)) {
            return $matches[1];
        }

        return null;
    }
}
