<script setup lang="ts">
import type { EmpresaDocumentoStats } from '~/types'

const props = defineProps<{
  stats: EmpresaDocumentoStats
}>()

const { formatBRL } = useDocumentoFormatters()

const statCards = computed(() => [
  { title: 'Total de XMLs', icon: 'i-lucide-file-text', value: String(props.stats.total), color: 'primary' },
  { title: 'XML Completo', icon: 'i-lucide-file-check', value: String(props.stats.xml_completo), color: 'success' },
  { title: 'XML Resumo', icon: 'i-lucide-file-warning', value: String(props.stats.xml_resumo), color: 'warning' },
  { title: 'Manifestados', icon: 'i-lucide-shield-check', value: String(props.stats.manifestados), color: 'info' },
  { title: 'Valor Total', icon: 'i-lucide-circle-dollar-sign', value: formatBRL(props.stats.valor_total ?? 0), color: 'success' }
])
</script>

<template>
  <UPageGrid class="grid-cols-2 lg:grid-cols-5">
    <UPageCard
      v-for="stat in statCards"
      :key="stat.title"
      :title="stat.title"
      :description="stat.value"
      :icon="stat.icon"
      variant="soft"
      :ui="{
        wrapper: 'items-start',
        title: 'text-xs text-muted uppercase font-medium',
        description: 'text-xl font-bold text-highlighted mt-1',
        leadingIcon: `text-${stat.color} size-8`
      }"
    />
  </UPageGrid>
</template>
