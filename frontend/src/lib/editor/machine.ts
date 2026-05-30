/**
 * ─── EDITOR STATE MACHINE ───────────────────────────────────────
 *
 * States, transitions, and context-menu options — declarative,
 * human-readable.
 *
 * Each entry lists every valid (action → target) from that state.
 * Guards (named) gate transitions in the reducer.
 *
 * Legend:
 *   state
 *     ├ ACTION ───→ target   (guardName)   // guarded transition
 *     └ ACTION ───→ target                 // unguarded
 */

export interface MenuOption {
  readonly label: string;
  readonly action: string;
  readonly target?: string;
  readonly guard?: string;
  readonly description: string;
}

export interface CanvasAction {
  readonly trigger: string;
  readonly action: string;
  readonly target?: string;
  readonly condition?: string;
  readonly description: string;
}

export interface SidePanelAction {
  readonly control: string;
  readonly action: string;
  readonly target?: string;
  readonly guard?: string;
  readonly description: string;
}

export interface StateSpec {
  readonly id: string;
  readonly label: string;
  readonly description: string;
  readonly menuOptions: readonly MenuOption[];
  readonly canvasActions: readonly CanvasAction[];
  readonly sidePanelActions: readonly SidePanelAction[];
}

// ── Machine table ──────────────────────────────────────────────

export const machineStates: readonly StateSpec[] = [
  // ══════════════════════════════════════════════════════════════
  //  NEUTRAL — nothing selected, nothing in progress
  // ══════════════════════════════════════════════════════════════
  {
    id: "neutral",
    label: "Neutral",
    description: "Nothing selected, nothing in progress.",
    menuOptions: [
      { label: "Nouvelle zone",         action: "CREATE_NEW_ZONE", target: "zone_new",    guard: "canCreateZone", description: "Create a brand new zone" },
      { label: "Nouvelle ligne de tir", action: "CREATE_NEW_LOF",  target: "lof_new",     guard: "canDraw",       description: "Create a new line of fire" },
    ],
    canvasActions: [
      { trigger: "Right-click on map", action: "OPEN_CONTEXT_MENU", description: "Show map context menu" },
    ],
    sidePanelActions: [
      { control: "Zone list item", action: "SELECT_ZONE", target: "zone_editing", guard: "canSelectZone", description: "Click a zone in the list to edit it" },
      { control: "LoF list item",  action: "SELECT_LOF",  target: "lof_viewing",  guard: "canSelectLoF",  description: "Click a line of fire in the list to view it" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE EDITING — zone being edited
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_editing",
    label: "Zone editing",
    description: "A zone is selected and being edited.",
    menuOptions: [
      { label: "Nouvelle zone",         action: "CREATE_NEW_ZONE", target: "zone_new",    guard: "canCreateZone", description: "Create a brand new zone" },
      { label: "Nouvelle forme",        action: "START_DRAWING",   target: "zone_drawing", guard: "canDraw",       description: "Start drawing a new polygon on the current zone" },
      { label: "Nouvelle ligne de tir", action: "CREATE_NEW_LOF",  target: "lof_new",     guard: "canDraw",       description: "Create a new line of fire" },
      { label: "Coller",                action: "PASTE_POLYGONS",  target: "zone_editing",  guard: "canPaste",       description: "Paste copied polygons" },
    ],
    canvasActions: [
      { trigger: "Left-click on background",   action: "DESELECT_POLYGONS",                description: "Deselect all polygons, clear temp points and undo stack" },
      { trigger: "Left-click on polygon",      action: "SELECT_POLYGON",                   description: "Select a polygon (replaces selection)" },
      { trigger: "Ctrl+click on polygon",      action: "TOGGLE_POLYGON_SELECTION",         description: "Add/remove polygon from multi-selection" },
      { trigger: "Right-click on polygon",     action: "OPEN_CONTEXT_MENU",              description: "Show polygon editing context menu (copy, paste, floor, flip)" },
      { trigger: "Right-click on map",         action: "OPEN_CONTEXT_MENU",               description: "Show map context menu" },
      { trigger: "Drag zone",                  action: "DRAG_ZONE",                        description: "Move the selected zone to a new position" },
      { trigger: "Double-click on point handle", action: "DELETE_POINT",                   description: "Delete a vertex from the polygon" },
    ],
    sidePanelActions: [
      { control: "Zone list item",         action: "SELECT_ZONE",            target: "zone_editing",  guard: "canSelectZone", description: "Switch to a different zone" },
      { control: "LoF list item",          action: "SELECT_LOF",             target: "lof_viewing",   guard: "canSelectLoF",  description: "Switch to a line of fire" },
      { control: "Delete zone button",     action: "DELETE_ZONE",                                                    description: "Delete the currently selected zone" },
      { control: "Rename zone (double-click)", action: "RENAME_ZONE",                                              description: "Rename the currently selected zone" },
      { control: "Version browser toggle", action: "ENTER_VERSION_BROWSING", target: "zone_version", guard: "canSelectZone", description: "Open version history for the selected zone" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE NEW — new zone being created (awaiting first polygon)
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_new",
    label: "New zone – awaiting first polygon",
    description: "A new zone is being created. Click the map to place the first polygon vertex.",
    menuOptions: [
      { label: "Annuler", action: "DESELECT_ALL", target: "neutral", description: "Cancel creating the new zone" },
    ],
    canvasActions: [
      { trigger: "Left-click on background", action: "FIRST_POLYGON_POINT", target: "zone_drawing", condition: "canDraw", description: "Place the first polygon vertex on the map" },
    ],
    sidePanelActions: [
      { control: "Zone list item",     action: "SELECT_ZONE",     target: "zone_editing", guard: "canSelectZone", description: "Switch to a different zone (discards new zone)" },
      { control: "LoF list item",      action: "SELECT_LOF",      target: "lof_viewing",  guard: "canSelectLoF",  description: "Switch to a line of fire (cancels new zone)" },
      { control: "Cancel zone button", action: "DESELECT_ALL",    target: "neutral",                                 description: "Cancel creating the new zone" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE DRAWING — polygon drawing in progress
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_drawing",
    label: "Drawing polygon",
    description: "A polygon is being drawn on the map.",
    menuOptions: [
      { label: "Terminer la forme", action: "FINISH_DRAWING", target: "zone_editing", description: "Complete the polygon being drawn (3+ vertices)" },
      { label: "Annuler le point",                    action: "REMOVE_LAST_POINT",                          description: "Remove the last placed vertex" },
      { label: "Annuler la forme",                    action: "SET_MODE",         target: "zone_editing",  description: "Cancel drawing and discard the incomplete polygon" },
      { label: "Annuler la forme (premier point)",    action: "DESELECT_ALL",     target: "neutral",       description: "Cancel the new zone entirely" },
    ],
    canvasActions: [
      { trigger: "Left-click on background (awaiting first point)", action: "FIRST_POLYGON_POINT", target: "zone_drawing", description: "Place the first vertex of the polygon" },
      { trigger: "Left-click on background (drawing in progress)",  action: "ADD_POLYGON_POINT",   target: "zone_drawing", description: "Add a vertex to the current polygon" },
      { trigger: "Ctrl+click on background (drawing in progress)",  action: "REMOVE_LAST_POINT",                            description: "Remove the last placed vertex" },
      { trigger: "Left-click on first point",                      action: "FINISH_DRAWING",      target: "zone_editing", description: "Close the polygon back to the first vertex" },
      { trigger: "Right-click",                                    action: "OPEN_CONTEXT_MENU",                                description: "Show drawing context menu (finish, undo point, cancel shape)" },
    ],
    sidePanelActions: [
      { control: "Zone list item", action: "SELECT_ZONE", target: "zone_editing", guard: "canSelectZone", description: "Switch to a different zone (cancels drawing)" },
      { control: "Cancel drawing", action: "DESELECT_ALL", target: "neutral",                           description: "Cancel drawing and discard the incomplete polygon" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE NAMING — naming dialog open for a new zone
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_naming",
    label: "Naming new zone",
    description: "A naming dialog is open for the new zone.",
    menuOptions: [],
    canvasActions: [
      { trigger: "(dialog)", action: "HIDE_ZONE_NAME_DIALOG", target: "zone_new", description: "Confirm or cancel the zone name in the dialog" },
    ],
    sidePanelActions: [
      { control: "Zone list item", action: "SELECT_ZONE", target: "zone_editing", description: "Switch to a different zone (closes naming dialog)" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF NEW — new line of fire being created
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_new",
    label: "New line of fire – awaiting first cone",
    description: "A new line of fire is being created. Click the map to place the first cone.",
    menuOptions: [
      { label: "Annuler", action: "DESELECT_ALL", target: "neutral", description: "Cancel creating the new line of fire" },
    ],
    canvasActions: [
      { trigger: "Left-click on background", action: "START_LOF_DRAWING", target: "lof_drawing", condition: "canDraw", description: "Place the first cone point on the map" },
    ],
    sidePanelActions: [
      { control: "Zone list item", action: "SELECT_ZONE", target: "zone_editing", guard: "canSelectZone", description: "Switch to a zone (cancels new LoF)" },
      { control: "LoF list item",  action: "SELECT_LOF",  target: "lof_viewing",  guard: "canSelectLoF",  description: "Switch to an existing LoF" },
      { control: "Cancel LoF",     action: "DESELECT_ALL", target: "neutral",                             description: "Cancel creating the new line of fire" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF DRAWING — line-of-fire cones being drawn
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_drawing",
    label: "Drawing line-of-fire cones",
    description: "Line-of-fire cones are being placed on the map.",
    menuOptions: [],
    canvasActions: [
      { trigger: "Left-click on background", action: "ADD_LOF_POINT", target: "lof_drawing", description: "Add a cone vertex to the line of fire" },
      { trigger: "Double-click",             action: "FINISH_LOF_DRAWING", target: "lof_naming", description: "Complete the line of fire and open naming dialog" },
    ],
    sidePanelActions: [
      { control: "Zone list item", action: "SELECT_ZONE", target: "zone_editing", guard: "canSelectZone", description: "Switch to a zone (cancels LoF drawing)" },
      { control: "Cancel drawing", action: "DESELECT_ALL", target: "neutral",                           description: "Cancel the current LoF drawing" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF NAMING — naming dialog open for a new LoF
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_naming",
    label: "Naming new line of fire",
    description: "A naming dialog is open for the new line of fire.",
    menuOptions: [],
    canvasActions: [
      { trigger: "(dialog)", action: "HIDE_LOF_NAME_DIALOG", target: "neutral", description: "Confirm or cancel the LoF name in the dialog" },
    ],
    sidePanelActions: [
      { control: "Zone list item", action: "SELECT_ZONE", target: "zone_editing", description: "Switch to a zone (closes naming dialog)" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF VIEWING — LoF selected, displayed
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_viewing",
    label: "Line of fire – viewing",
    description: "A line of fire is selected and displayed on the map.",
    menuOptions: [
      { label: "Nouvelle zone",         action: "CREATE_NEW_ZONE", target: "zone_new",    guard: "canCreateZone", description: "Create a brand new zone" },
      { label: "Nouvelle ligne de tir", action: "CREATE_NEW_LOF",  target: "lof_new",     guard: "canDraw",       description: "Create a new line of fire" },
      { label: "Coller",                action: "PASTE_POLYGONS",  target: "zone_editing",  guard: "canPaste",       description: "Paste copied polygons" },
    ],
    canvasActions: [
      { trigger: "Left-click on background", action: "DESELECT_ALL", target: "neutral", description: "Deselect the line of fire" },
      { trigger: "Right-click on map",       action: "OPEN_CONTEXT_MENU",                  description: "Show map context menu" },
    ],
    sidePanelActions: [
      { control: "Zone list item",         action: "SELECT_ZONE",            target: "zone_editing",  guard: "canSelectZone", description: "Switch to a zone" },
      { control: "LoF list item (different)", action: "SELECT_LOF",          target: "lof_viewing",                            description: "Switch to a different line of fire" },
      { control: "Delete LoF button",      action: "DELETE_LOF",                                                  description: "Delete the currently selected line of fire" },
      { control: "Rename LoF (double-click)", action: "RENAME_LOF",                                              description: "Rename the currently selected line of fire" },
      { control: "Version browser toggle", action: "ENTER_VERSION_BROWSING", target: "lof_version", guard: "canSelectLoF", description: "Open version history for the selected LoF" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE VERSION — browsing zone version history
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_version",
    label: "Browsing zone version history",
    description: "The version history browser is open for a zone.",
    menuOptions: [],
    canvasActions: [
      { trigger: "(version browser UI)", action: "EXIT_VERSION_BROWSING", target: "neutral", description: "Close the version browser" },
    ],
    sidePanelActions: [
      { control: "Version browser back button", action: "EXIT_VERSION_BROWSING", target: "neutral", description: "Return to editing mode from version browser" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF VERSION — browsing LoF version history
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_version",
    label: "Browsing LoF version history",
    description: "The version history browser is open for a line of fire.",
    menuOptions: [],
    canvasActions: [
      { trigger: "(version browser UI)", action: "EXIT_VERSION_BROWSING", target: "neutral", description: "Close the version browser" },
    ],
    sidePanelActions: [
      { control: "Version browser back button", action: "EXIT_VERSION_BROWSING", target: "neutral", description: "Return to editing mode from version browser" },
    ],
  },
] as const;

// ── Lookup helpers ─────────────────────────────────────────────

const stateMap = new Map<string, StateSpec>(
  machineStates.map((s) => [s.id, s]),
);

type TransitionInfo = { action: string; target?: string; guard?: string };

function collectTransitions(spec: StateSpec): TransitionInfo[] {
  const seen = new Set<string>();
  const result: TransitionInfo[] = [];
  const add = (t: TransitionInfo) => {
    const key = t.action + "|" + (t.target ?? "") + "|" + (t.guard ?? "");
    if (!seen.has(key)) { seen.add(key); result.push(t); }
  };
  for (const m of spec.menuOptions) {
    add({ action: m.action, target: m.target, guard: m.guard });
  }
  for (const c of spec.canvasActions) {
    add({ action: c.action, target: c.target });
  }
  for (const s of spec.sidePanelActions) {
    add({ action: s.action, target: s.target, guard: s.guard });
  }
  return result;
}

/** Return the spec for a given state id, or undefined. */
export function getStateSpec(id: string): StateSpec | undefined {
  return stateMap.get(id);
}

/** Find the first matching transition for (stateId × action). */
export function findTransition(
  stateId: string,
  action: string,
): TransitionInfo | undefined {
  const spec = stateMap.get(stateId);
  if (!spec) return undefined;
  return collectTransitions(spec).find((t) => t.action === action);
}

/** True if the action is listed for the state (in any UI list). */
export function isValidTransition(stateId: string, action: string): boolean {
  return findTransition(stateId, action) !== undefined;
}

/** Return the menu options for a given state. */
export function getMenuOptions(stateId: string): readonly MenuOption[] {
  return stateMap.get(stateId)?.menuOptions ?? [];
}

/** Return the canvas actions for a given state. */
export function getCanvasActions(stateId: string): readonly CanvasAction[] {
  return stateMap.get(stateId)?.canvasActions ?? [];
}

/** Return the side-panel actions for a given state. */
export function getSidePanelActions(stateId: string): readonly SidePanelAction[] {
  return stateMap.get(stateId)?.sidePanelActions ?? [];
}

// ── Debug / logging ────────────────────────────────────────────

/**
 * Human-readable label:  "neutral ──SELECT_ZONE──→ zone_editing"
 * Returns "<stateId> cannot handle <action>" when invalid.
 */
export function describeTransition(stateId: string, action: string): string {
  const transition = findTransition(stateId, action);
  if (!transition) return `${stateId} ⚠ cannot handle ${action}`;
  return `${stateId} ──${action}──→ ${transition.target ?? "(stay)"}`;
}

/**
 * Full table dump for debug / visual inspection.
 */
export function dumpMachine(): string {
  const lines: string[] = [];
  for (const state of machineStates) {
    const transitions = collectTransitions(state);
    lines.push(`${state.id}  (${state.description})`);
    for (const t of transitions) {
      const guard = t.guard ? `  (${t.guard})` : "";
      const target = t.target ? `──→ ${t.target}` : "(stay)";
      lines.push(`  ├ ${t.action.padEnd(24)} ${target.padEnd(16)} ${guard}`);
    }
    if (state.menuOptions.length > 0) {
      lines.push(`  ── menu ──`);
      for (const m of state.menuOptions) {
        const guard = m.guard ? ` (${m.guard})` : "";
        const target = m.target ? ` ──→ ${m.target}` : "";
        lines.push(`  • ${m.label.padEnd(35)} ${m.action}${target}${guard}`);
      }
    }
    if (state.canvasActions.length > 0) {
      lines.push(`  ── canvas ──`);
      for (const c of state.canvasActions) {
        const cond = c.condition ? ` (${c.condition})` : "";
        const target = c.target ? ` ──→ ${c.target}` : "";
        lines.push(`  • ${c.trigger.padEnd(50)} ${c.action}${target}${cond}`);
      }
    }
    if (state.sidePanelActions.length > 0) {
      lines.push(`  ── side panel ──`);
      for (const s of state.sidePanelActions) {
        const guard = s.guard ? ` (${s.guard})` : "";
        const target = s.target ? ` ──→ ${s.target}` : "";
        lines.push(`  • ${s.control.padEnd(40)} ${s.action}${target}${guard}`);
      }
    }
    lines.push("");
  }
  return lines.join("\n");
}

/**
 * Full machine dump as an mermaid state diagram (for docs).
 */
export function dumpMermaid(): string {
  const lines: string[] = ["stateDiagram-v2"];
  for (const state of machineStates) {
    for (const t of collectTransitions(state)) {
      if (!t.target) continue;
      const guard = t.guard ? ` |${t.guard}|` : "";
      lines.push(`  ${state.id} --> ${t.target} : ${t.action}${guard}`);
    }
  }
  return lines.join("\n");
}
