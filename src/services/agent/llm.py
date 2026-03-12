import json
import os


def classify_with_groq_llm(user_prompt: str, api_key: str, model_name: str = "llama-3.3-70b-versatile") -> dict:
    """Use ChatGroq (GROQ LLM) to parse the user's prompt into a JSON filters object.

    Returns a dict with keys: gender, age_min, age_max, classes (list), city.
    This helper is minimal and will raise SystemExit on any failure — the demo
    requires the LLM to succeed.
    """
    try:
        from langchain_groq import ChatGroq
    except Exception:
        raise SystemExit("Required package 'langchain_groq' is not installed. Install it (pip install langchain-groq) to use GROQ as LLM.")

    if not api_key:
        raise SystemExit("GROQ_API_KEY must be set in the environment to use GROQ as LLM.")

    router = ChatGroq(api_key=api_key, model_name=model_name, temperature=0.0)

    # Very small strict prompt: return ONLY a JSON object with keys we expect.
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
        raise SystemExit("GROQ LLM returned an empty filter object; ensure the prompt is valid.")

    return out
