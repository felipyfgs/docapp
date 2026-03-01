from pydantic import BaseModel, field_validator


class ConsultaRequest(BaseModel):
    cnpj: str

    @field_validator("cnpj")
    @classmethod
    def validar_cnpj(cls, v: str) -> str:
        limpo = v.replace(".", "").replace("/", "").replace("-", "")
        if len(limpo) != 14 or not limpo.isdigit():
            raise ValueError("CNPJ invalido")
        return v


UFS_VALIDAS = {
    "AC", "AL", "AM", "AP", "BA", "CE", "DF", "ES", "GO", "MA", "MG", "MS",
    "MT", "PA", "PB", "PE", "PI", "PR", "RJ", "RN", "RO", "RR", "RS", "SC",
    "SE", "SP", "TO",
}


class CccRequest(BaseModel):
    cnpj: str
    uf: str

    @field_validator("cnpj")
    @classmethod
    def validar_cnpj(cls, v: str) -> str:
        limpo = v.replace(".", "").replace("/", "").replace("-", "")
        if len(limpo) != 14 or not limpo.isdigit():
            raise ValueError("CNPJ invalido")
        return v

    @field_validator("uf")
    @classmethod
    def validar_uf(cls, v: str) -> str:
        if v.upper().strip() not in UFS_VALIDAS:
            raise ValueError(f"UF invalida: {v}")
        return v.upper().strip()


# --- CNPJ ---

class Endereco(BaseModel):
    logradouro: str | None = None
    numero: str | None = None
    complemento: str | None = None
    bairro: str | None = None
    municipio: str | None = None
    uf: str | None = None
    cep: str | None = None


class Contato(BaseModel):
    email: str | None = None
    telefone: str | None = None


class CadastroResponse(BaseModel):
    nome_empresarial: str | None = None
    nome_fantasia: str | None = None
    data_abertura: str | None = None
    porte: str | None = None
    atividade_principal: str | None = None
    atividades_secundarias: str | None = None
    natureza_juridica: str | None = None
    endereco: Endereco
    contato: Contato
    situacao_cadastral: str | None = None
    data_situacao_cadastral: str | None = None
    motivo_situacao_cadastral: str | None = None
    situacao_especial: str | None = None
    data_situacao_especial: str | None = None


# --- DASN ---

class Declaracao(BaseModel):
    ano: str
    tipo_declaracao: str | None = None
    status: str | None = None
    pendente: bool = False
    baixada: bool = False
    situacao_especial: str | None = None
    data_baixa: str | None = None


class DasnResponse(BaseModel):
    razao_social: str | None = None
    mei_baixada: bool = False
    data_baixa: str | None = None
    declaracoes: list[Declaracao] = []
    pendentes: list[str] = []


# --- Optantes ---

class Periodo(BaseModel):
    data_inicial: str
    data_final: str
    detalhamento: str | None = None


class RegimeInfo(BaseModel):
    situacao: str | None = None
    periodos_anteriores: list[Periodo] = []
    eventos_futuros: str | None = None


class OptantesResponse(BaseModel):
    simples_nacional: RegimeInfo
    simei: RegimeInfo


# --- Master ---

class CccEndereco(BaseModel):
    logradouro: str | None = None
    numero: str | None = None
    complemento: str | None = None
    bairro: str | None = None
    cep: str | None = None
    uf: str | None = None
    municipio_codigo: str | None = None
    municipio: str | None = None


class CccContribuinte(BaseModel):
    cnpj: str | None = None
    razao_social: str | None = None
    nome_fantasia: str | None = None
    uf: str | None = None
    ie: str | None = None
    tipo_ie: str | None = None
    situacao_ie: str | None = None
    situacao_cnpj: str | None = None
    data_situacao_uf: str | None = None
    data_inicio: str | None = None
    data_fim: str | None = None
    cnae_principal: str | None = None
    regime_tributacao: str | None = None
    ie_destinatario: str | None = None
    tipo_produtor: str | None = None
    endereco: CccEndereco | None = None


class CccResponse(BaseModel):
    cnpj: str
    uf: str
    encontrado: bool = False
    inscricao_estadual: str | None = None
    situacao_ie: str | None = None
    contribuintes: list[CccContribuinte] = []


class CccTodasUfsResponse(BaseModel):
    cnpj: str
    encontrado: bool = False
    total_ies: int = 0
    contribuintes: list[CccContribuinte] = []


# --- IE (SINTEGRA) ---

class IeEndereco(BaseModel):
    logradouro: str | None = None
    numero: str | None = None
    complemento: str | None = None
    bairro: str | None = None
    municipio: str | None = None
    uf: str | None = None
    cep: str | None = None


class IeObrigacoes(BaseModel):
    nfe: str | None = None
    cte: str | None = None
    edf: str | None = None


class IeResponse(BaseModel):
    uf: str
    cnpj: str
    encontrado: bool = False
    ie: str | None = None
    razao_social: str | None = None
    situacao: str | None = None
    data_situacao: str | None = None
    regime: str | None = None
    cnae_principal: str | None = None
    endereco: IeEndereco | None = None
    obrigacoes: IeObrigacoes | None = None


class IeRequest(BaseModel):
    cnpj: str
    uf: str

    @field_validator("cnpj")
    @classmethod
    def validar_cnpj(cls, v: str) -> str:
        limpo = v.replace(".", "").replace("/", "").replace("-", "")
        if len(limpo) != 14 or not limpo.isdigit():
            raise ValueError("CNPJ invalido")
        return v

    @field_validator("uf")
    @classmethod
    def validar_uf(cls, v: str) -> str:
        if v.upper().strip() not in UFS_VALIDAS:
            raise ValueError(f"UF invalida: {v}")
        return v.upper().strip()


class ConsultaResponse(BaseModel):
    cnpj: str
    cadastro: CadastroResponse
    dasn: DasnResponse
    optantes: OptantesResponse
