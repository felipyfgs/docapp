<script setup lang="ts">
import { VisXYContainer, VisGroupedBar, VisAxis, VisCrosshair, VisTooltip } from '@unovis/vue'
import { format, parseISO } from 'date-fns'
import { ptBR } from 'date-fns/locale'
import type { EmpresaOverview, CompetenciaCount } from '~/types'

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
  { title: 'Total', icon: 'i-lucide-file-text', value: stats.value.total, color: 'text-primary' },
  { title: 'XML Completo', icon: 'i-lucide-file-check', value: stats.value.xml_completo, color: 'text-success' },
  { title: 'XML Resumo', icon: 'i-lucide-file-warning', value: stats.value.xml_resumo, color: 'text-warning' },
  { title: 'Manifestados', icon: 'i-lucide-shield-check', value: stats.value.manifestados, color: 'text-info' }
])

const accordionItems = [
  { label: 'Dados Cadastrais', icon: 'i-lucide-building-2', slot: 'cadastro' },
  { label: 'Certificado & Sincronização', icon: 'i-lucide-shield', slot: 'certificado' },
  { label: 'Documentos Recentes', icon: 'i-lucide-files', slot: 'documentos' }
]

function tipoColor(tipo: string) {
  const m: Record<string, string> = { 'nf-e': 'primary', 'nfc-e': 'success', 'ct-e': 'warning', 'nfs-e': 'neutral' }
  return (m[tipo] ?? 'neutral') as 'primary' | 'success' | 'warning' | 'neutral'
}

function statusColor(s: string) {
  const m: Record<string, string> = { autorizada: 'success', cancelada: 'error', denegada: 'warning' }
  return (m[s] ?? 'neutral') as 'success' | 'error' | 'warning' | 'neutral'
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
        <div class="flex flex-wrap items-center gap-2 mb-6">
          <p class="text-sm text-muted font-mono">
            {{ formatCNPJ(empresa.cnpj) }}
          </p>
          <UBadge
            :color="empresa.situacao_cadastral === 'Ativa' ? 'success' : 'warning'"
            variant="subtle"
            size="sm"
          >
            {{ empresa.situacao_cadastral || 'Desconhecida' }}
          </UBadge>
          <UBadge :color="certColor as any" variant="subtle" size="sm">
            {{ certLabel }}
          </UBadge>
          <UBadge
            v-if="isBlocked"
            color="error"
            variant="subtle"
            size="sm"
            icon="i-lucide-lock"
          >
            Bloqueado até {{ formatDate(blockedUntilDate) }}
          </UBadge>
        </div>

        <!-- Stats cards -->
        <UPageGrid class="lg:grid-cols-4 gap-4 mb-6">
          <UPageCard
            v-for="stat in statCards"
            :key="stat.title"
            :icon="stat.icon"
            :title="stat.title"
            variant="subtle"
            :ui="{
              container: 'gap-y-1',
              wrapper: 'items-start',
              leading: 'p-2.5 rounded-full bg-primary/10 ring ring-inset ring-primary/25',
              title: 'font-normal text-muted text-xs uppercase'
            }"
          >
            <span class="text-2xl font-semibold text-highlighted">{{ stat.value }}</span>
          </UPageCard>
        </UPageGrid>

        <!-- Grafico por competencia -->
        <UCard
          v-if="porCompetencia.length > 0"
          class="mb-6"
          :ui="{ body: '!px-0 !pb-3 !pt-0' }"
        >
          <template #header>
            <p class="text-xs text-muted uppercase mb-0.5">
              Documentos por competência
            </p>
            <p class="text-2xl font-semibold text-highlighted">
              {{ stats.total }} docs
            </p>
          </template>

          <VisXYContainer
            :data="chartData"
            :padding="{ top: 20, left: 8, right: 8 }"
            class="h-48"
          >
            <VisGroupedBar :x="x" :y="[(d: CompetenciaCount) => d.count]" color="['var(--ui-primary)']" />
            <VisAxis type="x" :x="x" :tick-format="xTicks" />
            <VisCrosshair color="var(--ui-primary)" :template="template" />
            <VisTooltip />
          </VisXYContainer>
        </UCard>

        <!-- Accordions -->
        <!-- eslint-disable vue/no-unused-vars -->
        <UAccordion :items="(accordionItems as any)" type="multiple" :default-value="['cadastro', 'certificado', 'documentos']">
          <template #cadastro>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-x-6 gap-y-3 py-4 text-sm">
              <div>
                <p class="text-xs text-muted uppercase mb-0.5">
                  Razão Social
                </p>
                <p class="font-medium text-highlighted">
                  {{ empresa.razao_social }}
                </p>
              </div>
              <div v-if="empresa.nome_fantasia">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Nome Fantasia
                </p>
                <p>{{ empresa.nome_fantasia }}</p>
              </div>
              <div>
                <p class="text-xs text-muted uppercase mb-0.5">
                  CNPJ
                </p>
                <p class="font-mono">
                  {{ formatCNPJ(empresa.cnpj) }}
                </p>
              </div>
              <div v-if="empresa.cnae">
                <p class="text-xs text-muted uppercase mb-0.5">
                  CNAE
                </p>
                <p>{{ empresa.cnae }}</p>
              </div>
              <div v-if="empresa.porte">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Porte
                </p>
                <p>{{ empresa.porte }}</p>
              </div>
              <div v-if="empresa.natureza_juridica">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Natureza Jurídica
                </p>
                <p>{{ empresa.natureza_juridica }}</p>
              </div>
              <div v-if="empresa.data_inicio_atividade">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Início Atividade
                </p>
                <p>{{ formatDateOnly(empresa.data_inicio_atividade) }}</p>
              </div>
              <div v-if="empresa.email">
                <p class="text-xs text-muted uppercase mb-0.5">
                  E-mail
                </p>
                <p>{{ empresa.email }}</p>
              </div>
              <div v-if="empresa.telefone">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Telefone
                </p>
                <p>{{ empresa.telefone }}</p>
              </div>
              <div v-if="empresa.logradouro" class="col-span-2 md:col-span-3">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Endereço
                </p>
                <p>
                  {{ empresa.logradouro }}{{ empresa.numero ? `, ${empresa.numero}` : '' }}{{ empresa.complemento ? ` - ${empresa.complemento}` : '' }},
                  {{ empresa.bairro }} — {{ empresa.cidade }}/{{ empresa.estado }}, CEP {{ empresa.cep }}
                </p>
              </div>
            </div>
          </template>

          <template #certificado>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-x-6 gap-y-3 py-4 text-sm">
              <div>
                <p class="text-xs text-muted uppercase mb-0.5">
                  Status
                </p>
                <UBadge :color="certColor as any" variant="subtle" size="sm">
                  {{ certLabel }}
                </UBadge>
              </div>
              <div v-if="empresa.certificado_valido_ate">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Válido até
                </p>
                <p>{{ formatDateOnly(empresa.certificado_valido_ate) }}</p>
              </div>
              <div v-if="syncState">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Período de busca
                </p>
                <p>{{ syncState.lookback_days }} dias</p>
              </div>
              <div v-if="syncState">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Último NSU
                </p>
                <p class="font-mono text-xs">
                  {{ syncState.ult_nsu || '—' }}
                </p>
              </div>
              <div v-if="syncState">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Max NSU
                </p>
                <p class="font-mono text-xs">
                  {{ syncState.max_nsu || '—' }}
                </p>
              </div>
              <div v-if="syncState?.ultima_sincronizacao">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Última sync
                </p>
                <p>{{ formatDate(syncState.ultima_sincronizacao) }}</p>
              </div>
              <div v-if="syncState?.ultimo_cstat">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Último cStat
                </p>
                <UBadge
                  :color="syncState.ultimo_cstat === '138' ? 'success' : syncState.ultimo_cstat === '656' ? 'error' : 'neutral'"
                  variant="subtle"
                  size="sm"
                >
                  {{ syncState.ultimo_cstat }} — {{ syncState.ultimo_xmotivo || '—' }}
                </UBadge>
              </div>
              <div v-if="isBlocked">
                <p class="text-xs text-muted uppercase mb-0.5">
                  Bloqueado até
                </p>
                <p class="text-error">
                  {{ formatDate(blockedUntilDate) }}
                </p>
              </div>
            </div>
          </template>

          <template #documentos>
            <div class="py-2">
              <div
                v-if="recentes.length === 0"
                class="text-sm text-muted py-4 text-center"
              >
                Nenhum documento encontrado.
              </div>
              <table v-else class="w-full text-sm">
                <thead>
                  <tr class="text-xs text-muted uppercase border-b border-default">
                    <th class="text-left py-2 pr-4 font-normal">
                      Emitente
                    </th>
                    <th class="text-left py-2 pr-4 font-normal">
                      Número
                    </th>
                    <th class="text-left py-2 pr-4 font-normal">
                      Tipo
                    </th>
                    <th class="text-left py-2 pr-4 font-normal">
                      Status
                    </th>
                    <th class="text-left py-2 pr-4 font-normal">
                      Competência
                    </th>
                    <th class="text-left py-2 font-normal">
                      XML
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="doc in recentes"
                    :key="doc.id"
                    class="border-b border-default last:border-0"
                  >
                    <td class="py-2 pr-4 max-w-48 truncate">
                      <p class="truncate font-medium text-highlighted">
                        {{ doc.emitente_nome || '—' }}
                      </p>
                      <p class="text-xs text-muted">
                        {{ doc.emitente_cnpj || '' }}
                      </p>
                    </td>
                    <td class="py-2 pr-4 text-muted">
                      {{ doc.numero_documento || '—' }}
                    </td>
                    <td class="py-2 pr-4">
                      <UBadge :color="tipoColor(doc.tipo_documento)" variant="subtle" size="xs">
                        {{ doc.tipo_documento.toUpperCase() }}
                      </UBadge>
                    </td>
                    <td class="py-2 pr-4">
                      <UBadge :color="statusColor(doc.status_documento)" variant="subtle" size="xs">
                        {{ doc.status_documento }}
                      </UBadge>
                    </td>
                    <td class="py-2 pr-4 text-muted">
                      {{ doc.competencia || '—' }}
                    </td>
                    <td class="py-2">
                      <UBadge :color="doc.xml_resumo ? 'warning' : 'success'" variant="subtle" size="xs">
                        {{ doc.xml_resumo ? 'Resumo' : 'Completo' }}
                      </UBadge>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </template>
        </UAccordion>
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
