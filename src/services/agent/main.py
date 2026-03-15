import argparse
import os
import sys

import pandas as pd
from core.answer import generate_final_answer
from core.llm import parse_prompt
from core.report import build_report
from dotenv import load_dotenv

load_dotenv()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--prompt", required=True)
    parser.add_argument("--limit", type=int, default=10)
    args = parser.parse_args()

    token = os.getenv("GROQ_API_KEY")
    if not token:
        sys.exit("GROQ_API_KEY not set")

    user_prompt = args.prompt

    filters = parse_prompt(user_prompt, token)

    df = pd.read_csv("../../../data/claro.csv")

    rows = build_report(df, filters)

    city_fallback = False

    if not rows and "city" in filters:
        original_city = filters["city"]
        del filters["city"]
        rows = build_report(df, filters)
        city_fallback = True

    if not rows:
        print("Nenhum dado encontrado para os filtros informados.")
        return

    final_answer = generate_final_answer(
        user_prompt=user_prompt,
        filters=filters,
        ranking=rows[: args.limit],
        api_key=token,
        city_fallback=city_fallback,
    )

    print("\nResposta Estratégica:\n")
    print(final_answer)


if __name__ == "__main__":
    main()
