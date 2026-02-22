<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

const emit = defineEmits<{ created: [] }>()

const open = ref(false)
const toast = useToast()
const loadingCNPJ = ref(false)

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
  estado: z.string().optional(),
  telefone: z.string().optional(),
  email: z.email('Email inválido').optional().or(z.literal('')),
  cnae: z.string().optional(),
  porte: z.string().optional(),
  natureza_juridica: z.string().optional(),
  data_inicio_atividade: z.string().optional(),
  lookback_days: z.number().int().min(1).max(365).default(90)
})

type Schema = z.output<typeof schema>

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

async function buscarCNPJ() {
  const cnpj = state.cnpj
  if (!cnpj || cnpj.length !== 14) {
    toast.add({ title: 'Digite um CNPJ válido com 14 dígitos', color: 'warning' })
    return
  }

  loadingCNPJ.value = true
  try {
    const data: any = await $fetch(`/api/cnpj/${cnpj}`)
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
    await $fetch('/api/empresas', { method: 'POST', body: event.data })
    toast.add({ title: 'Empresa cadastrada com sucesso', color: 'success' })
    open.value = false
    emit('created')
    Object.assign(state, {
      cnpj: '', razao_social: '', nome_fantasia: '', lookback_days: 90
    })
  } catch {
    toast.add({ title: 'Erro ao cadastrar empresa', color: 'error' })
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="Nova Empresa"
    description="Preencha o CNPJ para preenchimento automático"
    :dismissible="false"
    :ui="{ content: 'sm:max-w-2xl h-[90vh]', body: 'overflow-y-auto flex-1 min-h-0' }"
  >
    <UButton label="Nova Empresa" icon="i-lucide-plus" />

    <template #body>
      <UForm id="empresa-form" :schema="schema" :state="state" class="space-y-4" @submit="onSubmit">
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

        <!-- Secoes organizadas em accordion -->
        <UAccordion
          type="multiple"
          :unmount-on-hide="false"
          :default-value="['identificacao']"
          :items="[
            { label: 'Identificação', value: 'identificacao', icon: 'i-lucide-building-2', slot: 'identificacao' },
            { label: 'Endereço', value: 'endereco', icon: 'i-lucide-map-pin', slot: 'endereco' },
            { label: 'Contato', value: 'contato', icon: 'i-lucide-phone', slot: 'contato' },
            { label: 'Dados Fiscais', value: 'fiscal', icon: 'i-lucide-file-text', slot: 'fiscal' },
            { label: 'Configuração', value: 'configuracao', icon: 'i-lucide-settings-2', slot: 'configuracao' }
          ]"
        >
          <template #identificacao>
            <div class="grid grid-cols-2 gap-4 px-1 pb-3">
              <UFormField label="Razão Social" name="razao_social" required class="col-span-2">
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
            <div class="grid grid-cols-2 gap-4 px-1 pb-3">
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
              <UFormField label="UF" name="estado">
                <UInput v-model="state.estado" maxlength="2" class="w-full" />
              </UFormField>
            </div>
          </template>

          <template #contato>
            <div class="grid grid-cols-2 gap-4 px-1 pb-3">
              <UFormField label="Telefone" name="telefone">
                <UInput v-model="state.telefone" class="w-full" />
              </UFormField>
              <UFormField label="Email" name="email">
                <UInput v-model="state.email" type="email" class="w-full" />
              </UFormField>
            </div>
          </template>

          <template #fiscal>
            <div class="grid grid-cols-2 gap-4 px-1 pb-3">
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
            <div class="px-1 pb-3">
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
            </div>
          </template>
        </UAccordion>

      </UForm>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton label="Cancelar" color="neutral" variant="subtle" @click="open = false" />
        <UButton form="empresa-form" label="Cadastrar" color="primary" type="submit" />
      </div>
    </template>
  </UModal>
</template>
