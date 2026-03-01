<script setup lang="ts">
import type { DocumentoFiscal, DocumentoXMLResponse, DocumentoItem, DocumentoItensResponse } from '~/types'

const toast = useToast()
const { formatBRL } = useDocumentoFormatters()
const open = ref(false)
const loading = ref(false)
const content = ref('')
const documento = ref<DocumentoFiscal | null>(null)
const activeTab = ref('itens')
const itens = ref<DocumentoItem[]>([])
const itensLoading = ref(false)

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
  itens.value = []
  activeTab.value = 'itens'

  try {
    const [xmlResponse, itensResponse] = await Promise.all([
      $fetch<DocumentoXMLResponse>(`/api/documentos/${doc.id}/xml`),
      $fetch<DocumentoItensResponse>(`/api/documentos/${doc.id}/itens`).catch(() => ({ items: [], total: 0 }))
    ])
    content.value = xmlResponse.xml
    itens.value = itensResponse.items || []
    if (itens.value.length === 0) {
      activeTab.value = 'xml'
    }
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

function formatQtd(value: number): string {
  if (value === Math.floor(value)) return value.toString()
  return value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 4 })
}

const tabs = computed(() => [
  { label: `Itens (${itens.value.length})`, value: 'itens', icon: 'i-lucide-package' },
  { label: 'XML', value: 'xml', icon: 'i-lucide-file-code-2' }
])

defineExpose({ show })
</script>

<template>
  <UModal
    v-model:open="open"
    :title="title"
    :description="description"
    :ui="{ content: 'sm:max-w-6xl', body: 'max-h-[72vh] overflow-auto' }"
  >
    <template #body>
      <div
        v-if="loading"
        class="flex items-center justify-center gap-2 py-8 text-muted"
      >
        <UIcon name="i-lucide-loader-circle" class="animate-spin size-4" />
        Carregando...
      </div>

      <template v-else>
        <UTabs
          v-model="activeTab"
          :items="tabs"
          class="w-full"
        />

        <div v-if="activeTab === 'itens'" class="mt-4">
          <div v-if="itens.length === 0" class="text-center py-8 text-muted">
            Nenhum item encontrado para este documento.
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-xs">
              <thead>
                <tr class="border-b border-default text-left text-muted">
                  <th class="py-2 px-2 font-medium">#</th>
                  <th class="py-2 px-2 font-medium">Cod</th>
                  <th class="py-2 px-2 font-medium min-w-48">Descricao</th>
                  <th class="py-2 px-2 font-medium">NCM</th>
                  <th class="py-2 px-2 font-medium">CFOP</th>
                  <th class="py-2 px-2 font-medium">Un</th>
                  <th class="py-2 px-2 font-medium text-right">Qtd</th>
                  <th class="py-2 px-2 font-medium text-right">Vl.Unit</th>
                  <th class="py-2 px-2 font-medium text-right">Vl.Total</th>
                  <th class="py-2 px-2 font-medium text-right">Desc</th>
                  <th class="py-2 px-2 font-medium text-right">ICMS</th>
                  <th class="py-2 px-2 font-medium text-right">PIS</th>
                  <th class="py-2 px-2 font-medium text-right">COFINS</th>
                  <th class="py-2 px-2 font-medium text-right">IPI</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in itens"
                  :key="item.id"
                  class="border-b border-default/50 hover:bg-elevated/50"
                >
                  <td class="py-1.5 px-2 text-muted">{{ item.n_item }}</td>
                  <td class="py-1.5 px-2 font-mono text-muted">{{ item.c_prod || '—' }}</td>
                  <td class="py-1.5 px-2 font-medium truncate max-w-64" :title="item.x_prod">{{ item.x_prod }}</td>
                  <td class="py-1.5 px-2 font-mono">{{ item.ncm || '—' }}</td>
                  <td class="py-1.5 px-2 font-mono">{{ item.cfop || '—' }}</td>
                  <td class="py-1.5 px-2">{{ item.u_com || '—' }}</td>
                  <td class="py-1.5 px-2 text-right tabular-nums">{{ formatQtd(item.q_com) }}</td>
                  <td class="py-1.5 px-2 text-right tabular-nums">{{ formatBRL(item.v_un_com) }}</td>
                  <td class="py-1.5 px-2 text-right tabular-nums font-medium">{{ formatBRL(item.v_prod) }}</td>
                  <td class="py-1.5 px-2 text-right tabular-nums text-muted">{{ item.v_desc > 0 ? formatBRL(item.v_desc) : '—' }}</td>
                  <td class="py-1.5 px-2 text-right tabular-nums">{{ item.icms_v_icms > 0 ? formatBRL(item.icms_v_icms) : '—' }}</td>
                  <td class="py-1.5 px-2 text-right tabular-nums">{{ item.pis_v_pis > 0 ? formatBRL(item.pis_v_pis) : '—' }}</td>
                  <td class="py-1.5 px-2 text-right tabular-nums">{{ item.cofins_v_cofins > 0 ? formatBRL(item.cofins_v_cofins) : '—' }}</td>
                  <td class="py-1.5 px-2 text-right tabular-nums">{{ item.ipi_v_ipi > 0 ? formatBRL(item.ipi_v_ipi) : '—' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div v-if="activeTab === 'xml'" class="mt-4">
          <pre class="text-xs leading-relaxed whitespace-pre-wrap break-all rounded-md border border-default bg-elevated/50 p-3">{{ content }}</pre>
        </div>
      </template>
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
          v-if="activeTab === 'xml'"
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
