import argparse
import glob
import json
import os
from datetime import datetime

import pandas as pd
import requests


def classify_with_groq_llm(user_prompt: str, api_key: str, model_name: str = "llama-3.3-70b-versatile") -> dict:
    try:
        from langchain_groq import ChatGroq
    except Exception:
        raise SystemExit("Required package 'langchain_groq' is not installed. Install it (pip install langchain-groq) to use GROQ as LLM.")

    router = ChatGroq(api_key=api_key, model_name=model_name, temperature=0.0)

    llm_prompt = (
        "You are an assistant that receives a marketing brief and must output a single JSON object "
        "with optional keys: gender (female/male), age_min (int), age_max (int), classes (array like [\"A\", \"B\"]), city (string). "
        "Return ONLY the JSON object with these keys or an empty object if nothing applies.\n\n"
        f"User brief: {user_prompt}\n"
    )

    resp = router.invoke(llm_prompt)
    raw = getattr(resp, "content", str(resp)).strip()
    try:
        data = json.loads(raw)
    except Exception as e:
        raise SystemExit(f"GROQ LLM returned non-JSON output; parsing failed: {e}\nRaw output:\n{raw}")

    out = {}
    if "gender" in data:
        out["gender"] = data["gender"]
    if "age_min" in data:
        try:
            out["age_min"] = int(data["age_min"])
        except Exception:
            pass
    if "age_max" in data:
        try:
            out["age_max"] = int(data["age_max"])
        except Exception:
            pass
    if "classes" in data and isinstance(data["classes"], (list, tuple)):
        out["classes"] = [str(x).strip() for x in data["classes"] if x]
    if "city" in data:
        out["city"] = data["city"]
    if not out:
        raise SystemExit("GROQ LLM returned an empty filter object; ensure the prompt is valid or adjust the LLM prompt/model.")
    return out


def find_latest_csv(data_dir="./data", pattern="bq-results-*.csv"):
    paths = glob.glob(os.path.join(data_dir, pattern))
    if not paths:
        raise FileNotFoundError(f"No CSV files found matching {data_dir}/{pattern}")
    paths.sort(key=os.path.getmtime, reverse=True)
    return paths[0]


def canonical_col(col):
    return col.strip().lower()


def detect_columns(df):
    cols = {canonical_col(c): c for c in df.columns}
    mapping = {}
    # age
    for name in ("idade", "age", "years"):
        if name in cols:
            mapping["age"] = cols[name]
            break
    # sex
    for name in ("sexo", "gender", "sex"):
        if name in cols:
            mapping["sex"] = cols[name]
            break
    # class
    for name in ("classe", "class", "social_class"):
        if name in cols:
            mapping["class"] = cols[name]
            break
    # city / state
    for name in ("cidade", "city", "municipio", "state", "estado"):
        if name in cols:
            mapping["city"] = cols[name]
            break
    # location / street
    for name in ("rua", "street", "location", "point", "endereco", "address"):
        if name in cols:
            mapping["street"] = cols[name]
            break
    # date
    for name in ("data", "date", "timestamp", "datetime"):
        if name in cols:
            mapping["date"] = cols[name]
            break
    return mapping


def normalize_date_series(s):
    return pd.to_datetime(s, errors="coerce").dt.floor("D")


def sex_is_female(v):
    if pd.isna(v):
        return False
    v = str(v).strip().lower()
    return v in ("f", "female", "mulher", "feminino")


def class_is_ab(v):
    if pd.isna(v):
        return False
    v = str(v).strip().upper()
    return v in ("A", "B", "AB")


def city_is_sao_paulo(v):
    if pd.isna(v):
        return False
    v = str(v).strip().lower()
    return "sao paulo" in v or "são paulo" in v or v in ("sp", "sao-paulo", "s\u00e3o paulo")


def compute_report(df, mapping, filters=None):
    # ensure we have required columns
    if "street" not in mapping:
        raise RuntimeError("Could not detect a street/location column in the dataset.")
    if "date" not in mapping:
        df["__day"] = pd.to_datetime("today").floor("D")
        date_col = "__day"
    else:
        date_col = mapping["date"]
        df[date_col] = normalize_date_series(df[date_col])

    street_col = mapping["street"]

    # Filters
    mask = pd.Series(True, index=df.index)
    if filters is None:
        filters = {}

    # gender
    gender = filters.get("gender")
    if gender is None and "sex" in mapping:
        pass
    elif gender is not None and "sex" in mapping:
        if gender.lower() in ("female", "mulher", "f", "feminino"):
            mask = mask & df[mapping["sex"]].apply(sex_is_female)
        else:
            mask = mask & (~df[mapping["sex"]].apply(sex_is_female))

    # age range
    age_min = filters.get("age_min")
    age_max = filters.get("age_max")
    if (age_min is not None or age_max is not None) and "age" in mapping:
        ages = pd.to_numeric(df[mapping["age"]], errors="coerce")
        if age_min is not None:
            mask = mask & (ages >= age_min)
        if age_max is not None:
            mask = mask & (ages <= age_max)

    # social class
    classes = filters.get("classes")
    if classes and "class" in mapping:
        def class_match(v):
            if pd.isna(v):
                return False
            vv = str(v).upper()
            for c in classes:
                if c.upper() in vv:
                    return True
            return False

        mask = mask & df[mapping["class"]].apply(class_match)

    # city
    city = filters.get("city")
    if city and "city" in mapping:
        def city_match(v):
            if pd.isna(v):
                return False
            return city.lower() in str(v).strip().lower()

        mask = mask & df[mapping["city"]].apply(city_match)

    filtered = df[mask].copy()

    if filtered.empty:
        return []

    # aggregate: count per street per day
    agg = (
        filtered
        .groupby([street_col, date_col])
        .size()
        .reset_index(name="count")
    )

    days_per_street = agg.groupby(street_col)[date_col].nunique().rename("days")
    total_per_street = agg.groupby(street_col)["count"].sum().rename("total")

    report = pd.concat([total_per_street, days_per_street], axis=1)
    report["avg_per_day"] = (report["total"] / report["days"]).fillna(0)

    if report["avg_per_day"].max() > 0:
        report["affinity"] = (report["avg_per_day"] / report["avg_per_day"].max()) * 100
    #!/usr/bin/env python3
    """Minimal demo CLI that uses GROQ as LLM to parse a campaign brief and
    reports best streets to invest in using local CSV (`data/claro.csv`).

    Design constraints (per user):
    - Use GROQ only as an LLM to parse prompts (mandatory).
    - No local heuristics fallback.
    - No GROQ execution against Sanity.
    - Minimal, modular, and focused on the demo functionality.
    """

    import argparse
    import os
    import sys
    import pandas as pd

    from src.services.agent.llm import classify_with_groq_llm
    from src.services.agent.data import detect_columns, compute_report, print_table


    def main():
        parser = argparse.ArgumentParser(description="Analyze campaign locations for a target audience")
        parser.add_argument("--prompt", required=True, help="Natural language prompt describing the campaign")
        parser.add_argument("--limit", type=int, default=10, help="Number of rows to show (default 10)")
        args = parser.parse_args()

        token = os.getenv("GROQ_API_KEY")
        if not token:
            print("Error: GROQ_API_KEY must be set to use GROQ as the LLM for parsing.")
            sys.exit(2)

        # Parse prompt using GROQ LLM (must succeed or exit)
        filters = classify_with_groq_llm(args.prompt, api_key=token)

        # Load local CSV dataset (this script always uses local CSV for analysis)
        csv_path = os.path.join(os.getcwd(), "data", "claro.csv")
        if not os.path.isfile(csv_path):
            print(f"Error: local CSV not found at {csv_path}. This demo requires the CSV.")
            sys.exit(2)

        df = pd.read_csv(csv_path)
        mapping = detect_columns(df)
        rows = compute_report(df, mapping, filters=filters)

        # Present human-readable table with counts and affinity
        print_table(rows, limit=args.limit)


    if __name__ == "__main__":
        main()
        out["gender"] = "female"
