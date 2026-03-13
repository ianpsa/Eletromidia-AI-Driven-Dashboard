import json
from langchain_groq import ChatGroq


def parse_prompt(prompt: str, api_key: str) -> dict:
    llm = ChatGroq(
        api_key=api_key,
        model_name="llama-3.3-70b-versatile",
        temperature=0.0
    )

    instruction = (
        "Você receberá uma descrição de campanha em português.\n"
        "Extraia os filtros da campanha e retorne APENAS um JSON válido.\n\n"
        "O JSON pode conter as seguintes chaves:\n"
        "- gender: 'female' ou 'male'\n"
        "- age_min: número inteiro\n"
        "- age_max: número inteiro\n"
        "- classes: lista como ['A', 'B']\n"
        "- city: string\n\n"
        "Não explique nada.\n"
        "Não escreva texto adicional.\n"
        "Retorne somente o JSON.\n\n"
        f"Campanha: {prompt}"
    )

    response = llm.invoke(
        instruction,
        response_format={"type": "json_object"}
    )

    content = response.content

    if not content:
        raise SystemExit("Resposta vazia do GROQ")

    return json.loads(content)