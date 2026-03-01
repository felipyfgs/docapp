<script setup lang="ts">
import type { Empresa } from '~/types'

defineProps<{
  empresa: Empresa
}>()
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex items-center gap-2">
        <UIcon name="i-lucide-building-2" class="size-5 text-primary" />
        <h3 class="font-semibold text-highlighted">
          Dados Cadastrais
        </h3>
      </div>
    </template>

    <div class="space-y-4 text-sm">
      <div v-if="empresa.nome_fantasia">
        <span class="block text-xs text-muted mb-1">Nome Fantasia</span>
        <span class="font-medium">{{ empresa.nome_fantasia }}</span>
      </div>
      <div v-if="empresa.cnae || empresa.natureza_juridica">
        <span class="block text-xs text-muted mb-1">Natureza / CNAE</span>
        <span>{{ empresa.natureza_juridica || '—' }} <span v-if="empresa.cnae" class="text-muted">({{ empresa.cnae }})</span></span>
      </div>
      <div v-if="empresa.email || empresa.telefone" class="grid grid-cols-2 gap-4">
        <div v-if="empresa.email">
          <span class="block text-xs text-muted mb-1">E-mail</span>
          <span class="truncate">{{ empresa.email }}</span>
        </div>
        <div v-if="empresa.telefone">
          <span class="block text-xs text-muted mb-1">Telefone</span>
          <span>{{ empresa.telefone }}</span>
        </div>
      </div>
      <div v-if="empresa.logradouro || empresa.cidade">
        <span class="block text-xs text-muted mb-1">Endereço</span>
        <span class="leading-relaxed text-muted">
          {{ empresa.logradouro }}{{ empresa.numero ? `, ${empresa.numero}` : '' }}{{ empresa.complemento ? ` - ${empresa.complemento}` : '' }}<br v-if="empresa.logradouro">
          <span v-if="empresa.bairro">{{ empresa.bairro }} — </span>{{ empresa.cidade }}<span v-if="empresa.estado">/{{ empresa.estado }}</span>
          <span v-if="empresa.cep"> (CEP {{ empresa.cep }})</span>
        </span>
      </div>
    </div>
  </UCard>
</template>
