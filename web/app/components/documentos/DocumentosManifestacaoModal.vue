<script setup lang="ts">
import type { DocumentoFiscal, ManifestacaoResult } from '~/types'

const props = defineProps<{
  selectedRows: DocumentoFiscal[]
}>()

const emit = defineEmits<{
  done: []
}>()

const toast = useToast()
const open = ref(false)
const processing = ref(false)
const tipoEvento = ref('210200')
const justificativa = ref('')
const result = ref<ManifestacaoResult | null>(null)

const tipoEventoOptions = [
  { label: 'Confirmação da Operação', value: '210200' },
  { label: 'Ciência da Operação', value: '210210' },
  { label: 'Desconhecimento da Operação', value: '210220' },
  { label: 'Operação Não Realizada', value: '210240' }
]

const requiresJustificativa = computed(() => tipoEvento.value === '210240')

function show() {
  if (props.selectedRows.length === 0) return
  result.value = null
  justificativa.value = ''
  open.value = true
}

async function handleManifestar() {
  const ids = props.selectedRows.map(d => d.id)
  if (ids.length === 0) return

  processing.value = true
  result.value = null

  try {
    const response = await $fetch<ManifestacaoResult>('/api/documentos/manifestar', {
      method: 'POST',
      body: {
        ids,
        tipo_evento: tipoEvento.value,
        justificativa: justificativa.value
      }
    })

    result.value = response

    if (response.erros === 0) {
      toast.add({ title: `${response.sucesso} documento(s) manifestado(s) com sucesso`, color: 'success' })
      open.value = false
      emit('done')
    } else if (response.sucesso > 0) {
      toast.add({ title: `${response.sucesso} sucesso, ${response.erros} erro(s)`, color: 'warning' })
      emit('done')
    } else {
      toast.add({ title: 'Nenhum documento manifestado', color: 'error' })
    }
  } catch (error: unknown) {
    const msg = (error as { data?: { message?: string } })?.data?.message ?? 'Erro ao manifestar documentos'
    toast.add({ title: msg, color: 'error' })
  } finally {
    processing.value = false
  }
}

defineExpose({ show })
</script>

<template>
  <UModal
    v-model:open="open"
    title="Manifestar Documentos"
    :description="`${selectedRows.length} documento(s) selecionado(s)`"
  >
    <template #body>
      <div class="space-y-4">
        <UFormField label="Tipo de Manifestação">
          <USelect
            v-model="tipoEvento"
            :items="tipoEventoOptions"
            class="w-full"
          />
        </UFormField>

        <UFormField v-if="requiresJustificativa" label="Justificativa">
          <UTextarea
            v-model="justificativa"
            placeholder="Informe a justificativa (obrigatório para Operação Não Realizada)"
            :rows="3"
            class="w-full"
          />
        </UFormField>

        <div v-if="result && result.erros > 0" class="space-y-2">
          <p class="text-sm font-medium text-error">
            {{ result.erros }} documento(s) com erro:
          </p>
          <div class="max-h-40 overflow-y-auto space-y-1">
            <div
              v-for="item in result.resultados.filter(r => r.status === 'erro')"
              :key="item.id"
              class="text-xs text-muted bg-elevated rounded px-2 py-1"
            >
              <span class="font-mono">{{ item.chave_acesso?.slice(-10) || item.id }}</span>
              <span class="ml-1">— {{ item.erro }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton
          label="Cancelar"
          color="neutral"
          variant="subtle"
          :disabled="processing"
          @click="open = false"
        />
        <UButton
          label="Manifestar"
          color="primary"
          :loading="processing"
          icon="i-lucide-stamp"
          :disabled="requiresJustificativa && !justificativa.trim()"
          @click="handleManifestar"
        />
      </div>
    </template>
  </UModal>
</template>
