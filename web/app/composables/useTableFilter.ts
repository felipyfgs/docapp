import type { MaybeRefOrGetter } from 'vue'

export type ColumnDataType = 'text' | 'option'

export interface ColumnOption {
  label: string
  value: string
  icon?: string
}

export interface ColumnConfigBase {
  id: string
  displayName: string
  icon: string
  type: ColumnDataType
  options?: MaybeRefOrGetter<ColumnOption[]>
}

export interface ColumnConfig<TData = unknown> extends ColumnConfigBase {
  accessor: (row: TData) => string | null | undefined
}

export type OptionFilterOperator = 'is' | 'is not'
export type TextFilterOperator = 'contains' | 'does not contain'
export type FilterOperator = OptionFilterOperator | TextFilterOperator

export interface FilterModel {
  columnId: string
  type: ColumnDataType
  operator: FilterOperator
  values: string[]
}

export type FiltersState = FilterModel[]

export interface DataTableFilterActions {
  addFilter: (columnId: string) => void
  removeFilter: (columnId: string) => void
  addFilterValue: (columnId: string, value: string) => void
  removeFilterValue: (columnId: string, value: string) => void
  setFilterValues: (columnId: string, values: string[]) => void
  setFilterOperator: (columnId: string, operator: FilterOperator) => void
  clearAll: () => void
}

function defaultOperator(type: ColumnDataType): FilterOperator {
  return type === 'text' ? 'contains' : 'is'
}

export function useTableFilter<TData>(
  columnsInput: ColumnConfig<TData>[],
  dataInput: TData[] | Ref<TData[]> | (() => TData[])
) {
  const filters = ref<FiltersState>([])

  const filteredData = computed(() => {
    const data = typeof dataInput === 'function'
      ? dataInput()
      : toValue(dataInput)

    if (filters.value.length === 0) return data

    return data.filter((row) => {
      return filters.value.every((filter) => {
        if (filter.values.length === 0) return true

        const col = columnsInput.find(c => c.id === filter.columnId)
        if (!col) return true

        const val = col.accessor(row)

        if (filter.type === 'option') {
          const match = filter.values.includes(String(val ?? ''))
          return filter.operator === 'is' ? match : !match
        }

        if (filter.type === 'text') {
          const haystack = String(val ?? '').toLowerCase()
          const needle = (filter.values[0] ?? '').toLowerCase()
          const match = haystack.includes(needle)
          return filter.operator === 'contains' ? match : !match
        }

        return true
      })
    })
  })

  const hasFilters = computed(() => filters.value.length > 0)

  const actions: DataTableFilterActions = {
    addFilter(columnId) {
      if (filters.value.find(f => f.columnId === columnId)) return
      const col = columnsInput.find(c => c.id === columnId)
      if (!col) return
      filters.value = [...filters.value, {
        columnId,
        type: col.type,
        operator: defaultOperator(col.type),
        values: []
      }]
    },

    removeFilter(columnId) {
      filters.value = filters.value.filter(f => f.columnId !== columnId)
    },

    addFilterValue(columnId, value) {
      const filter = filters.value.find(f => f.columnId === columnId)
      if (!filter) {
        actions.addFilter(columnId)
        const newFilter = filters.value.find(f => f.columnId === columnId)
        if (newFilter) newFilter.values = [value]
        return
      }
      if (!filter.values.includes(value)) {
        filter.values = [...filter.values, value]
      }
    },

    removeFilterValue(columnId, value) {
      const filter = filters.value.find(f => f.columnId === columnId)
      if (!filter) return
      filter.values = filter.values.filter(v => v !== value)
      if (filter.values.length === 0) {
        actions.removeFilter(columnId)
      }
    },

    setFilterValues(columnId, values) {
      const filter = filters.value.find(f => f.columnId === columnId)
      if (!filter) return
      filter.values = values
    },

    setFilterOperator(columnId, operator) {
      const filter = filters.value.find(f => f.columnId === columnId)
      if (!filter) return
      filter.operator = operator
    },

    clearAll() {
      filters.value = []
    }
  }

  return {
    filters: filters as Readonly<Ref<FiltersState>>,
    filteredData,
    hasFilters,
    actions
  }
}
