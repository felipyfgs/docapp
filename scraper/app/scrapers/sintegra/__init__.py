from .ma import consultar_ie_ma

SCRAPERS_POR_UF = {
    "MA": consultar_ie_ma,
}

UFS_DISPONIVEIS = set(SCRAPERS_POR_UF.keys())


async def consultar_ie(page, cnpj: str, uf: str) -> dict:
    uf = uf.upper().strip()
    if uf not in SCRAPERS_POR_UF:
        raise ValueError(f"UF {uf} ainda nao implementada. Disponiveis: {', '.join(sorted(UFS_DISPONIVEIS))}")
    return await SCRAPERS_POR_UF[uf](page, cnpj)
