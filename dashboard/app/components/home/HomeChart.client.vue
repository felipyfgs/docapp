<script setup lang="ts">
import { format } from 'date-fns'
import { ptBR } from 'date-fns/locale'
import { VisXYContainer, VisLine, VisAxis, VisArea, VisCrosshair, VisTooltip } from '@unovis/vue'
import type { DashboardChartPoint, Period } from '~/types'

const cardRef = useTemplateRef<HTMLElement | null>('cardRef')

const props = defineProps<{
  data: DashboardChartPoint[] | null
  period: Period
}>()

const { formatBRL } = useDocumentoFormatters()
const { width } = useElementSize(cardRef)

const chartData = computed(() => {
  if (!props.data?.length) return []
  return props.data.map(p => ({
    date: new Date(p.date),
    amount: p.valor_total,
    count: p.count
  }))
})

type DataRecord = { date: Date, amount: number, count: number }

const x = (_: DataRecord, i: number) => i
const y = (d: DataRecord) => d.amount

const total = computed(() => chartData.value.reduce((acc, { amount }) => acc + amount, 0))

const formatDate = (date: Date): string => {
  return ({
    daily: format(date, 'd MMM', { locale: ptBR }),
    weekly: format(date, 'd MMM', { locale: ptBR }),
    monthly: format(date, 'MMM yyyy', { locale: ptBR })
  })[props.period]
}

const xTicks = (i: number) => {
  if (i === 0 || i === chartData.value.length - 1 || !chartData.value[i]) return ''
  return formatDate(chartData.value[i].date)
}

const template = (d: DataRecord) => `${formatDate(d.date)}: ${formatBRL(d.amount)} (${d.count} docs)`
</script>

<template>
  <UCard ref="cardRef" :ui="{ root: 'overflow-visible', body: '!px-0 !pt-0 !pb-3' }">
    <template #header>
      <div>
        <p class="text-xs text-muted uppercase mb-1.5">
          Valor Movimentado
        </p>
        <p class="text-3xl text-highlighted font-semibold">
          {{ formatBRL(total) }}
        </p>
      </div>
    </template>

    <template v-if="chartData.length > 0">
      <VisXYContainer
        :data="chartData"
        :padding="{ top: 40 }"
        class="h-96"
        :width="width"
      >
        <VisLine :x="x" :y="y" color="var(--ui-primary)" />
        <VisArea :x="x" :y="y" color="var(--ui-primary)" :opacity="0.1" />
        <VisAxis type="x" :x="x" :tick-format="xTicks" />
        <VisCrosshair color="var(--ui-primary)" :template="template" />
        <VisTooltip />
      </VisXYContainer>
    </template>

    <div v-else class="h-96 flex items-center justify-center text-muted text-sm">
      Nenhum dado para o período selecionado
    </div>
  </UCard>
</template>

<style scoped>
.unovis-xy-container {
  --vis-crosshair-line-stroke-color: var(--ui-primary);
  --vis-crosshair-circle-stroke-color: var(--ui-bg);
  --vis-axis-grid-color: var(--ui-border);
  --vis-axis-tick-color: var(--ui-border);
  --vis-axis-tick-label-color: var(--ui-text-dimmed);
  --vis-tooltip-background-color: var(--ui-bg);
  --vis-tooltip-border-color: var(--ui-border);
  --vis-tooltip-text-color: var(--ui-text-highlighted);
}
</style>
