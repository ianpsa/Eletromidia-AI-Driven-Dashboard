import { useEffect, type RefObject } from "react";

type OutsideClickOptions = {
  ignoreRefs?: Array<RefObject<HTMLElement>>;
};

export function useOutsideClick<T extends HTMLElement>(
  ref: RefObject<T>,
  handler: (event: MouseEvent) => void,
  options?: OutsideClickOptions,
) {
  useEffect(() => {
    const listener = (event: MouseEvent) => {
      const target = event.target as Node;

      if (!ref.current || ref.current.contains(event.target as Node)) return;

      if (options?.ignoreRefs?.some((ignoredRef) => ignoredRef.current?.contains(target))) {
        return;
      }

      handler(event);
    };

    document.addEventListener("mousedown", listener);
    return () => document.removeEventListener("mousedown", listener);
  }, [ref, handler, options]);
}
