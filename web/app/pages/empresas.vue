<script setup lang="ts">
import type { Empresa } from '~/types'

const { data, status, refresh } = await useFetch<Empresa[]>('/api/empresas', { lazy: true })

const toast = useToast()
const empresaToDelete = ref<Empresa | null>(null)
const empresasToDelete = ref<Empresa[]>([])
const empresaToEdit = ref<Empresa | null>(null)
const tableRef = useTemplateRef('table')

function handleDelete(empresa: Empresa) {
  empresaToDelete.value = empresa
}

function handleBulkDelete() {
  const selected = tableRef.value?.selectedRows ?? []
  if (selected.length > 0) {
    empresasToDelete.value = selected
  }
}

function handleEdit(empresa: Empresa) {
  empresaToEdit.value = empresa
}

async function handleSync(empresa: Empresa) {
  try {
    toast.add({ title: `Sincronizando ${empresa.razao_social}...`, color: 'info' })
    await $fetch(`/api/empresas/${empresa.id}/sync`, { method: 'POST' })
    toast.add({ title: 'Sincronização concluída', color: 'success' })
    refresh()
  } catch (e: unknown) {
    const msg = (e as { data?: { message?: string } })?.data?.message ?? 'Erro ao sincronizar'
    toast.add({ title: msg, color: 'error' })
  }
}

function onDeleted() {
  empresaToDelete.value = null
  empresasToDelete.value = []
  refresh()
}

function onUpdated() {
  empresaToEdit.value = null
  refresh()
}

const selectedCount = computed(() => tableRef.value?.selectedRows?.length ?? 0)
</script>

<template>
  <UDashboardPanel id="empresas">
    <template #header>
      <UDashboardNavbar title="Empresas">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
        <template #right>
          <EmpresasAddModal
            :empresa="empresaToEdit"
            @created="refresh"
            @updated="onUpdated"
            @close="empresaToEdit = null"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <EmpresasTable
        ref="table"
        :data="data"
        :status="status"
        @delete="handleDelete"
        @edit="handleEdit"
        @sync="handleSync"
      >
        <template #actions>
          <EmpresasDeleteModal :empresas="empresasToDelete" @deleted="onDeleted">
            <UButton
              v-if="selectedCount > 0"
              label="Deletar"
              color="error"
              variant="subtle"
              icon="i-lucide-trash"
              @click="handleBulkDelete"
            >
              <template #trailing>
                <UKbd>{{ selectedCount }}</UKbd>
              </template>
            </UButton>
          </EmpresasDeleteModal>
        </template>
      </EmpresasTable>

      <EmpresasDeleteModal :empresa="empresaToDelete" :empresas="[]" @deleted="onDeleted" />
    </template>
  </UDashboardPanel>
</template>
