export function LookerDashboardPage() {
  return (
    <div className="content looker-content">
      <div className="looker-header">
        <h1 className="looker-title">Dashboard</h1>
        <p className="looker-subtitle">Visão geral de indicadores</p>
      </div>
      <div className="looker-embed-wrap">
        <iframe
          title="Looker Studio Dashboard"
          src="https://lookerstudio.google.com/embed/reporting/1776f716-b7de-4268-99ef-8107f950868d"
          className="looker-iframe"
          allowFullScreen
          sandbox="allow-storage-access-by-user-activation allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox"
        />
      </div>
    </div>
  );
}
