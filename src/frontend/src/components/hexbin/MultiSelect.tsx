import { useMemo, useRef, useState } from "react";
import "./MultiSelect.css";
import { useOutsideClick } from "../../hooks/useOutsideClick";
import { formatMultiSelectDisplay } from "../../utils/multiSelect";
import type { MultiSelectProps } from "../../types/hexbin";

export function MultiSelect({
  label,
  options,
  selected,
  onChange,
  placeholder = "Todos",
}: MultiSelectProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);

  useOutsideClick(rootRef, () => setOpen(false));

  const displayValue = useMemo(
    () => formatMultiSelectDisplay(selected, placeholder),
    [selected, placeholder],
  );

  const toggleOption = (value: string) => {
    onChange(
      selected.includes(value) ? selected.filter((item) => item !== value) : [...selected, value],
    );
  };

  return (
    <div ref={rootRef} className="multi-select">
      <span className="multi-select__label">{label}</span>

      <button
        type="button"
        className={`multi-select__control ${open ? "is-open" : ""}`}
        onClick={() => setOpen((prev) => !prev)}
      >
        <span className={selected.length === 0 ? "is-placeholder" : ""}>{displayValue}</span>
        <span className="multi-select__arrow">▾</span>
      </button>

      {open && (
        <div className="multi-select__dropdown">
          {options.map((option) => {
            const checked = selected.includes(option);

            return (
              <label key={option} className={`multi-select__option ${checked ? "is-checked" : ""}`}>
                <input type="checkbox" checked={checked} onChange={() => toggleOption(option)} />
                <span>{option}</span>
              </label>
            );
          })}
        </div>
      )}
    </div>
  );
}