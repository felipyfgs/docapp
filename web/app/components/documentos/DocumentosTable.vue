<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { getPaginationRowModel } from '@tanstack/table-core'
import type { Column, Row, Table } from '@tanstack/table-core'
import { upperFirst } from 'scule'
import type { DocumentoFiscal } from '~/types'
import type { ColumnConfig } from '~/composables/useTableFilter'

const props = defineProps<{
  data: DocumentoFiscal[] | null | undefined
  status: string
}>()

const emit = defineEmits<{
  viewXml: [documento: DocumentoFiscal]
}>()

const toast = useToast()

const UBadge = resolveComponent('UBadge')
const UButton = resolveComponent('UButton')
const UCheckbox = resolveComponent('UCheckbox')
const UDropdownMenu = resolveComponent('UDropdownMenu')

const table = useTemplateRef<{ tableApi: Table<DocumentoFiscal> }>('table')

const columnFilters = ref<{ id: string, value: string }[]>([])
const columnVisibility = ref<Record<string, boolean>>({
  chave_acesso: false,
  xml_resumo: false,
  manifestacao_status: false
})
const rowSelection = ref<Record<string, boolean>>({})

const pagination = ref({ pageIndex: 0, pageSize: 15 })

const search = defineModel<string>('search', { default: '' })

const competenciaOptions = computed(() => {
  const all = [...new Set((props.data ?? []).map(d => d.competencia).filter(Boolean))]
  return all.sort().reverse().map(c => ({ label: c!, value: c! }))
})

const filterColumns: ColumnConfig<DocumentoFiscal>[] = [
  {
    id: 'tipo_documento',
    accessor: row => row.tipo_documento,
    displayName: 'Tipo',
    icon: 'i-lucide-file-text',
    type: 'option',
    options: [
      { label: 'NF-e', value: 'nf-e' },
      { label: 'NFC-e', value: 'nfc-e' },
      { label: 'CT-e', value: 'ct-e' },
      { label: 'NFS-e', value: 'nfs-e' },
      { label: 'Desconhecido', value: 'desconhecido' }
    ]
  },
  {
    id: 'status_documento',
    accessor: row => row.status_documento,
    displayName: 'Status',
    icon: 'i-lucide-circle-check',
    type: 'option',
    options: [
      { label: 'Autorizada', value: 'autorizada' },
      { label: 'Cancelada', value: 'cancelada' },
      { label: 'Denegada', value: 'denegada' },
      { label: 'Desconhecido', value: 'desconhecido' }
    ]
  },
  {
    id: 'xml_resumo',
    accessor: row => row.xml_resumo ? 'resumo' : 'completo',
    displayName: 'XML',
    icon: 'i-lucide-file-code',
    type: 'option',
    options: [
      { label: 'Completo', value: 'completo' },
      { label: 'Resumo', value: 'resumo' }
    ]
  },
  {
    id: 'competencia',
    accessor: row => row.competencia ?? null,
    displayName: 'Competência',
    icon: 'i-lucide-calendar',
    type: 'option',
    options: competenciaOptions
  },
  {
    id: 'manifestacao_status',
    accessor: row => row.manifestacao_status ?? 'pendente',
    displayName: 'Manifestação',
    icon: 'i-lucide-stamp',
    type: 'option',
    options: [
      { label: 'Pendente', value: 'pendente' },
      { label: 'Ciência', value: 'ciencia' },
      { label: 'Confirmada', value: 'confirmada' },
      { label: 'Desconhecida', value: 'desconhecida' },
      { label: 'Não Realizada', value: 'nao_realizada' }
    ]
  }
]

const { filters, filteredData, actions: filterActions } = useTableFilter(filterColumns, () => props.data ?? [])

const filtered = computed(() => {
  const query = search.value.toLowerCase().trim()
  let result = filteredData.value

  if (query) {
    result = result.filter((documento) => {
      const values = [
        documento.chave_acesso,
        documento.numero_documento,
        documento.emitente_nome,
        documento.destinatario_nome,
        documento.emitente_cnpj,
        documento.destinatario_cnpj
      ]
      return values.some(value => value?.toLowerCase().includes(query))
    })
  }

  return result
})

const selectedRows = computed((): DocumentoFiscal[] => {
  if (!table.value?.tableApi) return []
  return table.value.tableApi.getFilteredSelectedRowModel().rows.map((row: Row<DocumentoFiscal>) => row.original)
})

defineExpose({ selectedRows, tableApi: computed(() => table.value?.tableApi) })

function rowItems(row: Row<DocumentoFiscal>) {
  return [
    [
      { type: 'label' as const, label: 'Ações' }
    ],
    [
      {
        label: 'Ver XML',
        icon: 'i-lucide-file-code-2',
        onSelect() {
          emit('viewXml', row.original)
        }
      },
      {
        label: 'Copiar chave',
        icon: 'i-lucide-copy',
        onSelect() {
          navigator.clipboard.writeText(row.original.chave_acesso)
          toast.add({ title: 'Chave de acesso copiada', color: 'success' })
        }
      }
    ]
  ]
}

function formatBRL(value: number): string {
  return value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

function formatCNPJ(cnpj: string | undefined): string {
  const digits = (cnpj || '').replace(/\D/g, '')
  if (digits.length === 11) {
    return `${digits.slice(0, 3)}.${digits.slice(3, 6)}.${digits.slice(6, 9)}-${digits.slice(9)}`
  }
  if (digits.length === 14) {
    return `${digits.slice(0, 2)}.${digits.slice(2, 5)}.${digits.slice(5, 8)}/${digits.slice(8, 12)}-${digits.slice(12)}`
  }
  return cnpj || '—'
}

function tipoBadge(documento: DocumentoFiscal): { label: string, color: 'primary' | 'success' | 'warning' | 'neutral' } {
  const map = {
    'nf-e': { label: 'NF-e', color: 'primary' },
    'nfc-e': { label: 'NFC-e', color: 'success' },
    'ct-e': { label: 'CT-e', color: 'warning' },
    'nfs-e': { label: 'NFS-e', color: 'neutral' },
    'desconhecido': { label: 'Desconhecido', color: 'neutral' }
  } as const

  return map[documento.tipo_documento as keyof typeof map] ?? map.desconhecido
}

function manifestacaoBadge(status: string | undefined): { label: string, color: 'success' | 'info' | 'error' | 'warning' | 'neutral' } {
  const map = {
    ciencia: { label: 'Ciência', color: 'info' },
    confirmada: { label: 'Confirmada', color: 'success' },
    desconhecida: { label: 'Desconhecida', color: 'error' },
    nao_realizada: { label: 'Não Realizada', color: 'warning' }
  } as const

  if (!status) return { label: 'Pendente', color: 'neutral' }
  return map[status as keyof typeof map] ?? { label: status, color: 'neutral' }
}

function statusBadge(documento: DocumentoFiscal): { label: string, color: 'success' | 'error' | 'warning' | 'neutral' } {
  const map = {
    autorizada: { label: 'Autorizada', color: 'success' },
    cancelada: { label: 'Cancelada', color: 'error' },
    denegada: { label: 'Denegada', color: 'warning' },
    desconhecido: { label: 'Desconhecido', color: 'neutral' }
  } as const

  return map[documento.status_documento as keyof typeof map] ?? map.desconhecido
}

const columns: TableColumn<DocumentoFiscal>[] = [
  {
    id: 'select',
    enableHiding: false,
    enableSorting: false,
    header: ({ table }) =>
      h(UCheckbox, {
        'modelValue': table.getIsSomePageRowsSelected()
          ? 'indeterminate'
          : table.getIsAllPageRowsSelected(),
        'onUpdate:modelValue': (value: boolean | 'indeterminate') => table.toggleAllPageRowsSelected(!!value),
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
    accessorKey: 'emitente_nome',
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Emitente',
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    },
    cell: ({ row }) => h('div', { class: 'min-w-48' }, [
      h('p', { class: 'font-medium text-highlighted truncate' }, row.original.emitente_nome || '—'),
      h('p', { class: 'text-xs text-muted truncate' }, formatCNPJ(row.original.emitente_cnpj))
    ])
  },
  {
    accessorKey: 'destinatario_nome',
    header: 'Destinatário',
    cell: ({ row }) => h('div', { class: 'min-w-48' }, [
      h('p', { class: 'font-medium text-highlighted truncate' }, row.original.destinatario_nome || '—'),
      h('p', { class: 'text-xs text-muted truncate' }, formatCNPJ(row.original.destinatario_cnpj))
    ])
  },
  {
    accessorKey: 'tipo_documento',
    header: 'Tipo',
    cell: ({ row }) => {
      const tipo = tipoBadge(row.original)
      return h(UBadge, { variant: 'subtle', color: tipo.color }, () => tipo.label)
    }
  },
  {
    accessorKey: 'status_documento',
    header: 'Status',
    cell: ({ row }) => {
      const status = statusBadge(row.original)
      return h(UBadge, { variant: 'subtle', color: status.color }, () => status.label)
    }
  },
  {
    accessorKey: 'chave_acesso',
    header: 'Chave',
    cell: ({ row }) => h('p', { class: 'font-mono text-xs truncate max-w-56' }, row.original.chave_acesso || '—')
  },
  {
    accessorKey: 'numero_documento',
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Número',
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    },
    cell: ({ row }) => row.original.numero_documento || '—'
  },
  {
    accessorKey: 'competencia',
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Competência',
        icon: isSorted
          ? isSorted === 'asc'
            ? 'i-lucide-arrow-up-narrow-wide'
            : 'i-lucide-arrow-down-wide-narrow'
          : 'i-lucide-arrow-up-down',
        class: '-mx-2.5',
        onClick: () => column.toggleSorting(column.getIsSorted() === 'asc')
      })
    },
    cell: ({ row }) => row.original.competencia || '—'
  },
  {
    accessorKey: 'xml_resumo',
    header: 'XML',
    cell: ({ row }) => h(UBadge, {
      variant: 'subtle',
      color: row.original.xml_resumo ? 'warning' : 'success'
    }, () => row.original.xml_resumo ? 'Resumo' : 'Completo')
  },
  {
    accessorKey: 'manifestacao_status',
    header: 'Manifestação',
    cell: ({ row }) => {
      const status = row.original.manifestacao_status
      const badge = manifestacaoBadge(status)
      return h(UBadge, { variant: 'subtle', color: badge.color }, () => badge.label)
    }
  },
  {
    accessorKey: 'valor_total',
    meta: { class: { th: 'text-right', td: 'text-right' } },
    header: ({ column }) => {
      const isSorted = column.getIsSorted()
      return h(UButton, {
        color: 'neutral',
        variant: 'ghost',
        label: 'Valor',
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
      const v = row.original.valor_total
      if (!v || v === 0) return h('span', { class: 'text-muted' }, '—')
      return h('span', { class: 'font-mono text-xs tabular-nums' }, formatBRL(v))
    }
  },
  {
    id: 'actions',
    enableHiding: false,
    meta: { class: { td: 'text-right' } },
    cell: ({ row }) =>
      h(UDropdownMenu, {
        content: { align: 'end' },
        items: rowItems(row)
      }, () => h(UButton, {
        icon: 'i-lucide-ellipsis-vertical',
        color: 'neutral',
        variant: 'ghost',
        class: 'ml-auto'
      }))
  }
]

function getVisibilityItems() {
  if (!table.value?.tableApi) return []

  return table.value.tableApi
    .getAllColumns()
    .filter((column: Column<DocumentoFiscal>) => column.getCanHide())
    .map((column: Column<DocumentoFiscal>) => ({
      label: upperFirst(column.id),
      type: 'checkbox' as const,
      checked: column.getIsVisible(),
      onUpdateChecked(checked: boolean) {
        table.value?.tableApi?.getColumn(column.id)?.toggleVisibility(!!checked)
      },
      onSelect(event?: Event) {
        event?.preventDefault()
      }
    }))
}
</script>

<template>
  <!-- Row 1: search + actions + column visibility -->
  <div class="flex flex-wrap items-center gap-2">
    <UInput
      v-model="search"
      icon="i-lucide-search"
      placeholder="Emitente, destinatário, chave, número..."
      class="flex-1 min-w-52"
    />

    <div class="flex items-center gap-2 ml-auto">
      <slot name="actions" :selected-rows="selectedRows" />

      <UDropdownMenu
        :items="getVisibilityItems()"
        :content="{ align: 'end' }"
      >
        <UButton
          color="neutral"
          variant="outline"
          icon="i-lucide-columns-3"
          label="Colunas"
        />
      </UDropdownMenu>
    </div>
  </div>

  <!-- Row 2: Linear-style filter chips -->
  <DocumentosDataTableFilter
    :columns="filterColumns"
    :filters="filters"
    :actions="filterActions"
  />

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
    sticky
    class="shrink-0"
    :ui="{
      base: 'table-fixed border-separate border-spacing-0',
      thead: '[&>tr]:bg-elevated/50 [&>tr]:after:content-none',
      tbody: '[&>tr]:last:[&>td]:border-b-0',
      th: 'px-3 py-2 text-xs first:rounded-l-lg last:rounded-r-lg border-y border-default first:border-l last:border-r',
      td: 'px-3 py-2 text-sm border-b border-default',
      separator: 'h-0'
    }"
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
      @update:page="(page: number) => table?.tableApi?.setPageIndex(page - 1)"
    />
  </div>
</template>
