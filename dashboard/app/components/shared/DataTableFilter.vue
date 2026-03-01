<script setup lang="ts">
import type { ColumnConfigBase, FilterModel, DataTableFilterActions } from '~/composables/useTableFilter'

defineProps<{
  columns: ColumnConfigBase[]
  filters: FilterModel[]
  actions: DataTableFilterActions
}>()
</script>

<template>
  <div class="flex items-center gap-2 flex-wrap">
    <SharedDataTableFilterSelector
      :columns="columns"
      :filters="filters"
      :actions="actions"
    />

    <div v-if="filters.length > 0" class="flex items-center gap-2 flex-wrap py-1">
      <SharedDataTableFilterChip
        v-for="filter in filters"
        :key="filter.columnId"
        :filter="filter"
        :column="columns.find(c => c.id === filter.columnId)!"
        :actions="actions"
        class="shrink-0"
      />
    </div>

    <UButton
      v-if="filters.length > 0"
      color="error"
      variant="soft"
      icon="i-lucide-x"
      class="shrink-0"
      @click="actions.clearAll()"
    />
  </div>
</template>
