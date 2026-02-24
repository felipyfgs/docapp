<script setup lang="ts">
import type { ColumnConfig, FilterModel, DataTableFilterActions, FilterOperator } from '~/composables/useTableFilter'

const props = defineProps<{
  filter: FilterModel
  column: ColumnConfig<unknown>
  actions: DataTableFilterActions
}>()

const OPERATOR_LABELS: Record<FilterOperator, string> = {
  'is': 'é',
  'is not': 'não é',
  'contains': 'contém',
  'does not contain': 'não contém'
}

const operatorLabel = computed(() => OPERATOR_LABELS[props.filter.operator] ?? props.filter.operator)

const operatorOptions = computed(() => {
  if (props.column.type === 'option') {
    return [
      { label: 'é', value: 'is' as FilterOperator },
      { label: 'não é', value: 'is not' as FilterOperator }
    ]
  }
  return [
    { label: 'contém', value: 'contains' as FilterOperator },
    { label: 'não contém', value: 'does not contain' as FilterOperator }
  ]
})

const valuesLabel = computed(() => {
  const vals = props.filter.values
  if (vals.length === 0) return '...'
  if (props.column.type === 'option') {
    const options = toValue(props.column.options) ?? []
    const selected = options.filter(o => vals.includes(o.value))
    if (selected.length === 0) return vals.join(', ')
    if (selected.length === 1) return selected[0]!.label
    return `${selected.length} ${props.column.displayName.toLowerCase()}`
  }
  return vals[0] ?? '...'
})

const operatorOpen = ref(false)
const valueOpen = ref(false)
</script>

<template>
  <div class="flex h-7 items-center rounded-full border border-border bg-background shadow-xs text-xs whitespace-nowrap">
    <!-- Subject -->
    <div class="flex items-center gap-1 px-2.5 select-none">
      <UIcon :name="column.icon" class="size-3.5 text-muted" />
      <span class="font-medium text-default">{{ column.displayName }}</span>
    </div>

    <!-- Separator -->
    <div class="w-px self-stretch bg-border" />

    <!-- Operator -->
    <UPopover v-model:open="operatorOpen" :content="{ align: 'start', side: 'bottom' }">
      <button
        class="px-2 h-full hover:bg-elevated/80 text-muted transition-colors"
        @click="operatorOpen = !operatorOpen"
      >
        {{ operatorLabel }}
      </button>
      <template #content>
        <ul class="py-1 w-36">
          <li
            v-for="op in operatorOptions"
            :key="op.value"
            class="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer hover:bg-elevated rounded-sm mx-1"
            :class="filter.operator === op.value ? 'text-primary font-medium' : ''"
            @click="() => { actions.setFilterOperator(filter.columnId, op.value); operatorOpen = false }"
          >
            {{ op.label }}
          </li>
        </ul>
      </template>
    </UPopover>

    <!-- Separator -->
    <div class="w-px self-stretch bg-border" />

    <!-- Values -->
    <UPopover v-model:open="valueOpen" :content="{ align: 'start', side: 'bottom' }">
      <button
        class="px-2 h-full hover:bg-elevated/80 transition-colors max-w-40 truncate"
        @click="valueOpen = !valueOpen"
      >
        {{ valuesLabel }}
      </button>
      <template #content>
        <FilterValuePicker :filter="filter" :column="column" :actions="actions" />
      </template>
    </UPopover>

    <!-- Separator -->
    <div class="w-px self-stretch bg-border" />

    <!-- Remove -->
    <button
      class="px-2 h-full hover:bg-elevated/80 rounded-r-full transition-colors text-muted hover:text-default"
      @click="actions.removeFilter(filter.columnId)"
    >
      <UIcon name="i-lucide-x" class="size-3.5" />
    </button>
  </div>
</template>
