<script setup lang="ts">
import type { ColumnConfigBase, DataTableFilterActions, FiltersState } from '~/composables/useTableFilter'

const props = defineProps<{
  column: ColumnConfigBase
  filters: FiltersState
  actions: DataTableFilterActions
  facetedCounts: Record<string, Map<string, number>>
}>()

const search = ref('')

const filterValues = computed(() => {
  const f = props.filters.find(f => f.columnId === props.column.id)
  return f?.values ?? []
})

const allOptions = computed(() => toValue(props.column.options) ?? [])

const filteredOptions = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return allOptions.value
  return allOptions.value.filter(o =>
    o.label.toLowerCase().includes(q) || o.value.toLowerCase().includes(q)
  )
})

const counts = computed(() => props.facetedCounts[props.column.id])

function isChecked(value: string) {
  return filterValues.value.includes(value)
}

function toggle(value: string) {
  if (isChecked(value)) {
    props.actions.removeFilterValue(props.column.id, value)
  } else {
    props.actions.addFilterValue(props.column.id, value)
  }
}

function setOnly(value: string) {
  props.actions.setFilterValues(props.column.id, [value])
}

function formatCount(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}
</script>

<template>
  <div class="grid gap-2">
    <UInput
      v-if="allOptions.length > 4"
      v-model="search"
      placeholder="Pesquisar..."
      size="xs"
      icon="i-lucide-search"
    />

    <div class="max-h-[200px] overflow-y-auto">
      <div
        v-for="option in filteredOptions"
        :key="String(option.value)"
        class="group flex items-center gap-2 px-1 py-1.5 rounded-md hover:bg-elevated/50 cursor-pointer"
        @click="toggle(option.value)"
      >
        <UCheckbox
          :model-value="isChecked(option.value)"
          readonly
          tabindex="-1"
          class="pointer-events-none"
        />

        <span class="flex-1 truncate text-sm">
          {{ option.label }}
        </span>

        <span class="font-mono text-xs text-muted tabular-nums">
          {{ counts?.has(option.value) ? formatCount(counts.get(option.value) || 0) : '' }}
        </span>

        <button
          type="button"
          class="hidden group-hover:inline text-xs text-muted hover:text-default"
          @click.stop="setOnly(option.value)"
        >
          only
        </button>
      </div>

      <p
        v-if="filteredOptions.length === 0"
        class="px-1 py-3 text-sm text-center text-muted"
      >
        Nenhum resultado
      </p>
    </div>
  </div>
</template>
