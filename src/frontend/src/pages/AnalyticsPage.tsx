const DEFAULT_LOOKER_URL =
  "https://lookerstudio.google.com/embed/reporting/1776f716-b7de-4268-99ef-8107f950868d";

interface AnalyticsPageProps {
  src?: string;
}

export function AnalyticsPage({ src }: AnalyticsPageProps) {
  return (
    <>
      <div className="looker-header">
        <h1 className="looker-title">Analytics</h1>
        <p className="looker-subtitle">Relatórios e métricas detalhadas</p>
      </div>
      <div className="looker-embed-wrap">
        <iframe
          key={src ?? DEFAULT_LOOKER_URL}
          title="Looker Studio Dashboard"
          src={src ?? DEFAULT_LOOKER_URL}
          className="looker-iframe"
          allowFullScreen
          sandbox="allow-storage-access-by-user-activation allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox"
        />
      </div>
    </>
  );
}
