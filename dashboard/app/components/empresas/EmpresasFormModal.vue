<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'
import type { Empresa } from '~/types'

const props = defineProps<{
  empresa?: Empresa | null
}>()

const emit = defineEmits<{ created: [], updated: [], close: [] }>()

const open = ref(false)
const toast = useToast()
const loadingCNPJ = ref(false)
const certFile = ref<File | null>(null)
const certSenha = ref('')

const isEditMode = computed(() => !!props.empresa)
const modalTitle = computed(() => isEditMode.value ? 'Editar Empresa' : 'Nova Empresa')
const submitLabel = computed(() => isEditMode.value ? 'Salvar' : 'Cadastrar')

watch(() => props.empresa, (empresa) => {
  if (empresa) {
    Object.assign(state, {
      cnpj: empresa.cnpj || '',
      razao_social: empresa.razao_social || '',
      nome_fantasia: empresa.nome_fantasia || '',
      situacao_cadastral: empresa.situacao_cadastral || '',
      logradouro: empresa.logradouro || '',
      numero: empresa.numero || '',
      complemento: empresa.complemento || '',
      bairro: empresa.bairro || '',
      cep: empresa.cep || '',
      cidade: empresa.cidade || '',
      estado: empresa.estado || '',
      telefone: empresa.telefone || '',
      email: empresa.email || '',
      cnae: empresa.cnae || '',
      porte: empresa.porte || '',
      natureza_juridica: empresa.natureza_juridica || '',
      data_inicio_atividade: empresa.data_inicio_atividade || '',
      lookback_days: empresa.lookback_days || 90
    })
    open.value = true
  }
})

watch(open, (isOpen) => {
  if (!isOpen) {
    resetForm()
    emit('close')
  }
})

function onCertFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  certFile.value = input.files?.[0] ?? null
}

const schema = z.object({
  cnpj: z.string().regex(/^\d{14}$/, 'CNPJ deve conter 14 dígitos'),
  razao_social: z.string().min(2, 'Obrigatório'),
  nome_fantasia: z.string().optional(),
  situacao_cadastral: z.string().optional(),
  logradouro: z.string().optional(),
  numero: z.string().optional(),
  complemento: z.string().optional(),
  bairro: z.string().optional(),
  cep: z.string().optional(),
  cidade: z.string().optional(),
  estado: z.string().regex(/^[A-Z]{2}$/, 'UF deve conter 2 letras'),
  telefone: z.string().optional(),
  email: z.email('Email inválido').optional().or(z.literal('')),
  cnae: z.string().optional(),
  porte: z.string().optional(),
  natureza_juridica: z.string().optional(),
  data_inicio_atividade: z.string().optional(),
  lookback_days: z.number().int().min(1).max(365).default(90)
})

type Schema = z.output<typeof schema>

type CnpjLookupResponse = {
  razao_social?: string
  porte?: { descricao?: string }
  natureza_juridica?: { descricao?: string }
  estabelecimento?: {
    nome_fantasia?: string
    situacao_cadastral?: string
    tipo_logradouro?: string
    logradouro?: string
    numero?: string
    complemento?: string
    bairro?: string
    cep?: string
    cidade?: { nome?: string }
    estado?: { sigla?: string }
    ddd1?: string
    telefone1?: string
    email?: string
    atividade_principal?: { id?: string }
    data_inicio_atividade?: string
  }
}

const state = reactive<Partial<Schema>>({
  cnpj: '',
  razao_social: '',
  nome_fantasia: '',
  situacao_cadastral: '',
  logradouro: '',
  numero: '',
  complemento: '',
  bairro: '',
  cep: '',
  cidade: '',
  estado: '',
  telefone: '',
  email: '',
  cnae: '',
  porte: '',
  natureza_juridica: '',
  data_inicio_atividade: '',
  lookback_days: 90
})

const cnpjFormatted = computed({
  get: () => {
    const d = (state.cnpj || '').replace(/\D/g, '')
    if (d.length <= 2) return d
    if (d.length <= 5) return `${d.slice(0, 2)}.${d.slice(2)}`
    if (d.length <= 8) return `${d.slice(0, 2)}.${d.slice(2, 5)}.${d.slice(5)}`
    if (d.length <= 12) return `${d.slice(0, 2)}.${d.slice(2, 5)}.${d.slice(5, 8)}/${d.slice(8)}`
    return `${d.slice(0, 2)}.${d.slice(2, 5)}.${d.slice(5, 8)}/${d.slice(8, 12)}-${d.slice(12, 14)}`
  },
  set: (val: string) => {
    const digits = val.replace(/\D/g, '').slice(0, 14)
    state.cnpj = digits
  }
})

const estadoFormatted = computed({
  get: () => (state.estado || '').toUpperCase(),
  set: (val: string) => {
    state.estado = val.replace(/[^a-zA-Z]/g, '').slice(0, 2).toUpperCase()
  }
})

async function buscarCNPJ() {
  const cnpj = state.cnpj
  if (!cnpj || cnpj.length !== 14) {
    toast.add({ title: 'Digite um CNPJ válido com 14 dígitos', color: 'warning' })
    return
  }

  loadingCNPJ.value = true
  try {
    const data = await $fetch<CnpjLookupResponse>(`/api/cnpj/${cnpj}`)
    const est = data.estabelecimento ?? {}
    state.razao_social = data.razao_social ?? ''
    state.nome_fantasia = est.nome_fantasia ?? ''
    state.situacao_cadastral = est.situacao_cadastral ?? ''
    state.logradouro = `${est.tipo_logradouro ?? ''} ${est.logradouro ?? ''}`.trim()
    state.numero = est.numero ?? ''
    state.complemento = est.complemento ?? ''
    state.bairro = est.bairro ?? ''
    state.cep = est.cep ?? ''
    state.cidade = est.cidade?.nome ?? ''
    state.estado = est.estado?.sigla ?? ''
    state.telefone = est.ddd1 && est.telefone1 ? `(${est.ddd1}) ${est.telefone1}` : ''
    state.email = est.email ?? ''
    state.cnae = est.atividade_principal?.id ?? ''
    state.porte = data.porte?.descricao ?? ''
    state.natureza_juridica = data.natureza_juridica?.descricao ?? ''
    state.data_inicio_atividade = est.data_inicio_atividade ?? ''
    toast.add({ title: 'Dados preenchidos com sucesso', color: 'success' })
  } catch {
    toast.add({ title: 'CNPJ não encontrado', color: 'warning' })
  } finally {
    loadingCNPJ.value = false
  }
}

async function onSubmit(event: FormSubmitEvent<Schema>) {
  try {
    let empresaId: number | undefined

    if (isEditMode.value && props.empresa) {
      await $fetch(`/api/empresas/${props.empresa.id}`, { method: 'PUT', body: event.data })
      empresaId = props.empresa.id
      toast.add({ title: 'Empresa atualizada com sucesso', color: 'success' })
      emit('updated')
    } else {
      const empresa: { id: number } = await $fetch('/api/empresas', { method: 'POST', body: event.data })
      empresaId = empresa.id
      toast.add({ title: 'Empresa cadastrada com sucesso', color: 'success' })
      emit('created')
    }

    if (certFile.value && empresaId) {
      const fd = new FormData()
      fd.append('certificado', certFile.value)
      fd.append('senha', certSenha.value)
      const siglaUF = String(state.estado || '').trim().toUpperCase()
      if (siglaUF) {
        fd.append('sigla_uf', siglaUF)
      }
      await $fetch(`/api/empresas/${empresaId}/certificado`, { method: 'POST', body: fd })
    }

    open.value = false
    resetForm()
  } catch (e: unknown) {
    const msg = (e as { data?: { message?: string } })?.data?.message ?? 'Erro ao salvar empresa'
    toast.add({ title: msg, color: 'error' })
  }
}

function resetForm() {
  Object.assign(state, {
    cnpj: '', razao_social: '', nome_fantasia: '', situacao_cadastral: '',
    logradouro: '', numero: '', complemento: '', bairro: '', cep: '',
    cidade: '', estado: '', telefone: '', email: '', cnae: '', porte: '',
    natureza_juridica: '', data_inicio_atividade: '', lookback_days: 90
  })
  certFile.value = null
  certSenha.value = ''
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="modalTitle"
    :description="isEditMode ? 'Edite os dados da empresa' : 'Preencha o CNPJ para preenchimento automático'"
    :dismissible="false"
    :ui="{ content: 'sm:max-w-2xl h-[90vh]', body: 'overflow-y-auto flex-1 min-h-0' }"
  >
    <UButton v-if="!isEditMode" label="Nova Empresa" icon="i-lucide-plus" />

    <template #body>
      <UForm
        id="empresa-form"
        :schema="schema"
        :state="state"
        class="space-y-4"
        @submit="onSubmit"
      >
        <!-- CNPJ com auto-fill -->
        <UFormField label="CNPJ" name="cnpj" required>
          <div class="flex items-center gap-2 w-full">
            <UInput
              v-model="cnpjFormatted"
              placeholder="00.000.000/0000-00"
              class="flex-1"
              maxlength="18"
              @keydown.enter.prevent="buscarCNPJ"
            />
            <UButton
              icon="i-lucide-search"
              color="neutral"
              variant="outline"
              :loading="loadingCNPJ"
              :disabled="(state.cnpj?.length ?? 0) < 14"
              @click="buscarCNPJ"
            />
          </div>
        </UFormField>

        <!-- Secoes organizadas em tabs -->
        <UTabs
          :unmount-on-hide="false"
          color="neutral"
          variant="link"
          :items="[
            { label: 'Identificação', icon: 'i-lucide-building-2', slot: 'identificacao' },
            { label: 'Endereço', icon: 'i-lucide-map-pin', slot: 'endereco' },
            { label: 'Contato', icon: 'i-lucide-phone', slot: 'contato' },
            { label: 'Fiscal', icon: 'i-lucide-file-text', slot: 'fiscal' },
            { label: 'Configuração', icon: 'i-lucide-settings-2', slot: 'configuracao' }
          ]"
        >
          <template #identificacao>
            <div class="grid grid-cols-2 gap-4 pt-4">
              <UFormField
                label="Razão Social"
                name="razao_social"
                required
                class="col-span-2"
              >
                <UInput v-model="state.razao_social" class="w-full" />
              </UFormField>
              <UFormField label="Nome Fantasia" name="nome_fantasia" class="col-span-2">
                <UInput v-model="state.nome_fantasia" class="w-full" />
              </UFormField>
              <UFormField label="Situação Cadastral" name="situacao_cadastral">
                <UInput v-model="state.situacao_cadastral" class="w-full" />
              </UFormField>
              <UFormField label="Início de Atividade" name="data_inicio_atividade">
                <UInput v-model="state.data_inicio_atividade" class="w-full" />
              </UFormField>
            </div>
          </template>

          <template #endereco>
            <div class="grid grid-cols-2 gap-4 pt-4">
              <UFormField label="Logradouro" name="logradouro" class="col-span-2">
                <UInput v-model="state.logradouro" class="w-full" />
              </UFormField>
              <UFormField label="Número" name="numero">
                <UInput v-model="state.numero" class="w-full" />
              </UFormField>
              <UFormField label="Complemento" name="complemento">
                <UInput v-model="state.complemento" class="w-full" />
              </UFormField>
              <UFormField label="Bairro" name="bairro">
                <UInput v-model="state.bairro" class="w-full" />
              </UFormField>
              <UFormField label="CEP" name="cep">
                <UInput v-model="state.cep" class="w-full" />
              </UFormField>
              <UFormField label="Cidade" name="cidade">
                <UInput v-model="state.cidade" class="w-full" />
              </UFormField>
              <UFormField label="UF" name="estado" required>
                <UInput v-model="estadoFormatted" maxlength="2" class="w-full" />
              </UFormField>
            </div>
          </template>

          <template #contato>
            <div class="grid grid-cols-2 gap-4 pt-4">
              <UFormField label="Telefone" name="telefone">
                <UInput v-model="state.telefone" class="w-full" />
              </UFormField>
              <UFormField label="Email" name="email">
                <UInput v-model="state.email" type="email" class="w-full" />
              </UFormField>
            </div>
          </template>

          <template #fiscal>
            <div class="grid grid-cols-2 gap-4 pt-4">
              <UFormField label="CNAE Principal" name="cnae">
                <UInput v-model="state.cnae" class="w-full" />
              </UFormField>
              <UFormField label="Porte" name="porte">
                <UInput v-model="state.porte" class="w-full" />
              </UFormField>
              <UFormField label="Natureza Jurídica" name="natureza_juridica" class="col-span-2">
                <UInput v-model="state.natureza_juridica" class="w-full" />
              </UFormField>
            </div>
          </template>

          <template #configuracao>
            <div class="flex flex-col gap-4 pt-4">
              <UFormField
                label="Lookback (dias)"
                name="lookback_days"
                description="Janela de busca retroativa de documentos fiscais"
              >
                <UInputNumber
                  v-model="state.lookback_days"
                  :min="1"
                  :max="365"
                  class="w-full"
                />
              </UFormField>

              <div class="border border-default rounded-lg p-4 flex flex-col gap-3">
                <p class="text-sm font-medium text-highlighted flex items-center gap-1.5">
                  <UIcon name="i-lucide-shield-check" class="size-4" />
                  Certificado Digital (A1)
                </p>
                <div class="flex items-center gap-2">
                  <label
                    class="flex-1 flex items-center gap-2 px-3 py-2 rounded-md border border-default bg-default text-sm text-muted cursor-pointer hover:bg-elevated transition-colors"
                  >
                    <UIcon name="i-lucide-upload" class="size-4 shrink-0" />
                    <span class="truncate">{{ certFile ? certFile.name : 'Selecionar arquivo .pfx' }}</span>
                    <input
                      type="file"
                      accept=".pfx,.p12"
                      class="sr-only"
                      @change="onCertFileChange"
                    >
                  </label>
                  <UButton
                    v-if="certFile"
                    icon="i-lucide-x"
                    color="neutral"
                    variant="ghost"
                    size="sm"
                    @click="certFile = null"
                  />
                </div>
                <UInput
                  v-if="certFile"
                  v-model="certSenha"
                  type="password"
                  placeholder="Senha do certificado"
                  class="w-full"
                />
              </div>
            </div>
          </template>
        </UTabs>
      </UForm>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton
          label="Cancelar"
          color="neutral"
          variant="subtle"
          @click="open = false; resetForm()"
        />
        <UButton
          form="empresa-form"
          :label="submitLabel"
          color="primary"
          type="submit"
        />
      </div>
    </template>
  </UModal>
</template>
