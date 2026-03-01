<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import type { DocumentoFiscal } from '~/types'
import { UBadge } from '#components'

defineProps<{
  documentos: DocumentoFiscal[]
}>()

const { formatBRL, tipoBadge, statusBadge } = useDocumentoFormatters()

const tableColumns: TableColumn<DocumentoFiscal>[] = [
  {
    accessorKey: 'emitente_nome',
    header: 'Emitente',
    cell: ({ row }) => {
      return h('div', { class: 'max-w-48 truncate' }, [
        h('p', { class: 'truncate font-medium text-highlighted' }, row.original.emitente_nome || '—'),
        h('p', { class: 'text-xs text-muted' }, row.original.emitente_cnpj || '')
      ])
    }
  },
  {
    accessorKey: 'numero_documento',
    header: 'Número',
    cell: ({ row }) => h('span', { class: 'text-muted' }, row.original.numero_documento || '—')
  },
  {
    accessorKey: 'tipo_documento',
    header: 'Tipo',
    cell: ({ row }) => {
      const tipo = tipoBadge(row.original)
      return h(UBadge, { color: tipo.color, variant: 'subtle', size: 'xs', class: 'uppercase' }, () => tipo.label)
    }
  },
  {
    accessorKey: 'status_documento',
    header: 'Status',
    cell: ({ row }) => {
      const status = statusBadge(row.original)
      return h(UBadge, { color: status.color, variant: 'subtle', size: 'xs', class: 'capitalize' }, () => status.label)
    }
  },
  {
    accessorKey: 'competencia',
    header: 'Competência',
    cell: ({ row }) => h('span', { class: 'text-muted' }, row.original.competencia || '—')
  },
  {
    id: 'xml_resumo',
    header: 'XML',
    cell: ({ row }) => h(UBadge, { color: row.original.xml_resumo ? 'warning' : 'success', variant: 'subtle', size: 'xs' }, () => row.original.xml_resumo ? 'Resumo' : 'Completo')
  },
  {
    accessorKey: 'valor_total',
    header: 'Valor',
    cell: ({ row }) => {
      const v = row.original.valor_total
      if (!v || v === 0) return h('span', { class: 'text-muted' }, '—')
      return h('span', { class: 'font-mono text-xs' }, formatBRL(v))
    }
  }
]
</script>

<template>
  <UCard :ui="{ body: 'p-0 sm:p-0', header: 'px-4 py-3 sm:px-6' }">
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-files" class="size-5 text-primary" />
          <h3 class="font-semibold text-highlighted">
            Documentos Recentes
          </h3>
        </div>
        <UButton
          to="/documentos"
          color="neutral"
          variant="ghost"
          size="xs"
          label="Ver todos"
          icon="i-lucide-arrow-right"
          trailing
        />
      </div>
    </template>

    <UTable
      v-if="documentos.length > 0"
      :data="documentos"
      :columns="tableColumns"
      class="w-full"
    />
    <div v-else class="text-sm text-muted py-8 text-center">
      Nenhum documento encontrado.
    </div>
  </UCard>
</template>
