<script setup lang="ts">
import type { EmpresaOverview } from '~/types'

const route = useRoute()
const toast = useToast()
const id = computed(() => route.params.id as string)
const importModalRef = useTemplateRef<{ show: () => void }>('importModal')

const { data, status, refresh } = await useFetch<EmpresaOverview>(
  () => id.value ? `/api/empresas/${id.value}/overview` : null as unknown as string,
  { lazy: true, server: false }
)

const empresa = computed(() => data.value?.empresa)
const syncState = computed(() => data.value?.sync_state)
const stats = computed(() => data.value?.stats ?? { total: 0, xml_completo: 0, xml_resumo: 0, manifestados: 0, valor_total: 0 })
const porCompetencia = computed(() => data.value?.documentos_por_competencia ?? [])
const recentes = computed(() => data.value?.documentos_recentes ?? [])

const syncing = ref(false)
async function handleSync() {
  syncing.value = true
  try {
    await $fetch(`/api/empresas/${id.value}/sync`, { method: 'POST' })
    toast.add({ title: 'Sincronização concluída', color: 'success' })
    refresh()
  } catch {
    toast.add({ title: 'Erro ao sincronizar', color: 'error' })
  } finally {
    syncing.value = false
  }
}

const togglingNFSe = ref(false)
async function toggleNFSe(val: boolean) {
  togglingNFSe.value = true
  try {
    await $fetch(`/api/empresas/${id.value}/nfse`, { method: 'PATCH', body: { habilitada: val } })
    toast.add({ title: val ? 'NFS-e habilitada' : 'NFS-e desabilitada', color: 'success' })
    refresh()
  } catch {
    toast.add({ title: 'Erro ao alterar NFS-e', color: 'error' })
  } finally {
    togglingNFSe.value = false
  }
}
</script>

<template>
  <UDashboardPanel :id="`empresa-${id}`">
    <template #header>
      <UDashboardNavbar :title="empresa?.razao_social ?? 'Carregando...'">
        <template #leading>
          <div class="flex items-center gap-2">
            <UDashboardSidebarCollapse />
            <UButton
              icon="i-lucide-arrow-left"
              color="neutral"
              variant="ghost"
              to="/empresas"
              size="sm"
            />
          </div>
        </template>
        <template #right>
          <UButton
            label="Importar"
            icon="i-lucide-upload"
            color="neutral"
            variant="outline"
            size="sm"
            @click="importModalRef?.show()"
          />
          <UButton
            label="Sincronizar"
            icon="i-lucide-refresh-cw"
            color="neutral"
            variant="outline"
            :loading="syncing"
            size="sm"
            @click="handleSync"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div
        v-if="status === 'pending'"
        class="flex items-center justify-center py-20 text-muted gap-2"
      >
        <UIcon name="i-lucide-loader-circle" class="animate-spin size-5" />
        Carregando...
      </div>

      <template v-else-if="empresa">
        <EmpresasOverviewHeader
          :empresa="empresa"
          :blocked-until="syncState?.blocked_until"
        />

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div class="space-y-6">
            <EmpresasCertSyncCard
              :empresa="empresa"
              :sync-state="syncState"
              :toggling-nfse="togglingNFSe"
              @toggle-nfse="toggleNFSe"
            />
            <EmpresasDadosCard :empresa="empresa" />
          </div>

          <div class="lg:col-span-2 space-y-6">
            <EmpresasOverviewStats :stats="stats" />
            <EmpresasOverviewCharts :por-competencia="porCompetencia" />
            <EmpresasRecentDocs :documentos="recentes" />
          </div>
        </div>
      </template>
    </template>
  </UDashboardPanel>

  <EmpresasImportModal
    ref="importModal"
    :empresa-id="id"
    @imported="refresh()"
  />
</template>
