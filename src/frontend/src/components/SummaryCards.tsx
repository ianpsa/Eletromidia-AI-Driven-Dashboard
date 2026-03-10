import type { ObjectItem } from "../types/api";
import { formatBytes } from "../utils/format";

interface SummaryCardsProps {
  folders: string[];
  files: ObjectItem[];
}

export function SummaryCards({ folders, files }: SummaryCardsProps) {
  const totalSize = files.reduce((sum, f) => sum + f.size, 0);

  return (
    <section className="summary-cards" aria-label="Resumo">
      <article>
        <span>Pastas</span>
        <strong>{folders.length}</strong>
      </article>
      <article>
        <span>Arquivos</span>
        <strong>{files.length}</strong>
      </article>
      <article>
        <span>Tamanho total</span>
        <strong>{formatBytes(totalSize)}</strong>
      </article>
    </section>
  );
}
