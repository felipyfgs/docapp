<script setup lang="ts">
import type { ColumnConfigBase, DataTableFilterActions } from '~/composables/useTableFilter'

const props = defineProps<{
  columns: ColumnConfigBase[]
  actions: DataTableFilterActions
}>()

const inputRef = useTemplateRef<{ inputRef: { $el: HTMLInputElement } }>('input')
const inputValue = ref('')
const open = ref(false)
const selectedIndex = ref(0)

const FIELD_DELIMITER = ':'
const VALUE_DELIMITER = ','
const RANGE_DELIMITER = '-'

const parsed = computed(() => {
  const raw = inputValue.value.trim()
  const colonIdx = raw.indexOf(FIELD_DELIMITER)
  if (colonIdx === -1) return { field: null, query: raw, valueQuery: '' }
  const fieldKey = raw.slice(0, colonIdx).trim().toLowerCase()
  const valueQuery = raw.slice(colonIdx + 1)
  const col = props.columns.find(c =>
    c.id.toLowerCase() === fieldKey
    || c.displayName.toLowerCase() === fieldKey
  )
  return { field: col ?? null, query: fieldKey, valueQuery }
})

const suggestions = computed(() => {
  const { field, query, valueQuery } = parsed.value

  if (field) {
    if (field.type === 'option') {
      const options = toValue(field.options) ?? []
      const vq = valueQuery.toLowerCase()
      return options
        .filter(o => !vq || o.label.toLowerCase().includes(vq) || o.value.toLowerCase().includes(vq))
        .slice(0, 10)
        .map(o => ({ type: 'value' as const, field, label: o.label, value: o.value }))
    }
    return []
  }

  const q = query.toLowerCase()
  return props.columns
    .filter(c => !c.commandDisabled)
    .filter(c => !q || c.id.toLowerCase().includes(q) || c.displayName.toLowerCase().includes(q))
    .map(c => ({ type: 'field' as const, field: c, label: c.displayName, value: c.id }))
})

function selectSuggestion(item: (typeof suggestions.value)[number]) {
  if (item.type === 'field') {
    inputValue.value = `${item.value}${FIELD_DELIMITER}`
    open.value = true
    selectedIndex.value = 0
  } else if (item.type === 'value') {
    const col = item.field as ColumnConfigBase
    const currentValues = inputValue.value.split(FIELD_DELIMITER)[1]?.split(VALUE_DELIMITER).filter(Boolean) ?? []
    if (!currentValues.includes(item.value)) {
      props.actions.addFilterValue(col.id, item.value)
    }
    inputValue.value = ''
    open.value = false
  }
}

function handleSubmit() {
  const { field, valueQuery } = parsed.value
  if (!field) return

  if (field.type === 'option') {
    const values = valueQuery.split(VALUE_DELIMITER).map(v => v.trim()).filter(Boolean)
    for (const v of values) {
      const options = toValue(field.options) ?? []
      const match = options.find(o => o.value.toLowerCase() === v.toLowerCase() || o.label.toLowerCase() === v.toLowerCase())
      if (match) {
        props.actions.addFilterValue(field.id, match.value)
      }
    }
  } else if (field.type === 'slider') {
    const parts = valueQuery.split(RANGE_DELIMITER).map(v => v.trim()).filter(Boolean)
    if (parts.length === 2) {
      props.actions.setFilterValues(field.id, parts)
    }
  } else if (field.type === 'timerange') {
    const parts = valueQuery.split(RANGE_DELIMITER).map(v => v.trim()).filter(Boolean)
    if (parts.length === 2) {
      props.actions.setFilterValues(field.id, parts)
    }
  }

  inputValue.value = ''
  open.value = false
}

function focusInput() {
  const el = inputRef.value?.inputRef?.$el as HTMLInputElement | undefined
  el?.focus()
}

function blurInput() {
  const el = inputRef.value?.inputRef?.$el as HTMLInputElement | undefined
  el?.blur()
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    if (suggestions.value.length > 0 && open.value) {
      selectSuggestion(suggestions.value[selectedIndex.value]!)
    } else {
      handleSubmit()
    }
  } else if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, suggestions.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
  } else if (e.key === 'Escape') {
    open.value = false
    blurInput()
  }
}

watch(inputValue, () => {
  open.value = inputValue.value.length > 0
  selectedIndex.value = 0
})

onMounted(() => {
  const handler = (e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      focusInput()
    }
  }
  window.addEventListener('keydown', handler)
  onUnmounted(() => window.removeEventListener('keydown', handler))
})

function onBlur() {
  setTimeout(() => {
    open.value = false
  }, 150)
}
</script>

<template>
  <div class="relative">
    <UInput
      ref="input"
      v-model="inputValue"
      icon="i-lucide-search"
      placeholder="Filtrar..."
      size="sm"
      @keydown="onKeydown"
      @focus="open = inputValue.length > 0"
      @blur="onBlur"
    >
      <template #trailing>
        <UKbd class="hidden sm:inline-flex">
          ⌘K
        </UKbd>
      </template>
    </UInput>

    <div
      v-if="open && suggestions.length > 0"
      class="absolute top-full left-0 right-0 z-50 mt-1 rounded-md border border-default bg-default shadow-lg overflow-hidden"
    >
      <div class="max-h-64 overflow-y-auto py-1">
        <button
          v-for="(item, index) in suggestions"
          :key="`${item.type}-${item.value}`"
          type="button"
          class="flex w-full items-center gap-2 px-3 py-1.5 text-sm text-left"
          :class="index === selectedIndex ? 'bg-elevated text-default' : 'text-muted hover:bg-elevated/50'"
          @mouseenter="selectedIndex = index"
          @mousedown.prevent="selectSuggestion(item)"
        >
          <UIcon
            v-if="item.type === 'field'"
            :name="item.field.icon"
            class="size-4 shrink-0"
          />
          <UIcon
            v-else
            name="i-lucide-corner-down-right"
            class="size-3.5 shrink-0 text-muted"
          />
          <span class="truncate">
            {{ item.label }}
          </span>
          <span
            v-if="item.type === 'field'"
            class="ml-auto font-mono text-[10px] text-muted"
          >
            {{ item.value }}
          </span>
        </button>
      </div>

      <div class="hidden sm:flex border-t border-default px-3 py-1.5 text-[10px] text-muted items-center gap-3">
        <span>
          <kbd class="font-mono">↑↓</kbd> navegar
        </span>
        <span>
          <kbd class="font-mono">Enter</kbd> selecionar
        </span>
        <span>
          <kbd class="font-mono">Esc</kbd> fechar
        </span>
      </div>
    </div>
  </div>
</template>
