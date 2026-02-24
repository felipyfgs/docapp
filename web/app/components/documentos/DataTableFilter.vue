<script setup lang="ts">
import type { ColumnConfigBase, FilterModel, DataTableFilterActions } from '~/composables/useTableFilter'

defineProps<{
  columns: ColumnConfigBase[]
  filters: FilterModel[]
  actions: DataTableFilterActions
}>()
</script>

<template>
  <div class="flex flex-wrap items-center gap-2">
    <DocumentosFilterSelector
      :columns="columns"
      :filters="filters"
      :actions="actions"
    />

    <DocumentosFilterChip
      v-for="filter in filters"
      :key="filter.columnId"
      :filter="filter"
      :column="columns.find(c => c.id === filter.columnId)!"
      :actions="actions"
    />

    <UButton
      v-if="filters.length > 0"
      color="neutral"
      variant="ghost"
      size="sm"
      icon="i-lucide-x"
      label="Limpar"
      class="h-7"
      @click="actions.clearAll()"
    />
  </div>
</template>
