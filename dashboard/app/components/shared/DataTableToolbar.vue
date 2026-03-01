<script setup lang="ts">
import type { ColumnConfigBase, DataTableFilterActions } from '~/composables/useTableFilter'

defineProps<{
  columns: ColumnConfigBase[]
  actions: DataTableFilterActions
  totalRows: number
  filteredRows: number
  hasFilters: boolean
  activeFilterCount: number
}>()

const { open, toggle } = useFilterControls()

onMounted(() => {
  const handler = (e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
      e.preventDefault()
      toggle()
    }
  }
  window.addEventListener('keydown', handler)
  onUnmounted(() => window.removeEventListener('keydown', handler))
})
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-1.5">
    <div class="flex items-center gap-1.5">
      <UTooltip text="Mostrar/ocultar filtros (⌘B)">
        <UButton
          size="sm"
          variant="ghost"
          color="neutral"
          :icon="open ? 'i-lucide-panel-left-close' : 'i-lucide-panel-left-open'"
          @click="toggle"
        />
      </UTooltip>

      <SharedDataTableFilterCommandInput
        :columns="columns"
        :actions="actions"
        class="w-48 sm:w-64"
      />

      <span class="hidden lg:inline text-sm text-muted whitespace-nowrap">
        {{ filteredRows }} de {{ totalRows }} linha(s)
      </span>
    </div>

    <div class="flex flex-wrap items-center gap-1.5">
      <UButton
        v-if="hasFilters"
        label="Limpar"
        color="neutral"
        variant="ghost"
        size="sm"
        icon="i-lucide-x"
        @click="actions.clearAll()"
      >
        <template #trailing>
          <UKbd>{{ activeFilterCount }}</UKbd>
        </template>
      </UButton>

      <slot name="actions" />
    </div>
  </div>
</template>
