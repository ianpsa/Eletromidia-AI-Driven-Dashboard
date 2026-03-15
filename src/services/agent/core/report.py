import pandas as pd


def parse_age_range(label):
    if pd.isna(label):
        return None, None
    parts = str(label).split("-")
    if len(parts) == 2:
        return int(parts[0]), int(parts[1])
    return None, None


def normalize_gender(v):
    v = str(v).strip().lower()
    if "fem" in v:
        return "female"
    if "masc" in v:
        return "male"
    return v


def build_report(df, filters):
    mask = pd.Series(True, index=df.index)

    if "gender" in filters:
        target_gender = filters["gender"].lower()
        gender_series = df["gender_label"].apply(normalize_gender)
        mask &= gender_series == target_gender

    if "classes" in filters:
        mask &= df["class_label"].isin(filters["classes"])

    if "city" in filters:
        mask &= df["cidade"].str.lower() == filters["city"].lower()

    if "age_min" in filters or "age_max" in filters:
        age_min = filters.get("age_min", 0)
        age_max = filters.get("age_max", 200)

        ages = df["age_label"].apply(parse_age_range)

        mask &= ages.apply(
            lambda x: x[0] is not None and x[0] >= age_min and x[1] <= age_max
        )

    filtered = df[mask].copy()

    if filtered.empty:
        return []

    agg = (
        filtered.groupby(["endereco", "numero"])
        .agg({"joint_count": "sum", "uniques": "max"})
        .reset_index()
    )

    agg["affinity"] = (agg["joint_count"] / agg["uniques"]) * 100

    agg = agg.sort_values("affinity", ascending=False)

    return [
        {
            "point": f"{r.endereco}, {r.numero}",
            "total_target": round(r.joint_count, 2),
            "total_flow": round(r.uniques, 2),
            "affinity": round(r.affinity, 2),
        }
        for r in agg.itertuples()
    ]


def print_table(rows, limit):
    if not rows:
        print("No matching audience found")
        return

    rows = rows[:limit]

    print("point | total_target | total_flow | affinity_%")
    print("------------------------------------------------")
    for r in rows:
        print(
            f"{r['point']} | {r['total_target']} | {r['total_flow']} | {r['affinity']}"
        )
