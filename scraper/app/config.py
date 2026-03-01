import os
import platform

IS_LINUX = platform.system() == "Linux"

DASN_URL = "https://www8.receita.fazenda.gov.br/SimplesNacional/Aplicacoes/ATSPO/dasnsimei.app"
OPTANTES_URL = "https://consopt.www8.receita.fazenda.gov.br/consultaoptantes"
CNPJ_URL = "https://solucoes.receita.fazenda.gov.br/servicos/cnpjreva/Cnpjreva_Solicitacao.asp"
CNPJ_HCAPTCHA_SITEKEY = "af4fc5a3-1ac5-4e6d-819d-324d412a5e9d"

NOPECHA_API_KEY = os.environ.get("NOPECHA_API_KEY", "")
NOPECHA_API_URL = "https://api.nopecha.com/v1"

CONSULTAR_IO_TOKEN = os.environ.get("CONSULTAR_IO_TOKEN", "")
CONSULTAR_IO_BASE_URL = "https://consultar.io/api/v2"

UFS_VALIDAS = {
    "AC", "AL", "AM", "AP", "BA", "CE", "DF", "ES", "GO", "MA", "MG", "MS",
    "MT", "PA", "PB", "PE", "PI", "PR", "RJ", "RN", "RO", "RR", "RS", "SC",
    "SE", "SP", "TO",
}
