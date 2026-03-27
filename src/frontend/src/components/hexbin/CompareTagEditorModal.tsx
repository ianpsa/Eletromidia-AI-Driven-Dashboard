import { useEffect, useMemo, useState } from "react";
import { useGeoFilterOptions } from "../../hooks/useGeoFilterOptions";
import type { CompareChartsConfig, CompareFilter, CompareMode } from "../../types/hexbin";
import type { CompareTagItem } from "../../utils/hexbinCompareTags";
import { HEXBIN_DISTANCE_MAX, HEXBIN_DISTANCE_MIN } from "../../utils/hexbinOptions";
import { MultiSelect } from "./MultiSelect";
import "./CompareChartsModal.css";

type CompareTagEditorModalProps = {
  open: boolean;
  config: CompareChartsConfig | null;
  targetTag: CompareTagItem | null;
  onClose: () => void;
  onSave: (nextConfig: CompareChartsConfig) => void;
};

function findFilter(config: CompareChartsConfig | null, targetTag: CompareTagItem | null) {
  if (!config || !targetTag || targetTag.kind !== "filter") return null;
  return config.filters.find((filter) => filter.key === targetTag.filterKey) ?? null;
}

export function CompareTagEditorModal({
  open,
  config,
  targetTag,
  onClose,
  onSave,
}: CompareTagEditorModalProps) {
  const filter = useMemo(() => findFilter(config, targetTag), [config, targetTag]);
  const [draftFilter, setDraftFilter] = useState<CompareFilter | null>(null);
  const [draftCompareMode, setDraftCompareMode] = useState<CompareMode | null>(null);

  const selectedState =
    draftFilter && draftFilter.key === "location" ? draftFilter.value.state[0] : undefined;
  const selectedCity =
    draftFilter && draftFilter.key === "location" ? draftFilter.value.city[0] : undefined;
  const { options: remoteOptions } = useGeoFilterOptions({
    selectedState,
    selectedCity,
  });

  useEffect(() => {
    if (!open || !targetTag) return;

    if (targetTag.kind === "filter") {
      setDraftFilter(filter ? { ...filter } : null);
      setDraftCompareMode(null);
      return;
    }

    setDraftFilter(null);
    setDraftCompareMode(config?.compareMode ?? null);
  }, [open, targetTag, filter, config]);

  if (!open || !targetTag || !config) return null;

  const handleSave = () => {
    if (targetTag.kind === "mode") {
      onSave({
        ...config,
        compareMode: draftCompareMode,
      });
      onClose();
      return;
    }

    if (!draftFilter) return;

    onSave({
      ...config,
      filters: config.filters.map((item) =>
        item.key === draftFilter.key ? { ...draftFilter, enabled: true } : item,
      ),
    });
    onClose();
  };

  return (
    <div className="compare-modal__backdrop" onClick={onClose}>
      <div
        className="compare-modal compare-tag-editor-modal"
        onClick={(event) => event.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="compare-tag-editor-title"
      >
        <div className="compare-modal__header">
          <div>
            <h2 id="compare-tag-editor-title">Editar {targetTag.label}</h2>
            <p>Atualize apenas os valores desta tag.</p>
          </div>

          <button className="compare-modal__close" onClick={onClose} type="button">
            ×
          </button>
        </div>

        <div className="compare-modal__body">
          {targetTag.kind === "mode" && (
            <section className="compare-modal__section">
              <h3>Comparar por</h3>
              <div className="compare-tag-editor__mode-options">
                <button
                  className={`compare-modal__tab ${draftCompareMode === "gender" ? "compare-modal__tab--active" : ""}`}
                  onClick={() => setDraftCompareMode("gender")}
                  type="button"
                >
                  Gênero
                </button>
                <button
                  className={`compare-modal__tab ${draftCompareMode === "age" ? "compare-modal__tab--active" : ""}`}
                  onClick={() => setDraftCompareMode("age")}
                  type="button"
                >
                  Idade
                </button>
                <button
                  className={`compare-modal__tab ${draftCompareMode === "socialClass" ? "compare-modal__tab--active" : ""}`}
                  onClick={() => setDraftCompareMode("socialClass")}
                  type="button"
                >
                  Classe
                </button>
              </div>
            </section>
          )}

          {targetTag.kind === "filter" && draftFilter?.key === "location" && (
            <section className="compare-modal__section compare-modal__stack">
              <h3>Valores de localização</h3>
              <MultiSelect
                label="Estados"
                options={remoteOptions.states}
                selected={draftFilter.value.state}
                onChange={(values) =>
                  setDraftFilter((current) =>
                    current && current.key === "location"
                      ? {
                          ...current,
                          value: {
                            ...current.value,
                            state: values,
                            city: [],
                            address: [],
                          },
                        }
                      : current,
                  )
                }
                placeholder="Todos"
              />
              <MultiSelect
                label="Cidades"
                options={remoteOptions.cities}
                selected={draftFilter.value.city}
                onChange={(values) =>
                  setDraftFilter((current) =>
                    current && current.key === "location"
                      ? {
                          ...current,
                          value: {
                            ...current.value,
                            city: values,
                            address: [],
                          },
                        }
                      : current,
                  )
                }
                placeholder="Todos"
              />
              <MultiSelect
                label="Endereços"
                options={remoteOptions.addresses}
                selected={draftFilter.value.address}
                onChange={(values) =>
                  setDraftFilter((current) =>
                    current && current.key === "location"
                      ? { ...current, value: { ...current.value, address: values } }
                      : current,
                  )
                }
                placeholder="Todos"
              />
            </section>
          )}

          {targetTag.kind === "filter" && draftFilter?.key === "hour" && (
            <section className="compare-modal__section">
              <h3>Valores de hora</h3>
              <MultiSelect
                label="Horários"
                options={remoteOptions.hours}
                selected={draftFilter.value}
                onChange={(values) =>
                  setDraftFilter((current) =>
                    current && current.key === "hour" ? { ...current, value: values } : current,
                  )
                }
                placeholder="Todos"
              />
            </section>
          )}

          {targetTag.kind === "filter" && draftFilter?.key === "distance" && (
            <section className="compare-modal__section">
              <h3>Valor de distância</h3>
              <div className="compare-modal__distance">
                <input
                  type="range"
                  min={HEXBIN_DISTANCE_MIN}
                  max={HEXBIN_DISTANCE_MAX}
                  value={draftFilter.value}
                  onChange={(event) =>
                    setDraftFilter((current) =>
                      current && current.key === "distance"
                        ? { ...current, value: Number(event.target.value) }
                        : current,
                    )
                  }
                />
                <strong>{draftFilter.value} km</strong>
              </div>
            </section>
          )}

          {targetTag.kind === "filter" && draftFilter?.key === "gender" && (
            <section className="compare-modal__section">
              <h3>Valores de gênero</h3>
              <MultiSelect
                label="Gêneros"
                options={remoteOptions.genders}
                selected={draftFilter.value}
                onChange={(values) =>
                  setDraftFilter((current) =>
                    current && current.key === "gender" ? { ...current, value: values } : current,
                  )
                }
                placeholder="Todos"
              />
            </section>
          )}

          {targetTag.kind === "filter" && draftFilter?.key === "age" && (
            <section className="compare-modal__section">
              <h3>Valores de idade</h3>
              <MultiSelect
                label="Faixas etárias"
                options={remoteOptions.ages}
                selected={draftFilter.value}
                onChange={(values) =>
                  setDraftFilter((current) =>
                    current && current.key === "age" ? { ...current, value: values } : current,
                  )
                }
                placeholder="Todos"
              />
            </section>
          )}

          {targetTag.kind === "filter" && draftFilter?.key === "socialClass" && (
            <section className="compare-modal__section">
              <h3>Valores de classe social</h3>
              <MultiSelect
                label="Classes sociais"
                options={remoteOptions.socialClasses}
                selected={draftFilter.value}
                onChange={(values) =>
                  setDraftFilter((current) =>
                    current && current.key === "socialClass"
                      ? { ...current, value: values }
                      : current,
                  )
                }
                placeholder="Todos"
              />
            </section>
          )}
        </div>

        <div className="compare-modal__footer">
          <button className="compare-modal__ghost" onClick={onClose} type="button">
            Cancelar
          </button>
          <button className="compare-modal__primary" onClick={handleSave} type="button">
            Salvar tag
          </button>
        </div>
      </div>
    </div>
  );
}
