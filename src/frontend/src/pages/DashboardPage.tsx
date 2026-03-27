import { useState } from "react";
import { HexbinChart } from "../components/hexbin/HexbinChart";
import { HexbinFilters } from "../components/hexbin/HexbinFilters";
import { CompareChartsModal } from "../components/hexbin/CompareChartsModal";
import { CompareTagEditorModal } from "../components/hexbin/CompareTagEditorModal";
import { generateMockHexbinData } from "../utils/mockHexbinData";
import {
  buildCompareTags,
  type CompareTagItem,
} from "../utils/hexbinCompareTags";
import type { CompareChartsConfig } from "../types/hexbin";

const data = generateMockHexbinData(4500);

export function DashboardPage() {
  const [compareOpen, setCompareOpen] = useState(false);
  const [compareConfig, setCompareConfig] = useState<CompareChartsConfig | null>(null);
  const [editingTag, setEditingTag] = useState<CompareTagItem | null>(null);

  const handleCompareConfirm = (config: CompareChartsConfig) => {
    setCompareConfig(config);
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
    setCompareConfig((current) =>
      current
        ? {
            ...current,
            compareMode: null,
          }
        : current,
    );
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
    setCompareConfig(nextConfig);
  };

  return (
    <>
      <div className="dashboard-grid dashboard-grid--hexbin">
        <div className="main-column">
          <HexbinChart
            title="Mapa de densidade de pessoas"
            data={data}
            height={520}
          />
        </div>

        <HexbinFilters />
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