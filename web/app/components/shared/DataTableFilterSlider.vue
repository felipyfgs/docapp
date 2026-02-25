<script setup lang="ts">
import type { ColumnConfigBase, DataTableFilterActions, FiltersState } from '~/composables/useTableFilter'

const props = defineProps<{
  column: ColumnConfigBase
  filters: FiltersState
  actions: DataTableFilterActions
  facetedMinMax: Record<string, [number, number]>
}>()

const filterValues = computed(() => {
  const f = props.filters.find(f => f.columnId === props.column.id)
  return f?.values ?? []
})

const range = computed<[number, number]>(() => {
  const faceted = props.facetedMinMax[props.column.id]
  if (faceted) return faceted
  return [props.column.min ?? 0, props.column.max ?? 100]
})
const min = computed(() => range.value[0])
const max = computed(() => range.value[1])

const localMin = ref(min.value)
const localMax = ref(max.value)

watch(filterValues, (vals) => {
  if (vals.length === 2) {
    localMin.value = Number(vals[0])
    localMax.value = Number(vals[1])
  } else {
    localMin.value = min.value
    localMax.value = max.value
  }
}, { immediate: true })

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function applyFilter() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    if (localMin.value <= min.value && localMax.value >= max.value) {
      props.actions.removeFilter(props.column.id)
    } else {
      props.actions.setFilterValues(props.column.id, [String(localMin.value), String(localMax.value)])
    }
  }, 400)
}

function onSliderChange(values: number[] | undefined) {
  if (!values || values.length !== 2) return
  localMin.value = values[0]!
  localMax.value = values[1]!
  applyFilter()
}

function onMinInput(e: Event) {
  const val = Number((e.target as HTMLInputElement).value)
  if (!Number.isNaN(val)) {
    localMin.value = val
    applyFilter()
  }
}

function onMaxInput(e: Event) {
  const val = Number((e.target as HTMLInputElement).value)
  if (!Number.isNaN(val)) {
    localMax.value = val
    applyFilter()
  }
}
</script>

<template>
  <div class="grid gap-2">
    <div class="grid grid-cols-2 gap-2">
      <div class="grid gap-1">
        <label class="text-xs text-muted">
          Min.
        </label>
        <UInput
          type="number"
          size="xs"
          :model-value="String(localMin)"
          :min="min"
          :max="max"
          @input="onMinInput"
        />
      </div>
      <div class="grid gap-1">
        <label class="text-xs text-muted">
          Max.
        </label>
        <UInput
          type="number"
          size="xs"
          :model-value="String(localMax)"
          :min="min"
          :max="max"
          @input="onMaxInput"
        />
      </div>
    </div>

    <USlider
      :model-value="[localMin, localMax]"
      :min="min"
      :max="max"
      :step="1"
      size="sm"
      @update:model-value="onSliderChange"
    />
  </div>
</template>
