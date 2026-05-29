import type { ZonePolygon } from "../api/types";

// ── Primary mode ──────────────────────────────────────
// Mutually exclusive top-level state — exactly one is active.
// zoneId = 0 means the zone doesn't exist yet (new).
export type EditorMode =
  | { readonly type: "neutral" }
  | { readonly type: "zone_editing"; readonly zoneId: number }
  | { readonly type: "zone_new" }
  | { readonly type: "zone_drawing"; readonly zoneId: number; readonly pointCount: number; readonly awaitingFirst: boolean }
  | { readonly type: "zone_naming" }
  | { readonly type: "lof_new" }
  | { readonly type: "lof_viewing"; readonly lofId: number }
  | { readonly type: "lof_drawing"; readonly pointCount: number }
  | { readonly type: "lof_naming" }
  | { readonly type: "zone_version"; readonly zoneId: number }
  | { readonly type: "lof_version"; readonly lofId: number };

// ── Lock state ────────────────────────────────────────
export type LockState =
  | { readonly type: "unlocked" }
  | { readonly type: "locked_by_me"; readonly expiresAt: string; readonly remainingSeconds: number }
  | { readonly type: "locked_by_other"; readonly editorName: string }
  | { readonly type: "expired" };

// ── Clipboard ─────────────────────────────────────────
export interface ClipboardState {
  readonly polygons: readonly ZonePolygon[];
  readonly sourceZoneId: number | null;
}

// ── Dialogs / overlays ────────────────────────────────
export interface ContextMenuState {
  readonly x: number;
  readonly y: number;
}

export interface DialogsState {
  readonly contextMenu: ContextMenuState | null;
  readonly contextMenuItems: string;
  readonly discardConfirm: boolean;
  readonly helpVisible: boolean;
}

// ── Full editor state ─────────────────────────────────
export interface EditorState {
  readonly mode: EditorMode;
  readonly lock: LockState;
  readonly isDirty: boolean;
  readonly isSaving: boolean;
  readonly clipboard: ClipboardState;
  readonly dialogs: DialogsState;
  readonly versionBrowsing: boolean;
}

// ── Status flags (derived from mode) ──────────────────

export function isZoneActive(mode: EditorMode): boolean {
  switch (mode.type) {
    case "zone_editing":
    case "zone_new":
    case "zone_drawing":
    case "zone_naming":
    case "zone_version":
      return true;
    default:
      return false;
  }
}

export function isLoFActive(mode: EditorMode): boolean {
  switch (mode.type) {
    case "lof_new":
    case "lof_viewing":
    case "lof_drawing":
    case "lof_naming":
    case "lof_version":
      return true;
    default:
      return false;
  }
}

export function isDrawing(mode: EditorMode): boolean {
  return mode.type === "zone_drawing" || mode.type === "lof_drawing";
}

export function isModalOpen(state: EditorState): boolean {
  return (
    state.dialogs.discardConfirm ||
    state.dialogs.contextMenu !== null ||
    state.mode.type === "zone_naming" ||
    state.mode.type === "lof_naming"
  );
}

export function canEdit(mode: EditorMode, lock: LockState): boolean {
  if (mode.type === "zone_version" || mode.type === "lof_version") return false;
  return lock.type === "locked_by_me";
}
