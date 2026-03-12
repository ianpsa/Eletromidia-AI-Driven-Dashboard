import os
import pandas as pd


def canonical_col(col):
    return col.strip().lower()


def detect_columns(df):
    cols = {canonical_col(c): c for c in df.columns}
    mapping = {}
    for name in ("idade", "age", "years"):
        if name in cols:
            mapping["age"] = cols[name]
            break
    for name in ("sexo", "gender", "sex"):
        if name in cols:
            mapping["sex"] = cols[name]
            break
    for name in ("classe", "class", "social_class"):
        if name in cols:
            mapping["class"] = cols[name]
            break
    for name in ("cidade", "city", "municipio", "state", "estado"):
        if name in cols:
            mapping["city"] = cols[name]
            break
    for name in ("rua", "street", "location", "point", "endereco", "address"):
        if name in cols:
            mapping["street"] = cols[name]
            break
    for name in ("data", "date", "timestamp", "datetime"):
        if name in cols:
            mapping["date"] = cols[name]
            break
    return mapping


def normalize_date_series(s):
    return pd.to_datetime(s, errors="coerce").dt.floor("D")


def compute_report(df, mapping, filters=None):
    if "street" not in mapping:
        raise RuntimeError("Could not detect a street/location column in the dataset.")
    if "date" not in mapping:
        df["__day"] = pd.to_datetime("today").floor("D")
        date_col = "__day"
    else:
        date_col = mapping["date"]
        df[date_col] = normalize_date_series(df[date_col])

    street_col = mapping["street"]
    mask = pd.Series(True, index=df.index)
    if filters is None:
        filters = {}

    gender = filters.get("gender")
    if gender is not None and "sex" in mapping:
        def sex_is_female(v):
            if pd.isna(v):
                return False
            v = str(v).strip().lower()
            return v in ("f", "female", "mulher", "feminino")

        if gender.lower() in ("female", "mulher", "f", "feminino"):
            mask = mask & df[mapping["sex"]].apply(sex_is_female)
        else:
            mask = mask & (~df[mapping["sex"]].apply(sex_is_female))

    age_min = filters.get("age_min")
    age_max = filters.get("age_max")
    if (age_min is not None or age_max is not None) and "age" in mapping:
        ages = pd.to_numeric(df[mapping["age"]], errors="coerce")
        if age_min is not None:
            mask = mask & (ages >= age_min)
        if age_max is not None:
            mask = mask & (ages <= age_max)

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
    else:
        report["affinity"] = 0

    report = report.sort_values("affinity", ascending=False).reset_index()

    rows = []
    for _, r in report.iterrows():
        rows.append({
            "street": r[street_col],
            "avg_people_per_day": float(round(r["avg_per_day"], 2)),
            "days_observed": int(r["days"]),
            "total_people": int(r["total"]),
            "affinity": float(round(r["affinity"], 2)),
        })
    return rows


def print_table(rows, limit=None):
    try:
        from tabulate import tabulate
    except Exception:
        tabulate = None

    if not rows:
        print("No matching records found for the target audience.")
        return

    if limit:
        rows = rows[:limit]

    table = [
        [r["street"], r["avg_people_per_day"], r["days_observed"], r["total_people"], r["affinity"]]
        for r in rows
    ]
    headers = ["Street", "Avg/day", "Days", "Total", "Affinity"]
    if tabulate:
        print(tabulate(table, headers=headers, tablefmt="github"))
    else:
        print("\t".join(headers))
        for row in table:
            print("\t".join(str(x) for x in row))
