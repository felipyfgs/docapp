<script setup lang="ts">
import { eachDayOfInterval } from 'date-fns'
import type { Period, Range } from '~/types'

const model = defineModel<Period>({ required: true })

const props = defineProps<{
  range: Range
}>()

const periodLabels: Record<Period, string> = {
  daily: 'Diário',
  weekly: 'Semanal',
  monthly: 'Mensal'
}

const days = computed(() => eachDayOfInterval(props.range))

const periods = computed(() => {
  if (days.value.length <= 8) {
    return [{ label: periodLabels.daily, value: 'daily' as Period }]
  }

  if (days.value.length <= 31) {
    return [
      { label: periodLabels.daily, value: 'daily' as Period },
      { label: periodLabels.weekly, value: 'weekly' as Period }
    ]
  }

  return [
    { label: periodLabels.weekly, value: 'weekly' as Period },
    { label: periodLabels.monthly, value: 'monthly' as Period }
  ]
})

watch(periods, () => {
  const validValues = periods.value.map(p => p.value)
  if (!validValues.includes(model.value)) {
    model.value = validValues[0]!
  }
})
</script>

<template>
  <USelect
    v-model="model"
    :items="periods"
    variant="ghost"
    class="data-[state=open]:bg-elevated"
    :ui="{ trailingIcon: 'group-data-[state=open]:rotate-180 transition-transform duration-200' }"
  />
</template>
