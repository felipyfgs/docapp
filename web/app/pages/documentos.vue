<script setup lang="ts">
import type { DocumentoExportResponse, DocumentoFiscal, DocumentoListResponse, DocumentoXMLResponse } from '~/types'

type DeliveryMode = 'proxy' | 'presigned'
type ExportFormat = 'xml' | 'danfe' | 'ambos'
type OrganizationMode = 'tipo/competencia/cnpj' | 'cnpj/competencia/tipo' | 'competencia/cnpj/tipo'

type BackfillResponse = {
  processed: number
  uploaded: number
  skipped: number
}

const toast = useToast()
const tableRef = useTemplateRef<{ selectedRows: DocumentoFiscal[] }>('table')

const { data, status, refresh } = await useFetch<DocumentoListResponse>('/api/documentos', {
  lazy: true,
  query: {
    page: 1,
    page_size: 500
  }
})

const documentos = computed(() => data.value?.items ?? [])
const selectedRows = computed(() => tableRef.value?.selectedRows ?? [])
const selectedCount = computed(() => selectedRows.value.length)

const exportOpen = ref(false)
const exportFormat = ref<ExportFormat>('xml')
const exportOrganization = ref<OrganizationMode>('tipo/competencia/cnpj')
const exportDeliveryMode = ref<DeliveryMode>('proxy')
const exporting = ref(false)

const backfilling = ref(false)

const xmlOpen = ref(false)
const xmlLoading = ref(false)
const xmlContent = ref('')
const xmlDocumento = ref<DocumentoFiscal | null>(null)

const xmlTitle = computed(() => {
  if (!xmlDocumento.value) {
    return 'XML do documento'
  }

  return xmlDocumento.value.chave_acesso
    ? `XML - ${xmlDocumento.value.chave_acesso}`
    : `XML - Documento #${xmlDocumento.value.id}`
})

function openExportModal() {
  if (selectedCount.value === 0) {
    return
  }

  exportOpen.value = true
}

async function handleViewXML(documento: DocumentoFiscal) {
  xmlDocumento.value = documento
  xmlOpen.value = true
  xmlLoading.value = true
  xmlContent.value = ''

  try {
    const response = await $fetch<DocumentoXMLResponse>(`/api/documentos/${documento.id}/xml`)
    xmlContent.value = response.xml
  } catch (error: unknown) {
    toast.add({
      title: getErrorMessage(error, 'Erro ao carregar XML'),
      color: 'error'
    })
  } finally {
    xmlLoading.value = false
  }
}

function copyXML() {
  if (!xmlContent.value) {
    return
  }

  navigator.clipboard.writeText(xmlContent.value)
  toast.add({ title: 'XML copiado', color: 'success' })
}

async function handleExport() {
  const ids = selectedRows.value.map(documento => documento.id)
  if (ids.length === 0) {
    return
  }

  exporting.value = true

  try {
    if (exportDeliveryMode.value === 'presigned') {
      const response = await $fetch<DocumentoExportResponse>('/api/documentos/export', {
        method: 'POST',
        body: {
          ids,
          format: exportFormat.value,
          organization: exportOrganization.value,
          delivery_mode: exportDeliveryMode.value
        }
      })

      if (response.presigned_url) {
        window.open(response.presigned_url, '_blank', 'noopener')
      }

      toast.add({ title: 'Exportação gerada com sucesso', color: 'success' })
      exportOpen.value = false
      return
    }

    const response = await fetch('/api/documentos/export', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        ids,
        format: exportFormat.value,
        organization: exportOrganization.value,
        delivery_mode: exportDeliveryMode.value
      })
    })

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || ''
      if (contentType.includes('application/json')) {
        const body = await response.json() as { message?: string }
        throw new Error(body.message || 'Erro ao exportar documentos')
      }

      throw new Error('Erro ao exportar documentos')
    }

    const fileName = extractFileName(response.headers.get('content-disposition'))
    const blob = await response.blob()
    downloadBlob(blob, fileName)

    toast.add({ title: 'Arquivo de exportação baixado', color: 'success' })
    exportOpen.value = false
  } catch (error: unknown) {
    toast.add({
      title: getErrorMessage(error, 'Erro ao exportar documentos'),
      color: 'error'
    })
  } finally {
    exporting.value = false
  }
}

async function runBackfill() {
  backfilling.value = true

  try {
    const response = await $fetch<BackfillResponse>('/api/documentos/backfill', {
      method: 'POST',
      body: { limit: 500 }
    })

    toast.add({
      title: 'Backfill concluído',
      description: `Processados: ${response.processed} • Enviados: ${response.uploaded} • Ignorados: ${response.skipped}`,
      color: 'success'
    })

    refresh()
  } catch (error: unknown) {
    toast.add({
      title: getErrorMessage(error, 'Erro ao executar backfill'),
      color: 'error'
    })
  } finally {
    backfilling.value = false
  }
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message) {
    return error.message
  }

  if (typeof error === 'object' && error !== null) {
    const maybeData = error as { data?: { message?: string }, statusMessage?: string, message?: string }
    return maybeData.data?.message || maybeData.statusMessage || maybeData.message || fallback
  }

  return fallback
}

function extractFileName(contentDisposition: string | null): string {
  if (!contentDisposition) {
    return `documentos_${Date.now()}.zip`
  }

  const match = contentDisposition.match(/filename="?([^";]+)"?/i)
  return match?.[1] || `documentos_${Date.now()}.zip`
}

function downloadBlob(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
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
        @view-xml="handleViewXML"
      >
        <template #actions>
          <UButton
            :loading="backfilling"
            label="Backfill XML"
            color="neutral"
            variant="outline"
            icon="i-lucide-database-backup"
            @click="runBackfill"
          />

          <UButton
            v-if="selectedCount > 0"
            label="Exportar"
            color="primary"
            icon="i-lucide-download"
            @click="openExportModal"
          >
            <template #trailing>
              <UKbd>{{ selectedCount }}</UKbd>
            </template>
          </UButton>
        </template>
      </DocumentosTable>
    </template>
  </UDashboardPanel>

  <UModal
    v-model:open="exportOpen"
    title="Exportar documentos"
    description="Selecione formato, organização das pastas e modo de entrega."
  >
    <template #body>
      <div class="space-y-4">
        <UFormField label="Formato">
          <USelect
            v-model="exportFormat"
            :items="[
              { label: 'XML', value: 'xml' },
              { label: 'DANFE', value: 'danfe' },
              { label: 'XML + DANFE', value: 'ambos' }
            ]"
            class="w-full"
          />
        </UFormField>

        <UFormField label="Organização de pastas">
          <USelect
            v-model="exportOrganization"
            :items="[
              { label: 'tipo/competencia/cnpj', value: 'tipo/competencia/cnpj' },
              { label: 'cnpj/competencia/tipo', value: 'cnpj/competencia/tipo' },
              { label: 'competencia/cnpj/tipo', value: 'competencia/cnpj/tipo' }
            ]"
            class="w-full"
          />
        </UFormField>

        <UFormField label="Entrega">
          <USelect
            v-model="exportDeliveryMode"
            :items="[
              { label: 'Download via proxy', value: 'proxy' },
              { label: 'URL pré-assinada', value: 'presigned' }
            ]"
            class="w-full"
          />
        </UFormField>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton
          label="Cancelar"
          color="neutral"
          variant="subtle"
          :disabled="exporting"
          @click="exportOpen = false"
        />
        <UButton
          label="Exportar"
          color="primary"
          :loading="exporting"
          icon="i-lucide-download"
          @click="handleExport"
        />
      </div>
    </template>
  </UModal>

  <UModal
    v-model:open="xmlOpen"
    :title="xmlTitle"
    :description="xmlDocumento?.xml_resumo ? 'Documento em modo resumo' : 'Documento completo'"
    :ui="{ content: 'sm:max-w-5xl', body: 'max-h-[72vh] overflow-auto' }"
  >
    <template #body>
      <div
        v-if="xmlLoading"
        class="flex items-center justify-center gap-2 py-8 text-muted"
      >
        <UIcon name="i-lucide-loader-circle" class="animate-spin size-4" />
        Carregando XML...
      </div>

      <pre
        v-else
        class="text-xs leading-relaxed whitespace-pre-wrap break-all rounded-md border border-default bg-elevated/50 p-3"
      >{{ xmlContent }}</pre>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton
          label="Fechar"
          color="neutral"
          variant="subtle"
          @click="xmlOpen = false"
        />
        <UButton
          label="Copiar XML"
          color="primary"
          icon="i-lucide-copy"
          :disabled="!xmlContent"
          @click="copyXML"
        />
      </div>
    </template>
  </UModal>
</template>
