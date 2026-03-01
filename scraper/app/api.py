from fastapi import FastAPI, HTTPException
from .models import (
    ConsultaRequest, ConsultaResponse,
    CadastroResponse, DasnResponse, OptantesResponse,
    CccRequest, CccResponse, CccTodasUfsResponse,
    IeRequest, IeResponse,
)
from .consultar import (
    consultar_mei,
    consultar_apenas_cnpj,
    consultar_apenas_dasn,
    consultar_apenas_optantes,
    consultar_apenas_ccc,
    consultar_apenas_ccc_todas_ufs,
    consultar_apenas_ie,
)

app = FastAPI(title="Simplix", description="API de consulta MEI/CNPJ")


async def _safe_call(fn, *args):
    try:
        return await fn(*args)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except RuntimeError as e:
        raise HTTPException(status_code=502, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/consultar", response_model=ConsultaResponse)
async def consultar(req: ConsultaRequest):
    return await _safe_call(consultar_mei, req.cnpj)


@app.post("/consultar/cnpj", response_model=CadastroResponse)
async def consultar_cnpj(req: ConsultaRequest):
    return await _safe_call(consultar_apenas_cnpj, req.cnpj)


@app.post("/consultar/dasn", response_model=DasnResponse)
async def consultar_dasn(req: ConsultaRequest):
    return await _safe_call(consultar_apenas_dasn, req.cnpj)


@app.post("/consultar/optantes", response_model=OptantesResponse)
async def consultar_optantes(req: ConsultaRequest):
    return await _safe_call(consultar_apenas_optantes, req.cnpj)


@app.post("/consultar/ccc", response_model=CccResponse)
async def consultar_ccc(req: CccRequest):
    return await _safe_call(consultar_apenas_ccc, req.cnpj, req.uf)


@app.post("/consultar/ccc/todas", response_model=CccTodasUfsResponse)
async def consultar_ccc_todas(req: ConsultaRequest):
    return await _safe_call(consultar_apenas_ccc_todas_ufs, req.cnpj)


@app.post("/consultar/ie", response_model=IeResponse)
async def consultar_ie(req: IeRequest):
    return await _safe_call(consultar_apenas_ie, req.cnpj, req.uf)


@app.get("/health")
async def health():
    return {"status": "ok"}
