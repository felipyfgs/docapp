<script setup lang="ts">
import type { DashboardStats } from '~/types'

const props = defineProps<{
  stats: DashboardStats | null
}>()

const { formatBRL } = useDocumentoFormatters()

const cards = computed(() => {
  const s = props.stats

  return [
    {
      title: 'Documentos',
      icon: 'i-lucide-file-text',
      value: s ? s.total_documentos.toLocaleString('pt-BR') : '—',
      to: '/documentos'
    },
    {
      title: 'Empresas Ativas',
      icon: 'i-lucide-building-2',
      value: s ? s.total_empresas.toLocaleString('pt-BR') : '—',
      to: '/empresas?filter=situacao_cadastral:Ativa'
    },
    {
      title: 'Valor Total',
      icon: 'i-lucide-circle-dollar-sign',
      value: s ? formatBRL(s.valor_total) : '—',
      to: '/documentos'
    },
    {
      title: 'Pendentes Manifestação',
      icon: 'i-lucide-stamp',
      value: s ? s.pendentes_manifestacao.toLocaleString('pt-BR') : '—',
      to: '/documentos?filter=manifestacao_status:pendente'
    }
  ]
})
</script>

<template>
  <UPageGrid class="lg:grid-cols-4 gap-4 sm:gap-6 lg:gap-px">
    <UPageCard
      v-for="(card, index) in cards"
      :key="index"
      :icon="card.icon"
      :title="card.title"
      :to="card.to"
      variant="subtle"
      :ui="{
        container: 'gap-y-1.5',
        wrapper: 'items-start',
        leading: 'p-2.5 rounded-full bg-primary/10 ring ring-inset ring-primary/25 flex-col',
        title: 'font-normal text-muted text-xs uppercase'
      }"
      class="lg:rounded-none first:rounded-l-lg last:rounded-r-lg hover:z-1"
    >
      <span class="text-2xl font-semibold text-highlighted">
        {{ card.value }}
      </span>
    </UPageCard>
  </UPageGrid>
</template>
