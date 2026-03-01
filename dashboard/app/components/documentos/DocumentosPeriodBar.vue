<script setup lang="ts">
import { CalendarDate, getLocalTimeZone } from '@internationalized/date'

type Granularity = 'monthly' | 'quarterly' | 'yearly' | 'custom'

interface PeriodRange {
  start: Date
  end: Date
}

const model = defineModel<PeriodRange>({ required: true })
const granularity = ref<Granularity>('monthly')

const granularityItems = [
  { label: 'Mensal', value: 'monthly' as Granularity },
  { label: 'Trimestral', value: 'quarterly' as Granularity },
  { label: 'Anual', value: 'yearly' as Granularity },
  { label: 'Personalizado', value: 'custom' as Granularity }
]

function fmtDate(d: Date): string {
  return d.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

const periodLabel = computed(() => `${fmtDate(model.value.start)} até ${fmtDate(model.value.end)}`)

function startOfMonth(date: Date): Date { return new Date(date.getFullYear(), date.getMonth(), 1) }
function endOfMonth(date: Date): Date { return new Date(date.getFullYear(), date.getMonth() + 1, 0) }
function startOfQuarter(date: Date): Date { const q = Math.floor(date.getMonth() / 3) * 3; return new Date(date.getFullYear(), q, 1) }
function endOfQuarter(date: Date): Date { const q = Math.floor(date.getMonth() / 3) * 3; return new Date(date.getFullYear(), q + 3, 0) }
function startOfYear(date: Date): Date { return new Date(date.getFullYear(), 0, 1) }
function endOfYear(date: Date): Date { return new Date(date.getFullYear(), 11, 31) }

function snapToGranularity(date: Date, g: Granularity): PeriodRange {
  switch (g) {
    case 'quarterly': return { start: startOfQuarter(date), end: endOfQuarter(date) }
    case 'yearly': return { start: startOfYear(date), end: endOfYear(date) }
    default: return { start: startOfMonth(date), end: endOfMonth(date) }
  }
}

function navigate(direction: -1 | 1) {
  const ref = model.value.start
  let next: Date
  switch (granularity.value) {
    case 'quarterly': next = new Date(ref.getFullYear(), ref.getMonth() + (3 * direction), 1); break
    case 'yearly': next = new Date(ref.getFullYear() + direction, 0, 1); break
    default: next = new Date(ref.getFullYear(), ref.getMonth() + direction, 1)
  }
  model.value = snapToGranularity(next, granularity.value)
}

watch(granularity, (g) => {
  if (g !== 'custom') model.value = snapToGranularity(model.value.start, g)
})

const toCalendarDate = (d: Date) => new CalendarDate(d.getFullYear(), d.getMonth() + 1, d.getDate())

const calendarRange = computed({
  get: () => ({
    start: model.value.start ? toCalendarDate(model.value.start) : undefined,
    end: model.value.end ? toCalendarDate(model.value.end) : undefined
  }),
  set: (v: { start: CalendarDate | null, end: CalendarDate | null }) => {
    model.value = {
      start: v.start ? v.start.toDate(getLocalTimeZone()) : new Date(),
      end: v.end ? v.end.toDate(getLocalTimeZone()) : new Date()
    }
  }
})
</script>

<template>
  <div class="flex items-center gap-2">
    <span class="text-sm text-muted">Período</span>

    <UButton
      v-if="granularity !== 'custom'"
      icon="i-lucide-chevron-left"
      color="neutral"
      variant="ghost"
      @click="navigate(-1)"
    />

    <UPopover :content="{ align: 'center' }">
      <UButton color="neutral" variant="subtle" icon="i-lucide-calendar">
        {{ periodLabel }}
      </UButton>

      <template #content>
        <UCalendar v-model="calendarRange" class="p-2" :number-of-months="2" range />
      </template>
    </UPopover>

    <UButton
      v-if="granularity !== 'custom'"
      icon="i-lucide-chevron-right"
      color="neutral"
      variant="ghost"
      @click="navigate(1)"
    />

    <USelect
      :model-value="granularity"
      :items="granularityItems"
      class="w-36"
      @update:model-value="(v: string) => { granularity = v as Granularity }"
    />
  </div>
</template>
