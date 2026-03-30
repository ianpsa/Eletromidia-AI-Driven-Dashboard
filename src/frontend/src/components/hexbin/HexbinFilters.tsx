import { MultiSelect } from "./MultiSelect";
import "./HexbinFilters.css";
import { useGeoFilterOptions } from "../../hooks/useGeoFilterOptions";
import { useHexbinFilters } from "../../hooks/useHexbinFilters";
import type { HexbinFiltersState } from "../../types/hexbin";
import {
  HEXBIN_DISTANCE_MAX,
  HEXBIN_DISTANCE_MIN,
} from "../../utils/hexbinOptions";

type HexbinFiltersProps = {
  onApplyFilters?: (filters: HexbinFiltersState) => void;
  onClearFilters?: () => void;
  initialFilters?: Partial<HexbinFiltersState>;
};

export function HexbinFilters({
  onApplyFilters,
  onClearFilters,
  initialFilters,
}: HexbinFiltersProps) {
  const {
    filters,
    setField,
    hasAnySelection,
    clearFilters,
    hasChangesSinceLastApply,
    markFiltersAsApplied,
  } = useHexbinFilters(initialFilters);

  const { options: remoteOptions } = useGeoFilterOptions({
    selectedState: filters.states[0],
    selectedCity: filters.cities[0],
  });

  const handleApply = () => {
    if (!hasChangesSinceLastApply) return;

    markFiltersAsApplied();

    onApplyFilters?.(filters);
  };

  const handleClear = () => {
    clearFilters();
    onClearFilters?.();
  };

  return (
    <aside className="hexbin-filters">
      <h3 className="hexbin-filters__title">Filtros</h3>

      <MultiSelect
        label="Estado"
        options={remoteOptions.states}
        selected={filters.states}
        onChange={(values) => {
          setField("states", values);
          setField("cities", []);
          setField("addresses", []);
        }}
      />

      <MultiSelect
        label="Cidade"
        options={remoteOptions.cities}
        selected={filters.cities}
        onChange={(values) => {
          setField("cities", values);
          setField("addresses", []);
        }}
      />

      <MultiSelect
        label="Endereço"
        options={remoteOptions.addresses}
        selected={filters.addresses}
        onChange={(values) => setField("addresses", values)}
      />

      <label className="hexbin-filters__field">
        <span>Distância máxima</span>
        <input
          type="range"
          min={HEXBIN_DISTANCE_MIN}
          max={HEXBIN_DISTANCE_MAX}
          step={1}
          value={filters.maxDistance}
          onChange={(e) => setField("maxDistance", Number(e.target.value))}
        />
        <strong>{filters.maxDistance} km</strong>
      </label>

      <MultiSelect
        label="Horário"
        options={remoteOptions.hours}
        selected={filters.hours}
        onChange={(values) => setField("hours", values)}
      />

      <MultiSelect
        label="Gênero"
        options={remoteOptions.genders}
        selected={filters.genders}
        onChange={(values) => setField("genders", values)}
      />

      <MultiSelect
        label="Faixa etária"
        options={remoteOptions.ages}
        selected={filters.ages}
        onChange={(values) => setField("ages", values)}
      />

      <MultiSelect
        label="Classe social"
        options={remoteOptions.socialClasses}
        selected={filters.socialClasses}
        onChange={(values) => setField("socialClasses", values)}
      />

      <div className="hexbin-filters__actions">
        <button
          type="button"
          className="hexbin-filters__primary"
          onClick={handleApply}
          disabled={!hasChangesSinceLastApply}
        >
          Aplicar filtros
        </button>
        <button
          type="button"
          className="hexbin-filters__secondary"
          onClick={handleClear}
          disabled={!hasAnySelection}
        >
          Limpar filtros
        </button>
      </div>
    </aside>
  );
}
