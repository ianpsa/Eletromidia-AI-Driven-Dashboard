import { useMemo, useState } from "react";
import { MultiSelect } from "./MultiSelect";
import "./CompareChartsModal.css";

type CompareMode = "gender" | "age" | "socialClass";

type CompareFilterKey = "location" | "hour" | "distance" | "gender" | "age" | "socialClass";

type LocationValue = {
  state: string[];
  city: string[];
  address: string[];
};

type CompareFilter =
  | {
      key: "location";
      label: string;
      enabled: boolean;
      value: LocationValue;
    }
  | {
      key: "hour";
      label: string;
      enabled: boolean;
      value: string[];
    }
  | {
      key: "distance";
      label: string;
      enabled: boolean;
      value: number;
    }
  | {
      key: "gender" | "age" | "socialClass";
      label: string;
      enabled: boolean;
      value: string[];
    };

type CompareChartsModalProps = {
  open: boolean;
  onClose: () => void;
  onConfirm?: (config: {
    compareMode: CompareMode | null;
    filters: CompareFilter[];
  }) => void;
};

const initialFilters: CompareFilter[] = [
  {
    key: "location",
    label: "Localização",
    enabled: false,
    value: {
      state: [],
      city: [],
      address: [],
    },
  },
  { key: "hour", label: "Hora", enabled: false, value: [] },
  { key: "distance", label: "Distância", enabled: false, value: 10 },
  { key: "gender", label: "Gênero", enabled: false, value: [] },
  { key: "age", label: "Idade", enabled: false, value: [] },
  { key: "socialClass", label: "Classe social", enabled: false, value: [] },
];

const STATE_OPTIONS = ["SP", "RJ", "MG", "PR"];
const CITY_OPTIONS = ["São Paulo", "Campinas", "Santos", "Guarulhos", "Osasco"];
const ADDRESS_OPTIONS = ["Rua M.M.D.C", "Avenida Paulista", "Rua da Consolação", "Rua Vergueiro"];
const HOUR_OPTIONS = ["00:00 - 06:00", "06:00 - 12:00", "12:00 - 18:00", "18:00 - 24:00"];
const GENDER_OPTIONS = ["Feminino", "Masculino", "Outro"];
const AGE_OPTIONS = ["18-19", "20-29", "30-39", "40-49", "50+"];
const CLASS_OPTIONS = ["Classe A/B", "Classe C", "Classe D/E"];

export function CompareChartsModal({
  open,
  onClose,
  onConfirm,
}: CompareChartsModalProps) {
  const [compareMode, setCompareMode] = useState<CompareMode | null>(null);
  const [filters, setFilters] = useState<CompareFilter[]>(initialFilters);

  const visibleFilters = useMemo(() => {
    if (!compareMode) return filters;
    return filters.filter((filter) => filter.key !== compareMode);
  }, [filters, compareMode]);

  if (!open) return null;

  function updateFilter(key: CompareFilterKey, patch: Partial<CompareFilter>) {
    setFilters((current) =>
      current.map((item) => (item.key === key ? ({ ...item, ...patch } as CompareFilter) : item)),
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
                type="button"
              >
                Gênero
              </button>
              <button
                className={`compare-modal__tab ${compareMode === "age" ? "compare-modal__tab--active" : ""}`}
                onClick={() => setCompareMode("age")}
                type="button"
              >
                Idade
              </button>
              <button
                className={`compare-modal__tab ${compareMode === "socialClass" ? "compare-modal__tab--active" : ""}`}
                onClick={() => setCompareMode("socialClass")}
                type="button"
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
              {visibleFilters.map((filter) => (
                <div key={filter.key} className="compare-modal__filter-card">
                  <label className="compare-modal__filter-toggle">
                    <input
                      type="checkbox"
                      checked={filter.enabled}
                      onChange={() => toggleFilter(filter.key)}
                    />
                    <span>{filter.label}</span>
                  </label>

                  <div
                    className={`compare-modal__filter-body ${filter.enabled ? "compare-modal__filter-body--active" : ""}`}
                  >
                    {filter.key === "location" && (
                      <div className="compare-modal__stack">
                        <MultiSelect
                          label="Estados"
                          options={STATE_OPTIONS}
                          selected={filter.value.state}
                          onChange={(values) =>
                            updateFilter("location", { value: { ...filter.value, state: values } })
                          }
                          placeholder="Todos"
                        />

                        <MultiSelect
                          label="Cidades"
                          options={CITY_OPTIONS}
                          selected={filter.value.city}
                          onChange={(values) =>
                            updateFilter("location", { value: { ...filter.value, city: values } })
                          }
                          placeholder="Todos"
                        />

                        <MultiSelect
                          label="Endereços"
                          options={ADDRESS_OPTIONS}
                          selected={filter.value.address}
                          onChange={(values) =>
                            updateFilter("location", { value: { ...filter.value, address: values } })
                          }
                          placeholder="Todos"
                        />
                      </div>
                    )}

                    {filter.key === "hour" && (
                      <div>
                        <MultiSelect
                          label="Horários"
                          options={HOUR_OPTIONS}
                          selected={filter.value}
                          onChange={(values) => updateFilter("hour", { value: values })}
                          placeholder="Todos"
                        />
                      </div>
                    )}

                    {filter.key === "distance" && (
                      <div className="compare-modal__distance">
                        <input
                          type="range"
                          min={2}
                          max={15}
                          value={filter.value}
                          onChange={(e) =>
                            updateFilter("distance", { value: Number(e.target.value) })
                          }
                        />
                        <strong>{filter.value} km</strong>
                      </div>
                    )}

                    {filter.key === "gender" && (
                      <MultiSelect
                        label="Selecionar gêneros"
                        options={GENDER_OPTIONS}
                        selected={filter.value}
                        onChange={(values) => updateFilter("gender", { value: values })}
                        placeholder="Todos"
                      />
                    )}

                    {filter.key === "age" && (
                      <MultiSelect
                        label="Selecionar faixas etárias"
                        options={AGE_OPTIONS}
                        selected={filter.value}
                        onChange={(values) => updateFilter("age", { value: values })}
                        placeholder="Todos"
                      />
                    )}

                    {filter.key === "socialClass" && (
                      <MultiSelect
                        label="Selecionar classes sociais"
                        options={CLASS_OPTIONS}
                        selected={filter.value}
                        onChange={(values) => updateFilter("socialClass", { value: values })}
                        placeholder="Todos"
                      />
                    )}
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>

        <div className="compare-modal__footer">
          <button className="compare-modal__ghost" onClick={onClose} type="button">
            Cancelar
          </button>
          <button className="compare-modal__primary" onClick={handleConfirm} type="button">
            Confirmar comparação
          </button>
        </div>
      </div>
    </div>
  );
}