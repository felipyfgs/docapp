<script setup lang="ts">
import type { Empresa } from '~/types'

const props = defineProps<{
  empresa: Empresa
  blockedUntil?: string | null
}>()

const { formatCNPJ, formatDate, certBadge } = useDocumentoFormatters()

const cert = computed(() => certBadge(props.empresa.certificado_status))

const isBlocked = computed(() => {
  if (!props.blockedUntil) return false
  return new Date(props.blockedUntil) > new Date()
})
</script>

<template>
  <div class="flex flex-wrap items-center gap-3 mb-6">
    <UBadge
      :color="empresa.situacao_cadastral === 'Ativa' ? 'success' : 'warning'"
      variant="subtle"
    >
      {{ empresa.situacao_cadastral || 'Desconhecida' }}
    </UBadge>
    <p class="text-sm text-muted font-mono bg-default rounded-md px-2 py-0.5 border border-default">
      {{ formatCNPJ(empresa.cnpj) }}
    </p>
    <UBadge :color="cert.color as any" variant="subtle">
      {{ cert.label }}
    </UBadge>
    <UBadge
      v-if="isBlocked"
      color="error"
      variant="subtle"
      icon="i-lucide-lock"
    >
      Bloqueado até {{ formatDate(blockedUntil!) }}
    </UBadge>
  </div>
</template>
