<script setup lang="ts">
const emit = defineEmits<{ imported: [] }>()

const toast = useToast()
const open = ref(false)
const importing = ref(false)
const importFiles = ref<File[]>([])
const fileInputRef = ref<HTMLInputElement | null>(null)

type AutoImportResult = {
  by_empresa: Record<string, { imported: number, failed: number, errors?: string[] }>
  unknown: number
  unknown_empresas?: {
    cnpj: string
    razao_social: string
    nome_fantasia: string
    logradouro: string
    numero: string
    bairro: string
    cep: string
    cidade: string
    estado: string
  }[]
}

const importResult = ref<AutoImportResult | null>(null)
const selectedUnknownEmpresas = ref<string[]>([])
const registeringEmpresas = ref(false)
const registerProgress = ref(0)

const importFileLabel = computed(() => {
  if (importFiles.value.length === 0) return 'Clique para selecionar arquivos'
  if (importFiles.value.length === 1) return importFiles.value[0]!.name
  return `${importFiles.value.length} arquivo(s) selecionado(s)`
})

function show() {
  open.value = true
}

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
    if (total > 0) emit('imported')
  } catch {
    toast.add({ title: 'Erro ao importar documentos', color: 'error' })
  } finally {
    importing.value = false
  }
}

function close() {
  open.value = false
  importFiles.value = []
  importResult.value = null
  selectedUnknownEmpresas.value = []
  registerProgress.value = 0
}

async function registerSelectedEmpresas() {
  if (selectedUnknownEmpresas.value.length === 0) return
  registeringEmpresas.value = true
  registerProgress.value = 0

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
            nome_fantasia: emp.nome_fantasia || '',
            logradouro: emp.logradouro || '',
            numero: emp.numero || '',
            bairro: emp.bairro || '',
            cep: emp.cep || '',
            cidade: emp.cidade || '',
            estado: emp.estado || 'SP',
            lookback_days: 90
          }
        })
        successCount++
      } catch (err) {
        console.error('Failed to create empresa:', emp.cnpj, err)
      }
      registerProgress.value++
    }

    if (successCount > 0) {
      toast.add({ title: `${successCount} empresa(s) cadastrada(s)`, color: 'success' })
      selectedUnknownEmpresas.value = []
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

defineExpose({ show })
</script>

<template>
  <UModal
    v-model:open="open"
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

        <div v-if="importing" class="space-y-1.5 pt-1">
          <UProgress animation="carousel" />
          <p class="text-xs text-muted text-center">
            Processando {{ importFiles.length }} arquivo(s)...
          </p>
        </div>

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
          <div v-if="importResult.unknown > 0 && !importResult.unknown_empresas?.length" class="text-xs text-warning px-1">
            {{ importResult.unknown }} arquivo(s) sem empresa correspondente — verifique se a empresa está cadastrada.
          </div>

          <UAlert
            v-if="importResult.unknown === 0 && !importResult.unknown_empresas?.length && Object.keys(importResult.by_empresa).length > 0"
            color="success"
            variant="subtle"
            title="Importação concluída!"
            description="Todos os arquivos válidos foram processados com sucesso."
            icon="i-lucide-check-circle"
          />

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
            <div v-if="registeringEmpresas" class="mt-3 space-y-1.5">
              <UProgress
                :model-value="registerProgress"
                :max="selectedUnknownEmpresas.length"
                status
              />
              <p class="text-xs text-muted text-center">
                Cadastrando empresa {{ registerProgress + 1 }} de {{ selectedUnknownEmpresas.length }}...
              </p>
            </div>
            <UButton
              v-else-if="selectedUnknownEmpresas.length > 0"
              class="mt-3 w-full"
              label="Cadastrar Selecionadas e Re-importar"
              color="primary"
              variant="outline"
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
        @click="close"
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
</template>
