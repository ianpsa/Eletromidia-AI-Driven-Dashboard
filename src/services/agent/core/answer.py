import json
import numpy as np
from langchain_groq import ChatGroq


def generate_final_answer(user_prompt, filters, ranking, api_key, city_fallback=False, used_age_range=False):
    llm = ChatGroq(
        api_key=api_key, model_name="llama-3.3-70b-versatile", temperature=0.3
    )

    def convert(o):
        if isinstance(o, np.integer):
            return int(o)
        if isinstance(o, np.floating):
            return float(o)
        return str(o)

    context = {
        "pergunta_usuario": user_prompt,
        "filtros_aplicados": filters,
        "cidade_nao_encontrada": city_fallback,
        "faixa_utilizada": used_age_range,
        "top_pontos": ranking[:5],
    }

    instruction = (
        "Você é um especialista em planejamento de mídia OOH.\n"
        "Com base na pergunta do usuário e nos dados calculados, "
        "Responda de forma natural e estratégica.\n\n"
        "Quando a faixa etária solicitada não existir exatamente nos dados, "
        "explique de forma simples que as informações estão organizadas em "
        "faixas e que a análise utilizou as faixas disponíveis mais próximas "
        "para cobrir o intervalo solicitado.\n\n"
        "Se cidade_nao_encontrada for true, explique que a cidade "
        "não possui dados e que a análise foi feita considerando todos os pontos.\n\n"
        "Use os dados dos top_pontos para justificar.\n"
        "Nunca mencione termos técnicos, nomes de campos ou funcionamento interno.\n"
        "Responda apenas como consultor.\n\n"
        "Responda em português.\n\n"
        f"Contexto:\n{json.dumps(context, ensure_ascii=False, default=convert)}"
    )

    response = llm.invoke(instruction)

    return response.content
