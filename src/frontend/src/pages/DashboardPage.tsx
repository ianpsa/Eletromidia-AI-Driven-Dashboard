import { useMemo, useState } from "react";
import { CompareChartsModal } from "../components/hexbin/CompareChartsModal";
import { CompareTagEditorModal } from "../components/hexbin/CompareTagEditorModal";
import { HexbinChart } from "../components/hexbin/HexbinChart";
import { HexbinFilters } from "../components/hexbin/HexbinFilters";
import { useGeoPoints } from "../hooks/useGeoPoints";
import { getEmptyCompareConfig } from "../hooks/useHexbinCompareModal";
import { DEFAULT_HEXBIN_FILTERS_STATE } from "../hooks/useHexbinFilters";
import type {
  CompareChartsConfig,
  CompareMode,
  HexbinFiltersState,
} from "../types/hexbin";
import {
  buildCompareTags,
  type CompareTagItem,
} from "../utils/hexbinCompareTags";
import { cloneHexbinFilters } from "../utils/hexbinFilters";

type CompareHexbinCardProps = {
  title: string;
  filters: HexbinFiltersState;
};

function CompareHexbinCard({ title, filters }: CompareHexbinCardProps) {
  const { points, loading, error } = useGeoPoints({ filters, limit: 2000 });

  return (
    <div className="hexbin-compare-chart">
      <HexbinChart
        title={title}
        data={points}
        height={360}
        maxDistanceKm={filters.maxDistance}
      />
      {loading && <p className="hexbin-map-feedback">Carregando...</p>}
      {!loading && error && (
        <p className="hexbin-map-feedback hexbin-map-feedback--error">
          {error}
        </p>
      )}
      {!loading && !error && points.length === 0 && (
        <p className="hexbin-map-feedback">
          Sem dados para os filtros deste gráfico.
        </p>
      )}
    </div>
  );
}

export function DashboardPage() {
  const [compareOpen, setCompareOpen] = useState(false);
  const [compareConfig, setCompareConfig] =
    useState<CompareChartsConfig | null>(null);
  const [editingTag, setEditingTag] = useState<CompareTagItem | null>(null);
  const [appliedFilters, setAppliedFilters] = useState<HexbinFiltersState>(
    cloneHexbinFilters(DEFAULT_HEXBIN_FILTERS_STATE),
  );
  const {
    points,
    loading: pointsLoading,
    error: pointsError,
  } = useGeoPoints({ filters: appliedFilters });

  function buildFiltersForGroup(groupValue: string) {
    const next = cloneHexbinFilters(DEFAULT_HEXBIN_FILTERS_STATE);

    if (!compareConfig) return next;

    const compareMode = compareConfig.compareMode;

    for (const filter of compareConfig.filters) {
      if (!filter.enabled) continue;

      switch (filter.key) {
        case "location":
          next.states = [...filter.value.state];
          next.cities = [...filter.value.city];
          next.addresses = [...filter.value.address];
          break;
        case "hour":
          next.hours = [...filter.value];
          break;
        case "distance":
          next.maxDistance = filter.value;
          break;
        case "gender":
          if (compareMode !== "gender") next.genders = [...filter.value];
          break;
        case "age":
          if (compareMode !== "age") next.ages = [...filter.value];
          break;
        case "socialClass":
          if (compareMode !== "socialClass")
            next.socialClasses = [...filter.value];
          break;
      }
    }

    if (compareMode === "gender") next.genders = [groupValue];
    if (compareMode === "age") next.ages = [groupValue];
    if (compareMode === "socialClass") next.socialClasses = [groupValue];

    return next;
  }

  const compareValues = useMemo(() => {
    const compareMode = compareConfig?.compareMode;
    if (!compareMode) return [] as string[];

    const modeFilter = compareConfig.filters.find((f) => f.key === compareMode);
    const selected =
      modeFilter && Array.isArray(modeFilter.value) ? modeFilter.value : [];

    const fallbackByMode: Record<CompareMode, string[]> = {
      gender: ["feminino", "masculino"],
      age: [
        "18-19",
        "20-29",
        "30-39",
        "40-49",
        "50-59",
        "60-69",
        "70-79",
        "80+",
      ],
      socialClass: ["a", "b1", "b2", "c1", "c2", "de"],
    };

    const raw = selected.length > 0 ? selected : fallbackByMode[compareMode];

    if (!raw) return [];

    const unique = Array.from(
      new Set(raw.map((v) => v.trim()).filter((v) => v.length > 0)),
    );
    return compareMode === "gender" ? unique.slice(0, 2) : unique;
  }, [compareConfig]);

  const compareCards = useMemo(() => {
    if (!compareConfig?.compareMode) return [] as CompareHexbinCardProps[];

    return compareValues.map((group) => ({
      title: `${compareConfig.compareMode} • ${group}`,
      filters: buildFiltersForGroup(group),
    }));
  }, [compareConfig, compareValues]);

  const handleCompareConfirm = (config: CompareChartsConfig) => {
    setCompareConfig(config.compareMode ? config : null);
  };

  const handleRemoveFilterTag = (tagToRemove: CompareTagItem) => {
    if (tagToRemove.kind !== "filter") return;

    setCompareConfig((current) => {
      if (!current) return current;

      return {
        ...current,
        filters: current.filters.map((filter) =>
          filter.key === tagToRemove.filterKey
            ? { ...filter, enabled: false }
            : filter,
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
            maxDistanceKm={appliedFilters.maxDistance}
          />

          {pointsLoading && (
            <p className="hexbin-map-feedback">Carregando dados do mapa...</p>
          )}
          {!pointsLoading && pointsError && (
            <p className="hexbin-map-feedback hexbin-map-feedback--error">
              {pointsError}
            </p>
          )}
        </div>

        <HexbinFilters
          initialFilters={appliedFilters}
          onApplyFilters={(filters) =>
            setAppliedFilters(cloneHexbinFilters(filters))
          }
          onClearFilters={() =>
            setAppliedFilters(cloneHexbinFilters(DEFAULT_HEXBIN_FILTERS_STATE))
          }
        />
      </div>

      <div className="hexbin-compare-bar">
        <button
          className="hexbin-compare-button"
          onClick={() => setCompareOpen(true)}
        >
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

      {compareConfig?.compareMode && compareCards.length > 0 && (
        <div className="hexbin-compare-charts">
          <div className="hexbin-compare-grid">
            {compareCards.map((card) => (
              <CompareHexbinCard
                key={card.title}
                title={card.title}
                filters={card.filters}
              />
            ))}
          </div>
        </div>
      )}

      {compareConfig?.compareMode && compareCards.length === 0 && (
        <p className="hexbin-map-feedback">
          Não foi possível montar os grupos de comparação para o modo
          selecionado.
        </p>
      )}
    </>
  );
}
