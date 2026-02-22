<script setup lang="ts">
import type { Empresa } from '~/types'

const { data, status, refresh } = await useFetch<Empresa[]>('/api/empresas', { lazy: true })

const empresaToDelete = ref<Empresa | null>(null)

function handleDelete(empresa: Empresa) {
  empresaToDelete.value = empresa
}

function onDeleted() {
  empresaToDelete.value = null
  refresh()
}
</script>

<template>
  <UDashboardPanel id="empresas">
    <template #header>
      <UDashboardNavbar title="Empresas">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
        <template #right>
          <EmpresasAddModal @created="refresh" />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <EmpresasTable :data="data" :status="status" @delete="handleDelete" />
      <EmpresasDeleteModal :empresa="empresaToDelete" @deleted="onDeleted" />
    </template>
  </UDashboardPanel>
</template>
