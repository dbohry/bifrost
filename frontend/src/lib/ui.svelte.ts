/**
 * Things one part of the screen asks another to do. The bottom bar on a phone
 * is outside the dashboard, but its Search and Add buttons act on it.
 */
class Ui {
  /** Set by the bottom bar. The dashboard does it, then clears it. */
  request = $state<'search' | 'add' | null>(null);
  /** A sentence for screen readers, such as where a tile was moved to. */
  announcement = $state('');
}

export const ui = new Ui();
