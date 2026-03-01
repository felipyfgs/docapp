<script setup lang="ts">
import type { ColumnConfigBase, DataTableFilterActions, FiltersState } from '~/composables/useTableFilter'

defineProps<{
  columns: ColumnConfigBase[]
  filters: FiltersState
  actions: DataTableFilterActions
  facetedCounts: Record<string, Map<string, number>>
  facetedMinMax: Record<string, [number, number]>
}>()

function activeCount(filters: FiltersState, columnId: string): number {
  const f = filters.find(f => f.columnId === columnId)
  return f?.values.length ?? 0
}
</script>

<template>
  <div class="h-full overflow-y-auto">
    <UAccordion
      type="multiple"
      :default-value="columns.filter(c => c.defaultOpen).map(c => c.id)"
      :ui="{
        item: 'border-none',
        trigger: 'px-3 py-2 hover:no-underline text-sm',
        content: 'px-0',
        body: 'px-3 pb-3 pt-0'
      }"
      :items="columns.map(col => ({
        label: col.displayName,
        icon: col.icon,
        value: col.id
      }))"
    >
      <template #leading="{ item }">
        <UIcon
          :name="columns.find(c => c.id === item.value)?.icon || 'i-lucide-filter'"
          class="size-4 text-muted shrink-0"
        />
      </template>

      <template #default="{ item }">
        <div class="flex w-full items-center justify-between gap-2 pr-1">
          <span class="text-sm font-medium truncate">
            {{ item.label }}
          </span>
          <UButton
            v-if="activeCount(filters, item.value!) > 0"
            size="xs"
            color="neutral"
            variant="ghost"
            icon="i-lucide-x"
            :label="String(activeCount(filters, item.value!))"
            @click.stop="actions.removeFilter(item.value!)"
          />
        </div>
      </template>

      <template #body="{ item }">
        <SharedDataTableFilterCheckbox
          v-if="columns.find(c => c.id === item.value)?.type === 'option'"
          :column="columns.find(c => c.id === item.value)!"
          :filters="filters"
          :actions="actions"
          :faceted-counts="facetedCounts"
        />
        <SharedDataTableFilterSlider
          v-else-if="columns.find(c => c.id === item.value)?.type === 'slider'"
          :column="columns.find(c => c.id === item.value)!"
          :filters="filters"
          :actions="actions"
          :faceted-min-max="facetedMinMax"
        />
        <SharedDataTableFilterTimerange
          v-else-if="columns.find(c => c.id === item.value)?.type === 'timerange'"
          :column="columns.find(c => c.id === item.value)!"
          :filters="filters"
          :actions="actions"
        />
      </template>
    </UAccordion>
  </div>
</template>
