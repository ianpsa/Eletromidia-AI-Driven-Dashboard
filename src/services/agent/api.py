import os
import pandas as pd
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from core.llm import parse_prompt
from core.report import build_report
from core.answer import generate_final_answer


load_dotenv()

app = FastAPI()

DATA_PATH = "../../../data/claro.csv"


class CampaignRequest(BaseModel):
    prompt: str
    limit: int = 5


@app.post("/analyze")
def analyze_campaign(request: CampaignRequest):
    token = os.getenv("GROQ_API_KEY")
    if not token:
        raise HTTPException(status_code=500, detail="GROQ_API_KEY not set")

    try:
        filters = parse_prompt(request.prompt, token)

        df = pd.read_csv(DATA_PATH)

        rows = build_report(df, filters)

        city_fallback = False

        if not rows and "city" in filters:
            del filters["city"]
            rows = build_report(df, filters)
            city_fallback = True

        if not rows:
            return {
                "success": False,
                "message": "Nenhum dado encontrado para os critérios informados."
            }

        top_points = rows[:request.limit]

        final_answer = generate_final_answer(
            user_prompt=request.prompt,
            filters=filters,
            ranking=top_points,
            api_key=token,
            city_fallback=city_fallback
        )

        return {
            "success": True,
            "analysis": final_answer,
            "top_points": top_points
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))