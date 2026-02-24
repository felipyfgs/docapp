import type { AvatarProps } from '@nuxt/ui'

export type UserStatus = 'subscribed' | 'unsubscribed' | 'bounced'
export type SaleStatus = 'paid' | 'failed' | 'refunded'

export interface User {
  id: number
  name: string
  email: string
  avatar?: AvatarProps
  status: UserStatus
  location: string
}

export interface Mail {
  id: number
  unread?: boolean
  from: User
  subject: string
  body: string
  date: string
}

export interface Member {
  name: string
  username: string
  role: 'member' | 'owner'
  avatar: AvatarProps
}

export interface Stat {
  title: string
  icon: string
  value: number | string
  variation: number
  formatter?: (value: number) => string
}

export interface Sale {
  id: string
  date: string
  status: SaleStatus
  email: string
  amount: number
}

export interface Notification {
  id: number
  unread?: boolean
  sender: User
  body: string
  date: string
}

export type Period = 'daily' | 'weekly' | 'monthly'

export interface Range {
  start: Date
  end: Date
}

export interface Empresa {
  id: number
  cnpj: string
  razao_social: string
  nome_fantasia: string
  situacao_cadastral: string
  logradouro: string
  numero: string
  complemento?: string
  bairro: string
  cep: string
  cidade: string
  estado: string
  telefone?: string
  email?: string
  cnae?: string
  porte?: string
  natureza_juridica?: string
  data_inicio_atividade?: string
  lookback_days: number
  ativo: boolean
  certificado_caminho?: string
  certificado_senha?: string
  certificado_valido_ate?: string
  certificado_status?: 'vencido' | 'prestes_a_vencer' | 'valido' | 'sem_certificado'
  created_at: string
}

export type DocumentoTipo = 'nf-e' | 'nfc-e' | 'ct-e' | 'nfs-e' | 'desconhecido'
export type DocumentoStatus = 'autorizada' | 'cancelada' | 'denegada' | 'desconhecido'

export interface DocumentoFiscal {
  id: number
  empresa_id: number
  empresa?: Empresa
  nsu: string
  chave_acesso: string
  tipo_documento: DocumentoTipo
  status_documento: DocumentoStatus
  numero_documento: string
  emitente_nome: string
  emitente_cnpj: string
  destinatario_nome: string
  destinatario_cnpj: string
  competencia: string
  schema: string
  xml_object_key?: string
  xml_resumo: boolean
  danfe_object_key?: string
  danfe_generated_at?: string
  data_emissao?: string
  manifestacao_status?: string
  manifestacao_at?: string
  created_at: string
  valor_total?: number
  valor_produtos?: number
}

export interface DocumentoListResponse {
  items: DocumentoFiscal[]
  total: number
  page: number
  page_size: number
}

export interface DocumentoExportResponse {
  file_name: string
  mode: 'proxy' | 'presigned'
  presigned_url?: string
  total: number
  xml_count: number
  danfe_count: number
  skipped_danfe: number
}

export interface DocumentoXMLResponse {
  id: number
  xml: string
  xml_resumo: boolean
  chave_acesso: string
}

export interface EmpresaSyncState {
  ult_nsu: string
  max_nsu: string
  ultima_sincronizacao?: string
  blocked_until?: string
  ultimo_cstat: string
  ultimo_xmotivo: string
  ativo: boolean
  lookback_days: number
}

export interface EmpresaDocumentoStats {
  total: number
  xml_completo: number
  xml_resumo: number
  manifestados: number
  valor_total: number
}

export interface CompetenciaCount {
  competencia: string
  count: number
  valor_total: number
}

export interface EmpresaOverview {
  empresa: Empresa
  sync_state?: EmpresaSyncState
  stats: EmpresaDocumentoStats
  documentos_por_competencia: CompetenciaCount[]
  documentos_recentes: DocumentoFiscal[]
}
