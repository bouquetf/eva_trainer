import { describe, it, expect } from "vitest";
import { createInitialState, editorReducer, isCanvasActionValid } from "./state";
import { getMenuOptions } from "./machine";

describe("editor state machine – lock sync (bug reproduction)", () => {
  it("rejects CREATE_NEW_ZONE when lock is unlocked", () => {
    const state = createInitialState();
    expect(state.mode.type).toBe("neutral");
    expect(state.lock.type).toBe("unlocked");

    const next = editorReducer(state, { type: "CREATE_NEW_ZONE" });

    expect(next.mode.type).toBe("neutral");
    expect(next.availableActions.menuOptions).toHaveLength(0);
  });

  it("neutral menu shows Nouvelle zone / Nouvelle ligne de tir after rejected transition", () => {
    const state = createInitialState();

    const menu = getMenuOptions(state.mode.type);
    const labels = menu.map((m) => m.label);

    expect(labels).toContain("Nouvelle zone");
    expect(labels).toContain("Nouvelle ligne de tir");
  });

  it("SET_LOCK re-derives availableActions so guarded menu options appear", () => {
    const state = createInitialState();
    expect(state.availableActions.menuOptions).toHaveLength(0);

    const locked = editorReducer(state, {
      type: "SET_LOCK",
      lock: { type: "locked_by_me", expiresAt: "2099-12-31T00:00:00Z", remainingSeconds: 3600 },
    });

    const labels = locked.availableActions.menuOptions.map((o) => o.action);
    expect(labels).toContain("CREATE_NEW_ZONE");
  });

  it("CREATE_NEW_ZONE succeeds after lock is synced to locked_by_me", () => {
    const state = createInitialState();

    const locked = editorReducer(state, {
      type: "SET_LOCK",
      lock: { type: "locked_by_me", expiresAt: "2099-12-31T00:00:00Z", remainingSeconds: 3600 },
    });
    expect(locked.lock.type).toBe("locked_by_me");

    const zoneNew = editorReducer(locked, { type: "CREATE_NEW_ZONE" });
    expect(zoneNew.mode.type).toBe("zone_new");
  });

  it("full drawing flow works after lock sync: neutral → zone_new → zone_drawing", () => {
    const state = createInitialState();

    const locked = editorReducer(state, {
      type: "SET_LOCK",
      lock: { type: "locked_by_me", expiresAt: "2099-12-31T00:00:00Z", remainingSeconds: 3600 },
    });

    const zoneNew = editorReducer(locked, { type: "CREATE_NEW_ZONE" });
    expect(zoneNew.mode.type).toBe("zone_new");
    expect(isCanvasActionValid(zoneNew, "FIRST_POLYGON_POINT")).toBe(true);

    const drawing = editorReducer(zoneNew, { type: "FIRST_POLYGON_POINT", zoneId: 0 });
    expect(drawing.mode.type).toBe("zone_drawing");

    const menu = getMenuOptions(drawing.mode.type);
    const labels = menu.map((m) => m.label);
    expect(labels).toContain("Annuler le point");
    expect(labels).toContain("Annuler la forme");
    expect(labels).not.toContain("Nouvelle zone");
  });

  it("FINISH_DRAWING from zone_drawing goes to zone_editing when dirty", () => {
    const state = createInitialState();

    const locked = editorReducer(state, {
      type: "SET_LOCK",
      lock: { type: "locked_by_me", expiresAt: "2099-12-31T00:00:00Z", remainingSeconds: 3600 },
    });
    const zoneNew = editorReducer(locked, { type: "CREATE_NEW_ZONE" });
    const drawing = editorReducer(zoneNew, { type: "FIRST_POLYGON_POINT", zoneId: 0 });

    expect(drawing.mode.type).toBe("zone_drawing");

    const dirty = editorReducer(drawing, { type: "MARK_DIRTY" });
    expect(dirty.isDirty).toBe(true);

    const finished = editorReducer(dirty, { type: "FINISH_DRAWING" });
    expect(finished.mode.type).toBe("zone_editing");
    expect(finished.mode).toHaveProperty("zoneId", 0);
  });

  it("FINISH_DRAWING from zone_drawing goes to zone_naming when clean + new zone", () => {
    const state = createInitialState();

    const locked = editorReducer(state, {
      type: "SET_LOCK",
      lock: { type: "locked_by_me", expiresAt: "2099-12-31T00:00:00Z", remainingSeconds: 3600 },
    });
    const zoneNew = editorReducer(locked, { type: "CREATE_NEW_ZONE" });
    const drawing = editorReducer(zoneNew, { type: "FIRST_POLYGON_POINT", zoneId: 0 });

    expect(drawing.mode.type).toBe("zone_drawing");
    expect(drawing.isDirty).toBe(false);

    const finished = editorReducer(drawing, { type: "FINISH_DRAWING" });
    expect(finished.mode.type).toBe("zone_naming");
  });
});
