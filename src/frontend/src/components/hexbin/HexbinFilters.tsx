import { MultiSelect } from "./MultiSelect";
import "./HexbinFilters.css";
import type { HexbinFiltersState } from "../../types/hexbin";
import { useHexbinFilters } from "../../hooks/useHexbinFilters";
import {
  STATE_OPTIONS,
  CITY_OPTIONS,
  ADDRESS_OPTIONS,
  HOUR_OPTIONS,
  GENDER_OPTIONS,
  AGE_OPTIONS,
  CLASS_OPTIONS,
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
  const { filters, setField, hasAnySelection, clearFilters } = useHexbinFilters(initialFilters);

  const handleApply = () => {
    if (onApplyFilters) {
      onApplyFilters(filters);
      return;
    }

    console.log("Filtros aplicados:", filters);
  };

  const handleClear = () => {
    clearFilters();

    if (onClearFilters) {
      onClearFilters();
      return;
    }

    console.log("Filtros limpos");
  };

  return (
    <aside className="hexbin-filters">
      <h3 className="hexbin-filters__title">Filtros</h3>

      <MultiSelect
        label="Estado"
        options={STATE_OPTIONS}
        selected={filters.states}
        onChange={(values) => setField("states", values)}
      />

      <MultiSelect
        label="Cidade"
        options={CITY_OPTIONS}
        selected={filters.cities}
        onChange={(values) => setField("cities", values)}
      />

      <MultiSelect
        label="Endereço"
        options={ADDRESS_OPTIONS}
        selected={filters.addresses}
        onChange={(values) => setField("addresses", values)}
      />

      <label className="hexbin-filters__field">
        <span>Distância máxima</span>
        <input
          type="range"
          min={2}
          max={15}
          value={filters.maxDistance}
          onChange={(e) => setField("maxDistance", Number(e.target.value))}
        />
        <strong>{filters.maxDistance} km</strong>
      </label>

      <MultiSelect
        label="Horário"
        options={HOUR_OPTIONS}
        selected={filters.hours}
        onChange={(values) => setField("hours", values)}
      />

      <MultiSelect
        label="Gênero"
        options={GENDER_OPTIONS}
        selected={filters.genders}
        onChange={(values) => setField("genders", values)}
      />

      <MultiSelect
        label="Faixa etária"
        options={AGE_OPTIONS}
        selected={filters.ages}
        onChange={(values) => setField("ages", values)}
      />

      <MultiSelect
        label="Classe social"
        options={CLASS_OPTIONS}
        selected={filters.socialClasses}
        onChange={(values) => setField("socialClasses", values)}
      />

      <div className="hexbin-filters__actions">
        <button type="button" className="hexbin-filters__primary" onClick={handleApply}>
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