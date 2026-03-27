import { useState } from "react";
import { HexbinChart } from "../components/hexbin/HexbinChart";
import { HexbinFilters } from "../components/hexbin/HexbinFilters";
import { CompareChartsModal } from "../components/hexbin/CompareChartsModal";
import { CompareTagEditorModal } from "../components/hexbin/CompareTagEditorModal";
import {
  buildCompareTags,
  type CompareTagItem,
} from "../utils/hexbinCompareTags";
import type { CompareChartsConfig, HexbinFiltersState } from "../types/hexbin";
import { getEmptyCompareConfig } from "../hooks/useHexbinCompareModal";
import {
  DEFAULT_HEXBIN_FILTERS_STATE,
} from "../hooks/useHexbinFilters";
import { useGeoPoints } from "../hooks/useGeoPoints";
import { cloneHexbinFilters } from "../utils/hexbinFilters";

export function DashboardPage() {
  const [compareOpen, setCompareOpen] = useState(false);
  const [compareConfig, setCompareConfig] = useState<CompareChartsConfig | null>(null);
  const [editingTag, setEditingTag] = useState<CompareTagItem | null>(null);
  const [appliedFilters, setAppliedFilters] = useState<HexbinFiltersState>(
    cloneHexbinFilters(DEFAULT_HEXBIN_FILTERS_STATE),
  );
  const {
    points,
    loading: pointsLoading,
    error: pointsError,
  } = useGeoPoints({ filters: appliedFilters });

  const handleCompareConfirm = (config: CompareChartsConfig) => {
    setCompareConfig(config.compareMode ? config : null);
    console.log("config de comparação:", config);
  };

  const handleRemoveFilterTag = (tagToRemove: CompareTagItem) => {
    if (tagToRemove.kind !== "filter") return;

    setCompareConfig((current) => {
      if (!current) return current;

      return {
        ...current,
        filters: current.filters.map((filter) =>
          filter.key === tagToRemove.filterKey ? { ...filter, enabled: false } : filter,
        ),
      };
    });
  };

  const handleRemoveCompareModeTag = () => {
    setCompareConfig(null);
    setEditingTag(null);
  };

  const { filterTags, compareTag } = compareConfig
    ? buildCompareTags(compareConfig)
    : { filterTags: [], compareTag: null };

  const handleOpenTagEditor = (tag: CompareTagItem) => {
    setEditingTag(tag);
  };

  const handleCloseTagEditor = () => {
    setEditingTag(null);
  };

  const handleSaveEditedTag = (nextConfig: CompareChartsConfig) => {
    setCompareConfig(nextConfig.compareMode ? nextConfig : null);
  };

  return (
    <>
      <div className="dashboard-grid dashboard-grid--hexbin">
        <div className="main-column">
          <HexbinChart
            title="Mapa de densidade de pessoas"
            data={points}
            height={520}
          />

          {pointsLoading && <p className="hexbin-map-feedback">Carregando dados do mapa...</p>}
          {!pointsLoading && pointsError && (
            <p className="hexbin-map-feedback hexbin-map-feedback--error">{pointsError}</p>
          )}
        </div>

        <HexbinFilters
          initialFilters={appliedFilters}
          onApplyFilters={(filters) => setAppliedFilters(cloneHexbinFilters(filters))}
          onClearFilters={() =>
            setAppliedFilters(cloneHexbinFilters(DEFAULT_HEXBIN_FILTERS_STATE))
          }
        />
      </div>

      <div className="hexbin-compare-bar">
        <button className="hexbin-compare-button" onClick={() => setCompareOpen(true)}>
          Comparar gráficos
        </button>

        {(filterTags.length > 0 || compareTag) && (
          <div className="hexbin-compare-tags" aria-live="polite">
            {filterTags.map((tag) => (
              <span key={tag.id} className="hexbin-compare-tag">
                <button
                  type="button"
                  className="hexbin-compare-tag__label"
                  onClick={() => handleOpenTagEditor(tag)}
                >
                  {tag.label}
                </button>
                <button
                  type="button"
                  className="hexbin-compare-tag__remove"
                  onClick={() => handleRemoveFilterTag(tag)}
                  aria-label={`Remover tag ${tag.label}`}
                >
                  ×
                </button>
              </span>
            ))}

            {compareTag && (
              <span className="hexbin-compare-tag hexbin-compare-tag--mode">
                <button
                  type="button"
                  className="hexbin-compare-tag__label"
                  onClick={() => handleOpenTagEditor(compareTag)}
                >
                  {compareTag.label}
                </button>
                <button
                  type="button"
                  className="hexbin-compare-tag__remove"
                  onClick={handleRemoveCompareModeTag}
                  aria-label="Remover tag de comparar por"
                >
                  ×
                </button>
              </span>
            )}
          </div>
        )}
      </div>

      <CompareChartsModal
        open={compareOpen}
        initialConfig={compareConfig ?? getEmptyCompareConfig()}
        onClose={() => setCompareOpen(false)}
        onConfirm={handleCompareConfirm}
      />

      <CompareTagEditorModal
        open={Boolean(editingTag)}
        config={compareConfig}
        targetTag={editingTag}
        onClose={handleCloseTagEditor}
        onSave={handleSaveEditedTag}
      />
    </>
  );
}