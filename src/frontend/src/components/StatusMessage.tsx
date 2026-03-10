interface StatusMessageProps {
  loading: boolean;
  error: string;
  empty: boolean;
  downloadError: string;
}

export function StatusMessage({
  loading,
  error,
  empty,
  downloadError,
}: StatusMessageProps) {
  return (
    <>
      {loading && <p className="status">Carregando...</p>}
      {error && <p className="status status-error">{error}</p>}
      {empty && <p className="status">Nenhum item encontrado.</p>}
      {downloadError && <p className="status status-error">{downloadError}</p>}
    </>
  );
}
