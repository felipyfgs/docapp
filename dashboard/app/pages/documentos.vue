<script setup lang="ts">
import type { DocumentoFiscal, DocumentoListResponse } from '~/types'

const PAGE_SIZE = 100

function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1)
}

function endOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0)
}

function formatDateParam(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

const period = ref({
  start: startOfMonth(new Date()),
  end: endOfMonth(new Date())
})

const page = ref(1)
const allDocs = ref<DocumentoFiscal[]>([])
const totalCount = ref(0)

const queryParams = computed(() => ({
  page: page.value,
  page_size: PAGE_SIZE,
  data_inicio: formatDateParam(period.value.start),
  data_fim: formatDateParam(period.value.end)
}))

const { data, status, refresh: fetchRefresh } = await useFetch<DocumentoListResponse>('/api/documentos', {
  lazy: true,
  query: queryParams
})

watch(data, (res) => {
  if (!res) return
  if (page.value === 1) {
    allDocs.value = res.items ?? []
  } else {
    allDocs.value = [...allDocs.value, ...(res.items ?? [])]
  }
  totalCount.value = res.total ?? 0
})

watch(period, () => {
  page.value = 1
  allDocs.value = []
}, { deep: true })

const documentos = computed(() => allDocs.value)
const hasMore = computed(() => allDocs.value.length < totalCount.value)

function loadMore() {
  page.value++
}

async function refresh() {
  page.value = 1
  allDocs.value = []
  await fetchRefresh()
}

provide('documentos', documentos)
provide('documentosStatus', status)
provide('documentosRefresh', refresh)
provide('documentosPeriod', period)

const importModalRef = useTemplateRef<{ show: () => void }>('importModal')
</script>

<template>
  <UDashboardPanel id="documentos">
    <template #header>
      <UDashboardNavbar title="Documentos">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <DocumentosPeriodBar v-model="period" />

          <UButton
            label="Importar"
            icon="i-lucide-upload"
            color="neutral"
            variant="outline"
            @click="importModalRef?.show()"
          />
          <UButton
            icon="i-lucide-refresh-cw"
            color="neutral"
            variant="ghost"
            @click="refresh()"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>

      <NuxtPage />

      <div v-if="hasMore" class="flex justify-center py-4">
        <UButton
          label="Carregar mais"
          color="neutral"
          variant="outline"
          icon="i-lucide-chevrons-down"
          :loading="status === 'pending' && page > 1"
          @click="loadMore"
        />
      </div>
    </template>
  </UDashboardPanel>

  <DocumentosImportModal ref="importModal" @imported="refresh()" />
</template>
