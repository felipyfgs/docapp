<script setup lang="ts">
import type { DocumentoFiscal } from '~/types'

const documentos = inject<ComputedRef<DocumentoFiscal[]>>('documentos', computed(() => []))
const status = inject<Ref<string>>('documentosStatus', ref('idle'))
const refresh = inject<() => Promise<void>>('documentosRefresh', async () => {})

const pendentes = computed(() =>
  documentos.value.filter(d =>
    d.xml_resumo
    && !d.manifestacao_status
    && d.chave_acesso
  )
)

const tableRef = useTemplateRef<{ selectedRows: DocumentoFiscal[] }>('table')
const manifestacaoModalRef = useTemplateRef<{ show: () => void }>('manifestacaoModal')
const xmlModalRef = useTemplateRef<{ show: (doc: DocumentoFiscal) => void }>('xmlModal')

const selectedRows = computed(() => tableRef.value?.selectedRows ?? [])
const selectedCount = computed(() => selectedRows.value.length)
</script>

<template>
  <div v-if="pendentes.length === 0 && status !== 'pending'" class="flex flex-col items-center justify-center gap-3 py-20">
    <UIcon name="i-lucide-check-circle" class="size-10 text-success" />
    <p class="text-sm text-muted">
      Nenhum documento pendente de manifestação.
    </p>
  </div>

  <template v-else>
    <DocumentosTable
      ref="table"
      :data="pendentes"
      :status="status"
      @view-xml="(doc: DocumentoFiscal) => xmlModalRef?.show(doc)"
    >
      <template #actions>
        <UButton
          v-if="selectedCount > 0"
          label="Manifestar"
          color="primary"
          icon="i-lucide-stamp"
          @click="manifestacaoModalRef?.show()"
        >
          <template #trailing>
            <UKbd>{{ selectedCount }}</UKbd>
          </template>
        </UButton>
      </template>
    </DocumentosTable>

    <DocumentosManifestacaoModal ref="manifestacaoModal" :selected-rows="selectedRows" @done="refresh()" />
    <DocumentosXmlModal ref="xmlModal" />
  </template>
</template>
