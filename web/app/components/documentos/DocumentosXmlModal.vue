<script setup lang="ts">
import type { DocumentoFiscal, DocumentoXMLResponse } from '~/types'

const toast = useToast()
const open = ref(false)
const loading = ref(false)
const content = ref('')
const documento = ref<DocumentoFiscal | null>(null)

const title = computed(() => {
  if (!documento.value) return 'XML do documento'
  return documento.value.chave_acesso
    ? `XML - ${documento.value.chave_acesso}`
    : `XML - Documento #${documento.value.id}`
})

const description = computed(() => {
  return documento.value?.xml_resumo ? 'Documento em modo resumo' : 'Documento completo'
})

async function show(doc: DocumentoFiscal) {
  documento.value = doc
  open.value = true
  loading.value = true
  content.value = ''

  try {
    const response = await $fetch<DocumentoXMLResponse>(`/api/documentos/${doc.id}/xml`)
    content.value = response.xml
  } catch (error: unknown) {
    toast.add({
      title: getErrorMessage(error, 'Erro ao carregar XML'),
      color: 'error'
    })
  } finally {
    loading.value = false
  }
}

function copyXML() {
  if (!content.value) return
  navigator.clipboard.writeText(content.value)
  toast.add({ title: 'XML copiado', color: 'success' })
}

defineExpose({ show })
</script>

<template>
  <UModal
    v-model:open="open"
    :title="title"
    :description="description"
    :ui="{ content: 'sm:max-w-5xl', body: 'max-h-[72vh] overflow-auto' }"
  >
    <template #body>
      <div
        v-if="loading"
        class="flex items-center justify-center gap-2 py-8 text-muted"
      >
        <UIcon name="i-lucide-loader-circle" class="animate-spin size-4" />
        Carregando XML...
      </div>

      <pre
        v-else
        class="text-xs leading-relaxed whitespace-pre-wrap break-all rounded-md border border-default bg-elevated/50 p-3"
      >{{ content }}</pre>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton
          label="Fechar"
          color="neutral"
          variant="subtle"
          @click="open = false"
        />
        <UButton
          label="Copiar XML"
          color="primary"
          icon="i-lucide-copy"
          :disabled="!content"
          @click="copyXML"
        />
      </div>
    </template>
  </UModal>
</template>
