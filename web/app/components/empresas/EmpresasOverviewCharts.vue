<script setup lang="ts">
import { VisXYContainer, VisGroupedBar, VisLine, VisAxis, VisCrosshair, VisTooltip } from '@unovis/vue'
import type { CompetenciaCount } from '~/types'

const props = defineProps<{
  porCompetencia: CompetenciaCount[]
}>()

const { formatBRL } = useDocumentoFormatters()

const chartData = computed(() => props.porCompetencia)
const hasValorData = computed(() => props.porCompetencia.some(d => d.valor_total > 0))

const x = (_: CompetenciaCount, i: number) => i
const xTicks = (i: number) => chartData.value[i]?.competencia ?? ''
const template = (d: CompetenciaCount) => `${d.competencia}: ${d.count} docs`
</script>

<template>
  <UCard v-if="porCompetencia.length > 0">
    <template #header>
      <div class="flex items-center gap-2">
        <UIcon name="i-lucide-bar-chart-3" class="size-5 text-primary" />
        <h3 class="font-semibold text-highlighted">
          Volume de XMLs por Competência
        </h3>
      </div>
    </template>

    <VisXYContainer
      :data="chartData"
      :padding="{ top: 20, left: 8, right: 8 }"
      class="h-48"
    >
      <VisGroupedBar :x="x" :y="[(d: CompetenciaCount) => d.count]" color="['var(--ui-primary)']" />
      <VisAxis type="x" :x="x" :tick-format="xTicks" />
      <VisCrosshair color="var(--ui-primary)" :template="template" />
      <VisTooltip />
    </VisXYContainer>
  </UCard>

  <UCard v-if="hasValorData">
    <template #header>
      <div class="flex items-center gap-2">
        <UIcon name="i-lucide-trending-up" class="size-5 text-success" />
        <h3 class="font-semibold text-highlighted">
          Valor Total por Competência
        </h3>
      </div>
    </template>

    <VisXYContainer
      :data="chartData"
      :padding="{ top: 20, left: 8, right: 8 }"
      class="h-48"
    >
      <VisLine :x="x" :y="(d: CompetenciaCount) => d.valor_total" color="var(--ui-color-success-500)" />
      <VisAxis type="x" :x="x" :tick-format="xTicks" />
      <VisCrosshair
        color="var(--ui-color-success-500)"
        :template="(d: CompetenciaCount) => `${d.competencia}: ${formatBRL(d.valor_total)}`"
      />
      <VisTooltip />
    </VisXYContainer>
  </UCard>
</template>

<style scoped>
.unovis-xy-container {
  --vis-axis-grid-color: var(--ui-border);
  --vis-axis-tick-color: var(--ui-border);
  --vis-axis-tick-label-color: var(--ui-text-dimmed);
  --vis-tooltip-background-color: var(--ui-bg);
  --vis-tooltip-border-color: var(--ui-border);
  --vis-tooltip-text-color: var(--ui-text-highlighted);
  --vis-crosshair-line-stroke-color: var(--ui-primary);
}
</style>
