<script setup lang="ts">
import type { Empresa } from '~/types'

const props = defineProps<{
  empresa?: Empresa | null
  empresas?: Empresa[] | null
}>()

const emit = defineEmits<{ deleted: [] }>()

const open = ref(false)
const loading = ref(false)
const toast = useToast()

const items = computed(() => {
  if (props.empresas && props.empresas.length > 0) {
    return props.empresas
  }
  if (props.empresa) {
    return [props.empresa]
  }
  return []
})

const title = computed(() => {
  if (items.value.length === 0) return 'Remover empresa'
  if (items.value.length === 1) return `Remover ${items.value[0]?.razao_social ?? 'empresa'}`
  return `Remover ${items.value.length} empresas`
})

watch(() => props.empresa, (val) => {
  if (val) open.value = true
})

watch(() => props.empresas, (val) => {
  if (val && val.length > 0) open.value = true
}, { deep: true })

watch(open, (val) => {
  if (!val) emit('deleted')
})

async function onConfirm() {
  if (items.value.length === 0) return

  loading.value = true
  try {
    await Promise.all(
      items.value.map(e => $fetch(`/api/empresas/${e.id}`, { method: 'DELETE' }))
    )
    const msg = items.value.length === 1
      ? 'Empresa removida'
      : `${items.value.length} empresas removidas`
    toast.add({ title: msg, color: 'success' })
    open.value = false
  } catch {
    toast.add({ title: 'Erro ao remover empresa(s)', color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="title"
    description="Tem certeza? Essa ação não pode ser desfeita."
  >
    <slot />

    <template #body>
      <div class="flex justify-end gap-2">
        <UButton
          label="Cancelar"
          color="neutral"
          variant="subtle"
          @click="open = false"
        />
        <UButton
          label="Remover"
          color="error"
          variant="solid"
          :loading="loading"
          @click="onConfirm"
        />
      </div>
    </template>
  </UModal>
</template>
