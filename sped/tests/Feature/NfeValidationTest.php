<?php

namespace Tests\Feature;

use Tests\TestCase;

class NfeValidationTest extends TestCase
{
    public function test_distdfe_requires_tenant(): void
    {
        $response = $this->postJson('/api/v1/nfe/distdfe', []);

        $response->assertStatus(422)
            ->assertJsonValidationErrors(['tenant']);
    }

    public function test_download_requires_chave_com_44_digitos(): void
    {
        $response = $this->postJson('/api/v1/nfe/download', [
            'tenant' => 'empresa_a',
            'chave' => '123',
        ]);

        $response->assertStatus(422)
            ->assertJsonValidationErrors(['chave']);
    }

    public function test_consulta_requires_chave(): void
    {
        $response = $this->postJson('/api/v1/nfe/consulta-chave', [
            'tenant' => 'empresa_a',
        ]);

        $response->assertStatus(422)
            ->assertJsonValidationErrors(['chave']);
    }
}
