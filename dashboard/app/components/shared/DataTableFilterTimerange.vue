<script setup lang="ts">
import type { ColumnConfigBase, DataTableFilterActions, FiltersState } from '~/composables/useTableFilter'

const props = defineProps<{
  column: ColumnConfigBase
  filters: FiltersState
  actions: DataTableFilterActions
}>()

const filterValues = computed(() => {
  const f = props.filters.find(f => f.columnId === props.column.id)
  return f?.values ?? []
})

const fromDate = ref('')
const toDate = ref('')

watch(filterValues, (vals) => {
  if (vals.length === 2) {
    fromDate.value = vals[0]!
    toDate.value = vals[1]!
  } else {
    fromDate.value = ''
    toDate.value = ''
  }
}, { immediate: true })

function apply() {
  if (fromDate.value && toDate.value) {
    props.actions.setFilterValues(props.column.id, [fromDate.value, toDate.value])
  } else if (!fromDate.value && !toDate.value) {
    props.actions.removeFilter(props.column.id)
  }
}

function onFromChange(e: Event) {
  fromDate.value = (e.target as HTMLInputElement).value
  apply()
}

function onToChange(e: Event) {
  toDate.value = (e.target as HTMLInputElement).value
  apply()
}

interface Preset {
  label: string
  from: string
  to: string
}

function formatDate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

const presets = computed<Preset[]>(() => {
  const now = new Date()
  const today = formatDate(now)

  const daysAgo = (n: number) => {
    const d = new Date(now)
    d.setDate(d.getDate() - n)
    return formatDate(d)
  }

  const startOfMonth = formatDate(new Date(now.getFullYear(), now.getMonth(), 1))
  const prevMonthStart = formatDate(new Date(now.getFullYear(), now.getMonth() - 1, 1))
  const prevMonthEnd = formatDate(new Date(now.getFullYear(), now.getMonth(), 0))

  const threeMonthsAgoStart = formatDate(new Date(now.getFullYear(), now.getMonth() - 3, 1))

  return [
    { label: 'Hoje', from: today, to: today },
    { label: '7 dias', from: daysAgo(7), to: today },
    { label: '30 dias', from: daysAgo(30), to: today },
    { label: 'Este mês', from: startOfMonth, to: today },
    { label: 'Mês passado', from: prevMonthStart, to: prevMonthEnd },
    { label: 'Últimos 3 meses', from: threeMonthsAgoStart, to: today }
  ]
})

function applyPreset(preset: Preset) {
  fromDate.value = preset.from
  toDate.value = preset.to
  props.actions.setFilterValues(props.column.id, [preset.from, preset.to])
}
</script>

<template>
  <div class="grid gap-2">
    <div class="grid gap-1.5">
      <label class="text-xs text-muted">
        De
      </label>
      <UInput
        type="date"
        size="xs"
        :model-value="fromDate"
        @change="onFromChange"
      />
    </div>
    <div class="grid gap-1.5">
      <label class="text-xs text-muted">
        Até
      </label>
      <UInput
        type="date"
        size="xs"
        :model-value="toDate"
        @change="onToChange"
      />
    </div>

    <div class="flex flex-wrap gap-1 pt-1">
      <UButton
        v-for="preset in presets"
        :key="preset.label"
        :label="preset.label"
        size="xs"
        :color="fromDate === preset.from && toDate === preset.to ? 'primary' : 'neutral'"
        :variant="fromDate === preset.from && toDate === preset.to ? 'subtle' : 'ghost'"
        @click="applyPreset(preset)"
      />
    </div>
  </div>
</template>
