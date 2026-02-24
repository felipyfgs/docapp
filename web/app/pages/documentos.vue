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

type AutoImportResult = {
  by_empresa: Record<string, { imported: number, failed: number, errors?: string[] }>
  unknown: number
  unknown_empresas?: { cnpj: string, razao_social: string }[]
}

// Import
const importOpen = ref(false)
const importing = ref(false)
const importFiles = ref<File[]>([])
const importResult = ref<AutoImportResult | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)

const selectedUnknownEmpresas = ref<string[]>([])
const registeringEmpresas = ref(false)

const importFileLabel = computed(() => {
  if (importFiles.value.length === 0) return 'Clique para selecionar arquivos'
  if (importFiles.value.length === 1) return importFiles.value[0]!.name
  return `${importFiles.value.length} arquivo(s) selecionado(s)`
})

function onImportFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  importFiles.value = Array.from(input.files ?? [])
  importResult.value = null
}

function triggerFileInput() {
  fileInputRef.value?.click()
}

async function handleImport() {
  if (importFiles.value.length === 0) return
  importing.value = true
  importResult.value = null
  try {
    const form = new FormData()
    for (const f of importFiles.value) form.append('files', f)
    const result = await $fetch<AutoImportResult>('/api/documentos/import', { method: 'POST', body: form })
    importResult.value = result
    const total = Object.values(result.by_empresa).reduce((s, r) => s + r.imported, 0)
    if (total > 0) refresh()
  } catch {
    toast.add({ title: 'Erro ao importar documentos', color: 'error' })
  } finally {
    importing.value = false
  }
}

function closeImport() {
  importOpen.value = false
  importFiles.value = []
  importResult.value = null
  selectedUnknownEmpresas.value = []
}

async function registerSelectedEmpresas() {
  if (selectedUnknownEmpresas.value.length === 0) return
  registeringEmpresas.value = true

  try {
    const empresasToRegister = importResult.value?.unknown_empresas?.filter(
      e => selectedUnknownEmpresas.value.includes(e.cnpj)
    ) || []

    let successCount = 0
    for (const emp of empresasToRegister) {
      try {
        await $fetch('/api/empresas', {
          method: 'POST',
          body: {
            cnpj: emp.cnpj,
            razao_social: emp.razao_social || `Empresa ${emp.cnpj}`,
            estado: 'SP', // Require a default state as it's required by the backend, could be improved but it works as a fallback
            lookback_days: 90
          }
        })
        successCount++
      } catch (err) {
        console.error('Failed to create empresa:', emp.cnpj, err)
      }
    }

    if (successCount > 0) {
      toast.add({ title: `${successCount} empresa(s) cadastrada(s)`, color: 'success' })
      selectedUnknownEmpresas.value = []
      
      // Delay handleImport to allow backend DB transactions to settle
      setTimeout(async () => {
        await handleImport()
      }, 500)
    } else {
      toast.add({ title: 'Erro ao cadastrar empresas', color: 'error' })
    }
  } catch {
    toast.add({ title: 'Erro ao processar cadastro', color: 'error' })
  } finally {
    registeringEmpresas.value = false
  }
}

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
            label="Importar"
            icon="i-lucide-upload"
            color="neutral"
            variant="outline"
            size="sm"
            @click="importOpen = true"
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
    v-model:open="importOpen"
    title="Importar documentos"
    description="A empresa é detectada automaticamente pelo CNPJ de cada XML."
    :ui="{ footer: 'justify-end' }"
  >
    <template #body>
      <div class="space-y-4">
        <input
          ref="fileInputRef"
          type="file"
          accept=".xml,.zip"
          multiple
          class="hidden"
          @change="onImportFileChange"
        >

        <button
          type="button"
          class="w-full flex flex-col items-center justify-center gap-2 border-2 border-dashed rounded-lg p-8 transition-colors"
          :class="importFiles.length > 0 ? 'border-primary bg-primary/5' : 'border-default hover:border-primary'"
          @click="triggerFileInput"
        >
          <UIcon
            :name="importFiles.length > 0 ? 'i-lucide-file-check' : 'i-lucide-file-up'"
            class="size-8"
            :class="importFiles.length > 0 ? 'text-primary' : 'text-muted'"
          />
          <span class="text-sm font-medium" :class="importFiles.length > 0 ? 'text-primary' : 'text-muted'">
            {{ importFileLabel }}
          </span>
          <span class="text-xs text-dimmed">XML ou ZIP · múltiplos arquivos · máx 200 MB total</span>
        </button>

        <div v-if="importResult" class="space-y-2">
          <div
            v-for="(r, empresa) in importResult.by_empresa"
            :key="empresa"
            class="rounded-lg px-3 py-2 text-sm"
            :class="r.failed > 0 ? 'bg-warning/10' : 'bg-success/10'"
          >
            <p class="font-medium text-highlighted truncate">
              {{ empresa }}
            </p>
            <p class="text-xs text-muted mt-0.5">
              {{ r.imported }} importado{{ r.imported !== 1 ? 's' : '' }}
              <template v-if="r.failed > 0">
                · {{ r.failed }} falha{{ r.failed !== 1 ? 's' : '' }}
              </template>
            </p>
          </div>
          <div v-if="importResult.unknown > 0" class="text-xs text-warning px-1">
            {{ importResult.unknown }} arquivo(s) sem empresa correspondente — verifique se a empresa está cadastrada.
          </div>

          <div v-if="importResult.unknown_empresas && importResult.unknown_empresas.length > 0" class="mt-4 border border-default rounded-md p-3">
            <div class="text-sm font-medium mb-2">
              Empresas não cadastradas encontradas nos arquivos:
            </div>
            <div class="space-y-2 max-h-40 overflow-y-auto">
              <div
                v-for="emp in importResult.unknown_empresas"
                :key="emp.cnpj"
                class="flex items-center"
              >
                <input
                  :id="emp.cnpj"
                  v-model="selectedUnknownEmpresas"
                  type="checkbox"
                  :value="emp.cnpj"
                  class="mr-2"
                >
                <label :for="emp.cnpj" class="text-sm">
                  {{ emp.cnpj }} - {{ emp.razao_social || 'Sem Razão Social' }}
                </label>
              </div>
            </div>
            <UButton
              v-if="selectedUnknownEmpresas.length > 0"
              class="mt-3 w-full"
              label="Cadastrar Selecionadas e Re-importar"
              color="primary"
              variant="outline"
              :loading="registeringEmpresas"
              @click="registerSelectedEmpresas"
            />
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <UButton
        :label="importResult ? 'Fechar' : 'Cancelar'"
        color="neutral"
        variant="ghost"
        @click="closeImport"
      />
      <UButton
        v-if="!importResult"
        label="Importar"
        icon="i-lucide-upload"
        :loading="importing"
        :disabled="importFiles.length === 0 || importing"
        @click="handleImport"
      />
    </template>
  </UModal>

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
