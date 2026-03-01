import asyncio
from ..config import CNPJ_URL, CNPJ_HCAPTCHA_SITEKEY, NOPECHA_API_KEY
from ..captcha import resolver_hcaptcha


def _limpar(valor: str) -> str | None:
    if not valor or valor.strip() in ("", "********"):
        return None
    return valor.strip()


async def consultar_cnpj(page, cnpj_limpo: str, max_retries: int = 3) -> dict:
    page_url = f"{CNPJ_URL}?cnpj={cnpj_limpo}"

    for attempt in range(max_retries):
        await page.goto(page_url, wait_until="domcontentloaded")

        captcha_resolved = False
        try:
            captcha_frame = page.frame_locator('iframe[title*="Widget contendo"]')
            checkbox = captcha_frame.locator("#checkbox")
            await checkbox.wait_for(timeout=5000)
            await checkbox.click()
            checked = captcha_frame.locator('[aria-checked="true"]')
            await checked.wait_for(timeout=5000)
            captcha_resolved = True
        except Exception:
            pass

        if not captcha_resolved and NOPECHA_API_KEY:
            try:
                token = await asyncio.to_thread(
                    resolver_hcaptcha, page_url, CNPJ_HCAPTCHA_SITEKEY
                )
                await page.evaluate("""(token) => {
                    document.querySelector('[name="h-captcha-response"]').value = token;
                    document.querySelector('[name="g-recaptcha-response"]').value = token;
                }""", token)
                captcha_resolved = True
            except Exception as e:
                if attempt == max_retries - 1:
                    raise RuntimeError(f"CNPJ - Captcha solver falhou: {e}")

        if not captcha_resolved:
            if attempt < max_retries - 1:
                await asyncio.sleep(2 + attempt * 2)
                continue
            raise RuntimeError("CNPJ - Captcha nao resolvido (configure NOPECHA_API_KEY)")

        await page.locator("button:has-text('Consultar')").click(force=True)
        try:
            await page.wait_for_url("**/Cnpjreva_Comprovante*", timeout=30000)
            break
        except Exception:
            if attempt < max_retries - 1:
                await asyncio.sleep(2 + attempt * 2)
                continue
            error_text = await page.evaluate("() => document.body.innerText.substring(0, 300)")
            raise RuntimeError(f"CNPJ - Falha apos captcha: {error_text[:200]}")

    dados = await page.evaluate("""() => {
        const data = {};
        const tables = document.querySelectorAll('table');
        tables.forEach(table => {
            const cells = table.querySelectorAll('td');
            cells.forEach(cell => {
                const parts = cell.textContent.trim().split('\\n').map(s => s.trim()).filter(Boolean);
                if (parts.length >= 2) {
                    data[parts[0]] = parts.slice(1).join(' ').trim();
                }
            });
        });
        return data;
    }""")

    return {
        "nome_empresarial": _limpar(dados.get("NOME EMPRESARIAL", "")),
        "nome_fantasia": _limpar(dados.get("TÍTULO DO ESTABELECIMENTO (NOME DE FANTASIA)", "")),
        "data_abertura": _limpar(dados.get("DATA DE ABERTURA", "")),
        "porte": _limpar(dados.get("PORTE", "")),
        "atividade_principal": _limpar(dados.get("CÓDIGO E DESCRIÇÃO DA ATIVIDADE ECONÔMICA PRINCIPAL", "")),
        "atividades_secundarias": _limpar(dados.get("CÓDIGO E DESCRIÇÃO DAS ATIVIDADES ECONÔMICAS SECUNDÁRIAS", "")),
        "natureza_juridica": _limpar(dados.get("CÓDIGO E DESCRIÇÃO DA NATUREZA JURÍDICA", "")),
        "endereco": {
            "logradouro": _limpar(dados.get("LOGRADOURO", "")),
            "numero": _limpar(dados.get("NÚMERO", "")),
            "complemento": _limpar(dados.get("COMPLEMENTO", "")),
            "bairro": _limpar(dados.get("BAIRRO/DISTRITO", "")),
            "municipio": _limpar(dados.get("MUNICÍPIO", "")),
            "uf": _limpar(dados.get("UF", "")),
            "cep": _limpar(dados.get("CEP", "")),
        },
        "contato": {
            "email": _limpar(dados.get("ENDEREÇO ELETRÔNICO", "")),
            "telefone": _limpar(dados.get("TELEFONE", "")),
        },
        "situacao_cadastral": _limpar(dados.get("SITUAÇÃO CADASTRAL", "")),
        "data_situacao_cadastral": _limpar(dados.get("DATA DA SITUAÇÃO CADASTRAL", "")),
        "motivo_situacao_cadastral": _limpar(dados.get("MOTIVO DE SITUAÇÃO CADASTRAL", "")),
        "situacao_especial": _limpar(dados.get("SITUAÇÃO ESPECIAL", "")),
        "data_situacao_especial": _limpar(dados.get("DATA DA SITUAÇÃO ESPECIAL", "")),
    }
