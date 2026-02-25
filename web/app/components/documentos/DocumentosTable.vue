<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { getPaginationRowModel } from '@tanstack/table-core'
import type { Row, Table } from '@tanstack/table-core'
import type { DocumentoFiscal } from '~/types'
import type { ColumnConfig } from '~/composables/useTableFilter'
import { tableUI } from '~/composables/useTableHelpers'
import { UBadge, UTooltip, UButton, UCheckbox, UDropdownMenu } from '#components'

const props = defineProps<{
  data: DocumentoFiscal[] | null | undefined
  status: string
}>()

const emit = defineEmits<{
  viewXml: [documento: DocumentoFiscal]
}>()

const toast = useToast()
const { formatBRL, formatCNPJ, tipoBadge, statusBadge, manifestacaoBadge } = useDocumentoFormatters()
const { sortableHeader, getVisibilityItems } = useTableHelpers()



const table = useTemplateRef<{ tableApi: Table<DocumentoFiscal> }>('table')

const columnFilters = ref<{ id: string, value: string }[]>([])
const columnVisibility = ref<Record<string, boolean>>({
  manifestacao_status: false,
  created_at: false
})
const rowSelection = ref<Record<string, boolean>>({})
const pagination = ref({ pageIndex: 0, pageSize: 15 })

function dynamicOptions(extractor: (d: DocumentoFiscal) => string | undefined) {
  const all = [...new Set((props.data ?? []).map(extractor).filter(Boolean))]
  return all.sort().map(v => ({ label: v!, value: v! }))
}

const filterColumns = computed<ColumnConfig<DocumentoFiscal>[]>(() => [
  {
    id: 'tipo_documento',
    accessor: row => row.tipo_documento,
    displayName: 'Tipo',
    icon: 'i-lucide-file-text',
    type: 'option',
    defaultOpen: true,
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
    defaultOpen: true,
    options: [
      { label: 'Autorizada', value: 'autorizada' },
      { label: 'Cancelada', value: 'cancelada' },
      { label: 'Denegada', value: 'denegada' },
      { label: 'Desconhecido', value: 'desconhecido' }
    ]
  },
  {
    id: 'data_emissao',
    accessor: row => row.data_emissao ?? null,
    displayName: 'Emissão',
    icon: 'i-lucide-calendar',
    type: 'timerange',
    defaultOpen: true,
    commandDisabled: true
  },
  {
    id: 'emitente',
    accessor: row => row.emitente_nome || '',
    displayName: 'Emitente',
    icon: 'i-lucide-building',
    type: 'option',
    options: dynamicOptions(d => d.emitente_nome || undefined)
  },
  {
    id: 'destinatario',
    accessor: row => row.destinatario_nome || '',
    displayName: 'Destinatário',
    icon: 'i-lucide-user',
    type: 'option',
    options: dynamicOptions(d => d.destinatario_nome || undefined)
  },
  {
    id: 'valor_total',
    accessor: row => row.valor_total ?? 0,
    displayName: 'Valor',
    icon: 'i-lucide-dollar-sign',
    type: 'slider',
    min: 0,
    max: 1000000,
    commandDisabled: true
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
])

const { filters, filteredData, hasFilters, activeFilterCount, facetedCounts, facetedMinMax, actions: filterActions } = useTableFilter(filterColumns, () => props.data ?? [])

const filtered = computed(() => filteredData.value)

const selectedRows = computed((): DocumentoFiscal[] => {
  if (!table.value?.tableApi) return []
  return table.value.tableApi.getFilteredSelectedRowModel().rows.map((row: Row<DocumentoFiscal>) => row.original)
})

const totalRows = computed(() => (props.data ?? []).length)
const totalFilteredRows = computed(() => table.value?.tableApi?.getFilteredRowModel().rows.length ?? 0)
const totalSelectedRows = computed(() => table.value?.tableApi?.getFilteredSelectedRowModel().rows.length ?? 0)

defineExpose({
  selectedRows,
  tableApi: computed(() => table.value?.tableApi),
  filterColumns,
  filters,
  filterActions,
  hasFilters,
  activeFilterCount,
  facetedCounts,
  facetedMinMax,
  totalRows,
  totalFilteredRows
})

function rowItems(row: Row<DocumentoFiscal>) {
  return [
    [{ type: 'label' as const, label: 'Ações' }],
    [
      {
        label: 'Ver XML',
        icon: 'i-lucide-file-code-2',
        onSelect() { emit('viewXml', row.original) }
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

const columns: TableColumn<DocumentoFiscal>[] = [
  {
    id: 'select',
    enableHiding: false,
    enableSorting: false,
    header: ({ table }) =>
      h(UCheckbox, {
        'modelValue': table.getIsSomePageRowsSelected() ? 'indeterminate' : table.getIsAllPageRowsSelected(),
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
    accessorKey: 'tipo_documento',
    header: sortableHeader('Tipo'),
    cell: ({ row }) => {
      const tipo = tipoBadge(row.original)
      return h(UBadge, { variant: 'subtle', color: tipo.color }, () => tipo.label)
    }
  },
  {
    accessorKey: 'numero_documento',
    header: sortableHeader('Número'),
    cell: ({ row }) => {
      const num = row.original.numero_documento
      const chave = row.original.chave_acesso
      return h('div', { class: 'flex items-center gap-1.5 whitespace-nowrap' }, [
        num ? h('span', { class: 'font-medium' }, `#${num}`) : h('span', { class: 'text-muted' }, '—'),
        chave
          ? h(UTooltip, { text: chave, content: { side: 'right' } }, () =>
              h(UButton, {
                icon: 'i-lucide-key-round',
                color: 'neutral',
                variant: 'ghost',
                size: 'xs',
                class: 'size-5',
                onClick: () => {
                  navigator.clipboard.writeText(chave)
                  toast.add({ title: 'Chave copiada', color: 'success' })
                }
              })
            )
          : null
      ])
    }
  },
  {
    accessorKey: 'emitente_nome',
    header: sortableHeader('Emitente'),
    cell: ({ row }) => h('div', { class: 'min-w-48' }, [
      h('p', { class: 'font-medium text-highlighted truncate' }, row.original.emitente_nome || '—'),
      h('p', { class: 'text-xs text-muted truncate' }, formatCNPJ(row.original.emitente_cnpj))
    ])
  },
  {
    accessorKey: 'destinatario_cnpj',
    header: sortableHeader('Destinatário'),
    cell: ({ row }) => h('span', { class: 'text-sm font-mono whitespace-nowrap' }, formatCNPJ(row.original.destinatario_cnpj) || '—')
  },
  {
    accessorKey: 'status_documento',
    header: sortableHeader('Status'),
    cell: ({ row }) => {
      const status = statusBadge(row.original)
      return h(UBadge, { variant: 'subtle', color: status.color }, () => status.label)
    }
  },
  {
    accessorKey: 'manifestacao_status',
    header: sortableHeader('Manifestação'),
    cell: ({ row }) => {
      const status = row.original.manifestacao_status
      const badge = manifestacaoBadge(status)
      return h(UBadge, { variant: 'subtle', color: badge.color }, () => badge.label)
    }
  },
  {
    accessorKey: 'valor_total',
    meta: { class: { th: 'text-right', td: 'text-right' } },
    header: sortableHeader('Valor'),
    cell: ({ row }) => {
      const v = row.original.valor_total
      if (!v || v === 0) return h('span', { class: 'text-muted' }, '—')
      return h('span', { class: 'font-mono text-xs tabular-nums' }, formatBRL(v))
    }
  },
  {
    accessorKey: 'data_emissao',
    header: sortableHeader('Emissão'),
    cell: ({ row }) => {
      const d = row.original.data_emissao
      if (!d) return h('span', { class: 'text-muted' }, '—')
      return h('span', { class: 'whitespace-nowrap' }, new Date(d).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' }))
    }
  },
  {
    accessorKey: 'created_at',
    header: sortableHeader('Recebido'),
    cell: ({ row }) => {
      const d = row.original.created_at
      if (!d) return h('span', { class: 'text-muted' }, '—')
      return h('span', { class: 'whitespace-nowrap' }, new Date(d).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' }))
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

const columnLabels: Record<string, string> = {
  tipo_documento: 'Tipo',
  numero_documento: 'Número',
  emitente_nome: 'Emitente',
  destinatario_cnpj: 'Destinatário',
  status_documento: 'Status',
  manifestacao_status: 'Manifestação',
  valor_total: 'Valor',
  data_emissao: 'Emissão',
  created_at: 'Recebido'
}
</script>

<template>
  <SharedDataTableToolbar
    :columns="filterColumns"
    :actions="filterActions"
    :total-rows="totalRows"
    :filtered-rows="filtered.length"
    :has-filters="hasFilters"
    :active-filter-count="activeFilterCount"
  >
    <template #actions>
      <slot name="actions" :selected-rows="selectedRows" />

      <UDropdownMenu
        :items="getVisibilityItems(table?.tableApi, columnLabels)"
        :content="{ align: 'end' }"
      >
        <UButton
          color="neutral"
          variant="outline"
          icon="i-lucide-columns-3"
          label="Colunas"
        />
      </UDropdownMenu>
    </template>
  </SharedDataTableToolbar>

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
    :ui="tableUI"
  />

  <ClientOnly>
    <div class="flex items-center justify-between gap-3 border-t border-default pt-4 mt-auto">
      <div class="flex items-center gap-3 text-sm text-muted">
        <span>
          {{ totalSelectedRows }} de
          {{ totalFilteredRows }} linha(s) selecionada(s).
        </span>

        <div class="flex items-center gap-1.5">
          <span>Linhas</span>
          <USelect
            :model-value="String(pagination.pageSize)"
            :items="[
              { label: '10', value: '10' },
              { label: '15', value: '15' },
              { label: '25', value: '25' },
              { label: '50', value: '50' },
              { label: '100', value: '100' },
              { label: '200', value: '200' },
              { label: 'Todos', value: String(totalFilteredRows || 9999) }
            ]"
            class="w-20"
            @update:model-value="(val: string) => { pagination = { pageIndex: 0, pageSize: Number(val) } }"
          />
        </div>
      </div>

      <UPagination
        :page="pagination.pageIndex + 1"
        :items-per-page="pagination.pageSize"
        :total="totalFilteredRows"
        @update:page="(page: number) => { pagination = { ...pagination, pageIndex: page - 1 } }"
      />
    </div>
  </ClientOnly>
</template>
