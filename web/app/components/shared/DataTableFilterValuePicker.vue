<script setup lang="ts">
import type { ColumnConfigBase, FilterModel, DataTableFilterActions } from '~/composables/useTableFilter'

const props = defineProps<{
  filter: FilterModel
  column: ColumnConfigBase
  actions: DataTableFilterActions
}>()

const search = ref('')

const allOptions = computed(() => toValue(props.column.options) ?? [])

const filteredOptions = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return allOptions.value
  return allOptions.value.filter(o =>
    o.label.toLowerCase().includes(q) || o.value.toLowerCase().includes(q)
  )
})

const selectedValues = computed(() => props.filter.values)

const selectedOptions = computed(() =>
  filteredOptions.value.filter(o => selectedValues.value.includes(o.value))
)
const unselectedOptions = computed(() =>
  filteredOptions.value.filter(o => !selectedValues.value.includes(o.value))
)

function toggleOption(value: string) {
  if (selectedValues.value.includes(value)) {
    props.actions.removeFilterValue(props.filter.columnId, value)
  } else {
    props.actions.addFilterValue(props.filter.columnId, value)
  }
}
</script>

<template>
  <div class="w-72">
    <div class="p-2 border-b border-default">
      <UInput
        v-model="search"
        autofocus
        placeholder="Pesquisar..."
        size="sm"
        icon="i-lucide-search"
        variant="none"
      />
    </div>
    <ul class="max-h-56 overflow-y-auto py-1">
      <template v-if="selectedOptions.length > 0">
        <li
          v-for="opt in selectedOptions"
          :key="opt.value"
          class="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer hover:bg-elevated rounded-sm mx-1"
          @click="toggleOption(opt.value)"
        >
          <UCheckbox
            :model-value="true"
            readonly
            tabindex="-1"
            class="pointer-events-none"
          />
          <UIcon v-if="opt.icon" :name="opt.icon" class="size-4 text-primary" />
          <span>{{ opt.label }}</span>
        </li>
        <li v-if="unselectedOptions.length > 0" class="h-px bg-border mx-2 my-1" role="separator" />
      </template>
      <li
        v-for="opt in unselectedOptions"
        :key="opt.value"
        class="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer hover:bg-elevated rounded-sm mx-1"
        @click="toggleOption(opt.value)"
      >
        <UCheckbox
          :model-value="false"
          readonly
          tabindex="-1"
          class="pointer-events-none"
        />
        <UIcon v-if="opt.icon" :name="opt.icon" class="size-4 text-muted" />
        <span>{{ opt.label }}</span>
      </li>
      <li
        v-if="filteredOptions.length === 0"
        class="px-3 py-4 text-sm text-center text-muted"
      >
        Nenhum resultado
      </li>
    </ul>
  </div>
</template>
