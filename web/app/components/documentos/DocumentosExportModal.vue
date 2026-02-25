<script setup lang="ts">
import type { DocumentoFiscal, DocumentoExportResponse } from '~/types'

type DeliveryMode = 'proxy' | 'presigned'
type ExportFormat = 'xml' | 'danfe' | 'ambos'
type OrganizationMode = 'tipo/competencia/cnpj' | 'cnpj/competencia/tipo' | 'competencia/cnpj/tipo'

const props = defineProps<{
  selectedRows: DocumentoFiscal[]
}>()

const toast = useToast()
const open = ref(false)
const exporting = ref(false)
const exportFormat = ref<ExportFormat>('xml')
const exportOrganization = ref<OrganizationMode>('tipo/competencia/cnpj')
const exportDeliveryMode = ref<DeliveryMode>('proxy')

function show() {
  if (props.selectedRows.length === 0) return
  open.value = true
}

async function handleExport() {
  const ids = props.selectedRows.map(d => d.id)
  if (ids.length === 0) return

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
      open.value = false
      return
    }

    const response = await fetch('/api/documentos/export', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
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

    const fileName = extractFileName(response.headers.get('content-disposition'), 'documentos')
    const blob = await response.blob()
    downloadBlob(blob, fileName)

    toast.add({ title: 'Arquivo de exportação baixado', color: 'success' })
    open.value = false
  } catch (error: unknown) {
    toast.add({
      title: getErrorMessage(error, 'Erro ao exportar documentos'),
      color: 'error'
    })
  } finally {
    exporting.value = false
  }
}

defineExpose({ show })
</script>

<template>
  <UModal
    v-model:open="open"
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
          @click="open = false"
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
</template>
