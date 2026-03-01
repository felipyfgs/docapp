import asyncio
from concurrent.futures import ThreadPoolExecutor
from ...captcha import submit_recaptcha_v2, poll_recaptcha_token

SINTEGRA_URL = "https://sistemas1.sefaz.ma.gov.br/sintegra/jsp/consultaSintegra/consultaSintegraFiltro.jsf"
RECAPTCHA_SITEKEY = "6LfKD1MUAAAAAOJxT7nt0qoTMOyII0tcs2CbYsmc"
ENTERPRISE = True

_executor = ThreadPoolExecutor(max_workers=2)


def _texto_celula(cells, label):
    for i, cell in enumerate(cells):
        if label.lower() in cell.lower() and i + 1 < len(cells):
            return cells[i + 1].strip()
    return ""


async def _injetar_token(page, token):
    await page.evaluate("""(token) => {
        const ta = document.getElementById('g-recaptcha-response');
        ta.value = token;
    }""", token)


async def consultar_ie_ma(page, cnpj: str) -> dict:
    resultado = {
        "uf": "MA",
        "cnpj": cnpj,
        "encontrado": False,
        "ie": None,
        "razao_social": None,
        "situacao": None,
        "data_situacao": None,
        "regime": None,
        "cnae_principal": None,
        "endereco": None,
        "obrigacoes": None,
    }

    # Submeter job NoPeCHA com enterprise=True ANTES de carregar a pagina
    loop = asyncio.get_event_loop()
    job_id = await loop.run_in_executor(
        _executor, submit_recaptcha_v2, SINTEGRA_URL, RECAPTCHA_SITEKEY, ENTERPRISE
    )

    await page.goto(SINTEGRA_URL, wait_until="networkidle", timeout=30000)
    await asyncio.sleep(1)

    # Selecionar CPF/CNPJ e preencher
    await page.locator("#form1\\:tipoEmissao\\:1").click()
    await asyncio.sleep(0.5)
    await page.locator("#form1\\:cpfCnpj").fill(cnpj)
    await asyncio.sleep(0.5)

    # Aguardar token do NoPeCHA (com enterprise=True resolve em ~1s)
    token = await loop.run_in_executor(_executor, poll_recaptcha_token, job_id, 60)
    await _injetar_token(page, token)

    # Clicar Consulta
    await page.locator("#form1\\:pnlPrincipal4 input:nth-of-type(2)").click()

    try:
        await page.wait_for_url("**/consultaSintegraResultado*", timeout=15000)
    except Exception:
        await asyncio.sleep(3)

    content = await page.content()
    if "Nenhum registro" in content or "não encontrado" in content.lower():
        return resultado
    if "Captcha inválido" in content:
        raise RuntimeError("Captcha invalido - token rejeitado pelo servidor")

    # Se na lista, clicar para ir aos detalhes
    if "ResultadoListaConsulta" in page.url:
        try:
            await page.locator("#j_id6\\:listaDados\\:0\\:j_id36 img").first.click()
            await page.wait_for_url("**/consultaSintegraResultadoConsulta*", timeout=10000)
        except Exception:
            try:
                await page.locator("#j_id6\\:pnlCadastro img").first.click()
                await page.wait_for_url("**/consultaSintegraResultadoConsulta*", timeout=10000)
            except Exception:
                await asyncio.sleep(2)

    resultado["encontrado"] = True
    cells = await page.locator("td").all_text_contents()

    resultado["ie"] = _texto_celula(cells, "Inscrição Estadual:")
    if resultado["ie"]:
        resultado["ie"] = resultado["ie"].replace(".", "").replace("-", "").strip()

    cgc = _texto_celula(cells, "CGC:")
    if not cgc:
        cgc = _texto_celula(cells, "CNPJ:")
    if cgc:
        resultado["cnpj"] = cgc.replace(".", "").replace("/", "").replace("-", "").strip()

    resultado["razao_social"] = _texto_celula(cells, "Razão Social:")
    resultado["regime"] = _texto_celula(cells, "Regime Apuração:")
    resultado["situacao"] = _texto_celula(cells, "Situação Cadastral Vigente:")
    resultado["data_situacao"] = _texto_celula(cells, "Data desta Situação Cadastral:")
    resultado["cnae_principal"] = _texto_celula(cells, "CNAE Principal:")

    resultado["endereco"] = {
        "logradouro": _texto_celula(cells, "Logradouro:"),
        "numero": _texto_celula(cells, "Número:"),
        "complemento": _texto_celula(cells, "Complemento:"),
        "bairro": _texto_celula(cells, "Bairro:"),
        "municipio": _texto_celula(cells, "Município:"),
        "uf": "MA",
        "cep": _texto_celula(cells, "CEP:"),
    }

    nfe = _texto_celula(cells, "NFe a partir de")
    cte = _texto_celula(cells, "CTE a partir de:")
    edf = _texto_celula(cells, "EDF a partir de:")
    resultado["obrigacoes"] = {
        "nfe": nfe or None,
        "cte": cte or None,
        "edf": edf or None,
    }

    return resultado
