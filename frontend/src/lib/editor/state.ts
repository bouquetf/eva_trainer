import { describeTransition } from "./machine";
import type { EditorState, EditorMode, LockState, ClipboardState } from "./types";

// ── Factory: initial state ────────────────────────────

export function createInitialState(): EditorState {
  return {
    mode: { type: "neutral" },
    lock: { type: "unlocked" },
    isDirty: false,
    isSaving: false,
    clipboard: { polygons: [], sourceZoneId: null },
    dialogs: {
      contextMenu: null,
      contextMenuItems: "",
      discardConfirm: false,
      helpVisible: false,
    },
    versionBrowsing: false,
  };
}

// ── Actions ───────────────────────────────────────────

export type EditorAction =
  | { type: "SET_MODE"; mode: EditorMode }
  | { type: "SELECT_ZONE"; zoneId: number }
  | { type: "DESELECT_ALL" }
  | { type: "CREATE_NEW_ZONE" }
  | { type: "CREATE_NEW_LOF" }
  | { type: "MARK_DIRTY" }
  | { type: "MARK_CLEAN" }
  | { type: "START_SAVING" }
  | { type: "STOP_SAVING" }
  | { type: "FIRST_POLYGON_POINT"; zoneId: number }
  | { type: "START_DRAWING"; zoneId: number; awaitingFirst?: boolean }
  | { type: "ADD_POLYGON_POINT" }
  | { type: "FINISH_DRAWING" }
  | { type: "SHOW_ZONE_NAME_DIALOG" }
  | { type: "HIDE_ZONE_NAME_DIALOG" }
  | { type: "SELECT_LOF"; lofId: number }
  | { type: "START_LOF_DRAWING" }
  | { type: "ADD_LOF_POINT" }
  | { type: "FINISH_LOF_DRAWING" }
  | { type: "SHOW_LOF_NAME_DIALOG" }
  | { type: "HIDE_LOF_NAME_DIALOG" }
  | { type: "SET_LOCK"; lock: LockState }
  | { type: "SET_CLIPBOARD"; clipboard: ClipboardState }
  | { type: "CLEAR_CLIPBOARD" }
  | { type: "OPEN_CONTEXT_MENU"; x: number; y: number; itemsLabel: string }
  | { type: "CLOSE_CONTEXT_MENU" }
  | { type: "SHOW_DISCARD_DIALOG" }
  | { type: "HIDE_DISCARD_DIALOG" }
  | { type: "TOGGLE_HELP" }
  | { type: "SHOW_HELP" }
  | { type: "HIDE_HELP" }
  | { type: "ENTER_VERSION_BROWSING" }
  | { type: "EXIT_VERSION_BROWSING" };

// ── Guards (return false if transition is invalid) ────

export function canSelectZone(state: EditorState, _zoneId: number): boolean {
  return !(state.mode.type === "zone_version" || state.mode.type === "lof_version");
}

export function canCreateZone(state: EditorState): boolean {
  if (state.mode.type === "zone_version" || state.mode.type === "lof_version") return false;
  return state.lock.type === "locked_by_me";
}

export function canDraw(state: EditorState): boolean {
  if (state.mode.type === "lof_version" || state.mode.type === "zone_version") return false;
  return state.lock.type === "locked_by_me";
}

export function canSelectLoF(state: EditorState, _lofId: number): boolean {
  return !(state.mode.type === "zone_version" || state.mode.type === "lof_version");
}

// ── Reducer ───────────────────────────────────────────

export function editorReducer(state: EditorState, action: EditorAction): EditorState {
  switch (action.type) {
    // ── Zone selection ──
    case "SELECT_ZONE":
      if (!canSelectZone(state, action.zoneId)) return state;
      return {
        ...state,
        mode: { type: "zone_editing", zoneId: action.zoneId },
        isDirty: false,
        dialogs: { ...state.dialogs, contextMenu: null },
      };

    case "DESELECT_ALL":
      return {
        ...state,
        mode: { type: "neutral" },
        isDirty: false,
        dialogs: { ...state.dialogs, contextMenu: null },
      };

    case "CREATE_NEW_ZONE":
      if (!canCreateZone(state)) return state;
      return {
        ...state,
        mode: { type: "zone_new" },
        isDirty: false,
        dialogs: { ...state.dialogs, contextMenu: null },
      };

    case "CREATE_NEW_LOF":
      if (!canDraw(state)) return state;
      return {
        ...state,
        mode: { type: "lof_new" },
        isDirty: false,
        dialogs: { ...state.dialogs, contextMenu: null },
      };

    case "MARK_DIRTY":
      return { ...state, isDirty: true };

    case "MARK_CLEAN":
      return { ...state, isDirty: false };

    case "START_SAVING":
      return { ...state, isSaving: true };

    case "STOP_SAVING":
      return { ...state, isSaving: false };

    case "FIRST_POLYGON_POINT": {
      if (state.mode.type === "zone_drawing" || state.mode.type === "lof_drawing") return state;
      if (!canDraw(state)) return state;
      return {
        ...state,
        mode: {
          type: "zone_drawing",
          zoneId: action.zoneId,
          pointCount: 1,
          awaitingFirst: false,
        },
      };
    }

    case "START_DRAWING":
      if (!canDraw(state)) return state;
      return {
        ...state,
        mode: {
          type: "zone_drawing",
          zoneId: action.zoneId,
          pointCount: 0,
          awaitingFirst: action.awaitingFirst ?? false,
        },
      };

    case "ADD_POLYGON_POINT": {
      const m = state.mode;
      if (m.type !== "zone_drawing") return state;
      return {
        ...state,
        mode: { ...m, pointCount: m.pointCount + 1, awaitingFirst: false },
      };
    }

    case "FINISH_DRAWING": {
      const m = state.mode;
      if (m.type !== "zone_drawing") return state;
      // First polygon of a brand new zone → open naming dialog
      if (!state.isDirty && m.zoneId === 0) {
        return {
          ...state,
          mode: { type: "zone_naming" },
        };
      }
      return {
        ...state,
        mode: { type: "zone_editing", zoneId: m.zoneId },
        isDirty: true,
      };
    }

    case "SHOW_ZONE_NAME_DIALOG":
      return {
        ...state,
        mode: { type: "zone_naming" },
      };

    case "HIDE_ZONE_NAME_DIALOG":
      return {
        ...state,
        mode: { type: "zone_new" },
      };

    // ── LoF lifecycle ──
    case "SELECT_LOF":
      if (!canSelectLoF(state, action.lofId)) return state;
      return {
        ...state,
        mode: { type: "lof_viewing", lofId: action.lofId },
        isDirty: false,
        dialogs: { ...state.dialogs, contextMenu: null },
      };

    case "START_LOF_DRAWING":
      if (!canDraw(state)) return state;
      return {
        ...state,
        mode: { type: "lof_drawing", pointCount: 0 },
      };

    case "ADD_LOF_POINT": {
      const m = state.mode;
      if (m.type !== "lof_drawing") return state;
      return {
        ...state,
        mode: { ...m, pointCount: m.pointCount + 1 },
      };
    }

    case "FINISH_LOF_DRAWING":
      return {
        ...state,
        mode: { type: "lof_naming" },
        isDirty: true,
      };

    case "SHOW_LOF_NAME_DIALOG":
      return {
        ...state,
        mode: { type: "lof_naming" },
      };

    case "HIDE_LOF_NAME_DIALOG":
      return {
        ...state,
        mode: { type: "neutral" },
      };

    // ── Lock ──
    case "SET_LOCK":
      return { ...state, lock: action.lock };

    // ── Clipboard ──
    case "SET_CLIPBOARD":
      return { ...state, clipboard: action.clipboard };

    case "CLEAR_CLIPBOARD":
      return { ...state, clipboard: { polygons: [], sourceZoneId: null } };

    // ── Dialogs ──
    case "OPEN_CONTEXT_MENU":
      return {
        ...state,
        dialogs: {
          ...state.dialogs,
          contextMenu: { x: action.x, y: action.y },
          contextMenuItems: action.itemsLabel,
        },
      };

    case "CLOSE_CONTEXT_MENU":
      return {
        ...state,
        dialogs: { ...state.dialogs, contextMenu: null, contextMenuItems: "" },
      };

    case "SHOW_DISCARD_DIALOG":
      return { ...state, dialogs: { ...state.dialogs, discardConfirm: true } };

    case "HIDE_DISCARD_DIALOG":
      return { ...state, dialogs: { ...state.dialogs, discardConfirm: false } };

    case "TOGGLE_HELP":
      return { ...state, dialogs: { ...state.dialogs, helpVisible: !state.dialogs.helpVisible } };

    case "SHOW_HELP":
      return { ...state, dialogs: { ...state.dialogs, helpVisible: true } };

    case "HIDE_HELP":
      return { ...state, dialogs: { ...state.dialogs, helpVisible: false } };

    // ── Version browsing ──
    case "ENTER_VERSION_BROWSING":
      return { ...state, versionBrowsing: true };

    case "EXIT_VERSION_BROWSING":
      return { ...state, versionBrowsing: false };

    // ── Direct mode override ──
    case "SET_MODE":
      return { ...state, mode: action.mode };

    default:
      return state;
  }
}

// ── Pure helper: determine context menu items label ──
// (lightweight key so we can diff stale menus without holding JSX)

export function contextMenuKey(mode: EditorMode): string {
  switch (mode.type) {
    case "neutral":
      return "neutral";
    case "zone_drawing":
      return "zone_drawing";
    case "lof_viewing":
      return "lof_viewing";
    default:
      if (mode.type.startsWith("zone_")) return "zone_active";
      if (mode.type.startsWith("lof_")) return "lof_active";
      return "neutral";
  }
}

// ── Transition labels (for debugging / logging) ───────

/** Human-readable description of the transition we are about to take. */
export function transitionLabel(state: EditorState, action: EditorAction): string {
  return describeTransition(state.mode.type, action.type);
}
