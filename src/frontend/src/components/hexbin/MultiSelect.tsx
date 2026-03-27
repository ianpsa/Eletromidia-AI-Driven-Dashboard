import { useEffect, useMemo, useRef, useState } from "react";
import "./MultiSelect.css";

type MultiSelectProps = {
  label: string;
  options: string[];
  selected: string[];
  onChange: (values: string[]) => void;
  placeholder?: string;
};

export function MultiSelect({
  label,
  options,
  selected,
  onChange,
  placeholder = "Todos",
}: MultiSelectProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const handleOutsideClick = (event: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    };

    document.addEventListener("mousedown", handleOutsideClick);
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, []);

  const displayValue = useMemo(() => {
    if (selected.length === 0) return placeholder;
    if (selected.length <= 2) return selected.join(", ");
    return `${selected.slice(0, 2).join(", ")} (+${selected.length - 2})`;
  }, [selected, placeholder]);

  const toggleOption = (value: string) => {
    onChange(
      selected.includes(value)
        ? selected.filter((item) => item !== value)
        : [...selected, value],
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
        <span className={selected.length === 0 ? "is-placeholder" : ""}>
          {displayValue}
        </span>
        <span className="multi-select__arrow">▾</span>
      </button>

      {open && (
        <div className="multi-select__dropdown">
          {options.map((option) => {
            const checked = selected.includes(option);

            return (
              <label
                key={option}
                className={`multi-select__option ${checked ? "is-checked" : ""}`}
              >
                <input
                  type="checkbox"
                  checked={checked}
                  onChange={() => toggleOption(option)}
                />
                <span>{option}</span>
              </label>
            );
          })}
        </div>
      )}
    </div>
  );
}