<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { getPaginationRowModel } from '@tanstack/table-core'
import type { Row } from '@tanstack/table-core'
import { upperFirst } from 'scule'
import type { DocumentoFiscal } from '~/types'

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

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const table = useTemplateRef<{ tableApi: any }>('table')

const columnFilters = ref<{ id: string, value: string }[]>([])
const columnVisibility = ref<Record<string, boolean>>({})
const rowSelection = ref<Record<string, boolean>>({})

const pagination = ref({ pageIndex: 0, pageSize: 10 })

const search = defineModel<string>('search', { default: '' })
const tipoFilter = defineModel<string>('tipoFilter', { default: 'all' })
const statusFilter = defineModel<string>('statusFilter', { default: 'all' })
const resumoFilter = defineModel<string>('resumoFilter', { default: 'all' })

const filtered = computed(() => {
  const query = search.value.toLowerCase().trim()
  let result = props.data ?? []

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

  if (tipoFilter.value !== 'all') {
    result = result.filter(documento => documento.tipo_documento === tipoFilter.value)
  }

  if (statusFilter.value !== 'all') {
    result = result.filter(documento => documento.status_documento === statusFilter.value)
  }

  if (resumoFilter.value === 'resumo') {
    result = result.filter(documento => documento.xml_resumo)
  }

  if (resumoFilter.value === 'completo') {
    result = result.filter(documento => !documento.xml_resumo)
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
    header: 'Emitente',
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
    header: 'Número',
    cell: ({ row }) => row.original.numero_documento || '—'
  },
  {
    accessorKey: 'competencia',
    header: 'Competência',
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

function getVisibilityItems() {
  if (!table.value?.tableApi) return []

  return table.value.tableApi
    .getAllColumns()
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    .filter((column: any) => column.getCanHide())
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    .map((column: any) => ({
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
  <div class="flex flex-wrap items-center justify-between gap-1.5">
    <UInput
      v-model="search"
      icon="i-lucide-search"
      placeholder="Filtrar por chave, número, emitente, destinatário..."
      class="max-w-md"
    />

    <div class="flex flex-wrap items-center gap-1.5">
      <slot name="actions" :selected-rows="selectedRows" />

      <USelect
        v-model="tipoFilter"
        :items="[
          { label: 'Todos os tipos', value: 'all' },
          { label: 'NF-e', value: 'nf-e' },
          { label: 'NFC-e', value: 'nfc-e' },
          { label: 'CT-e', value: 'ct-e' },
          { label: 'NFS-e', value: 'nfs-e' },
          { label: 'Desconhecido', value: 'desconhecido' }
        ]"
        :ui="{ trailingIcon: 'group-data-[state=open]:rotate-180 transition-transform duration-200' }"
        placeholder="Tipo"
        class="min-w-36"
      />

      <USelect
        v-model="statusFilter"
        :items="[
          { label: 'Todos os status', value: 'all' },
          { label: 'Autorizada', value: 'autorizada' },
          { label: 'Cancelada', value: 'cancelada' },
          { label: 'Denegada', value: 'denegada' },
          { label: 'Desconhecido', value: 'desconhecido' }
        ]"
        :ui="{ trailingIcon: 'group-data-[state=open]:rotate-180 transition-transform duration-200' }"
        placeholder="Status"
        class="min-w-36"
      />

      <USelect
        v-model="resumoFilter"
        :items="[
          { label: 'XML completo e resumo', value: 'all' },
          { label: 'Apenas XML completo', value: 'completo' },
          { label: 'Apenas XML resumo', value: 'resumo' }
        ]"
        :ui="{ trailingIcon: 'group-data-[state=open]:rotate-180 transition-transform duration-200' }"
        placeholder="XML"
        class="min-w-48"
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
