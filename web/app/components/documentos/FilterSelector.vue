<script setup lang="ts">
import type { ColumnConfigBase, FilterModel, DataTableFilterActions } from '~/composables/useTableFilter'

const props = defineProps<{
  columns: ColumnConfigBase[]
  filters: FilterModel[]
  actions: DataTableFilterActions
}>()

const open = ref(false)
const search = ref('')
const selectedColumnId = ref<string | null>(null)

const hasFilters = computed(() => props.filters.length > 0)

const selectedColumn = computed(() =>
  selectedColumnId.value ? props.columns.find(c => c.id === selectedColumnId.value) ?? null : null
)

const selectedFilter = computed(() =>
  selectedColumnId.value ? props.filters.find(f => f.columnId === selectedColumnId.value) ?? null : null
)

const filteredColumns = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return props.columns
  return props.columns.filter(c =>
    c.displayName.toLowerCase().includes(q) || c.id.toLowerCase().includes(q)
  )
})

const quickSearchResults = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (q.length < 2) return []

  const results: Array<{ column: ColumnConfigBase, option: { label: string, value: string } }> = []

  for (const col of props.columns) {
    if (col.type !== 'option') continue
    const options = toValue(col.options) ?? []
    for (const opt of options) {
      if (opt.label.toLowerCase().includes(q) || opt.value.toLowerCase().includes(q)) {
        results.push({ column: col, option: opt })
      }
    }
  }

  return results
})

const showQuickSearch = computed(() => search.value.trim().length >= 2 && quickSearchResults.value.length > 0)

function selectColumn(columnId: string) {
  selectedColumnId.value = columnId
  if (!props.filters.find(f => f.columnId === columnId)) {
    props.actions.addFilter(columnId)
  }
  search.value = ''
}

function quickSelectOption(column: ColumnConfigBase, optionValue: string) {
  props.actions.addFilterValue(column.id, optionValue)
  open.value = false
}

function handleOpenChange(val: boolean) {
  open.value = val
  if (!val) {
    setTimeout(() => {
      search.value = ''
      selectedColumnId.value = null
    }, 150)
  }
}

// When a filter is added via quickSearch, ensure a matching filter chip exists
watch(selectedFilter, (filter) => {
  if (selectedColumn.value && !filter) {
    props.actions.addFilter(selectedColumn.value.id)
  }
})
</script>

<template>
  <UPopover :open="open" :content="{ align: 'start', side: 'bottom' }" @update:open="handleOpenChange">
    <UButton
      color="neutral"
      variant="outline"
      icon="i-lucide-filter"
      :label="hasFilters ? undefined : 'Filtrar'"
      size="sm"
      class="h-7"
      @click="open = !open"
    />

    <template #content>
      <!-- Column value picker (drill-down) -->
      <template v-if="selectedColumn && selectedFilter">
        <div class="flex items-center gap-1 px-2 py-1.5 border-b border-default">
          <button
            class="flex items-center gap-1.5 text-sm text-muted hover:text-default transition-colors"
            @click="selectedColumnId = null"
          >
            <UIcon name="i-lucide-arrow-left" class="size-3.5" />
            <UIcon :name="selectedColumn.icon" class="size-3.5" />
            <span>{{ selectedColumn.displayName }}</span>
          </button>
        </div>
        <DocumentosFilterValuePicker
          :filter="selectedFilter"
          :column="selectedColumn"
          :actions="actions"
        />
      </template>

      <!-- Column picker -->
      <template v-else>
        <div class="p-2 border-b border-default">
          <UInput
            v-model="search"
            autofocus
            placeholder="Pesquisar..."
            size="sm"
            icon="i-lucide-search"
            variant="none"
          />
        </div>

        <ul class="max-h-64 overflow-y-auto py-1 w-52">
          <!-- QuickSearch: show matching option values -->
          <template v-if="showQuickSearch">
            <li
              v-for="item in quickSearchResults"
              :key="`${item.column.id}:${item.option.value}`"
              class="flex items-center gap-1.5 px-3 py-1.5 text-sm cursor-pointer hover:bg-elevated rounded-sm mx-1"
              @click="quickSelectOption(item.column, item.option.value)"
            >
              <UIcon :name="item.column.icon" class="size-4 text-muted shrink-0" />
              <span class="text-muted">{{ item.column.displayName }}</span>
              <UIcon name="i-lucide-chevron-right" class="size-3 text-muted/60 shrink-0" />
              <span>{{ item.option.label }}</span>
            </li>
          </template>

          <!-- Regular column list -->
          <template v-else>
            <li
              v-for="col in filteredColumns"
              :key="col.id"
              class="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer hover:bg-elevated rounded-sm mx-1"
              @click="selectColumn(col.id)"
            >
              <UIcon :name="col.icon" class="size-4 text-muted" />
              <span class="flex-1">{{ col.displayName }}</span>
              <UIcon name="i-lucide-chevron-right" class="size-3.5 text-muted" />
            </li>
            <li
              v-if="filteredColumns.length === 0"
              class="px-3 py-4 text-sm text-center text-muted"
            >
              Nenhum resultado
            </li>
          </template>
        </ul>
      </template>
    </template>
  </UPopover>
</template>
