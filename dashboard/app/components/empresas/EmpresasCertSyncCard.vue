<script setup lang="ts">
import type { Empresa, EmpresaSyncState } from '~/types'

const props = defineProps<{
  empresa: Empresa
  syncState?: EmpresaSyncState | null
  togglingNfse?: boolean
}>()

const emit = defineEmits<{
  toggleNfse: [value: boolean]
}>()

const { formatDate, formatDateOnly, certBadge } = useDocumentoFormatters()

const cert = computed(() => certBadge(props.empresa.certificado_status))
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex items-center gap-2">
        <UIcon name="i-lucide-shield-check" class="size-5 text-primary" />
        <h3 class="font-semibold text-highlighted">
          Certificado & Sincronização
        </h3>
      </div>
    </template>

    <div class="space-y-4 text-sm">
      <div class="flex items-center justify-between">
        <span class="text-muted">Status</span>
        <UBadge :color="cert.color as any" variant="subtle" size="sm">
          {{ cert.label }}
        </UBadge>
      </div>

      <div v-if="empresa.certificado_valido_ate" class="flex items-center justify-between">
        <span class="text-muted">Válido até</span>
        <span class="font-medium">{{ formatDateOnly(empresa.certificado_valido_ate) }}</span>
      </div>

      <USeparator />

      <div class="grid grid-cols-2 gap-4">
        <div v-if="syncState">
          <span class="block text-xs text-muted mb-1">Última sync</span>
          <span class="font-medium">{{ formatDate(syncState.ultima_sincronizacao) }}</span>
        </div>
        <div v-if="syncState">
          <span class="block text-xs text-muted mb-1">Período</span>
          <span class="font-medium">{{ syncState.lookback_days }} dias</span>
        </div>
        <div v-if="syncState">
          <span class="block text-xs text-muted mb-1">Último NSU</span>
          <span class="font-mono text-xs">{{ syncState.ult_nsu || '—' }}</span>
        </div>
        <div v-if="syncState?.ultimo_cstat">
          <span class="block text-xs text-muted mb-1">Último cStat</span>
          <UBadge
            :color="syncState.ultimo_cstat === '138' ? 'success' : syncState.ultimo_cstat === '656' ? 'error' : 'neutral'"
            variant="subtle"
            size="sm"
          >
            {{ syncState.ultimo_cstat }}
          </UBadge>
        </div>
      </div>

      <USeparator />

      <div class="flex items-center justify-between">
        <div>
          <span class="font-medium">NFS-e Nacional</span>
          <span class="block text-xs text-muted">Distribuição via ADN (API REST)</span>
        </div>
        <USwitch
          :model-value="syncState?.nfse_habilitada ?? false"
          :loading="togglingNfse"
          @update:model-value="(val: boolean) => emit('toggleNfse', val)"
        />
      </div>
      <div v-if="syncState?.nfse_habilitada" class="grid grid-cols-2 gap-4">
        <div>
          <span class="block text-xs text-muted mb-1">Última sync NFS-e</span>
          <span class="font-medium">{{ formatDate(syncState.ultima_sync_nfse) }}</span>
        </div>
        <div>
          <span class="block text-xs text-muted mb-1">NSU NFS-e</span>
          <span class="font-mono text-xs">{{ syncState.ult_nsu_nfse || '—' }}</span>
        </div>
      </div>
    </div>
  </UCard>
</template>
