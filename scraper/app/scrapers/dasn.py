from ..config import DASN_URL


async def consultar_dasn(page, cnpj_limpo: str) -> dict:
    await page.goto(f"{DASN_URL}/Identificacao", wait_until="domcontentloaded")
    await page.wait_for_selector("#identificacao-cnpj", timeout=10000)
    await page.fill("#identificacao-cnpj", cnpj_limpo)

    await page.click("#identificacao-continuar")
    try:
        await page.wait_for_url("**/dasnsimei.app/", timeout=30000)
    except Exception:
        error_el = await page.query_selector(".feedback:not(.d-none)")
        error_text = await error_el.inner_text() if error_el else "Resposta inesperada"
        raise RuntimeError(f"DASN - Falha (captcha ou erro): {error_text[:200]}")

    await page.wait_for_selector("#iniciar-ano-calendario", timeout=10000)

    razao_social = await page.evaluate("""() => {
        const ps = document.querySelectorAll('p');
        for (let i = 0; i < ps.length; i++) {
            const strong = ps[i].querySelector('strong');
            if (strong && strong.textContent.includes('Raz') && i + 1 < ps.length)
                return ps[i + 1].textContent.trim();
        }
        return '';
    }""")

    declaracoes = await page.evaluate("""() => {
        const container = document.querySelector('#iniciar-ano-calendario');
        if (!container) return [];
        const radios = container.querySelectorAll('input[type="radio"]');
        return Array.from(radios).map(radio => {
            const parent = radio.parentElement;
            const spans = parent.querySelectorAll('span');
            let statusTexto = '', acao = '';
            spans.forEach(span => {
                const cls = span.className || '';
                const text = span.textContent.trim();
                if (cls.includes('br-tag')) acao = text;
                else if (text) statusTexto = text;
            });
            const baixada = acao.toLowerCase() === 'baixada' || statusTexto.toLowerCase() === 'baixada';
            const pendente = !radio.disabled && statusTexto.toLowerCase().includes('não apresentada');
            return {
                ano: radio.value,
                tipo_declaracao: radio.dataset.tipoDeclaracao || '',
                situacao_especial: radio.dataset.situacaoEspecialTipo || '-',
                data_baixa: radio.dataset.situacaoEspecialEventobaixa || '-',
                status: statusTexto || acao,
                pendente, baixada,
            };
        });
    }""")

    mei_baixada = any(d["baixada"] for d in declaracoes)
    data_baixa = next(
        (d["data_baixa"] for d in declaracoes
         if d["data_baixa"] != "-" and d["situacao_especial"] == "Extinção"),
        None,
    )
    pendentes = [d["ano"] for d in declaracoes if d["pendente"]]

    cleaned = []
    for d in declaracoes:
        sit = d["situacao_especial"] if d["situacao_especial"] != "-" else None
        db = d["data_baixa"] if d["data_baixa"] != "-" else None
        cleaned.append({
            "ano": d["ano"],
            "tipo_declaracao": d["tipo_declaracao"],
            "status": d["status"],
            "pendente": d["pendente"],
            "baixada": d["baixada"],
            "situacao_especial": sit,
            "data_baixa": db,
        })

    return {
        "razao_social": razao_social or None,
        "mei_baixada": mei_baixada,
        "data_baixa": data_baixa,
        "declaracoes": cleaned,
        "pendentes": pendentes,
    }
