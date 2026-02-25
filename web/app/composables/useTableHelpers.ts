import type { Column, Table } from '@tanstack/table-core'
import { upperFirst } from 'scule'

export const tableUI = {
  base: 'table-fixed border-separate border-spacing-0',
  thead: '[&>tr]:bg-elevated/50 [&>tr]:after:content-none',
  tbody: '[&>tr]:last:[&>td]:border-b-0',
  th: 'px-3 py-2 text-xs first:rounded-l-lg last:rounded-r-lg border-y border-default first:border-l last:border-r',
  td: 'px-3 py-2 text-sm border-b border-default',
  separator: 'h-0'
}

export function useTableHelpers() {
  function sortableHeader<T>(label: string) {
    const UButton = resolveComponent('UButton')
    return ({ column }: { column: Column<T> }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label,
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    }
  }

  function getVisibilityItems<T>(tableApi: Table<T> | undefined, labels: Record<string, string>) {
    if (!tableApi) return []
    return tableApi
      .getAllColumns()
      .filter((column: Column<T>) => column.getCanHide())
      .map((column: Column<T>) => ({
        label: labels[column.id] || upperFirst(column.id),
        type: 'checkbox' as const,
        checked: column.getIsVisible(),
        onUpdateChecked(checked: boolean) {
          tableApi.getColumn(column.id)?.toggleVisibility(!!checked)
        },
        onSelect(e?: Event) {
          e?.preventDefault()
        }
      }))
  }

  return { sortableHeader, getVisibilityItems }
}
