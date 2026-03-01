import asyncio
from playwright.async_api import async_playwright
from .browser import create_browser
from .scrapers import consultar_cnpj, consultar_dasn, consultar_optantes
from .scrapers.ccc import consultar_ccc, consultar_ccc_todas_ufs
from .scrapers.sintegra import consultar_ie


def _limpar_cnpj(cnpj: str) -> str:
    cnpj_limpo = cnpj.replace(".", "").replace("/", "").replace("-", "")
    if len(cnpj_limpo) != 14 or not cnpj_limpo.isdigit():
        raise ValueError(f"CNPJ invalido: {cnpj}")
    return cnpj_limpo


async def consultar_mei(cnpj: str) -> dict:
    cnpj_limpo = _limpar_cnpj(cnpj)

    async with async_playwright() as p:
        browser, context = await create_browser(p)
        page = await context.new_page()

        cadastro = await consultar_cnpj(page, cnpj_limpo)
        await asyncio.sleep(1)
        dasn = await consultar_dasn(page, cnpj_limpo)
        await asyncio.sleep(1)
        optantes = await consultar_optantes(page, cnpj_limpo)

        await browser.close()

    return {
        "cnpj": cnpj,
        "cadastro": cadastro,
        "dasn": dasn,
        "optantes": optantes,
    }


async def consultar_apenas_cnpj(cnpj: str) -> dict:
    cnpj_limpo = _limpar_cnpj(cnpj)

    async with async_playwright() as p:
        browser, context = await create_browser(p)
        page = await context.new_page()
        resultado = await consultar_cnpj(page, cnpj_limpo)
        await browser.close()

    return resultado


async def consultar_apenas_dasn(cnpj: str) -> dict:
    cnpj_limpo = _limpar_cnpj(cnpj)

    async with async_playwright() as p:
        browser, context = await create_browser(p)
        page = await context.new_page()
        resultado = await consultar_dasn(page, cnpj_limpo)
        await browser.close()

    return resultado


async def consultar_apenas_optantes(cnpj: str) -> dict:
    cnpj_limpo = _limpar_cnpj(cnpj)

    async with async_playwright() as p:
        browser, context = await create_browser(p)
        page = await context.new_page()
        resultado = await consultar_optantes(page, cnpj_limpo)
        await browser.close()

    return resultado


async def consultar_apenas_ccc(cnpj: str, uf: str) -> dict:
    cnpj_limpo = _limpar_cnpj(cnpj)
    return await consultar_ccc(cnpj_limpo, uf)


async def consultar_apenas_ccc_todas_ufs(cnpj: str) -> dict:
    cnpj_limpo = _limpar_cnpj(cnpj)
    return await consultar_ccc_todas_ufs(cnpj_limpo)


async def consultar_apenas_ie(cnpj: str, uf: str) -> dict:
    cnpj_limpo = _limpar_cnpj(cnpj)

    async with async_playwright() as p:
        browser, context = await create_browser(p)
        page = await context.new_page()
        resultado = await consultar_ie(page, cnpj_limpo, uf)
        await browser.close()

    return resultado
