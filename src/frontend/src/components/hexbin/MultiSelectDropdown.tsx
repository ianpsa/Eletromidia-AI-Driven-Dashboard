import { createPortal } from "react-dom";
import type { MultiSelectDropdownProps } from "../../types/multiSelect";

export function MultiSelectDropdown({
  open,
  options,
  selected,
  onToggleOption,
  dropdownRef,
  positionStyle,
  maxHeight,
  isScrollable,
}: MultiSelectDropdownProps) {
  if (!open || typeof document === "undefined") return null;

  return createPortal(
    <div
      ref={dropdownRef}
      className={`multi-select__dropdown ${isScrollable ? "multi-select__dropdown--scrollable" : ""}`}
      style={{
        ...positionStyle,
        maxHeight,
      }}
    >
      {options.map((option) => {
        const checked = selected.includes(option);

        return (
          <label key={option} className={`multi-select__option ${checked ? "is-checked" : ""}`}>
            <input type="checkbox" checked={checked} onChange={() => onToggleOption(option)} />
            <span>{option}</span>
          </label>
        );
      })}
    </div>,
    document.body,
  );
}
