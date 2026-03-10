interface TopBarProps {
  bucketName: string;
  query: string;
  onQueryChange: (q: string) => void;
  onRefresh: () => void;
}

export function TopBar({
  bucketName,
  query,
  onQueryChange,
  onRefresh,
}: TopBarProps) {
  return (
    <header className="topbar">
      <div>
        <h1>Cloud Storage</h1>
        <p>
          Bucket: <strong>{bucketName}</strong>
        </p>
      </div>

      <div className="topbar-controls">
        <input
          type="search"
          placeholder="Buscar nesta pasta"
          value={query}
          onChange={(e) => onQueryChange(e.target.value)}
          aria-label="Buscar nesta pasta"
        />
        <button type="button" onClick={onRefresh}>
          Atualizar
        </button>
      </div>
    </header>
  );
}
