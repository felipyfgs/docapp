<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { getPaginationRowModel } from '@tanstack/table-core'
import type { Row } from '@tanstack/table-core'
import type { Empresa } from '~/types'

const props = defineProps<{
  data: Empresa[] | null
  status: string
}>()

const emit = defineEmits<{
  delete: [empresa: Empresa]
}>()

const UBadge = resolveComponent('UBadge')
const UButton = resolveComponent('UButton')
const UDropdownMenu = resolveComponent('UDropdownMenu')

const table = useTemplateRef('table')

const search = ref('')

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  if (!q || !props.data) return props.data ?? []
  return props.data.filter(e =>
    e.cnpj.includes(q)
    || e.razao_social.toLowerCase().includes(q)
    || e.nome_fantasia?.toLowerCase().includes(q)
    || e.cidade?.toLowerCase().includes(q)
  )
})

const pagination = ref({ pageIndex: 0, pageSize: 10 })

function rowItems(row: Row<Empresa>) {
  return [[{
    label: 'Deletar',
    icon: 'i-lucide-trash',
    color: 'error' as const,
    onSelect() {
      emit('delete', row.original)
    }
  }]]
}

function formatCNPJ(cnpj: string): string {
  const d = cnpj.replace(/\D/g, '')
  if (d.length !== 14) return cnpj
  return `${d.slice(0, 2)}.${d.slice(2, 5)}.${d.slice(5, 8)}/${d.slice(8, 12)}-${d.slice(12)}`
}

const columns: TableColumn<Empresa>[] = [
  {
    accessorKey: 'cnpj',
    header: 'CNPJ',
    cell: ({ row }) => formatCNPJ(row.original.cnpj)
  },
  {
    accessorKey: 'razao_social',
    header: 'Razão Social'
  },
  {
    accessorKey: 'nome_fantasia',
    header: 'Nome Fantasia',
    cell: ({ row }) => row.original.nome_fantasia || '—'
  },
  {
    accessorKey: 'situacao_cadastral',
    header: 'Situação',
    cell: ({ row }) => {
      const s = row.original.situacao_cadastral
      const color = s === 'Ativa' ? 'success' as const : 'warning' as const
      return h(UBadge, { variant: 'subtle', color, class: 'capitalize' }, () => s || '—')
    }
  },
  {
    id: 'localidade',
    header: 'Cidade/UF',
    cell: ({ row }) => {
      const { cidade, estado } = row.original
      if (!cidade && !estado) return '—'
      return `${cidade || ''}${cidade && estado ? '/' : ''}${estado || ''}`
    }
  },
  {
    accessorKey: 'lookback_days',
    header: 'Lookback',
    cell: ({ row }) => `${row.original.lookback_days}d`
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
</script>

<template>
  <div class="flex flex-col gap-4">
    <UInput
      v-model="search"
      icon="i-lucide-search"
      placeholder="Filtrar por CNPJ, razão social, cidade..."
      class="max-w-sm"
    />

    <UTable
      ref="table"
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
        {{ filtered.length }} empresa{{ filtered.length !== 1 ? 's' : '' }}
      </div>
      <UPagination
        :default-page="(table?.tableApi?.getState().pagination.pageIndex || 0) + 1"
        :items-per-page="table?.tableApi?.getState().pagination.pageSize"
        :total="filtered.length"
        @update:page="(p: number) => table?.tableApi?.setPageIndex(p - 1)"
      />
    </div>
  </div>
</template>
