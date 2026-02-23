<script setup lang="ts">
import type { TableColumn, TableRow } from '@nuxt/ui'
import { getPaginationRowModel } from '@tanstack/table-core'
import type { Row } from '@tanstack/table-core'
import type { Empresa } from '~/types'

const props = defineProps<{
  data: Empresa[] | null | undefined
  status: string
}>()

const emit = defineEmits<{
  delete: [empresa: Empresa]
  edit: [empresa: Empresa]
  sync: [empresa: Empresa]
}>()

const toast = useToast()

const UBadge = resolveComponent('UBadge')
const UButton = resolveComponent('UButton')
const UCheckbox = resolveComponent('UCheckbox')
const UDropdownMenu = resolveComponent('UDropdownMenu')

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const table = useTemplateRef<{ tableApi: any }>('table')

const columnFilters = ref<{ id: string, value: string }[]>([])
const columnVisibility = ref<Record<string, boolean>>({})
const rowSelection = ref<Record<string, boolean>>({})

const pagination = ref({ pageIndex: 0, pageSize: 10 })

const search = defineModel<string>('search', { default: '' })
const situacaoFilter = defineModel<string>('situacaoFilter', { default: 'all' })

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  let result = props.data ?? []

  if (q) {
    result = result.filter(e =>
      e.cnpj.includes(q)
      || e.razao_social.toLowerCase().includes(q)
      || e.nome_fantasia?.toLowerCase().includes(q)
      || e.cidade?.toLowerCase().includes(q)
    )
  }

  if (situacaoFilter.value && situacaoFilter.value !== 'all') {
    result = result.filter(e => e.situacao_cadastral === situacaoFilter.value)
  }

  return result
})

const selectedRows = computed((): Empresa[] => {
  if (!table.value?.tableApi) return []
  return table.value.tableApi.getFilteredSelectedRowModel().rows.map((r: Row<Empresa>) => r.original)
})

defineExpose({ selectedRows, tableApi: computed(() => table.value?.tableApi) })

function rowItems(row: Row<Empresa>) {
  return [
    [
      { type: 'label' as const, label: 'Ações' }
    ],
    [
      {
        label: 'Editar',
        icon: 'i-lucide-pencil',
        onSelect() {
          emit('edit', row.original)
        }
      },
      {
        label: 'Copiar CNPJ',
        icon: 'i-lucide-copy',
        onSelect() {
          navigator.clipboard.writeText(row.original.cnpj)
          toast.add({ title: 'CNPJ copiado', color: 'success' })
        }
      }
    ],
    [
      {
        label: 'Sincronizar',
        icon: 'i-lucide-refresh-cw',
        onSelect() {
          emit('sync', row.original)
        }
      }
    ],
    [
      {
        label: 'Deletar',
        icon: 'i-lucide-trash',
        color: 'error' as const,
        onSelect() {
          emit('delete', row.original)
        }
      }
    ]
  ]
}

function formatCNPJ(cnpj: string): string {
  const d = cnpj.replace(/\D/g, '')
  if (d.length !== 14) return cnpj
  return `${d.slice(0, 2)}.${d.slice(2, 5)}.${d.slice(5, 8)}/${d.slice(8, 12)}-${d.slice(12)}`
}

const columns: TableColumn<Empresa>[] = [
  {
    id: 'select',
    header: ({ table }) =>
      h(UCheckbox, {
        'modelValue': table.getIsSomePageRowsSelected()
          ? 'indeterminate'
          : table.getIsAllPageRowsSelected(),
        'onUpdate:modelValue': (value: boolean | 'indeterminate') =>
          table.toggleAllPageRowsSelected(!!value),
        'ariaLabel': 'Selecionar todos'
      }),
    cell: ({ row }) =>
      h(UCheckbox, {
        'modelValue': row.getIsSelected(),
        'onUpdate:modelValue': (value: boolean | 'indeterminate') => row.toggleSelected(!!value),
        'ariaLabel': 'Selecionar linha'
      })
  },
  {
    accessorKey: 'razao_social',
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Empresa',
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    },
    cell: ({ row }) => h('div', { class: 'min-w-48 max-w-64' }, [
      h('p', { class: 'font-medium text-highlighted truncate' }, row.original.razao_social),
      h('p', { class: 'text-xs text-muted' }, formatCNPJ(row.original.cnpj))
    ])
  },
  {
    accessorKey: 'situacao_cadastral',
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Situação',
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    },
    cell: ({ row }) => {
      const s = row.original.situacao_cadastral
      const isAtiva = s === 'Ativa'
      return h(UBadge, {
        variant: 'subtle',
        color: isAtiva ? 'success' as const : 'warning' as const,
        leadingIcon: isAtiva ? 'i-lucide-circle-check' : 'i-lucide-circle-alert'
      }, () => s || '—')
    }
  },
  {
    id: 'localidade',
    accessorFn: row => `${row.cidade || ''}${row.estado || ''}`,
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Cidade/UF',
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    },
    cell: ({ row }) => {
      const { cidade, estado } = row.original
      if (!cidade && !estado) return '—'
      const texto = `${cidade || ''}${cidade && estado ? '/' : ''}${estado || ''}`
      return h('div', { class: 'flex items-center gap-1' }, [
        h('span', { class: 'i-lucide-map-pin size-3 text-muted shrink-0' }),
        h('span', texto)
      ])
    }
  },
  {
    accessorKey: 'lookback_days',
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Período',
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    },
    cell: ({ row }) => {
      return h('div', { class: 'flex items-center gap-1' }, [
        h('span', { class: 'i-lucide-calendar size-3 text-muted shrink-0' }),
        h('span', `${row.original.lookback_days} dias`)
      ])
    }
  },
  {
    id: 'certificado',
    accessorFn: row => row.certificado_status || 'sem_certificado',
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Certificado',
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    },
    cell: ({ row }) => {
      const status = row.original.certificado_status
      type CertColor = 'error' | 'warning' | 'success' | 'neutral'
      const colorMap: Record<string, CertColor> = {
        vencido: 'error',
        prestes_a_vencer: 'warning',
        valido: 'success',
        sem_certificado: 'neutral'
      }
      const labelMap: Record<string, string> = {
        vencido: 'Vencido',
        prestes_a_vencer: 'Prestes a vencer',
        valido: 'Válido',
        sem_certificado: 'Sem certificado'
      }
      const iconMap: Record<string, string> = {
        vencido: 'i-lucide-shield-x',
        prestes_a_vencer: 'i-lucide-shield-alert',
        valido: 'i-lucide-shield-check',
        sem_certificado: 'i-lucide-shield-off'
      }
      const key = status || 'sem_certificado'
      return h(UBadge, {
        variant: 'subtle',
        color: colorMap[key] || 'neutral',
        leadingIcon: iconMap[key]
      }, () => labelMap[key] || 'Sem certificado')
    }
  },
  {
    id: 'actions',
    cell: ({ row }) =>
      h('div', { class: 'text-right' },
        h(UDropdownMenu, {
          content: { align: 'end' },
          items: rowItems(row)
        }, () => h(UButton, {
          icon: 'i-lucide-ellipsis-vertical',
          color: 'neutral',
          variant: 'ghost',
          class: 'ml-auto'
        }))
      )
  }
]

const columnLabels: Record<string, string> = {
  razao_social: 'Empresa',
  situacao_cadastral: 'Situação',
  localidade: 'Cidade/UF',
  lookback_days: 'Período',
  certificado: 'Certificado'
}

function getVisibilityItems() {
  if (!table.value?.tableApi) return []
  return table.value.tableApi
    .getAllColumns()
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    .filter((column: any) => column.getCanHide())
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    .map((column: any) => ({
      label: columnLabels[column.id] ?? column.id,
      type: 'checkbox' as const,
      checked: column.getIsVisible(),
      onUpdateChecked(checked: boolean) {
        table.value?.tableApi?.getColumn(column.id)?.toggleVisibility(!!checked)
      },
      onSelect(e?: Event) {
        e?.preventDefault()
      }
    }))
}
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-1.5">
    <UInput
      v-model="search"
      icon="i-lucide-search"
      placeholder="Filtrar por CNPJ, razão social, cidade..."
      class="max-w-sm"
    />

    <div class="flex flex-wrap items-center gap-1.5">
      <slot name="actions" :selected-rows="selectedRows" />

      <USelect
        v-model="situacaoFilter"
        :items="[
          { label: 'Todas', value: 'all' },
          { label: 'Ativa', value: 'Ativa' },
          { label: 'Baixada', value: 'Baixada' },
          { label: 'Suspensa', value: 'Suspensa' },
          { label: 'Inapta', value: 'Inapta' }
        ]"
        :ui="{ trailingIcon: 'group-data-[state=open]:rotate-180 transition-transform duration-200' }"
        placeholder="Situação"
        class="min-w-28"
      />

      <UDropdownMenu
        :items="getVisibilityItems()"
        :content="{ align: 'end' }"
      >
        <UButton
          label="Exibir"
          color="neutral"
          variant="outline"
          trailing-icon="i-lucide-settings-2"
        />
      </UDropdownMenu>
    </div>
  </div>

  <UTable
    ref="table"
    v-model:column-filters="columnFilters"
    v-model:column-visibility="columnVisibility"
    v-model:row-selection="rowSelection"
    v-model:pagination="pagination"
    :pagination-options="{ getPaginationRowModel: getPaginationRowModel() }"
    :data="filtered"
    :columns="columns"
    :loading="status === 'pending'"
    class="shrink-0"
    :ui="{
      base: 'table-fixed border-separate border-spacing-0',
      thead: '[&>tr]:bg-elevated/50 [&>tr]:after:content-none',
      tbody: '[&>tr]:last:[&>td]:border-b-0',
      th: 'py-2 first:rounded-l-lg last:rounded-r-lg border-y border-default first:border-l last:border-r',
      td: 'border-b border-default',
      separator: 'h-0',
      tr: 'cursor-pointer hover:bg-elevated/40 transition-colors'
    }"
    @select="(_: Event, row: TableRow<Empresa>) => navigateTo(`/empresas/${row.original.id}`)"
  />

  <div class="flex items-center justify-between gap-3 border-t border-default pt-4 mt-auto">
    <div class="text-sm text-muted">
      {{ table?.tableApi?.getFilteredSelectedRowModel().rows.length || 0 }} de
      {{ table?.tableApi?.getFilteredRowModel().rows.length || 0 }} linha(s) selecionada(s).
    </div>
    <UPagination
      :default-page="(table?.tableApi?.getState().pagination.pageIndex || 0) + 1"
      :items-per-page="table?.tableApi?.getState().pagination.pageSize"
      :total="table?.tableApi?.getFilteredRowModel().rows.length"
      @update:page="(p: number) => table?.tableApi?.setPageIndex(p - 1)"
    />
  </div>
</template>
