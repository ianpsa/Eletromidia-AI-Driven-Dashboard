import { MultiSelect } from "./MultiSelect";
import "./CompareChartsModal.css";
import { useGeoFilterOptions } from "../../hooks/useGeoFilterOptions";
import { useHexbinCompareModal } from "../../hooks/useHexbinCompareModal";
import type { CompareChartsConfig } from "../../types/hexbin";
import { canConfirmCompareConfig } from "../../utils/compareValidation";
import { HEXBIN_DISTANCE_MAX, HEXBIN_DISTANCE_MIN } from "../../utils/hexbinOptions";

type CompareChartsModalProps = {
  open: boolean;
  initialConfig?: CompareChartsConfig | null;
  onClose: () => void;
  onConfirm?: (config: CompareChartsConfig) => void;
};

export function CompareChartsModal({
  open,
  initialConfig,
  onClose,
  onConfirm,
}: CompareChartsModalProps) {
  const {
    compareMode,
    setCompareMode,
    filters,
    visibleFilters,
    updateFilter,
    toggleFilter,
  } = useHexbinCompareModal({ open, initialConfig });

  const locationFilter = filters.find((filter) => filter.key === "location");
  const selectedState =
    locationFilter && locationFilter.key === "location"
      ? locationFilter.value.state[0]
      : undefined;
  const selectedCity =
    locationFilter && locationFilter.key === "location"
      ? locationFilter.value.city[0]
      : undefined;

  const { options: remoteOptions } = useGeoFilterOptions({
    selectedState,
    selectedCity,
  });

  const isConfirmDisabled = !canConfirmCompareConfig(compareMode, visibleFilters);

  if (!open) return null;

  function handleConfirm() {
    if (isConfirmDisabled) return;
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

          <button className="compare-modal__close" onClick={onClose} type="button">
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
                          options={remoteOptions.states}
                          selected={filter.value.state}
                          onChange={(values) =>
                            updateFilter("location", {
                              value: {
                                ...filter.value,
                                state: values,
                                city: [],
                                address: [],
                              },
                            })
                          }
                          placeholder="Todos"
                        />

                        <MultiSelect
                          label="Cidades"
                          options={remoteOptions.cities}
                          selected={filter.value.city}
                          onChange={(values) =>
                            updateFilter("location", {
                              value: {
                                ...filter.value,
                                city: values,
                                address: [],
                              },
                            })
                          }
                          placeholder="Todos"
                        />

                        <MultiSelect
                          label="Endereços"
                          options={remoteOptions.addresses}
                          selected={filter.value.address}
                          onChange={(values) =>
                            updateFilter("location", {
                              value: { ...filter.value, address: values },
                            })
                          }
                          placeholder="Todos"
                        />
                      </div>
                    )}

                    {filter.key === "hour" && (
                      <div>
                        <MultiSelect
                          label="Horários"
                          options={remoteOptions.hours}
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
                          min={HEXBIN_DISTANCE_MIN}
                          max={HEXBIN_DISTANCE_MAX}
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
                        options={remoteOptions.genders}
                        selected={filter.value}
                        onChange={(values) => updateFilter("gender", { value: values })}
                        placeholder="Todos"
                      />
                    )}

                    {filter.key === "age" && (
                      <MultiSelect
                        label="Selecionar faixas etárias"
                        options={remoteOptions.ages}
                        selected={filter.value}
                        onChange={(values) => updateFilter("age", { value: values })}
                        placeholder="Todos"
                      />
                    )}

                    {filter.key === "socialClass" && (
                      <MultiSelect
                        label="Selecionar classes sociais"
                        options={remoteOptions.socialClasses}
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
          <button
            className="compare-modal__primary"
            onClick={handleConfirm}
            type="button"
            disabled={isConfirmDisabled}
          >
            Confirmar comparação
          </button>
        </div>
      </div>
    </div>
  );
}