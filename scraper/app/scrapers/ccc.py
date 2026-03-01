import requests
from ..config import CONSULTAR_IO_TOKEN, CONSULTAR_IO_BASE_URL


def _consultar_ie_por_uf(cnpj: str, uf: str) -> list[dict]:
    resp = requests.get(
        f"{CONSULTAR_IO_BASE_URL}/ie/consultar",
        params={"uf": uf, "cnpj": cnpj},
        headers={"Authorization": f"Token {CONSULTAR_IO_TOKEN}"},
        timeout=50,
    )
    if resp.status_code == 404:
        return []
    if resp.status_code != 200:
        data = resp.json() if resp.headers.get("content-type", "").startswith("application/json") else {}
        erro = data.get("error", "")
        msg = data.get("message", resp.text[:300])
        raise RuntimeError(f"Consultar.IO erro ({resp.status_code}): {erro} - {msg}")
    return resp.json()


def _consultar_ie_todas_ufs(cnpj: str) -> list[dict]:
    resp = requests.get(
        f"{CONSULTAR_IO_BASE_URL}/ie/consultar/todas",
        params={"cnpj": cnpj},
        headers={"Authorization": f"Token {CONSULTAR_IO_TOKEN}"},
        timeout=120,
    )
    if resp.status_code != 200:
        data = resp.json() if resp.headers.get("content-type", "").startswith("application/json") else {}
        erro = data.get("error", "")
        msg = data.get("message", resp.text[:300])
        raise RuntimeError(f"Consultar.IO erro ({resp.status_code}): {erro} - {msg}")
    return resp.json()


def _mapear_contribuinte(item: dict) -> dict:
    return {
        "cnpj": item.get("cnpj", ""),
        "razao_social": item.get("razao_social", ""),
        "nome_fantasia": item.get("nome_fantasia", ""),
        "uf": item.get("uf_ie", ""),
        "ie": item.get("ie", ""),
        "tipo_ie": item.get("tipo_ie", ""),
        "situacao_ie": item.get("situacao_ie", ""),
        "situacao_cnpj": item.get("situacao_cnpj", ""),
        "data_situacao_uf": item.get("data_situacao_uf", ""),
        "data_inicio": item.get("data_inicio", ""),
        "data_fim": item.get("data_fim", ""),
        "cnae_principal": item.get("cnae_principal_codigo", ""),
        "regime_tributacao": item.get("regime_tributacao", ""),
        "ie_destinatario": item.get("ie_destinatario", ""),
        "tipo_produtor": item.get("tipo_produtor", ""),
        "endereco": {
            "logradouro": item.get("logradouro", ""),
            "numero": item.get("numero", ""),
            "complemento": item.get("complemento", ""),
            "bairro": item.get("bairro", ""),
            "cep": item.get("cep", ""),
            "uf": item.get("uf", ""),
            "municipio_codigo": item.get("municipio_codigo", ""),
            "municipio": item.get("municipio_descricao", ""),
        },
    }


async def consultar_ccc(cnpj: str, uf: str) -> dict:
    if not CONSULTAR_IO_TOKEN:
        raise RuntimeError("CONSULTAR_IO_TOKEN nao configurado")

    resultados = _consultar_ie_por_uf(cnpj, uf.upper())
    contribuintes = [_mapear_contribuinte(r) for r in resultados]

    ie_principal = contribuintes[0]["ie"] if contribuintes else None
    sit_ie = contribuintes[0]["situacao_ie"] if contribuintes else None

    return {
        "cnpj": cnpj,
        "uf": uf.upper(),
        "encontrado": len(contribuintes) > 0,
        "inscricao_estadual": ie_principal,
        "situacao_ie": sit_ie,
        "contribuintes": contribuintes,
    }


async def consultar_ccc_todas_ufs(cnpj: str) -> dict:
    if not CONSULTAR_IO_TOKEN:
        raise RuntimeError("CONSULTAR_IO_TOKEN nao configurado")

    resultados = _consultar_ie_todas_ufs(cnpj)
    ufs_encontradas = []

    for uf_result in resultados:
        if uf_result.get("status_code") == 200 and uf_result.get("results"):
            for item in uf_result["results"]:
                ufs_encontradas.append(_mapear_contribuinte(item))

    return {
        "cnpj": cnpj,
        "encontrado": len(ufs_encontradas) > 0,
        "total_ies": len(ufs_encontradas),
        "contribuintes": ufs_encontradas,
    }
