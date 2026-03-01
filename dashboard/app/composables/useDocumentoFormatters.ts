import { format, parseISO } from 'date-fns'
import { ptBR } from 'date-fns/locale'
import type { DocumentoFiscal } from '~/types'

export function useDocumentoFormatters() {
  function formatBRL(value: number): string {
    return value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
  }

  function formatCNPJ(cnpj: string | undefined): string {
    const digits = (cnpj || '').replace(/\D/g, '')
    if (digits.length === 11) {
      return `${digits.slice(0, 3)}.${digits.slice(3, 6)}.${digits.slice(6, 9)}-${digits.slice(9)}`
    }
    if (digits.length === 14) {
      return `${digits.slice(0, 2)}.${digits.slice(2, 5)}.${digits.slice(5, 8)}/${digits.slice(8, 12)}-${digits.slice(12)}`
    }
    return cnpj || '—'
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

  function tipoBadge(documento: DocumentoFiscal) {
    const map = {
      'nf-e': { label: 'NF-e', color: 'primary' },
      'nfc-e': { label: 'NFC-e', color: 'success' },
      'ct-e': { label: 'CT-e', color: 'warning' },
      'nfs-e': { label: 'NFS-e', color: 'neutral' },
      'desconhecido': { label: 'Desconhecido', color: 'neutral' }
    } as const
    return map[documento.tipo_documento as keyof typeof map] ?? map.desconhecido
  }

  function statusBadge(documento: DocumentoFiscal) {
    const map = {
      autorizada: { label: 'Autorizada', color: 'success' },
      cancelada: { label: 'Cancelada', color: 'error' },
      denegada: { label: 'Denegada', color: 'warning' },
      desconhecido: { label: 'Desconhecido', color: 'neutral' }
    } as const
    return map[documento.status_documento as keyof typeof map] ?? map.desconhecido
  }

  function manifestacaoBadge(status: string | undefined) {
    const map = {
      ciencia: { label: 'Ciência', color: 'info' },
      confirmada: { label: 'Confirmada', color: 'success' },
      desconhecida: { label: 'Desconhecida', color: 'error' },
      nao_realizada: { label: 'Não Realizada', color: 'warning' }
    } as const
    if (!status) return { label: 'Pendente', color: 'neutral' as const }
    return map[status as keyof typeof map] ?? { label: status, color: 'neutral' as const }
  }

  function certBadge(status: string | undefined) {
    const key = status ?? 'sem_certificado'
    const colorMap: Record<string, string> = {
      valido: 'success',
      prestes_a_vencer: 'warning',
      vencido: 'error',
      sem_certificado: 'neutral'
    }
    const labelMap: Record<string, string> = {
      valido: 'Válido',
      prestes_a_vencer: 'Prestes a vencer',
      vencido: 'Vencido',
      sem_certificado: 'Sem certificado'
    }
    return {
      color: colorMap[key] ?? 'neutral',
      label: labelMap[key] ?? '—'
    }
  }

  return { formatBRL, formatCNPJ, formatDate, formatDateOnly, tipoBadge, statusBadge, manifestacaoBadge, certBadge }
}
