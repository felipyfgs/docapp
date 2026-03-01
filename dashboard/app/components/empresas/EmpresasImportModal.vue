<script setup lang="ts">
const props = defineProps<{
  empresaId: string | number
}>()

const emit = defineEmits<{ imported: [] }>()

const toast = useToast()
const open = ref(false)
const importing = ref(false)
const importFile = ref<File | null>(null)
const importResult = ref<{ imported: number, failed: number, errors?: string[] } | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)

function show() {
  open.value = true
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  importFile.value = input.files?.[0] ?? null
  importResult.value = null
}

function triggerFileInput() {
  fileInputRef.value?.click()
}

async function handleImport() {
  if (!importFile.value) return
  importing.value = true
  importResult.value = null
  try {
    const form = new FormData()
    form.append('file', importFile.value)
    const result = await $fetch<{ imported: number, failed: number, errors?: string[] }>(
      `/api/empresas/${props.empresaId}/import`,
      { method: 'POST', body: form }
    )
    importResult.value = result
    if (result.imported > 0) emit('imported')
  } catch {
    toast.add({ title: 'Erro ao importar documentos', color: 'error' })
  } finally {
    importing.value = false
  }
}

function close() {
  open.value = false
  importFile.value = null
  importResult.value = null
}

defineExpose({ show })
</script>

<template>
  <UModal v-model:open="open" title="Importar documentos" :ui="{ footer: 'justify-end' }">
    <template #body>
      <p class="text-sm text-muted mb-4">
        Selecione um arquivo <strong>.xml</strong> (NF-e individual) ou <strong>.zip</strong> com múltiplos XMLs.
      </p>

      <input
        ref="fileInputRef"
        type="file"
        accept=".xml,.zip"
        class="hidden"
        @change="onFileChange"
      >

      <button
        type="button"
        class="w-full flex flex-col items-center justify-center gap-2 border-2 border-dashed rounded-lg p-8 transition-colors"
        :class="importFile ? 'border-primary bg-primary/5' : 'border-default hover:border-primary'"
        @click="triggerFileInput"
      >
        <UIcon
          :name="importFile ? 'i-lucide-file-check' : 'i-lucide-file-up'"
          class="size-8"
          :class="importFile ? 'text-primary' : 'text-muted'"
        />
        <span class="text-sm font-medium" :class="importFile ? 'text-primary' : 'text-muted'">
          {{ importFile ? importFile.name : 'Clique para selecionar um arquivo' }}
        </span>
        <span v-if="importFile" class="text-xs text-muted">
          {{ (importFile.size / 1024).toFixed(0) }} KB · clique para trocar
        </span>
        <span v-else class="text-xs text-dimmed">XML ou ZIP · máx 200 MB</span>
      </button>

      <div v-if="importing" class="space-y-1.5 pt-1">
        <UProgress animation="carousel" />
        <p class="text-xs text-muted text-center">
          Importando...
        </p>
      </div>

      <div v-if="importResult" class="mt-4 rounded-lg p-3 text-sm" :class="importResult.failed > 0 ? 'bg-warning/10' : 'bg-success/10'">
        <p class="font-medium text-highlighted">
          {{ importResult.imported }} documento{{ importResult.imported !== 1 ? 's' : '' }} importado{{ importResult.imported !== 1 ? 's' : '' }}
          <template v-if="importResult.failed > 0">
            · {{ importResult.failed }} falha{{ importResult.failed !== 1 ? 's' : '' }}
          </template>
        </p>
        <ul v-if="importResult.errors?.length" class="mt-1 list-disc list-inside text-xs text-muted space-y-0.5">
          <li v-for="(err, i) in importResult.errors.slice(0, 5)" :key="i">
            {{ err }}
          </li>
          <li v-if="importResult.errors.length > 5">
            ...e mais {{ importResult.errors.length - 5 }} erro(s)
          </li>
        </ul>
      </div>
    </template>

    <template #footer>
      <UButton
        label="Cancelar"
        color="neutral"
        variant="ghost"
        @click="close"
      />
      <UButton
        label="Importar"
        icon="i-lucide-upload"
        :loading="importing"
        :disabled="!importFile || importing"
        @click="handleImport"
      />
    </template>
  </UModal>
</template>
