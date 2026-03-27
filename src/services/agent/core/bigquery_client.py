from __future__ import annotations

import os
import re
import threading

from google.cloud import bigquery
from google.oauth2 import service_account

_BQ_IDENTIFIER_RE = re.compile(r"^[a-zA-Z0-9_\-]+$")

_client: bigquery.Client | None = None
_client_lock = threading.Lock()


def _get_config() -> tuple[str, str, str]:
    """Return (project_id, dataset_id, credentials_path)."""
    project = os.environ.get("BQ_PROJECT_ID", "")
    dataset = os.environ.get("BQ_DATASET_ID", "")
    creds_path = os.environ.get("BQ_SA_CREDENTIALS", "")
    if not project or not dataset or not creds_path:
        raise RuntimeError(
            "BigQuery not configured. Set BQ_PROJECT_ID, "
            "BQ_DATASET_ID, and BQ_SA_CREDENTIALS env vars."
        )
    return project, dataset, creds_path


def get_bq_client() -> bigquery.Client:
    """Return a thread-safe singleton BigQuery client."""
    global _client
    if _client is not None:
        return _client

    with _client_lock:
        if _client is not None:
            return _client
        project, _, creds_path = _get_config()
        credentials = service_account.Credentials.from_service_account_file(creds_path)
        _client = bigquery.Client(project=project, credentials=credentials)
        return _client


def get_dataset_ref() -> str:
    """Return the fully-qualified dataset reference: `project.dataset`."""
    project, dataset, _ = _get_config()
    if not _BQ_IDENTIFIER_RE.match(project) or not _BQ_IDENTIFIER_RE.match(dataset):
        raise RuntimeError("Invalid BigQuery project or dataset identifier.")
    return f"{project}.{dataset}"


def run_query(sql: str) -> list[dict]:
    """Execute a SQL query and return results as a list of dicts."""
    client = get_bq_client()
    job_config = bigquery.QueryJobConfig(maximum_bytes_billed=1_000_000_000)
    result = client.query(sql, job_config=job_config).result()
    return [dict(row) for row in result]


def run_query_with_params(
    sql: str, params: list[bigquery.ScalarQueryParameter]
) -> list[dict]:
    """Execute a parameterized SQL query and return results as a list of dicts."""
    client = get_bq_client()
    job_config = bigquery.QueryJobConfig(
        query_parameters=params,
        maximum_bytes_billed=1_000_000_000,
    )
    result = client.query(sql, job_config=job_config).result()
    return [dict(row) for row in result]
