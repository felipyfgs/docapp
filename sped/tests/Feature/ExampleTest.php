<?php

namespace Tests\Feature;

// use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class ExampleTest extends TestCase
{
    /**
     * A basic test example.
     */
    public function test_the_health_route_returns_success(): void
    {
        $response = $this->get('/health');

        $response->assertStatus(200);
    }
}
