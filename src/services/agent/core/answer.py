import json
from langchain_groq import ChatGroq


def generate_final_answer(
    user_prompt,
    filters,
    ranking,
    api_key,
    city_fallback=False
):
    llm = ChatGroq(
        api_key=api_key,
        model_name="llama-3.3-70b-versatile",
        temperature=0.3
    )

    context = {
        "pergunta_usuario": user_prompt,
        "filtros_aplicados": filters,
        "cidade_nao_encontrada": city_fallback,
        "top_pontos": ranking[:5]
    }

    instruction = (
        "Você é um especialista em planejamento de mídia OOH.\n"
        "Com base na pergunta do usuário e nos dados calculados, "
        "gere uma resposta estratégica clara.\n\n"
        "Se cidade_nao_encontrada for true, explique que a cidade "
        "não possui dados e que a análise foi feita considerando todos os pontos.\n\n"
        "Use os dados dos top_pontos para justificar.\n"
        "Responda em português.\n\n"
        f"Contexto:\n{json.dumps(context, ensure_ascii=False)}"
    )

    response = llm.invoke(instruction)

    return response.content