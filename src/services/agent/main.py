import os
import sys
import argparse
import pandas as pd
from dotenv import load_dotenv

from core.llm import parse_prompt
from core.report import build_report, print_table


load_dotenv()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--prompt", required=True)
    parser.add_argument("--limit", type=int, default=10)
    args = parser.parse_args()

    token = os.getenv("GROQ_API_KEY")
    if not token:
        sys.exit("GROQ_API_KEY not set")

    filters = parse_prompt(args.prompt, token)

    df = pd.read_csv("../../../data/claro.csv")
    rows = build_report(df, filters)

    print_table(rows, args.limit)


if __name__ == "__main__":
    main()