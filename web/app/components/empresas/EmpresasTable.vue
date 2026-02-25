<script setup lang="ts">
import type { TableColumn, TableRow } from '@nuxt/ui'
import { getPaginationRowModel } from '@tanstack/table-core'
import type { Row, Table } from '@tanstack/table-core'
import type { Empresa } from '~/types'
import type { ColumnConfig } from '~/composables/useTableFilter'
import { tableUI } from '~/composables/useTableHelpers'

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
const { formatCNPJ } = useDocumentoFormatters()
const { sortableHeader, getVisibilityItems } = useTableHelpers()

const UBadge = resolveComponent('UBadge')
const UButton = resolveComponent('UButton')
const UCheckbox = resolveComponent('UCheckbox')
const UDropdownMenu = resolveComponent('UDropdownMenu')

const table = useTemplateRef<{ tableApi: Table<Empresa> }>('table')

const columnFilters = ref<{ id: string, value: string }[]>([])
const columnVisibility = ref<Record<string, boolean>>({})
const rowSelection = ref<Record<string, boolean>>({})

const pagination = ref({ pageIndex: 0, pageSize: 10 })

const search = defineModel<string>('search', { default: '' })

function dynamicOptions(extractor: (e: Empresa) => string | undefined) {
  const all = [...new Set((props.data ?? []).map(extractor).filter(Boolean))]
  return all.sort().map(v => ({ label: v!, value: v! }))
}

const filterColumns = computed<ColumnConfig<Empresa>[]>(() => [
  {
    id: 'situacao_cadastral',
    accessor: row => row.situacao_cadastral,
    displayName: 'Situação',
    icon: 'i-lucide-circle-check',
    type: 'option',
    options: [
      { label: 'Ativa', value: 'Ativa' },
      { label: 'Baixada', value: 'Baixada' },
      { label: 'Suspensa', value: 'Suspensa' },
      { label: 'Inapta', value: 'Inapta' }
    ]
  },
  {
    id: 'certificado',
    accessor: row => row.certificado_status || 'sem_certificado',
    displayName: 'Certificado',
    icon: 'i-lucide-shield',
    type: 'option',
    options: [
      { label: 'Válido', value: 'valido' },
      { label: 'Vencido', value: 'vencido' },
      { label: 'Prestes a vencer', value: 'prestes_a_vencer' },
      { label: 'Sem certificado', value: 'sem_certificado' }
    ]
  },
  {
    id: 'estado',
    accessor: row => row.estado || '',
    displayName: 'Estado',
    icon: 'i-lucide-map-pin',
    type: 'option',
    options: dynamicOptions(e => e.estado || undefined)
  }
])

const { filters, filteredData, actions: filterActions } = useTableFilter(filterColumns, () => props.data ?? [])

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return filteredData.value

  return filteredData.value.filter(e =>
    e.cnpj.includes(q)
    || e.razao_social.toLowerCase().includes(q)
    || e.nome_fantasia?.toLowerCase().includes(q)
    || e.cidade?.toLowerCase().includes(q)
  )
})

const selectedRows = computed((): Empresa[] => {
  if (!table.value?.tableApi) return []
  return table.value.tableApi.getFilteredSelectedRowModel().rows.map((r: Row<Empresa>) => r.original)
})

const totalFilteredRows = computed(() => table.value?.tableApi?.getFilteredRowModel().rows.length ?? 0)
const totalSelectedRows = computed(() => table.value?.tableApi?.getFilteredSelectedRowModel().rows.length ?? 0)

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

const columns: TableColumn<Empresa>[] = [
  {
    id: 'select',
    enableHiding: false,
    enableSorting: false,
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
    header: sortableHeader('Empresa'),
    cell: ({ row }) => h('div', { class: 'min-w-48 max-w-64' }, [
      h('p', { class: 'font-medium text-highlighted truncate' }, row.original.razao_social),
      h('p', { class: 'text-xs text-muted' }, formatCNPJ(row.original.cnpj))
    ])
  },
  {
    accessorKey: 'situacao_cadastral',
    header: sortableHeader('Situação'),
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
    header: sortableHeader('Cidade/UF'),
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
    header: sortableHeader('Período'),
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
    header: sortableHeader('Certificado'),
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
    enableHiding: false,
    meta: { class: { td: 'text-right' } },
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
</script>

<template>
  <div class="flex items-center gap-2">
    <UInput
      v-model="search"
      icon="i-lucide-search"
      placeholder="Filtrar por CNPJ, razão social, cidade..."
      class="w-64 shrink-0"
    />

    <SharedDataTableFilter
      :columns="filterColumns"
      :filters="filters"
      :actions="filterActions"
      class="min-w-0"
    />

    <div class="flex items-center gap-2 ml-auto shrink-0">
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
    sticky
    class="shrink-0"
    :ui="{
      ...tableUI,
      tr: 'cursor-pointer hover:bg-elevated/40 transition-colors'
    }"
    @select="(_: Event, row: TableRow<Empresa>) => navigateTo(`/empresas/${row.original.id}`)"
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
              { label: '100', value: '100' }
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
