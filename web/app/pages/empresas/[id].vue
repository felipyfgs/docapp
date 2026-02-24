<script setup lang="ts">
import { h, resolveComponent } from 'vue'
import { VisXYContainer, VisGroupedBar, VisAxis, VisCrosshair, VisTooltip } from '@unovis/vue'
import { format, parseISO } from 'date-fns'
import { ptBR } from 'date-fns/locale'
import type { TableColumn } from '@nuxt/ui'
import type { EmpresaOverview, CompetenciaCount, DocumentoFiscal } from '~/types'

const route = useRoute()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data, status, refresh } = await useFetch<EmpresaOverview>(
  () => id.value ? `/api/empresas/${id.value}/overview` : null as unknown as string,
  { lazy: true, server: false }
)

const empresa = computed(() => data.value?.empresa)
const syncState = computed(() => data.value?.sync_state)
const stats = computed(() => data.value?.stats ?? { total: 0, xml_completo: 0, xml_resumo: 0, manifestados: 0 })
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

const importOpen = ref(false)
const importing = ref(false)
const importFile = ref<File | null>(null)
const importResult = ref<{ imported: number, failed: number, errors?: string[] } | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  importFile.value = input.files?.[0] ?? null
  importResult.value = null
}

function triggerFileInput() {
  fileInputRef.value?.click()
}

async function handleImport() {
  if (!importFile.value) return
  importing.value = true
  importResult.value = null
  try {
    const form = new FormData()
    form.append('file', importFile.value)
    const result = await $fetch<{ imported: number, failed: number, errors?: string[] }>(
      `/api/empresas/${id.value}/import`,
      { method: 'POST', body: form }
    )
    importResult.value = result
    if (result.imported > 0) refresh()
  } catch {
    toast.add({ title: 'Erro ao importar documentos', color: 'error' })
  } finally {
    importing.value = false
  }
}

function closeImport() {
  importOpen.value = false
  importFile.value = null
  importResult.value = null
}

function formatCNPJ(cnpj: string | undefined): string {
  const d = (cnpj || '').replace(/\D/g, '')
  if (d.length !== 14) return cnpj || ''
  return `${d.slice(0, 2)}.${d.slice(2, 5)}.${d.slice(5, 8)}/${d.slice(8, 12)}-${d.slice(12)}`
}

function formatDate(iso: string | undefined): string {
  if (!iso) return '—'
  try {
    return format(parseISO(iso), 'dd/MM/yyyy HH:mm', { locale: ptBR })
  } catch {
    return iso
  }
}

function formatDateOnly(iso: string | undefined): string {
  if (!iso) return '—'
  try {
    return format(parseISO(iso), 'dd/MM/yyyy', { locale: ptBR })
  } catch {
    return iso
  }
}

const certColor = computed(() => {
  const map: Record<string, string> = {
    valido: 'success',
    prestes_a_vencer: 'warning',
    vencido: 'error',
    sem_certificado: 'neutral'
  }
  return map[empresa.value?.certificado_status ?? 'sem_certificado'] ?? 'neutral'
})

const certLabel = computed(() => {
  const map: Record<string, string> = {
    valido: 'Válido',
    prestes_a_vencer: 'Prestes a vencer',
    vencido: 'Vencido',
    sem_certificado: 'Sem certificado'
  }
  return map[empresa.value?.certificado_status ?? 'sem_certificado'] ?? '—'
})

const blockedUntilDate = computed(() => syncState.value?.blocked_until)
const isBlocked = computed(() => {
  if (!blockedUntilDate.value) return false
  return new Date(blockedUntilDate.value) > new Date()
})

// Chart
const chartData = computed(() => porCompetencia.value)
const x = (_: CompetenciaCount, i: number) => i
const xTicks = (i: number) => chartData.value[i]?.competencia ?? ''
const template = (d: CompetenciaCount) => `${d.competencia}: ${d.count} docs`

const statCards = computed(() => [
  { title: 'Total de XMLs', icon: 'i-lucide-file-text', value: stats.value.total, color: 'primary' },
  { title: 'XML Completo', icon: 'i-lucide-file-check', value: stats.value.xml_completo, color: 'success' },
  { title: 'XML Resumo', icon: 'i-lucide-file-warning', value: stats.value.xml_resumo, color: 'warning' },
  { title: 'Manifestados', icon: 'i-lucide-shield-check', value: stats.value.manifestados, color: 'info' }
])

function tipoColor(tipo: string) {
  const m: Record<string, string> = { 'nf-e': 'primary', 'nfc-e': 'success', 'ct-e': 'warning', 'nfs-e': 'neutral' }
  return (m[tipo] ?? 'neutral') as 'primary' | 'success' | 'warning' | 'neutral'
}

function statusColor(s: string) {
  const m: Record<string, string> = { autorizada: 'success', cancelada: 'error', denegada: 'warning' }
  return (m[s] ?? 'neutral') as 'success' | 'error' | 'warning' | 'neutral'
}

const UBadge = resolveComponent('UBadge')
const tableColumns: TableColumn<DocumentoFiscal>[] = [
  {
    accessorKey: 'emitente_nome',
    header: 'Emitente',
    cell: ({ row }) => {
      return h('div', { class: 'max-w-48 truncate' }, [
        h('p', { class: 'truncate font-medium text-highlighted' }, row.original.emitente_nome || '—'),
        h('p', { class: 'text-xs text-muted' }, row.original.emitente_cnpj || '')
      ])
    }
  },
  {
    accessorKey: 'numero_documento',
    header: 'Número',
    cell: ({ row }) => h('span', { class: 'text-muted' }, row.original.numero_documento || '—')
  },
  {
    accessorKey: 'tipo_documento',
    header: 'Tipo',
    cell: ({ row }) => h(UBadge, { color: tipoColor(row.original.tipo_documento), variant: 'subtle', size: 'xs', class: 'uppercase' }, () => row.original.tipo_documento)
  },
  {
    accessorKey: 'status_documento',
    header: 'Status',
    cell: ({ row }) => h(UBadge, { color: statusColor(row.original.status_documento), variant: 'subtle', size: 'xs', class: 'capitalize' }, () => row.original.status_documento)
  },
  {
    accessorKey: 'competencia',
    header: 'Competência',
    cell: ({ row }) => h('span', { class: 'text-muted' }, row.original.competencia || '—')
  },
  {
    id: 'xml_resumo',
    header: 'XML',
    cell: ({ row }) => h(UBadge, { color: row.original.xml_resumo ? 'warning' : 'success', variant: 'subtle', size: 'xs' }, () => row.original.xml_resumo ? 'Resumo' : 'Completo')
  }
]
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
            @click="importOpen = true"
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
        <!-- Empresa header -->
        <div class="flex flex-wrap items-center gap-3 mb-6">
          <UBadge
            :color="empresa.situacao_cadastral === 'Ativa' ? 'success' : 'warning'"
            variant="subtle"
          >
            {{ empresa.situacao_cadastral || 'Desconhecida' }}
          </UBadge>
          <p class="text-sm text-muted font-mono bg-default rounded-md px-2 py-0.5 border border-default">
            {{ formatCNPJ(empresa.cnpj) }}
          </p>
          <UBadge :color="certColor as any" variant="subtle">
            {{ certLabel }}
          </UBadge>
          <UBadge
            v-if="isBlocked"
            color="error"
            variant="subtle"
            icon="i-lucide-lock"
          >
            Bloqueado até {{ formatDate(blockedUntilDate) }}
          </UBadge>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <!-- Left Column: Cert and Company Info -->
          <div class="space-y-6">
            <!-- Certificado e Sync -->
            <UCard>
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-shield-check" class="size-5 text-primary" />
                  <h3 class="font-semibold text-highlighted">
                    Certificado & Sincronização
                  </h3>
                </div>
              </template>

              <div class="space-y-4 text-sm">
                <div class="flex items-center justify-between">
                  <span class="text-muted">Status</span>
                  <UBadge :color="certColor as any" variant="subtle" size="sm">
                    {{ certLabel }}
                  </UBadge>
                </div>

                <div v-if="empresa.certificado_valido_ate" class="flex items-center justify-between">
                  <span class="text-muted">Válido até</span>
                  <span class="font-medium">{{ formatDateOnly(empresa.certificado_valido_ate) }}</span>
                </div>

                <USeparator />

                <div class="grid grid-cols-2 gap-4">
                  <div v-if="syncState">
                    <span class="block text-xs text-muted mb-1">Última sync</span>
                    <span class="font-medium">{{ formatDate(syncState.ultima_sincronizacao) }}</span>
                  </div>
                  <div v-if="syncState">
                    <span class="block text-xs text-muted mb-1">Período</span>
                    <span class="font-medium">{{ syncState.lookback_days }} dias</span>
                  </div>
                  <div v-if="syncState">
                    <span class="block text-xs text-muted mb-1">Último NSU</span>
                    <span class="font-mono text-xs">{{ syncState.ult_nsu || '—' }}</span>
                  </div>
                  <div v-if="syncState?.ultimo_cstat">
                    <span class="block text-xs text-muted mb-1">Último cStat</span>
                    <UBadge
                      :color="syncState.ultimo_cstat === '138' ? 'success' : syncState.ultimo_cstat === '656' ? 'error' : 'neutral'"
                      variant="subtle"
                      size="sm"
                    >
                      {{ syncState.ultimo_cstat }}
                    </UBadge>
                  </div>
                </div>
              </div>
            </UCard>

            <!-- Dados Cadastrais -->
            <UCard>
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-building-2" class="size-5 text-primary" />
                  <h3 class="font-semibold text-highlighted">
                    Dados Cadastrais
                  </h3>
                </div>
              </template>

              <div class="space-y-4 text-sm">
                <div v-if="empresa.nome_fantasia">
                  <span class="block text-xs text-muted mb-1">Nome Fantasia</span>
                  <span class="font-medium">{{ empresa.nome_fantasia }}</span>
                </div>
                <div v-if="empresa.cnae || empresa.natureza_juridica">
                  <span class="block text-xs text-muted mb-1">Natureza / CNAE</span>
                  <span>{{ empresa.natureza_juridica || '—' }} <span v-if="empresa.cnae" class="text-muted">({{ empresa.cnae }})</span></span>
                </div>
                <div v-if="empresa.email || empresa.telefone" class="grid grid-cols-2 gap-4">
                  <div v-if="empresa.email">
                    <span class="block text-xs text-muted mb-1">E-mail</span>
                    <span class="truncate">{{ empresa.email }}</span>
                  </div>
                  <div v-if="empresa.telefone">
                    <span class="block text-xs text-muted mb-1">Telefone</span>
                    <span>{{ empresa.telefone }}</span>
                  </div>
                </div>
                <div v-if="empresa.logradouro || empresa.cidade">
                  <span class="block text-xs text-muted mb-1">Endereço</span>
                  <span class="leading-relaxed text-muted">
                    {{ empresa.logradouro }}{{ empresa.numero ? `, ${empresa.numero}` : '' }}{{ empresa.complemento ? ` - ${empresa.complemento}` : '' }}<br v-if="empresa.logradouro">
                    <span v-if="empresa.bairro">{{ empresa.bairro }} — </span>{{ empresa.cidade }}<span v-if="empresa.estado">/{{ empresa.estado }}</span>
                    <span v-if="empresa.cep"> (CEP {{ empresa.cep }})</span>
                  </span>
                </div>
              </div>
            </UCard>
          </div>

          <!-- Right Column: Stats and Lists -->
          <div class="lg:col-span-2 space-y-6">
            <!-- Stats Cards -->
            <UPageGrid class="grid-cols-2 lg:grid-cols-4">
              <UPageCard
                v-for="stat in statCards"
                :key="stat.title"
                :title="stat.title"
                :description="String(stat.value)"
                :icon="stat.icon"
                variant="soft"
                :ui="{
                  wrapper: 'items-start',
                  title: 'text-xs text-muted uppercase font-medium',
                  description: 'text-2xl font-bold text-highlighted mt-1',
                  leadingIcon: `text-${stat.color} size-8`
                }"
              />
            </UPageGrid>

            <!-- Chart -->
            <UCard v-if="porCompetencia.length > 0">
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-bar-chart-3" class="size-5 text-primary" />
                  <h3 class="font-semibold text-highlighted">
                    Volume de XMLs por Competência
                  </h3>
                </div>
              </template>

              <VisXYContainer
                :data="chartData"
                :padding="{ top: 20, left: 8, right: 8 }"
                class="h-56"
              >
                <VisGroupedBar :x="x" :y="[(d: CompetenciaCount) => d.count]" color="['var(--ui-primary)']" />
                <VisAxis type="x" :x="x" :tick-format="xTicks" />
                <VisCrosshair color="var(--ui-primary)" :template="template" />
                <VisTooltip />
              </VisXYContainer>
            </UCard>

            <!-- Documentos Recentes Table -->
            <UCard :ui="{ body: 'p-0 sm:p-0', header: 'px-4 py-3 sm:px-6' }">
              <template #header>
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <UIcon name="i-lucide-files" class="size-5 text-primary" />
                    <h3 class="font-semibold text-highlighted">
                      Documentos Recentes
                    </h3>
                  </div>
                  <UButton
                    to="/documentos"
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    label="Ver todos"
                    icon="i-lucide-arrow-right"
                    trailing
                  />
                </div>
              </template>

              <UTable
                v-if="recentes.length > 0"
                :data="recentes"
                :columns="tableColumns"
                class="w-full"
              />
              <div v-else class="text-sm text-muted py-8 text-center">
                Nenhum documento encontrado.
              </div>
            </UCard>
          </div>
        </div>
      </template>
    </template>
  </UDashboardPanel>

  <UModal v-model:open="importOpen" title="Importar documentos" :ui="{ footer: 'justify-end' }">
    <template #body>
      <p class="text-sm text-muted mb-4">
        Selecione um arquivo <strong>.xml</strong> (NF-e individual) ou <strong>.zip</strong> com múltiplos XMLs.
      </p>

      <!-- Input nativo oculto, acionado por ref -->
      <input
        ref="fileInputRef"
        type="file"
        accept=".xml,.zip"
        class="hidden"
        @change="onFileChange"
      >

      <!-- Área clicável que aciona o input -->
      <button
        type="button"
        class="w-full flex flex-col items-center justify-center gap-2 border-2 border-dashed rounded-lg p-8 transition-colors"
        :class="importFile ? 'border-primary bg-primary/5' : 'border-default hover:border-primary'"
        @click="triggerFileInput"
      >
        <UIcon
          :name="importFile ? 'i-lucide-file-check' : 'i-lucide-file-up'"
          class="size-8"
          :class="importFile ? 'text-primary' : 'text-muted'"
        />
        <span class="text-sm font-medium" :class="importFile ? 'text-primary' : 'text-muted'">
          {{ importFile ? importFile.name : 'Clique para selecionar um arquivo' }}
        </span>
        <span v-if="importFile" class="text-xs text-muted">
          {{ (importFile.size / 1024).toFixed(0) }} KB · clique para trocar
        </span>
        <span v-else class="text-xs text-dimmed">XML ou ZIP · máx 200 MB</span>
      </button>

      <div v-if="importing" class="space-y-1.5 pt-1">
        <UProgress animation="carousel" />
        <p class="text-xs text-muted text-center">
          Importando...
        </p>
      </div>

      <div v-if="importResult" class="mt-4 rounded-lg p-3 text-sm" :class="importResult.failed > 0 ? 'bg-warning/10' : 'bg-success/10'">
        <p class="font-medium text-highlighted">
          {{ importResult.imported }} documento{{ importResult.imported !== 1 ? 's' : '' }} importado{{ importResult.imported !== 1 ? 's' : '' }}
          <template v-if="importResult.failed > 0">
            · {{ importResult.failed }} falha{{ importResult.failed !== 1 ? 's' : '' }}
          </template>
        </p>
        <ul v-if="importResult.errors?.length" class="mt-1 list-disc list-inside text-xs text-muted space-y-0.5">
          <li v-for="(err, i) in importResult.errors.slice(0, 5)" :key="i">
            {{ err }}
          </li>
          <li v-if="importResult.errors.length > 5">
            ...e mais {{ importResult.errors.length - 5 }} erro(s)
          </li>
        </ul>
      </div>
    </template>

    <template #footer>
      <UButton
        label="Cancelar"
        color="neutral"
        variant="ghost"
        @click="closeImport"
      />
      <UButton
        label="Importar"
        icon="i-lucide-upload"
        :loading="importing"
        :disabled="!importFile || importing"
        @click="handleImport"
      />
    </template>
  </UModal>
</template>

<style scoped>
.unovis-xy-container {
  --vis-axis-grid-color: var(--ui-border);
  --vis-axis-tick-color: var(--ui-border);
  --vis-axis-tick-label-color: var(--ui-text-dimmed);
  --vis-tooltip-background-color: var(--ui-bg);
  --vis-tooltip-border-color: var(--ui-border);
  --vis-tooltip-text-color: var(--ui-text-highlighted);
  --vis-crosshair-line-stroke-color: var(--ui-primary);
}
</style>
