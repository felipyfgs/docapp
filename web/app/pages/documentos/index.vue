<script setup lang="ts">
import type { DocumentoFiscal } from '~/types'

const documentos = inject<ComputedRef<DocumentoFiscal[]>>('documentos', computed(() => []))
const status = inject<Ref<string>>('documentosStatus', ref('idle'))
const refresh = inject<() => Promise<void>>('documentosRefresh', async () => {})

const tableRef = useTemplateRef<{ selectedRows: DocumentoFiscal[] }>('table')
const exportModalRef = useTemplateRef<{ show: () => void }>('exportModal')
const manifestacaoModalRef = useTemplateRef<{ show: () => void }>('manifestacaoModal')
const xmlModalRef = useTemplateRef<{ show: (doc: DocumentoFiscal) => void }>('xmlModal')

const selectedRows = computed(() => tableRef.value?.selectedRows ?? [])
const selectedCount = computed(() => selectedRows.value.length)
</script>

<template>
  <DocumentosTable
    ref="table"
    :data="documentos"
    :status="status"
    @view-xml="(doc: DocumentoFiscal) => xmlModalRef?.show(doc)"
  >
    <template #actions>
      <UButton
        v-if="selectedCount > 0"
        label="Manifestar"
        color="neutral"
        variant="outline"
        icon="i-lucide-stamp"
        @click="manifestacaoModalRef?.show()"
      >
        <template #trailing>
          <UKbd>{{ selectedCount }}</UKbd>
        </template>
      </UButton>
      <UButton
        v-if="selectedCount > 0"
        label="Exportar"
        color="primary"
        icon="i-lucide-download"
        @click="exportModalRef?.show()"
      >
        <template #trailing>
          <UKbd>{{ selectedCount }}</UKbd>
        </template>
      </UButton>
    </template>
  </DocumentosTable>

  <DocumentosExportModal ref="exportModal" :selected-rows="selectedRows" />
  <DocumentosManifestacaoModal ref="manifestacaoModal" :selected-rows="selectedRows" @done="refresh()" />
  <DocumentosXmlModal ref="xmlModal" />
</template>
