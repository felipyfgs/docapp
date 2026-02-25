import type { MaybeRefOrGetter } from 'vue'

export type ColumnDataType = 'text' | 'option' | 'slider' | 'timerange'

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
  defaultOpen?: boolean
  commandDisabled?: boolean
  min?: number
  max?: number
}

export interface ColumnConfig<TData = unknown> extends ColumnConfigBase {
  accessor: (row: TData) => string | number | null | undefined
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
  if (type === 'text') return 'contains'
  return 'is'
}

export function useTableFilter<TData>(
  columnsInput: MaybeRefOrGetter<ColumnConfig<TData>[]>,
  dataInput: TData[] | Ref<TData[]> | (() => TData[])
) {
  const filters = ref<FiltersState>([])

  function getData(): TData[] {
    return typeof dataInput === 'function'
      ? dataInput()
      : toValue(dataInput)
  }

  const filteredData = computed(() => {
    const data = getData()
    const columns = toValue(columnsInput)

    if (filters.value.length === 0) return data

    return data.filter((row) => {
      return filters.value.every((filter) => {
        if (filter.values.length === 0) return true

        const col = columns.find(c => c.id === filter.columnId)
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

        if (filter.type === 'slider') {
          if (filter.values.length !== 2) return true
          const num = Number(val)
          if (Number.isNaN(num)) return false
          const min = Number(filter.values[0])
          const max = Number(filter.values[1])
          return num >= min && num <= max
        }

        if (filter.type === 'timerange') {
          if (filter.values.length !== 2) return true
          const dateStr = String(val ?? '')
          if (!dateStr) return false
          const d = new Date(dateStr).getTime()
          const from = new Date(filter.values[0]!).getTime()
          const to = new Date(filter.values[1]!).getTime() + 86400000 // inclusive end of day
          return d >= from && d <= to
        }

        return true
      })
    })
  })

  const hasFilters = computed(() => filters.value.length > 0)

  const facetedCounts = computed(() => {
    const data = getData()
    const columns = toValue(columnsInput)
    const counts: Record<string, Map<string, number>> = {}

    for (const col of columns) {
      if (col.type !== 'option') continue
      const map = new Map<string, number>()
      for (const row of data) {
        const val = String(col.accessor(row) ?? '')
        if (val) map.set(val, (map.get(val) || 0) + 1)
      }
      counts[col.id] = map
    }

    return counts
  })

  const facetedMinMax = computed(() => {
    const data = getData()
    const columns = toValue(columnsInput)
    const result: Record<string, [number, number]> = {}

    for (const col of columns) {
      if (col.type !== 'slider') continue
      let min = Infinity
      let max = -Infinity
      for (const row of data) {
        const num = Number(col.accessor(row))
        if (!Number.isNaN(num)) {
          if (num < min) min = num
          if (num > max) max = num
        }
      }
      if (min !== Infinity) {
        result[col.id] = [min, max]
      } else {
        result[col.id] = [col.min ?? 0, col.max ?? 100]
      }
    }

    return result
  })

  const activeFilterCount = computed(() => {
    return filters.value.reduce((count, f) => count + (f.values.length > 0 ? 1 : 0), 0)
  })

  const actions: DataTableFilterActions = {
    addFilter(columnId) {
      if (filters.value.find(f => f.columnId === columnId)) return
      const col = toValue(columnsInput).find(c => c.id === columnId)
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
      let filter = filters.value.find(f => f.columnId === columnId)
      if (!filter) {
        actions.addFilter(columnId)
        filter = filters.value.find(f => f.columnId === columnId)
      }
      if (filter) filter.values = values
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
    activeFilterCount,
    facetedCounts,
    facetedMinMax,
    actions
  }
}
