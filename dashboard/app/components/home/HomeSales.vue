<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import type { DocumentoFiscal } from '~/types'
import { tableUI } from '~/composables/useTableHelpers'
import { UBadge } from '#components'

defineProps<{
  data: DocumentoFiscal[] | null
}>()

const { formatBRL, formatCNPJ, tipoBadge, statusBadge } = useDocumentoFormatters()

const columns: TableColumn<DocumentoFiscal>[] = [
  {
    accessorKey: 'tipo_documento',
    header: 'Tipo',
    cell: ({ row }) => {
      const tipo = tipoBadge(row.original)
      return h(UBadge, { variant: 'subtle', color: tipo.color }, () => tipo.label)
    }
  },
  {
    accessorKey: 'numero_documento',
    header: 'Número',
    cell: ({ row }) => {
      const num = row.original.numero_documento
      return num ? `#${num}` : '—'
    }
  },
  {
    accessorKey: 'emitente_nome',
    header: 'Emitente',
    cell: ({ row }) => h('div', { class: 'min-w-40' }, [
      h('p', { class: 'font-medium text-highlighted truncate' }, row.original.emitente_nome || '—'),
      h('p', { class: 'text-xs text-muted truncate' }, formatCNPJ(row.original.emitente_cnpj))
    ])
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
    accessorKey: 'valor_total',
    header: () => h('div', { class: 'text-right' }, 'Valor'),
    cell: ({ row }) => {
      const v = row.original.valor_total
      if (!v || v === 0) return h('span', { class: 'text-muted' }, '—')
      return h('div', { class: 'text-right font-mono text-xs tabular-nums' }, formatBRL(v))
    }
  },
  {
    accessorKey: 'data_emissao',
    header: 'Emissão',
    cell: ({ row }) => {
      const d = row.original.data_emissao
      if (!d) return h('span', { class: 'text-muted' }, '—')
      return h('span', { class: 'whitespace-nowrap' }, new Date(d).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' }))
    }
  }
]
</script>

<template>
  <div>
    <h3 class="text-sm font-medium text-muted uppercase mb-3">
      Documentos Recentes
    </h3>
    <UTable
      :data="data ?? []"
      :columns="columns"
      class="shrink-0"
      :ui="tableUI"
    />
  </div>
</template>
