import asyncio
from ..config import OPTANTES_URL


async def consultar_optantes(page, cnpj_limpo: str, max_retries: int = 3) -> dict:
    for attempt in range(max_retries):
        await page.goto(OPTANTES_URL, wait_until="domcontentloaded")
        await page.wait_for_selector("#Cnpj", timeout=10000)
        await page.fill("#Cnpj", cnpj_limpo)

        await page.click("text=Consultar")
        try:
            await page.wait_for_selector("h3:has-text('Situação Atual')", timeout=30000)
            break
        except Exception:
            if attempt < max_retries - 1:
                await asyncio.sleep(2 + attempt * 2)
                continue
            error_text = await page.evaluate("() => document.body.innerText.substring(0, 300)")
            raise RuntimeError(f"Optantes - Falha (captcha ou erro): {error_text[:200]}")

    situacao = await page.evaluate("""() => {
        const body = document.body.innerText;
        const simples = body.match(/Situação no Simples Nacional:\\s*(.+)/);
        const simei = body.match(/Situação no SIMEI:\\s*(.+)/);
        return {
            simples_nacional: simples ? simples[1].trim() : '',
            simei: simei ? simei[1].trim() : '',
        };
    }""")

    btn = await page.query_selector("button:has-text('Mais informações'):not([disabled])")
    if btn:
        await btn.click()
        await page.wait_for_selector("h3:has-text('Períodos Anteriores')", timeout=10000)

    periodos = await page.evaluate("""() => {
        const result = { simples_anteriores: [], simei_anteriores: [], eventos_simples: '', eventos_simei: '' };
        const tables = document.querySelectorAll('table');
        const body = document.body.innerText;

        const parseTable = (table) => {
            const rows = table.querySelectorAll('tbody tr');
            return Array.from(rows).map(row => {
                const cells = row.querySelectorAll('td');
                return {
                    data_inicial: cells[0] ? cells[0].textContent.trim() : '',
                    data_final: cells[1] ? cells[1].textContent.trim() : '',
                    detalhamento: cells[2] ? cells[2].textContent.trim() : '',
                };
            });
        };

        const panels = document.querySelectorAll('.panel-body, div');
        let simplesTable = null, simeiTable = null;

        panels.forEach(panel => {
            const text = panel.textContent;
            if (text.includes('Opções pelo Simples Nacional em Períodos Anteriores')) {
                const t = panel.querySelector('table');
                if (t) simplesTable = t;
            }
            if (text.includes('Enquadramentos no SIMEI em Períodos Anteriores')) {
                const allTables = panel.querySelectorAll('table');
                if (allTables.length >= 2) simeiTable = allTables[1];
                else if (allTables.length === 1 && !simplesTable) simeiTable = allTables[0];
                else if (allTables.length === 1) simeiTable = allTables[0];
            }
        });

        if (simplesTable) result.simples_anteriores = parseTable(simplesTable);
        if (simeiTable) result.simei_anteriores = parseTable(simeiTable);

        if (body.includes('Eventos Futuros (Simples Nacional)')) {
            const match = body.match(/Eventos Futuros \\(Simples Nacional\\)\\s*([\\s\\S]*?)(?:Eventos Futuros \\(SIMEI\\)|$)/);
            result.eventos_simples = match ? match[1].trim() : '';
        }
        if (body.includes('Eventos Futuros (SIMEI)')) {
            const match = body.match(/Eventos Futuros \\(SIMEI\\)\\s*([\\s\\S]*?)(?:Informações|Voltar|$)/);
            result.eventos_simei = match ? match[1].trim() : '';
        }

        return result;
    }""")

    def _eventos_ou_null(val: str) -> str | None:
        if not val or val.lower() in ("não existem", "nao existem", ""):
            return None
        return val

    return {
        "simples_nacional": {
            "situacao": situacao["simples_nacional"] or None,
            "periodos_anteriores": periodos["simples_anteriores"] or [],
            "eventos_futuros": _eventos_ou_null(periodos["eventos_simples"]),
        },
        "simei": {
            "situacao": situacao["simei"] or None,
            "periodos_anteriores": periodos["simei_anteriores"] or [],
            "eventos_futuros": _eventos_ou_null(periodos["eventos_simei"]),
        },
    }
