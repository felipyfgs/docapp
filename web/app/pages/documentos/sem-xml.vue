<script setup lang="ts">
import type { DocumentoFiscal } from '~/types'

const documentos = inject<ComputedRef<DocumentoFiscal[]>>('documentos', computed(() => []))
const status = inject<Ref<string>>('documentosStatus', ref('idle'))

const semXml = computed(() =>
  documentos.value.filter(d => d.xml_resumo)
)

const xmlModalRef = useTemplateRef<{ show: (doc: DocumentoFiscal) => void }>('xmlModal')
</script>

<template>
  <div v-if="semXml.length === 0 && status !== 'pending'" class="flex flex-col items-center justify-center gap-3 py-20">
    <UIcon name="i-lucide-check-circle" class="size-10 text-success" />
    <p class="text-sm text-muted">
      Todos os documentos possuem XML completo.
    </p>
  </div>

  <template v-else>
    <DocumentosTable
      :data="semXml"
      :status="status"
      @view-xml="(doc: DocumentoFiscal) => xmlModalRef?.show(doc)"
    />
    <DocumentosXmlModal ref="xmlModal" />
  </template>
</template>
