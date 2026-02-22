<?php

return [
    'certs_path' => env('FISCAL_CERTS_PATH', storage_path('certs')),

    'download_limit_per_hour' => (int) env('SEFAZ_DOWNLOAD_LIMIT_PER_HOUR', 20),

    'distdfe_max_loops' => (int) env('SEFAZ_DISTDFE_MAX_LOOPS', 12),

    'distdfe_sleep_seconds' => (int) env('SEFAZ_DISTDFE_SLEEP_SECONDS', 2),

    'sefaz_retry_after_seconds' => (int) env('SEFAZ_RETRY_AFTER_SECONDS', 3600),
];
