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

export interface TransitionSpec {
  readonly action: string;
  readonly target: string;
  readonly guard?: string;
  readonly label: string;
}

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

export interface StateSpec {
  readonly id: string;
  readonly label: string;
  readonly description: string;
  readonly transitions: readonly TransitionSpec[];
  readonly menuOptions: readonly MenuOption[];
  readonly canvasActions: readonly CanvasAction[];
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
    transitions: [
      { action: "SELECT_ZONE",       target: "zone_editing", guard: "canSelectZone", label: "Select existing zone" },
      { action: "SELECT_LOF",        target: "lof_viewing",  guard: "canSelectLoF",  label: "Select existing line of fire" },
      { action: "CREATE_NEW_ZONE",   target: "zone_new",     guard: "canCreateZone", label: "Create new zone" },
      { action: "CREATE_NEW_LOF",    target: "lof_new",      guard: "canDraw",       label: "Create new line of fire" },
    ],
    menuOptions: [
      { label: "Nouvelle zone",         action: "CREATE_NEW_ZONE", target: "zone_new",    guard: "canCreateZone", description: "Create a brand new zone" },
      { label: "Nouvelle ligne de tir", action: "CREATE_NEW_LOF",  target: "lof_new",     guard: "canDraw",       description: "Create a new line of fire" },
      { label: "Coller",                action: "PASTE_POLYGONS",  target: "zone_editing",                           description: "Paste copied polygons" },
    ],
    canvasActions: [
      { trigger: "Left-click on existing zone polygon", action: "SELECT_ZONE", target: "zone_editing", description: "Select and enter editing for the clicked zone" },
      { trigger: "Right-click on map",                  action: "(context menu)",                    description: "Show map context menu" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE EDITING — zone being edited
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_editing",
    label: "Zone editing",
    description: "A zone is selected and being edited.",
    transitions: [
      { action: "START_DRAWING",     target: "zone_drawing",  guard: "canDraw",           label: "Add / edit polygon shape" },
      { action: "DESELECT_ALL",      target: "neutral",                                   label: "Deselect zone" },
    ],
    menuOptions: [
      { label: "Nouvelle zone",         action: "CREATE_NEW_ZONE", target: "zone_new",    guard: "canCreateZone", description: "Create a brand new zone" },
      { label: "Nouvelle forme",        action: "START_DRAWING",   target: "zone_drawing", guard: "canDraw",       description: "Start drawing a new polygon on the current zone" },
      { label: "Nouvelle ligne de tir", action: "CREATE_NEW_LOF",  target: "lof_new",     guard: "canDraw",       description: "Create a new line of fire" },
      { label: "Coller",                action: "PASTE_POLYGONS",  target: "zone_editing",                           description: "Paste copied polygons" },
    ],
    canvasActions: [
      { trigger: "Left-click on background",   action: "DESELECT_POLYGONS",                description: "Deselect all polygons, clear temp points and undo stack" },
      { trigger: "Left-click on polygon",      action: "SELECT_POLYGON",                   description: "Select a polygon (replaces selection)" },
      { trigger: "Ctrl+click on polygon",      action: "TOGGLE_POLYGON_SELECTION",         description: "Add/remove polygon from multi-selection" },
      { trigger: "Right-click on polygon",     action: "(zone context menu)",              description: "Show polygon editing context menu (copy, paste, floor, flip)" },
      { trigger: "Right-click on map",         action: "(map context menu)",               description: "Show map context menu" },
      { trigger: "Drag zone",                  action: "DRAG_ZONE",                        description: "Move the selected zone to a new position" },
      { trigger: "Double-click on point handle", action: "DELETE_POINT",                   description: "Delete a vertex from the polygon" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE NEW — new zone being created (awaiting first polygon)
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_new",
    label: "New zone – awaiting first polygon",
    description: "A new zone is being created. Click the map to place the first polygon vertex.",
    transitions: [
      { action: "FIRST_POLYGON_POINT",  target: "zone_drawing", guard: "canDraw",         label: "Place first vertex" },
      { action: "SHOW_ZONE_NAME_DIALOG", target: "zone_naming",                           label: "Rename immediately (no polygon)" },
      { action: "DESELECT_ALL",         target: "neutral",                                label: "Cancel new zone" },
    ],
    menuOptions: [
      { label: "Annuler", action: "DESELECT_ALL", target: "neutral", description: "Cancel creating the new zone" },
    ],
    canvasActions: [
      { trigger: "Left-click on background", action: "FIRST_POLYGON_POINT", target: "zone_drawing", condition: "canDraw", description: "Place the first polygon vertex on the map" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE DRAWING — polygon drawing in progress
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_drawing",
    label: "Drawing polygon",
    description: "A polygon is being drawn on the map.",
    transitions: [
      { action: "ADD_POLYGON_POINT", target: "zone_drawing",                              label: "Add vertex to current polygon" },
      { action: "FIRST_POLYGON_POINT", target: "zone_drawing", guard: "canDraw",          label: "Start next polygon (new shape)" },
      { action: "FINISH_DRAWING",    target: "zone_editing",                              label: "Complete polygon (existing zone)" },
      { action: "FINISH_DRAWING",    target: "zone_naming",                               label: "Complete polygon (new zone – name it)" },
    ],
    menuOptions: [
      { label: "Terminer la zone / Terminer la forme", action: "FINISH_DRAWING", target: "zone_editing", description: "Complete the polygon being drawn (3+ vertices)" },
      { label: "Annuler le point",                    action: "REMOVE_LAST_POINT",                          description: "Remove the last placed vertex" },
      { label: "Annuler la forme",                    action: "SET_MODE",         target: "zone_editing",  description: "Cancel drawing and discard the incomplete polygon" },
      { label: "Annuler la forme (premier point)",    action: "DESELECT_ALL",     target: "neutral",       description: "Cancel the new zone entirely" },
    ],
    canvasActions: [
      { trigger: "Left-click on background (awaiting first point)", action: "FIRST_POLYGON_POINT", target: "zone_drawing", description: "Place the first vertex of the polygon" },
      { trigger: "Left-click on background (drawing in progress)",  action: "ADD_POLYGON_POINT",   target: "zone_drawing", description: "Add a vertex to the current polygon" },
      { trigger: "Ctrl+click on background (drawing in progress)",  action: "REMOVE_LAST_POINT",                            description: "Remove the last placed vertex" },
      { trigger: "Left-click on first point",                      action: "FINISH_DRAWING",      target: "zone_editing", description: "Close the polygon back to the first vertex" },
      { trigger: "Right-click",                                    action: "(context menu)",                                description: "Show drawing context menu (finish, undo point, cancel shape)" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE NAMING — naming dialog open for a new zone
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_naming",
    label: "Naming new zone",
    description: "A naming dialog is open for the new zone.",
    transitions: [
      { action: "HIDE_ZONE_NAME_DIALOG", target: "zone_new",                              label: "Close naming dialog" },
    ],
    menuOptions: [],
    canvasActions: [
      { trigger: "(dialog)", action: "CONFIRM_NAME / HIDE_ZONE_NAME_DIALOG", target: "zone_new", description: "Confirm or cancel the zone name in the dialog" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF NEW — new line of fire being created
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_new",
    label: "New line of fire – awaiting first cone",
    description: "A new line of fire is being created. Click the map to place the first cone.",
    transitions: [
      { action: "START_LOF_DRAWING", target: "lof_drawing", guard: "canDraw",             label: "Start drawing cones" },
      { action: "DESELECT_ALL",      target: "neutral",                                   label: "Cancel new LoF" },
    ],
    menuOptions: [
      { label: "Annuler", action: "DESELECT_ALL", target: "neutral", description: "Cancel creating the new line of fire" },
    ],
    canvasActions: [
      { trigger: "Left-click on background", action: "START_LOF_DRAWING", target: "lof_drawing", condition: "canDraw", description: "Place the first cone point on the map" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF DRAWING — line-of-fire cones being drawn
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_drawing",
    label: "Drawing line-of-fire cones",
    description: "Line-of-fire cones are being placed on the map.",
    transitions: [
      { action: "ADD_LOF_POINT",       target: "lof_drawing",                             label: "Add cone vertex" },
      { action: "FINISH_LOF_DRAWING",  target: "lof_naming",                              label: "Complete LoF – name it" },
    ],
    menuOptions: [],
    canvasActions: [
      { trigger: "Left-click on background", action: "ADD_LOF_POINT", target: "lof_drawing", description: "Add a cone vertex to the line of fire" },
      { trigger: "Double-click",             action: "FINISH_LOF_DRAWING", target: "lof_naming", description: "Complete the line of fire and open naming dialog" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF NAMING — naming dialog open for a new LoF
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_naming",
    label: "Naming new line of fire",
    description: "A naming dialog is open for the new line of fire.",
    transitions: [
      { action: "HIDE_LOF_NAME_DIALOG", target: "neutral",                                label: "Save / cancel LoF" },
    ],
    menuOptions: [],
    canvasActions: [
      { trigger: "(dialog)", action: "CONFIRM_NAME / HIDE_LOF_NAME_DIALOG", target: "neutral", description: "Confirm or cancel the LoF name in the dialog" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF VIEWING — LoF selected, displayed
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_viewing",
    label: "Line of fire – viewing",
    description: "A line of fire is selected and displayed on the map.",
    transitions: [
      { action: "SELECT_LOF",          target: "lof_viewing",                             label: "Switch to different LoF" },
      { action: "DESELECT_ALL",        target: "neutral",                                 label: "Deselect LoF" },
    ],
    menuOptions: [
      { label: "Nouvelle zone",         action: "CREATE_NEW_ZONE", target: "zone_new",    guard: "canCreateZone", description: "Create a brand new zone" },
      { label: "Nouvelle ligne de tir", action: "CREATE_NEW_LOF",  target: "lof_new",     guard: "canDraw",       description: "Create a new line of fire" },
      { label: "Coller",                action: "PASTE_POLYGONS",  target: "zone_editing",                           description: "Paste copied polygons" },
    ],
    canvasActions: [
      { trigger: "Left-click on background", action: "DESELECT_ALL", target: "neutral", description: "Deselect the line of fire" },
      { trigger: "Right-click on map",       action: "(context menu)",                  description: "Show map context menu" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  ZONE VERSION — browsing zone version history
  // ══════════════════════════════════════════════════════════════
  {
    id: "zone_version",
    label: "Browsing zone version history",
    description: "The version history browser is open for a zone.",
    transitions: [
      { action: "EXIT_VERSION_BROWSING", target: "neutral",                               label: "Exit version browser" },
    ],
    menuOptions: [],
    canvasActions: [
      { trigger: "(version browser UI)", action: "EXIT_VERSION_BROWSING", target: "neutral", description: "Close the version browser" },
    ],
  },

  // ══════════════════════════════════════════════════════════════
  //  LOF VERSION — browsing LoF version history
  // ══════════════════════════════════════════════════════════════
  {
    id: "lof_version",
    label: "Browsing LoF version history",
    description: "The version history browser is open for a line of fire.",
    transitions: [
      { action: "EXIT_VERSION_BROWSING", target: "neutral",                               label: "Exit version browser" },
    ],
    menuOptions: [],
    canvasActions: [
      { trigger: "(version browser UI)", action: "EXIT_VERSION_BROWSING", target: "neutral", description: "Close the version browser" },
    ],
  },
] as const;

// ── Lookup helpers ─────────────────────────────────────────────

const stateMap = new Map<string, StateSpec>(
  machineStates.map((s) => [s.id, s]),
);

/** Return the spec for a given state id, or undefined. */
export function getStateSpec(id: string): StateSpec | undefined {
  return stateMap.get(id);
}

/** Find the first matching transition for (stateId × action). */
export function findTransition(
  stateId: string,
  action: string,
): TransitionSpec | undefined {
  return stateMap.get(stateId)?.transitions.find((t) => t.action === action);
}

/** True if the transition is listed in the machine table. */
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

// ── Debug / logging ────────────────────────────────────────────

/**
 * Human-readable label:  "neutral ──SELECT_ZONE──→ zone_editing"
 * Returns "<stateId> cannot handle <action>" when invalid.
 */
export function describeTransition(stateId: string, action: string): string {
  const transition = findTransition(stateId, action);
  if (!transition) return `${stateId} ⚠ cannot handle ${action}`;
  return `${stateId} ──${action}──→ ${transition.target}`;
}

/**
 * Full table dump for debug / visual inspection.
 */
export function dumpMachine(): string {
  const lines: string[] = [];
  for (const state of machineStates) {
    lines.push(`${state.id}  (${state.description})`);
    for (const t of state.transitions) {
      const guard = t.guard ? `  (${t.guard})` : "";
      lines.push(`  ├ ${t.action.padEnd(24)} ──→ ${t.target.padEnd(16)} ${guard}   ${t.label}`);
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
    for (const t of state.transitions) {
      const guard = t.guard ? ` |${t.guard}|` : "";
      lines.push(`  ${state.id} --> ${t.target} : ${t.action}${guard}`);
    }
  }
  return lines.join("\n");
}
