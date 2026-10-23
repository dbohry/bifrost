import Sortable from 'sortablejs';

export type SortableOptions = {
  /** Turn dragging off, for example while searching. */
  enabled: boolean;
  /** Called with the dragged item's old and new place among the draggable items. */
  onmove: (from: number, to: number) => void;
};

/**
 * Drag-to-reorder for a list of elements marked `data-sortable-item`.
 *
 * Svelte owns the elements, so after a drag the element is put back where
 * Svelte left it and the new order is handed to `onmove`. Svelte then draws
 * the new order itself, which keeps the page and the data in step.
 *
 * The built-in browser drag and drop is not used: it does not work on touch
 * screens. On touch, a drag starts after a short press so the page can still
 * be scrolled.
 */
export function sortable(node: HTMLElement, options: SortableOptions) {
  let current = options;
  let justDragged = false;

  const instance = Sortable.create(node, {
    draggable: '[data-sortable-item]',
    animation: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 150,
    forceFallback: true,
    fallbackTolerance: 4,
    delay: 250,
    delayOnTouchOnly: true,
    touchStartThreshold: 4,
    disabled: !options.enabled,
    ghostClass: 'sortable-ghost',
    chosenClass: 'sortable-chosen',
    onStart() {
      justDragged = true;
    },
    onEnd(evt) {
      // The click that ends a drag must not open the app.
      setTimeout(() => (justDragged = false), 0);

      const { item, from, oldIndex, oldDraggableIndex, newDraggableIndex } = evt;
      if (oldIndex === undefined || oldDraggableIndex === undefined || newDraggableIndex === undefined) return;
      if (oldDraggableIndex === newDraggableIndex) return;

      from.removeChild(item);
      from.insertBefore(item, from.children[oldIndex] ?? null);
      current.onmove(oldDraggableIndex, newDraggableIndex);
    },
  });

  const swallowClick = (e: Event) => {
    if (justDragged) {
      e.preventDefault();
      e.stopPropagation();
    }
  };
  node.addEventListener('click', swallowClick, true);

  return {
    update(next: SortableOptions) {
      current = next;
      instance.option('disabled', !next.enabled);
    },
    destroy() {
      node.removeEventListener('click', swallowClick, true);
      instance.destroy();
    },
  };
}
