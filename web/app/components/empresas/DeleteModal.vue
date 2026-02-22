<script setup lang="ts">
import type { Empresa } from '~/types'

const props = defineProps<{ empresa: Empresa | null }>()
const emit = defineEmits<{ deleted: [] }>()

const open = ref(false)
const loading = ref(false)
const toast = useToast()

watch(() => props.empresa, (val) => {
  if (val) open.value = true
})

watch(open, (val) => {
  if (!val) emit('deleted')
})

async function onConfirm() {
  if (!props.empresa) return
  loading.value = true
  try {
    await $fetch(`/api/empresas/${props.empresa.id}`, { method: 'DELETE' })
    toast.add({ title: 'Empresa removida', color: 'success' })
    open.value = false
  } catch {
    toast.add({ title: 'Erro ao remover empresa', color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="`Remover ${empresa?.razao_social ?? 'empresa'}`"
    description="Tem certeza? Essa ação não pode ser desfeita."
  >
    <template #body>
      <div class="flex justify-end gap-2">
        <UButton label="Cancelar" color="neutral" variant="subtle" @click="open = false" />
        <UButton label="Remover" color="error" variant="solid" :loading="loading" @click="onConfirm" />
      </div>
    </template>
  </UModal>
</template>
