import { useState } from "react";
import "./CompareChartsModal.css";

type CompareMode = "gender" | "age" | "socialClass";

type CompareFilterKey = "location" | "hour" | "distance" | "gender" | "age" | "socialClass";

type CompareFilter = {
  key: CompareFilterKey;
  label: string;
  enabled: boolean;
  value: string;
};

type CompareChartsModalProps = {
  open: boolean;
  onClose: () => void;
  onConfirm?: (config: {
    compareMode: CompareMode;
    filters: CompareFilter[];
  }) => void;
};

const initialFilters: CompareFilter[] = [
  { key: "location", label: "Localização", enabled: false, value: "SP / São Paulo / Rua M.M.D.C" },
  { key: "hour", label: "Hora", enabled: false, value: "" },
  { key: "distance", label: "Distância", enabled: false, value: "10 km" },
  { key: "gender", label: "Gênero", enabled: false, value: "Todos" },
  { key: "age", label: "Idade", enabled: false, value: "Todos" },
  { key: "socialClass", label: "Classe social", enabled: false, value: "Todos" },
];



export function CompareChartsModal({
  open,
  onClose,
  onConfirm,
}: CompareChartsModalProps) {
  const [compareMode, setCompareMode] = useState<CompareMode>("gender");
  const [filters, setFilters] = useState<CompareFilter[]>(initialFilters);

  

  if (!open) return null;

  function updateFilter(key: CompareFilterKey, patch: Partial<CompareFilter>) {
    setFilters((current) =>
      current.map((item) =>
        item.key === key ? { ...item, ...patch } : item,
      ),
    );
  }

  function toggleFilter(key: CompareFilterKey) {
    setFilters((current) =>
      current.map((item) =>
        item.key === key ? { ...item, enabled: !item.enabled } : item,
      ),
    );
  }

  function handleConfirm() {
    onConfirm?.({ compareMode, filters });
    onClose();
  }

  

  return (
    <div className="compare-modal__backdrop" onClick={onClose}>
      <div
        className="compare-modal"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="compare-modal-title"
      >
        <div className="compare-modal__header">
          <div>
            <h2 id="compare-modal-title">Comparar gráficos</h2>
          </div>

          <button className="compare-modal__close" onClick={onClose}>
            ×
          </button>
        </div>

        <div className="compare-modal__body">
          <section className="compare-modal__section">
            <h3>Comparar por</h3>

            <div className="compare-modal__tabs">
              <button
                className={`compare-modal__tab ${compareMode === "gender" ? "compare-modal__tab--active" : ""}`}
                onClick={() => setCompareMode("gender")}
              >
                Gênero
              </button>
              <button
                className={`compare-modal__tab ${compareMode === "age" ? "compare-modal__tab--active" : ""}`}
                onClick={() => setCompareMode("age")}
              >
                Idade
              </button>
              <button
                className={`compare-modal__tab ${compareMode === "socialClass" ? "compare-modal__tab--active" : ""}`}
                onClick={() => setCompareMode("socialClass")}
              >
                Classe
              </button>
            </div>

          </section>

          <section className="compare-modal__section">
            <div className="compare-modal__section-head">
              <h3>Filtros extras</h3>
              <p>Ative os filtros que vão compor a comparação futura.</p>
            </div>

            <div className="compare-modal__filters">
              {filters
                .filter((f) => f.key !== compareMode) // não mostrar o filtro que já está sendo usado na comparação
                .map((filter) => (
                <div key={filter.key} className="compare-modal__filter-card">
                  <label className="compare-modal__filter-toggle">
                    <input
                      type="checkbox"
                      checked={filter.enabled}
                      onChange={() => toggleFilter(filter.key)}
                    />
                    <span>{filter.label}</span>
                  </label>

                  <div className={`compare-modal__filter-body ${filter.enabled ? "compare-modal__filter-body--active" : ""}`}>
                    {filter.key === "location" && (
                      <div className="compare-modal__stack">
                        <select
                          value={filter.value.split(" / ")[0]}
                          onChange={(e) =>
                            updateFilter(filter.key, {
                              value: `${e.target.value} / São Paulo / Rua M.M.D.C`,
                            })
                          }
                        >
                          <option>SP</option>
                          <option>RJ</option>
                          <option>MG</option>
                          <option>PR</option>
                        </select>

                        <select
                          value="São Paulo"
                          onChange={(e) =>
                            updateFilter(filter.key, {
                              value: `SP / ${e.target.value} / Rua M.M.D.C`,
                            })
                          }
                        >
                          <option>São Paulo</option>
                          <option>Campinas</option>
                          <option>Santos</option>
                          <option>Guarulhos</option>
                          <option>Osasco</option>
                        </select>

                        <select
                          value="Rua M.M.D.C"
                          onChange={(e) =>
                            updateFilter(filter.key, {
                              value: `SP / São Paulo / ${e.target.value}`,
                            })
                          }
                        >
                          <option>Rua M.M.D.C</option>
                          <option>Avenida Paulista</option>
                          <option>Rua da Consolação</option>
                          <option>Rua Vergueiro</option>
                        </select>
                      </div>
                    )}

                    {filter.key === "hour" && (
                      <div>
                        <select
                          value={filter.value}
                          onChange={(e) => updateFilter(filter.key, { value: e.target.value })}
                        >
                          <option value="">Todos</option>
                          <option value="0-6">00:00 - 06:00</option>
                          <option value="6-12">06:00 - 12:00</option>
                          <option value="12-18">12:00 - 18:00</option>
                          <option value="18-24">18:00 - 24:00</option>
                        </select>
                      </div>
                    )}

                    {filter.key === "distance" && (
                      <div className="compare-modal__distance">
                        <input type="range" min={2} max={15} defaultValue={10} />
                        <strong>10 km</strong>
                      </div>
                    )}

                    {filter.key === "gender" && (
                      <select defaultValue="Todos">
                        <option>Todos</option>
                        <option>Feminino</option>
                        <option>Masculino</option>
                        <option>Outro</option>
                      </select>
                    )}

                    {filter.key === "age" && (
                      <select defaultValue="Todos">
                        <option>Todos</option>
                        <option>18-19</option>
                        <option>20-29</option>
                        <option>30-39</option>
                        <option>40-49</option>
                        <option>50+</option>
                      </select>
                    )}

                    {filter.key === "socialClass" && (
                      <select defaultValue="Todos">
                        <option>Todos</option>
                        <option>Classe A/B</option>
                        <option>Classe C</option>
                        <option>Classe D/E</option>
                      </select>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>

        <div className="compare-modal__footer">
          <button className="compare-modal__ghost" onClick={onClose}>
            Cancelar
          </button>
          <button className="compare-modal__primary" onClick={handleConfirm}>
            Confirmar comparação
          </button>
        </div>
      </div>
    </div>
  );
}