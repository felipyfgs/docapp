<?php

namespace App\Services\Sped\Exceptions;

use RuntimeException;

class SefazThrottleException extends RuntimeException
{
    public function __construct(
        string $message,
        private readonly int $retryAfter,
        int $code = 0,
        ?\Throwable $previous = null
    ) {
        parent::__construct($message, $code, $previous);
    }

    public function retryAfter(): int
    {
        return $this->retryAfter;
    }
}
