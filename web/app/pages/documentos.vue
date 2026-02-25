<script setup lang="ts">
import type { DocumentoFiscal, DocumentoListResponse } from '~/types'

const tableRef = useTemplateRef<{ selectedRows: DocumentoFiscal[] }>('table')
const importModalRef = useTemplateRef<{ show: () => void }>('importModal')
const exportModalRef = useTemplateRef<{ show: () => void }>('exportModal')
const manifestacaoModalRef = useTemplateRef<{ show: () => void }>('manifestacaoModal')
const xmlModalRef = useTemplateRef<{ show: (doc: DocumentoFiscal) => void }>('xmlModal')

const { data, status, refresh } = await useFetch<DocumentoListResponse>('/api/documentos', {
  lazy: true,
  query: { page: 1, page_size: 500 }
})

const documentos = computed(() => (data.value?.items ?? []).filter(d => !d.xml_resumo))
const selectedRows = computed(() => tableRef.value?.selectedRows ?? [])
const selectedCount = computed(() => selectedRows.value.length)
</script>

<template>
  <UDashboardPanel id="documentos">
    <template #header>
      <UDashboardNavbar title="Documentos">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <UButton
            label="Importar"
            icon="i-lucide-upload"
            color="neutral"
            variant="outline"
            size="sm"
            @click="importModalRef?.show()"
          />
          <UButton
            label="Recarregar"
            color="neutral"
            variant="outline"
            icon="i-lucide-refresh-cw"
            @click="refresh()"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
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
    </template>
  </UDashboardPanel>

  <DocumentosImportModal ref="importModal" @imported="refresh()" />
  <DocumentosExportModal ref="exportModal" :selected-rows="selectedRows" />
  <DocumentosManifestacaoModal ref="manifestacaoModal" :selected-rows="selectedRows" @done="refresh()" />
  <DocumentosXmlModal ref="xmlModal" />
</template>
