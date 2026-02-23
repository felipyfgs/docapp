<?php

namespace App\Http\Controllers;

use App\Services\Sped\ConsultaChaveService;
use App\Services\Sped\DistDFeService;
use App\Services\Sped\DownloadByKeyService;
use App\Services\Sped\Exceptions\SefazThrottleException;
use App\Services\Sped\Exceptions\TenantCredentialException;
use App\Services\Sped\Exceptions\TenantProfileNotFoundException;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\RateLimiter;

class NfeController extends Controller
{
    public function __construct(
        private readonly DistDFeService $distDFeService,
        private readonly DownloadByKeyService $downloadByKeyService,
        private readonly ConsultaChaveService $consultaChaveService
    ) {}

    public function distDFe(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'tenant' => ['required', 'string', 'max:120'],
            'ult_nsu' => ['nullable', 'string', 'regex:/^\d{1,15}$/'],
            'max_loops' => ['nullable', 'integer', 'min:1', 'max:50'],
            'include_xml' => ['nullable', 'boolean'],
        ]);
        $includeXml = $request->boolean('include_xml', true);

        try {
            $result = $this->distDFeService->execute(
                (string) $validated['tenant'],
                (string) ($validated['ult_nsu'] ?? '0'),
                (int) ($validated['max_loops'] ?? 1),
                $includeXml
            );

            return response()->json($result, $this->mapSefazStatus((string) ($result['cstat'] ?? '')));
        } catch (SefazThrottleException $exception) {
            return response()
                ->json([
                    'message' => $exception->getMessage(),
                    'retry_after' => $exception->retryAfter(),
                ], 429)
                ->header('Retry-After', (string) $exception->retryAfter());
        } catch (TenantProfileNotFoundException|TenantCredentialException $exception) {
            return response()->json(['message' => $exception->getMessage()], 404);
        } catch (\Throwable $exception) {
            report($exception);

            return response()->json(['message' => 'Falha na comunicação com a SEFAZ.'], 502);
        }
    }

    public function downloadByKey(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'tenant' => ['required', 'string', 'max:120'],
            'chave' => ['required', 'string', 'regex:/^\d{44}$/'],
        ]);

        $rateLimit = $this->checkDownloadRateLimit((string) $validated['tenant']);
        if ($rateLimit instanceof JsonResponse) {
            return $rateLimit;
        }

        try {
            $result = $this->downloadByKeyService->execute(
                (string) $validated['tenant'],
                (string) $validated['chave']
            );

            return response()->json($result, $this->mapSefazStatus((string) ($result['cstat'] ?? '')));
        } catch (SefazThrottleException $exception) {
            return response()
                ->json([
                    'message' => $exception->getMessage(),
                    'retry_after' => $exception->retryAfter(),
                ], 429)
                ->header('Retry-After', (string) $exception->retryAfter());
        } catch (TenantProfileNotFoundException|TenantCredentialException $exception) {
            return response()->json(['message' => $exception->getMessage()], 404);
        } catch (\Throwable $exception) {
            report($exception);

            return response()->json(['message' => 'Falha ao baixar XML na SEFAZ.'], 502);
        }
    }

    public function consultaChave(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'tenant' => ['required', 'string', 'max:120'],
            'chave' => ['required', 'string', 'regex:/^\d{44}$/'],
        ]);

        try {
            $result = $this->consultaChaveService->execute(
                (string) $validated['tenant'],
                (string) $validated['chave']
            );

            return response()->json($result, $this->mapSefazStatus((string) ($result['cstat'] ?? '')));
        } catch (TenantProfileNotFoundException|TenantCredentialException $exception) {
            return response()->json(['message' => $exception->getMessage()], 404);
        } catch (\Throwable $exception) {
            report($exception);

            return response()->json(['message' => 'Falha ao consultar chave na SEFAZ.'], 502);
        }
    }

    private function checkDownloadRateLimit(string $tenant): ?JsonResponse
    {
        $maxAttempts = max(1, (int) config('sped.download_limit_per_hour', 20));
        $normalizedTenant = preg_replace('/[^a-z0-9_-]/', '', strtolower($tenant)) ?? '';
        $key = 'sped:download:'.($normalizedTenant !== '' ? $normalizedTenant : md5($tenant));

        if (RateLimiter::tooManyAttempts($key, $maxAttempts)) {
            $retryAfter = RateLimiter::availableIn($key);

            return response()
                ->json([
                    'message' => 'Limite por hora para download por chave atingido.',
                    'retry_after' => $retryAfter,
                ], 429)
                ->header('Retry-After', (string) $retryAfter);
        }

        RateLimiter::hit($key, 3600);

        return null;
    }

    private function mapSefazStatus(string $cstat): int
    {
        return match ($cstat) {
            '137' => 404,
            '640' => 403,
            default => 200,
        };
    }
}
