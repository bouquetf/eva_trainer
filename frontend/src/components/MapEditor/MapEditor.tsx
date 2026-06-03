import React, { useState, useEffect, useCallback, useRef, useReducer } from "react";
import { useNavigate, Link } from "react-router-dom";
import { editorReducer, createInitialState, hasSidePanelControl, isCanvasActionValid, getMenuOptions } from "../../lib/editor";
import type { Zone, Point, ZonePolygon } from "../../lib/api/types";
import {
  listZones,
  batchActions,
  type BatchAction,
} from "../../api/zones";
import { acquireLock, releaseLock, refreshLock, getLockStatus, releaseLockSync } from "../../api/locks";
import { getMapImageUrl, listTeamMaps } from "../../api/maps";
import { useAuthStore } from "../../store/auth";
import type { Map as MapInfo } from "../../api/maps";
import { ApiError } from "../../api/client";
import { Toolbar } from "./Toolbar";
import { MapCanvas } from "./MapCanvas";
import { ZoneList } from "./ZoneList";
import { EditorContextMenu } from "./EditorContextMenu";
import { MapSelector } from "./MapSelector";
import type { ContextMenuActionItem } from "../../lib/editor";
import ConfirmDialog from "../ConfirmDialog";
import { EditorHelpPopup } from "./EditorHelpPopup";
import VersionManager, { type VersionManagerHandle } from "./VersionManager";
import {
  PASTE_OFFSET,
  translateSelectedPolygons,
  moveZonePoint,
  removeZonePoint,
  removePolygonsByIndex,
  addZonePointAt,
  mirrorSelectedPolygonsH,
  mirrorSelectedPolygonsV,
} from "../../lib/geometry/transforms";
import { listLinesOfFire, createLineOfFire, deleteLineOfFire, addLineOfFireCone, setLineOfFireName, batchLinesOfFireActions } from "../../api/lines_of_fire";
import type { LineOfFire, ConeData, BatchLineOfFireAction } from "../../lib/lines_of_fire/types";
import { LineOfFireNameDialog } from "./LineOfFireNameDialog";
import { ZoneNameDialog } from "./ZoneNameDialog";



type PendingSwitch = {
  action: 'select';
  zoneId: number;
} | {
  action: 'create';
  name: string;
} | {
  action: 'navigate';
  mapId: number;
}

interface MapEditorProps {
  teamId: number;
  mapId: number;
  readOnly?: boolean;
}

const LOCK_REFRESH_MS = 600000;

function showError(message: string) {
  alert(message);
}

export const MapEditor: React.FC<MapEditorProps> = ({ teamId, mapId, readOnly = false }) => {
  const navigate = useNavigate();
  const [mapsList, setMapsList] = useState<MapInfo[]>([]);
  const [zones, setZones] = useState<Zone[]>([]);
  const [imageSize, setImageSize] = useState<{ width: number; height: number } | null>(null);
  const [selectedZoneId, setSelectedZoneId] = useState<number | null>(null);
  const [editingZone, setEditingZone] = useState<Zone | null>(null);
  const [originalSnapshot, setOriginalSnapshot] = useState<string | null>(null);
  const [tempPoints, setTempPoints] = useState<Point[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedPolygonIndexes, setSelectedPolygonIndexes] = useState<Set<number>>(new Set());
  const [contextLofId, setContextLofId] = useState<number | null>(null);
  const [contextConeIndex, setContextConeIndex] = useState<number | undefined>(undefined);
  const [canEdit, setCanEdit] = useState(false);
  const [isLocked, setIsLocked] = useState(false);
  const [lockedByMe, setLockedByMe] = useState(false);
  const [editorName, setEditorName] = useState<string | undefined>();
  const [remainingSeconds, setRemainingSeconds] = useState<number | undefined>();
  const [showReacquireButton, setShowReacquireButton] = useState(false);
  // (showNameDialog removed - polygon completion adds directly)
  const [undoStack, setUndoStack] = useState<string[]>([]);
  const [versionRefreshKey, setVersionRefreshKey] = useState(0);
  const [lofVersionRefreshKey, setLofVersionRefreshKey] = useState(0);
  const [browsingLoFVersion, setBrowsingLoFVersion] = useState(false);
  const versionManagerRef = useRef<VersionManagerHandle>(null);
  const [awaitingFirstPoint, setAwaitingFirstPoint] = useState(false);

  const [editorState, dispatch] = useReducer(editorReducer, undefined, createInitialState);
  const [linesOfFire, setLinesOfFire] = useState<LineOfFire[]>([]);
  const [selectedLineOfFireId, setSelectedLineOfFireId] = useState<number | null>(null);
  const [drawingLineOfFire, setDrawingLineOfFire] = useState(false);
  const [tempConePoints, setTempConePoints] = useState<Point[]>([]);
  const [pendingLineOfFireData, setPendingLineOfFireData] = useState<ConeData | null>(null);
  const [renamingLineOfFireId, setRenamingLineOfFireId] = useState<number | null>(null);
  const [renamingLineOfFireName, setRenamingLineOfFireName] = useState("");
  const [lofSnapshot, setLofSnapshot] = useState<string | null>(null);
  const [, setLofIsSaving] = useState(false);
  const [extendingConeLofId, setExtendingConeLofId] = useState<number | null>(null);
  const linesOfFireRef = useRef<LineOfFire[]>([]);
  const lofSnapshotRef = useRef<string | null>(null);
  const pendingZoneSwitch = useRef<PendingSwitch | null>(null);
  const editorStateRef = useRef(editorState);
  editorStateRef.current = editorState;
  const tempPointsRef = useRef(tempPoints);
  tempPointsRef.current = tempPoints;
  const editingZoneRef = useRef(editingZone);
  editingZoneRef.current = editingZone;
  const drawingLineOfFireRef = useRef(drawingLineOfFire);
  drawingLineOfFireRef.current = drawingLineOfFire;
  const lockedByMeRef = useRef(lockedByMe);
  lockedByMeRef.current = lockedByMe;
  const remainingSecondsRef = useRef(remainingSeconds);
  remainingSecondsRef.current = remainingSeconds;
  const lockExpiresAtRef = useRef("");

  const lockRefreshRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const mountedRef = useRef(true);
  const autoSaveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const handleSaveRef = useRef<(() => Promise<boolean>) | null>(null);

  linesOfFireRef.current = linesOfFire;
  lofSnapshotRef.current = lofSnapshot;

  const getEditorId = () => useAuthStore.getState().user?.id;

  const startCountdown = useCallback((expiresAt: string) => {
    if (countdownRef.current) clearInterval(countdownRef.current);
    const tick = () => {
      const now = Date.now();
      const exp = new Date(expiresAt).getTime();
      const rem = Math.max(0, Math.floor((exp - now) / 1000));
      setRemainingSeconds(rem);
      if (rem <= 0 && mountedRef.current) {
        setCanEdit(false);
        setLockedByMe(false);
        setIsLocked(false);
        setShowReacquireButton(true);
        if (countdownRef.current) clearInterval(countdownRef.current);
      }
    };
    tick();
    countdownRef.current = setInterval(tick, 1000);
  }, []);

  const ensureLock = useCallback(async (): Promise<boolean> => {
    if (lockedByMe) return true;
    try {
      const result = await acquireLock(teamId, mapId);
      setCanEdit(true);
      setIsLocked(true);
      setLockedByMe(true);
      setShowReacquireButton(false);
      lockExpiresAtRef.current = result.lock?.expires_at ?? "";
      if (result.lock?.expires_at) startCountdown(result.lock.expires_at);

      if (!lockRefreshRef.current) {
        lockRefreshRef.current = setInterval(async () => {
          try {
            const refreshed = await refreshLock(teamId, mapId);
            if (refreshed.lock?.expires_at && mountedRef.current) {
              startCountdown(refreshed.lock.expires_at);
            }
          } catch {
            if (mountedRef.current) {
              setCanEdit(false);
              setLockedByMe(false);
              setShowReacquireButton(true);
              showError("Verrou d'édition perdu. Passage en lecture seule.");
            }
          }
        }, LOCK_REFRESH_MS);
      }

      return true;
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        const body = err.body as any;
        setIsLocked(true);
        setLockedByMe(false);
        setEditorName(body.editor_name ?? `éditeur #${body.locked_by}`);
        showError("Cette carte est en cours d'édition par un autre éditeur.");
        return false;
      }
      if (err instanceof ApiError && err.status === 403) {
        showError("Vous n'avez pas la permission d'éditer les zones.");
        return false;
      }
      showError("Impossible d'acquérir le verrou d'édition.");
      return false;
    }
  }, [teamId, mapId, lockedByMe, startCountdown]);

  const handleBrowsingChange = useCallback((_browsingZone: boolean, browsingLoF: boolean) => {
    setBrowsingLoFVersion(browsingLoF);
  }, []);

  const handleRestoreComplete = useCallback(() => {
    setVersionRefreshKey((k) => k + 1);
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    const init = async () => {
      setIsLoading(true);

      try {
        const zonesData = await listZones(teamId, mapId);
        if (mountedRef.current) {
          setZones(zonesData.zones);
        }
      } catch (err) {
        if (err instanceof ApiError) {
          if (err.status === 403) showError("Permission insuffisante pour voir les zones.");
          else if (err.status === 401) showError("Session expirée. Veuillez vous reconnecter.");
          else showError("Erreur lors du chargement des zones.");
        } else {
          showError("Impossible de contacter le serveur.");
        }
      }

      try {
        const mapsData = await listTeamMaps(teamId);
        if (mountedRef.current) {
          setMapsList(mapsData.maps);
        }
      } catch {
        // non-critical
      }

      if (!readOnly) {
        try {
          const lofData = await listLinesOfFire(teamId, mapId);
          if (mountedRef.current) {
            setLinesOfFire(lofData);
          }
        } catch {
          // non-critical
        }
      }

      try {
        const img = new Image();
        img.onload = () => {
          if (mountedRef.current) {
            setImageSize({ width: img.naturalWidth, height: img.naturalHeight });
            setIsLoading(false);
          }
        };
        img.onerror = () => {
          if (mountedRef.current) {
            setImageSize({ width: 2000, height: 1500 });
            setIsLoading(false);
          }
        };
        img.src = getMapImageUrl(mapId);
      } catch {
        if (mountedRef.current) {
          setImageSize({ width: 2000, height: 1500 });
          setIsLoading(false);
        }
      }
    };

    init();

    if (readOnly) return;

    const beforeUnloadHandler = () => {
      releaseLockSync(teamId, mapId);
    };
    window.addEventListener("beforeunload", beforeUnloadHandler);

    const pageHideHandler = () => {
      releaseLockSync(teamId, mapId);
    };
    window.addEventListener("pagehide", pageHideHandler);

    const visibilityHandler = () => {
      if (document.visibilityState === "hidden") {
        if (lockRefreshRef.current) {
          clearInterval(lockRefreshRef.current);
          lockRefreshRef.current = null;
        }
      } else if (document.visibilityState === "visible") {
        getLockStatus(teamId, mapId).then((status) => {
          if (!mountedRef.current) return;
          if (status.locked && status.lock) {
            if (status.lock.editor_id === getEditorId()) {
              setCanEdit(true);
              setLockedByMe(true);
              setIsLocked(true);
              setShowReacquireButton(false);
              setEditorName(status.lock.editor_name);
              lockExpiresAtRef.current = status.lock.expires_at ?? "";
              if (status.lock.expires_at) startCountdown(status.lock.expires_at);
              if (!lockRefreshRef.current) {
                lockRefreshRef.current = setInterval(async () => {
                  try {
                    const refreshed = await refreshLock(teamId, mapId);
                    if (refreshed.lock?.expires_at && mountedRef.current) {
                      startCountdown(refreshed.lock.expires_at);
                    }
                  } catch {
                    if (mountedRef.current) {
                      setCanEdit(false);
                      setLockedByMe(false);
                      setShowReacquireButton(true);
                    }
                  }
                }, LOCK_REFRESH_MS);
              }
            } else {
              setCanEdit(false);
              setLockedByMe(false);
              setIsLocked(true);
              setEditorName(status.lock.editor_name ?? `éditeur #${status.lock.editor_id}`);
              showError(`Le verrou d'édition a été repris par ${status.lock.editor_name ?? `éditeur #${status.lock.editor_id}`}.`);
            }
          } else {
            setCanEdit(false);
            setLockedByMe(false);
            setIsLocked(false);
            setShowReacquireButton(true);
          }
        }).catch(() => {});
      }
    };
    document.addEventListener("visibilitychange", visibilityHandler);

    return () => {
      mountedRef.current = false;
      window.removeEventListener("beforeunload", beforeUnloadHandler);
      window.removeEventListener("pagehide", pageHideHandler);
      document.removeEventListener("visibilitychange", visibilityHandler);
      if (lockRefreshRef.current) {
        clearInterval(lockRefreshRef.current);
        lockRefreshRef.current = null;
      }
      if (countdownRef.current) {
        clearInterval(countdownRef.current);
        countdownRef.current = null;
      }
      if (autoSaveTimerRef.current) {
        clearTimeout(autoSaveTimerRef.current);
        autoSaveTimerRef.current = null;
      }
      releaseLockSync(teamId, mapId);
      releaseLock(teamId, mapId).catch(() => {});
    };
  }, [teamId, mapId, startCountdown, readOnly]);

  useEffect(() => {
    if (lockedByMe) {
      dispatch({ type: "SET_LOCK", lock: { type: "locked_by_me", expiresAt: lockExpiresAtRef.current, remainingSeconds: remainingSeconds ?? 0 } });
    } else if (isLocked) {
      dispatch({ type: "SET_LOCK", lock: { type: "locked_by_other", editorName: editorName ?? "Autre éditeur" } });
    } else {
      dispatch({ type: "SET_LOCK", lock: { type: "unlocked" } });
    }
  }, [lockedByMe, isLocked, remainingSeconds, editorName, dispatch]);

  useEffect(() => {
    const handleHelpKeyDown = (e: KeyboardEvent) => {
      if (e.key === "F1") {
        e.preventDefault();
        dispatch({ type: "TOGGLE_HELP" });
        return;
      }
      if (e.key === "?" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        dispatch({ type: "TOGGLE_HELP" });
        return;
      }
    };
    window.addEventListener("keydown", handleHelpKeyDown);
    return () => window.removeEventListener("keydown", handleHelpKeyDown);
  }, []);

  useEffect(() => {
    if (readOnly) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (document.activeElement?.tagName === "INPUT") return;

      if ((e.ctrlKey || e.metaKey) && e.key === "s") {
        e.preventDefault();
        if (!editingZone) return;
        if (editingZone.id === 0) handleSaveNewZone();
        else handleSave();
        return;
      }

      if ((e.ctrlKey || e.metaKey) && e.key === "z") {
        e.preventDefault();
        handleUndo();
        return;
      }

      if (e.key === "Escape") {
        if (awaitingFirstPoint || tempPoints.length > 0) {
          setTempPoints([]);
          setAwaitingFirstPoint(false);
          return;
        }
        if (extendingConeLofId !== null || drawingLineOfFire) {
          setExtendingConeLofId(null);
          setDrawingLineOfFire(false);
          setTempConePoints([]);
          dispatch({ type: "DESELECT_ALL" });
          return;
        }
      }

      if (e.key === "Delete" || e.key === "Suppr") {
        if (selectedLineOfFireId !== null && selectedPolygonIndexes.size === 0) {
          e.preventDefault();
          handleLineOfFireDelete(selectedLineOfFireId);
          return;
        }
        if (selectedPolygonIndexes.size === 0 || !editingZone) return;
        e.preventDefault();
        updateEditingPolygon((z) => removePolygonsByIndex(z, selectedPolygonIndexes));
        setSelectedPolygonIndexes(new Set());
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  });

  function computeDiff(original: Zone, edited: Zone): BatchAction[] {
    const actions: BatchAction[] = [];

    if (original.name !== edited.name) {
      actions.push({ type: "rename", zone_id: original.id, name: edited.name });
    }

    const origPolys = new Map(original.polygons.map(p => [p.id, p]));
    const editPolys = new Map(edited.polygons.map(p => [p.id, p]));

    for (const [id] of origPolys) {
      if (!editPolys.has(id) && id > 0) {
        actions.push({ type: "remove_polygon", zone_id: original.id, polygon_id: id });
      }
    }

    for (const [id, poly] of editPolys) {
      if (id === 0) {
        actions.push({ type: "add_polygon", zone_id: original.id, points: poly.points, floor: poly.floor });
      } else {
        const orig = origPolys.get(id);
        if (orig) {
          const pointsChanged = orig.points.length !== poly.points.length ||
              orig.points.some((p, i) => p.x !== poly.points[i]?.x || p.y !== poly.points[i]?.y);
          if (pointsChanged) {
            actions.push({ type: "update_points", zone_id: original.id, polygon_id: id, points: poly.points });
          }
          if (orig.floor !== poly.floor) {
            actions.push({ type: "set_floor", zone_id: original.id, polygon_id: id, floor: poly.floor ?? null });
          }
        }
      }
    }

    return actions;
  }

  const handleSave = async (): Promise<boolean> => {
    if (!editingZone || !originalSnapshot) return false;
    if (!(await ensureLock())) return false;

    const original = JSON.parse(originalSnapshot) as Zone;
    const actions = computeDiff(original, editingZone);
    if (actions.length === 0) {
      dispatch({ type: "MARK_CLEAN" });
      return true;
    }

    dispatch({ type: "START_SAVING" });
    try {
      const result = await batchActions(teamId, mapId, actions);

      // Collect server-assigned polygon IDs from add_polygon results
      const newPolyIds: number[] = [];
      if (result.results) {
        let idx = 0;
        for (const action of actions) {
          if (action.type === "add_polygon") {
            const res = result.results[idx] as { polygon?: { id: number } } | undefined;
            if (res?.polygon?.id) {
              newPolyIds.push(res.polygon.id);
            }
          }
          idx++;
        }
      }

      // Build updated zone with server polygon IDs so computeDiff won't re-add them
      const zoneWithServerIds: Zone | undefined = newPolyIds.length > 0
        ? (() => {
            let addIdx = 0;
            return {
              ...editingZone,
              polygons: editingZone!.polygons.map((p) => {
                if (p.id !== 0) return p;
                const newId = newPolyIds[addIdx];
                addIdx++;
                return newId ? { ...p, id: newId } : p;
              }),
            };
          })()
        : undefined;

      const savedZone = zoneWithServerIds ?? editingZone;
      const fresh = JSON.stringify(savedZone);
      if (zoneWithServerIds) {
        setEditingZone(zoneWithServerIds);
      }
      setOriginalSnapshot(fresh);
      dispatch({ type: "MARK_CLEAN" });
      setUndoStack([]);
      setZones((prev) => prev.map((z) => (z.id === savedZone.id ? { ...savedZone } : z)));
      setVersionRefreshKey(k => k + 1);
      return true;
    } catch (err) {
      showError("Erreur lors de la sauvegarde.");
      return false;
    } finally {
      dispatch({ type: "STOP_SAVING" });
    }
  };

  handleSaveRef.current = handleSave;

  const handleDiscard = () => {
    if (!originalSnapshot) return;
    const original = JSON.parse(originalSnapshot) as Zone;
    setEditingZone(original);
    setZones((prev) => prev.map((z) => (z.id === original.id ? original : z)));
    dispatch({ type: "MARK_CLEAN" });
    setUndoStack([]);
  };

  function startEditZone(zone: Zone) {
    setEditingZone(zone);
    setOriginalSnapshot(JSON.stringify(zone));
    dispatch({ type: "MARK_CLEAN" });
    setSelectedZoneId(zone.id);
    setSelectedLineOfFireId(null);
    setSelectedPolygonIndexes(new Set());
    setTempPoints([]);
    setUndoStack([]);
  }

  function handleSelectZone(zoneId: number) {
    if (editorState.isDirty) {
      pendingZoneSwitch.current = { action: 'select', zoneId };
      dispatch({ type: "SHOW_DISCARD_DIALOG" });
      return;
    }
    const zone = zones.find((z) => z.id === zoneId) || editingZone;
    if (zone) {
      dispatch({ type: "SELECT_ZONE", zoneId });
      startEditZone(zone);
    }
  }

  function handleCreateZone(name: string) {
    if (!name.trim()) return;
    if (editorState.isDirty) {
      pendingZoneSwitch.current = { action: 'create', name: name.trim() };
      dispatch({ type: "SHOW_DISCARD_DIALOG" });
      return;
    }
    const newZone: Zone = {
      id: 0,
      team_id: teamId,
      map_id: mapId,
      polygons: [],
      name: name.trim(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setZones((prev) => [...prev, newZone]);
    startEditZone(newZone);
  }

  async function handleSaveNewZone(name?: string): Promise<boolean> {
    if (!editingZone) return false;
    if (!(await ensureLock())) return false;
    if (editingZone.id !== 0) {
      return handleSave();
    }
    if (editingZone.polygons.length === 0) {
      showError("Ajoutez au moins une forme avant de sauvegarder.");
      return false;
    }

    dispatch({ type: "START_SAVING" });
    try {
      const zoneName = name ?? editingZone.name;
      const actions: BatchAction[] = [
        {
          type: "create_zone",
          temp_id: "new",
          polygons: editingZone.polygons.map(p => ({ points: p.points, floor: p.floor })),
          name: zoneName,
        },
      ];
      const result = await batchActions(teamId, mapId, actions);
      const firstResult = result.results?.[0] as { zone_id?: number; polygons?: ZonePolygon[] } | undefined;
      if (firstResult?.zone_id) {
        const newId = firstResult.zone_id;
        const serverPolys = firstResult.polygons;
        const saved = {
          ...editingZone,
          name: zoneName,
          id: newId,
          polygons: serverPolys && serverPolys.length > 0
              ? editingZone.polygons.map((local, i) => ({
                ...local,
                id: serverPolys[i]?.id ?? local.id,
                zone_id: newId,
              }))
              : editingZone.polygons,
        };
        setZones((prev) => {
          const idx = prev.findIndex((z) => z.id === editingZone.id);
          if (idx !== -1) {
            const copy = [...prev];
            copy[idx] = saved;
            return copy;
          }
          return [...prev, saved];
        });
        setEditingZone(saved);
        setOriginalSnapshot(JSON.stringify(saved));
        dispatch({ type: "MARK_CLEAN" });
        dispatch({ type: "SELECT_ZONE", zoneId: newId });
        setSelectedZoneId(newId);
        setVersionRefreshKey(k => k + 1);
        return true;
      }
      return false;
    } catch {
      showError("Erreur lors de la sauvegarde.");
      return false;
    } finally {
      dispatch({ type: "STOP_SAVING" });
    }
  }

  const handleDeleteZone = async (id: number) => {
    if (id === 0) {
      setZones((prev) => prev.filter((z) => z.id !== id));
      if (selectedZoneId === id) {
        setSelectedZoneId(null);
        setEditingZone(null);
        setOriginalSnapshot(null);
        dispatch({ type: "MARK_CLEAN" });
      }
      return;
    }
    if (!(await ensureLock())) return;
    try {
      await batchActions(teamId, mapId, [{ type: "delete_zone", zone_id: id }]);
      setZones((prev) => prev.filter((z) => z.id !== id));
      if (selectedZoneId === id) {
        setSelectedZoneId(null);
        setEditingZone(null);
        setOriginalSnapshot(null);
        dispatch({ type: "MARK_CLEAN" });
      }
    } catch {
      showError("Erreur lors de la suppression.");
    }
  };

  const executePending = () => {
    const pending = pendingZoneSwitch.current;
    pendingZoneSwitch.current = null;
    if (!pending) return;
    if (pending.action === 'navigate') {
      navigate(`/teams/${teamId}/editor/${pending.mapId}`);
    } else if (pending.action === 'select') {
      const zone = zones.find((z) => z.id === pending.zoneId);
      if (zone) startEditZone(zone);
    } else {
      const newZone: Zone = {
        id: 0,
        team_id: teamId,
        map_id: mapId,
        polygons: [],
        name: pending.name.trim(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      startEditZone(newZone);
      setAwaitingFirstPoint(true);
    }
  };

  const handleDiscardConfirm = () => {
    dispatch({ type: "HIDE_DISCARD_DIALOG" });
    handleDiscard();
    executePending();
  };

  const handleDiscardSave = async () => {
    dispatch({ type: "HIDE_DISCARD_DIALOG" });
    const ok = editingZone && editingZone.id === 0
        ? await handleSaveNewZone()
        : await handleSave();
    if (ok) {
      executePending();
    }
  };

  const handleDiscardCancel = () => {
    dispatch({ type: "HIDE_DISCARD_DIALOG" });
    pendingZoneSwitch.current = null;
  };

  const queueAutoSave = () => {
    if (autoSaveTimerRef.current) {
      clearTimeout(autoSaveTimerRef.current);
    }
    autoSaveTimerRef.current = setTimeout(() => {
      autoSaveTimerRef.current = null;
      handleSaveRef.current?.();
    }, 500);
  };

  function updateEditingPolygon(updater: (zone: Zone) => Zone | null) {
    pushUndo();
    setEditingZone((prev) => {
      if (!prev) return prev;
      return updater(prev) || prev;
    });
    dispatch({ type: "MARK_DIRTY" });
    if (editingZone && editingZone.id > 0) {
      queueAutoSave();
    }
  }

  function pushUndo() {
    if (!editingZone) return;
    setUndoStack((prev) => [...prev.slice(-50), JSON.stringify(editingZone)]);
  }

  const handleUndo = () => {
    if (undoStack.length === 0) return;
    const prevSnapshot = undoStack[undoStack.length - 1];
    setUndoStack((s) => s.slice(0, -1));
    const restored = JSON.parse(prevSnapshot) as Zone;
    setEditingZone(restored);
    setZones((z) => z.map((zone) => (zone.id === restored.id ? restored : zone)));
    setSelectedPolygonIndexes(new Set(restored.polygons.map((_, i) => i)));
    dispatch(originalSnapshot !== null && prevSnapshot !== originalSnapshot ? { type: "MARK_DIRTY" } : { type: "MARK_CLEAN" });
  };

  const handlePointDrag = async (
      zoneId: number,
      polygonIndex: number,
      pointIndex: number,
      newPoint: Point,
  ) => {
    if (!(await ensureLock()) || editingZone?.id !== zoneId) return;
    updateEditingPolygon((z) => moveZonePoint(z, polygonIndex, pointIndex, newPoint));
  };

  const handlePointDelete = async (
      zoneId: number,
      polygonIndex: number,
      pointIndex: number,
  ) => {
    if (!isCanvasActionValid(editorState, "DELETE_POINT")) return;
    if (!(await ensureLock()) || editingZone?.id !== zoneId) return;
    const poly = editingZone.polygons[polygonIndex];
    if (!poly || poly.points.length <= 3) return;
    updateEditingPolygon((z) => removeZonePoint(z, polygonIndex, pointIndex));
  };

  const handlePointInsert = async (
      zoneId: number,
      polygonIndex: number,
      pointIndexAfter: number,
      newPoint: Point,
  ) => {
    if (!(await ensureLock()) || editingZone?.id !== zoneId) return;
    updateEditingPolygon((z) => addZonePointAt(z, polygonIndex, pointIndexAfter, newPoint));
  };

  const handleZoneDrag = async (ids: number[], dx: number, dy: number) => {
    if (!isCanvasActionValid(editorState, "DRAG_ZONE")) return;
    if (!(await ensureLock())) return;
    if (selectedPolygonIndexes.size === 0) return;
    for (const id of ids) {
      if (editingZone?.id === id) {
        updateEditingPolygon((z) => translateSelectedPolygons(z, selectedPolygonIndexes, dx, dy));
      }
    }
  };

  const handleRenameZone = (id: number, name: string) => {
    setZones((prev) => prev.map((z) => (z.id === id ? { ...z, name } : z)));
    if (editingZone?.id === id) {
      pushUndo();
      setEditingZone((prev) => (prev ? { ...prev, name } : prev));
    } else if (!editingZone) {
      const zone = zones.find((z) => z.id === id);
      if (zone) {
        setOriginalSnapshot(JSON.stringify(zone));
        setEditingZone({ ...zone, name });
        setSelectedZoneId(id);
        setSelectedPolygonIndexes(new Set());
        setTempPoints([]);
        setUndoStack([]);
      }
    }
    dispatch({ type: "MARK_DIRTY" });
  };

  const handleSetFloor = useCallback(
      (floor: 'haut' | 'bas' | null) => {
        if (!editingZone || selectedPolygonIndexes.size === 0) return;

        const selectedIndices = Array.from(selectedPolygonIndexes);

        pushUndo();
        const updatedPolygons = editingZone.polygons.map((p, i) =>
            selectedIndices.includes(i) ? { ...p, floor: floor ?? undefined } : p,
        );
        const updatedZone = { ...editingZone, polygons: updatedPolygons };

        setEditingZone(updatedZone);
        setZones((prev) =>
            prev.map((z) =>
                z.id === editingZone.id ? updatedZone : z,
            ),
        );
        dispatch({ type: "MARK_DIRTY" });
        if (editingZone.id > 0) queueAutoSave();
      },
      [editingZone, selectedPolygonIndexes],
  );

  const handleCopy = useCallback(() => {
    if (!editingZone || selectedPolygonIndexes.size === 0) return;
    const copied = Array.from(selectedPolygonIndexes)
        .sort((a, b) => a - b)
        .map((i) => editingZone.polygons[i])
        .filter(Boolean)
        .map((poly) => ({
          ...poly,
          id: 0,
          points: poly.points.map((p) => ({ ...p })),
          created_at: new Date().toISOString(),
        }));
    dispatch({ type: "SET_CLIPBOARD", clipboard: { polygons: copied, sourceZoneId: editingZone.id } });
  }, [editingZone, selectedPolygonIndexes, dispatch]);

  const handlePaste = useCallback(() => {
    const { polygons: cbPolys, sourceZoneId: cbSourceZoneId } = editorState.clipboard;
    if (!editingZone || cbPolys.length === 0 || cbSourceZoneId === null) return;

    if (editingZone.id === cbSourceZoneId) {
      updateEditingPolygon((z) => {
        const maxOrder = z.polygons.reduce((m, p) => Math.max(m, p.order_index), -1);
        const newPolys = cbPolys.map((poly, i) => {
          const offset = (i + 1) * PASTE_OFFSET;
          return {
            ...poly,
            zone_id: z.id,
            order_index: maxOrder + 1 + i,
            points: poly.points.map((p) => ({
              x: Math.round(p.x + offset),
              y: Math.round(p.y + offset),
            })),
          };
        });
        const updated = {
          ...z,
          polygons: [...z.polygons, ...newPolys],
          updated_at: new Date().toISOString(),
        };
        setZones((prev) => prev.map((zone) => (zone.id === updated.id ? updated : zone)));
        return updated;
      });
      setSelectedPolygonIndexes(
          new Set(cbPolys.map((_, i) => editingZone.polygons.length + i)),
      );
    } else {
      const sourceZone = zones.find((z) => z.id === cbSourceZoneId);
      if (!sourceZone) return;

      const maxOrder = sourceZone.polygons.reduce((m, p) => Math.max(m, p.order_index), -1);
      const newPolys = cbPolys.map((poly, i) => {
        const offset = (i + 1) * PASTE_OFFSET;
        return {
          ...poly,
          zone_id: sourceZone.id,
          order_index: maxOrder + 1 + i,
          points: poly.points.map((p) => ({
            x: Math.round(p.x + offset),
            y: Math.round(p.y + offset),
          })),
        };
      });
      const updatedZone: Zone = {
        ...sourceZone,
        polygons: [...sourceZone.polygons, ...newPolys],
        updated_at: new Date().toISOString(),
      };

      setEditingZone(updatedZone);
      setOriginalSnapshot(JSON.stringify(sourceZone));
      setZones((prev) => prev.map((z) => (z.id === updatedZone.id ? updatedZone : z)));
      dispatch({ type: "MARK_DIRTY" });
      setSelectedZoneId(updatedZone.id);
      setSelectedPolygonIndexes(
          new Set(newPolys.map((_, i) => sourceZone.polygons.length + i)),
      );
      setTempPoints([]);
      setUndoStack([]);
    }
  }, [editingZone, editorState.clipboard, zones, dispatch]);

  const handlePolygonClick = (zoneId: number, polygonIndex: number, e: React.MouseEvent) => {
    if (zoneId !== editingZone?.id) {
      handleSelectZone(zoneId);
      return;
    }
    if (!isCanvasActionValid(editorState, "SELECT_POLYGON")) return;
    if (e.ctrlKey || e.metaKey) {
      setSelectedPolygonIndexes((prev) => {
        const next = new Set(prev);
        if (next.has(polygonIndex)) next.delete(polygonIndex);
        else next.add(polygonIndex);
        return next;
      });
    } else {
      setSelectedPolygonIndexes(new Set([polygonIndex]));
    }
  };

  const handleFlipV = async () => {
    if (!(await ensureLock()) || !editingZone) return;
    if (selectedPolygonIndexes.size === 0) return;
    updateEditingPolygon((z) => mirrorSelectedPolygonsH(z, selectedPolygonIndexes));
  };

  const handleFlipH = async () => {
    if (!(await ensureLock()) || !editingZone) return;
    if (selectedPolygonIndexes.size === 0) return;
    updateEditingPolygon((z) => mirrorSelectedPolygonsV(z, selectedPolygonIndexes));
  };

  const handlePolygonCompleted = async () => {
    if (!(await ensureLock())) return;
    if (tempPoints.length < 3) return;
    const isNewZone = editingZone?.id === 0;
    const wasEmpty = (editingZone?.polygons.length ?? 0) === 0;
    pushUndo();
    const points = [...tempPoints];
    const newIndex = editingZone?.polygons.length ?? 0;
    setTempPoints([]);
    const newPoly = {
      id: 0,
      zone_id: editingZone?.id ?? 0,
      points,
      order_index: 0,
      created_at: new Date().toISOString(),
    };
    setEditingZone((prev) => {
      if (!prev) return prev;
      const maxOrder = prev.polygons.reduce((m, p) => Math.max(m, p.order_index), -1);
      return {
        ...prev,
        polygons: [...prev.polygons, { ...newPoly, order_index: maxOrder + 1 }],
      };
    });
    setSelectedPolygonIndexes(new Set([newIndex]));
    if (!isNewZone || !wasEmpty) {
      dispatch({ type: "MARK_DIRTY" });
    }
    dispatch({ type: "FINISH_DRAWING" });
    if (isNewZone && wasEmpty) {
      setAwaitingFirstPoint(false);
    } else {
      setAwaitingFirstPoint(false);
      queueAutoSave();
    }
  };

  const handlePolygonPointAdded = (p: Point) => {
    const isFirst = tempPoints.length === 0;
    setTempPoints((prev) => [...prev, p]);
    setAwaitingFirstPoint(false);
    const action = isFirst ? "FIRST_POLYGON_POINT" : "ADD_POLYGON_POINT";
    if (!isCanvasActionValid(editorState, action)) return;
    if (isFirst && editingZone) {
      dispatch({ type: "FIRST_POLYGON_POINT", zoneId: editingZone.id });
    } else if (!isFirst) {
      dispatch({ type: "ADD_POLYGON_POINT" });
    }
  };

  const handlePolygonLastPointRemoved = () => {
    setTempPoints((prev) => prev.slice(0, -1));
  };

  const handleLineOfFirePointAdd = useCallback(
      async (p: Point) => {
        if (!(await ensureLock())) return;
        const newPoints = [...tempConePoints, p];
        setTempConePoints(newPoints);

        if (newPoints.length === 3) {
          setDrawingLineOfFire(false);
          setTempConePoints([]);

          if (extendingConeLofId !== null) {
            const lofId = extendingConeLofId;
            setExtendingConeLofId(null);
            dispatch({ type: "SELECT_LOF", lofId });
            try {
              await addLineOfFireCone(teamId, lofId, {
                origin: newPoints[0],
                left_end: newPoints[1],
                right_end: newPoints[2],
              });
              const fresh = await listLinesOfFire(teamId, mapId);
              setLinesOfFire(fresh);
            } catch {
              showError("Erreur lors de l'ajout du cône.");
            }
          } else {
            const coneData: ConeData = {
              cones: [{ origin: newPoints[0], left_end: newPoints[1], right_end: newPoints[2] }],
            };
            setPendingLineOfFireData(coneData);
            dispatch({ type: "FINISH_LOF_DRAWING" });
          }
        }
      },
      [ensureLock, tempConePoints, extendingConeLofId, teamId, mapId, dispatch],
  );

  const handleLineOfFireNameConfirm = useCallback(
      async (name: string) => {
        if (!pendingLineOfFireData) return;
        if (!(await ensureLock())) return;
        try {
          const line = await createLineOfFire(teamId, mapId, name, pendingLineOfFireData);
          setLinesOfFire((prev) => [...prev, line]);
          setSelectedLineOfFireId(line.id);
        } catch {
          showError("Erreur lors de la création de la ligne de tir.");
        } finally {
          setPendingLineOfFireData(null);
        }
      },
      [teamId, mapId, pendingLineOfFireData, ensureLock],
  );

  const handleLineOfFireNameCancel = useCallback(() => {
    setPendingLineOfFireData(null);
  }, []);

  const handleLineOfFireClick = useCallback((lofId: number) => {
    versionManagerRef.current?.clearBrowsing();
    setSelectedLineOfFireId(lofId);
    setSelectedZoneId(null);
    setSelectedPolygonIndexes(new Set());
    setEditingZone(null);
    dispatch({ type: "SELECT_LOF", lofId });
  }, [dispatch]);

  const handleLineOfFireDelete = useCallback(async (lofId: number) => {
    if (!(await ensureLock())) return;
    if (!confirm("Supprimer cette ligne de tir ?")) return;
    try {
      await deleteLineOfFire(teamId, lofId);
      setLinesOfFire((prev) => prev.filter((l) => l.id !== lofId));
      if (selectedLineOfFireId === lofId) setSelectedLineOfFireId(null);
    } catch {
      showError("Erreur lors de la suppression.");
    }
  }, [teamId, selectedLineOfFireId, ensureLock]);

  const handleLineOfFireDeleteCone = useCallback(async (lofId: number, coneIndex: number) => {
    if (!(await ensureLock())) return;
    try {
      await batchLinesOfFireActions(teamId, mapId, [
        { type: "remove_cone", line_of_fire_id: lofId, cone_index: coneIndex },
      ]);
      setLinesOfFire((prev) =>
          prev.map((l) => {
            if (l.id !== lofId) return l;
            return { ...l, polygon_data: { cones: l.polygon_data.cones.filter((_, i) => i !== coneIndex) } };
          }),
      );
      setVersionRefreshKey((k) => k + 1);
      setLofVersionRefreshKey((k) => k + 1);
    } catch {
      showError("Erreur lors de la suppression du cône.");
    }
  }, [teamId, mapId, ensureLock]);

  const handleLineOfFireOriginDoubleClick = useCallback(async (lofId: number, coneIndex?: number) => {
    if (browsingLoFVersion) return;
    const lof = linesOfFire.find((l) => l.id === lofId);
    if (!lof || lof.polygon_data.cones.length === 0) return;
    const cone = lof.polygon_data.cones[coneIndex ?? 0];
    if (!cone) return;
    setTempConePoints([cone.origin]);
    setDrawingLineOfFire(true);
    dispatch({ type: "SET_MODE", mode: { type: "lof_drawing", pointCount: 0 } });
    setExtendingConeLofId(lofId);
    setLofSnapshot(null);
  }, [linesOfFire, browsingLoFVersion, dispatch]);

  const handleLineOfFireRenameAction = useCallback((lofId: number) => {
    const lof = linesOfFire.find((l) => l.id === lofId);
    setRenamingLineOfFireId(lofId);
    setRenamingLineOfFireName(lof?.name ?? "");
  }, [linesOfFire]);

  const handleLineOfFireRename = useCallback(
      async (lofId: number, name: string) => {
        if (!(await ensureLock())) return;
        try {
          await setLineOfFireName(teamId, lofId, name);
          setLinesOfFire((prev) => prev.map((l) => (l.id === lofId ? { ...l, name } : l)));
        } catch {
          showError("Erreur lors du renommage.");
        }
      },
      [teamId, ensureLock],
  );

  const handleLineOfFireRightClick = useCallback(
    (lofId: number, clientX: number, clientY: number, coneIndex?: number) => {
        setContextLofId(lofId);
        setContextConeIndex(coneIndex);
        const items: ContextMenuActionItem[] = [
          { label: "Renommer", action: "RENAME_LOF" },
          { label: "Supprimer", action: "DELETE_LOF" },
          { label: "Ajouter un cône", action: "ADD_CONE" },
        ];
        if (coneIndex !== undefined) {
          items.push({
            label: "Supprimer ce cône",
            action: "DELETE_CONE",
          });
        }
        dispatch({ type: "OPEN_CONTEXT_MENU", x: clientX, y: clientY, itemsLabel: "lof", items });
      },
      [dispatch],
  );

  const handleRenameLineOfFireConfirm = useCallback(
      async (name: string) => {
        if (renamingLineOfFireId === null) return;
        await handleLineOfFireRename(renamingLineOfFireId, name);
        setRenamingLineOfFireId(null);
        setRenamingLineOfFireName("");
      },
      [renamingLineOfFireId, handleLineOfFireRename],
  );

  const handleRenameLineOfFireCancel = useCallback(() => {
    setRenamingLineOfFireId(null);
    setRenamingLineOfFireName("");
  }, []);

  const captureLofSnapshot = useCallback(() => {
    setLofSnapshot((prev) => {
      if (prev !== null) return prev;
      const json = JSON.stringify(linesOfFireRef.current);
      lofSnapshotRef.current = json;
      return json;
    });
  }, []);

  const handleConeOriginDrag = useCallback(
      (lofId: number, _coneIndex: number, newOrigin: Point) => {
        captureLofSnapshot();
        setLinesOfFire((prev) =>
            prev.map((l) => {
              if (l.id !== lofId) return l;
              const cones = l.polygon_data.cones.map((c) => ({ ...c, origin: newOrigin }));
              return { ...l, polygon_data: { cones } };
            }),
        );
      },
      [captureLofSnapshot],
  );

  const handleConeBodyDrag = useCallback(
      (lofId: number, coneIndex: number, newOrigin: Point) => {
        captureLofSnapshot();
        setLinesOfFire((prev) =>
            prev.map((l) => {
              if (l.id !== lofId) return l;
              const dragged = l.polygon_data.cones[coneIndex];
              if (!dragged) return l;
              const dx = newOrigin.x - dragged.origin.x;
              const dy = newOrigin.y - dragged.origin.y;
              const cones = l.polygon_data.cones.map((c) => ({
                ...c,
                origin: { x: c.origin.x + dx, y: c.origin.y + dy },
                left_end: { x: c.left_end.x + dx, y: c.left_end.y + dy },
                right_end: { x: c.right_end.x + dx, y: c.right_end.y + dy },
              }));
              return { ...l, polygon_data: { cones } };
            }),
        );
      },
      [captureLofSnapshot],
  );

  const handleConeEndDrag = useCallback(
      (lofId: number, coneIndex: number, _which: "left_end" | "right_end", newPos: Point) => {
        captureLofSnapshot();
        setLinesOfFire((prev) =>
            prev.map((l) => {
              if (l.id !== lofId) return l;
              const cones = l.polygon_data.cones.map((c, i) => {
                if (i !== coneIndex) return c;
                return { ...c, [_which]: newPos };
              });
              return { ...l, polygon_data: { cones } };
            }),
        );
      },
      [captureLofSnapshot],
  );

  const handleConeArcDrag = useCallback(
      (lofId: number, coneIndex: number, arcBulge: number) => {
        captureLofSnapshot();
        setLinesOfFire((prev) =>
            prev.map((l) => {
              if (l.id !== lofId) return l;
              const cones = l.polygon_data.cones.map((c, i) => {
                if (i !== coneIndex) return c;
                return { ...c, arc_bulge: arcBulge };
              });
              return { ...l, polygon_data: { cones } };
            }),
        );
      },
      [captureLofSnapshot],
  );

  const handleLineOfFireDragEnd = useCallback(
      async (_lofId: number, _coneIndex: number) => {
        const snapshot = lofSnapshotRef.current;
        if (!snapshot) return;
        if (!(await ensureLock())) return;

        const actions = computeLofDiff(snapshot, linesOfFireRef.current);
        if (actions.length === 0) {
          setLofSnapshot(null);
          return;
        }

        setLofIsSaving(true);
        try {
          await batchLinesOfFireActions(teamId, mapId, actions);
          setLofSnapshot(null);
          setVersionRefreshKey((k) => k + 1);
          setLofVersionRefreshKey((k) => k + 1);
        } catch {
          showError("Erreur lors de la sauvegarde automatique.");
        } finally {
          setLofIsSaving(false);
        }
      },
      [teamId, mapId, ensureLock],
  );

  function computeLofDiff(snapshotJson: string, current: LineOfFire[]): BatchLineOfFireAction[] {
    const oldList: LineOfFire[] = JSON.parse(snapshotJson);
    const oldMap = new Map(oldList.map((l) => [l.id, l]));
    const actions: BatchLineOfFireAction[] = [];

    for (const lof of current) {
      const old = oldMap.get(lof.id);
      if (!old) continue;
      if (JSON.stringify(old.polygon_data.cones) !== JSON.stringify(lof.polygon_data.cones)) {
        actions.push({ type: "update_polygon_data", line_of_fire_id: lof.id, polygon_data: lof.polygon_data });
      }
    }

    return actions;
  }

  const handleBackgroundClick = useCallback(() => {
    if (!isCanvasActionValid(editorState, "DESELECT_POLYGONS")) return;
    setSelectedPolygonIndexes(new Set());
    setTempPoints([]);
    setUndoStack([]);
    setAwaitingFirstPoint(false);
  }, [editorState]);

  const handleContextNewZone = useCallback(async () => {
    if (editorState.isDirty) {
      pendingZoneSwitch.current = { action: 'create', name: '' };
      dispatch({ type: "SHOW_DISCARD_DIALOG" });
      return;
    }
    if (!(await ensureLock())) return;
    dispatch({ type: "SET_LOCK", lock: { type: "locked_by_me", expiresAt: lockExpiresAtRef.current, remainingSeconds: remainingSecondsRef.current ?? 0 } });
    const newZone: Zone = {
      id: 0,
      team_id: teamId,
      map_id: mapId,
      polygons: [],
      name: '',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setEditingZone(newZone);
    setOriginalSnapshot(JSON.stringify(newZone));
    dispatch({ type: "MARK_CLEAN" });
    setSelectedZoneId(newZone.id);
    setSelectedLineOfFireId(null);
    setSelectedPolygonIndexes(new Set());
    setTempPoints([]);
    setUndoStack([]);
    setAwaitingFirstPoint(true);
    dispatch({ type: "CREATE_NEW_ZONE" });
  }, [teamId, mapId, ensureLock, setEditingZone, setOriginalSnapshot, setSelectedZoneId, setSelectedLineOfFireId, setSelectedPolygonIndexes, setTempPoints, setUndoStack, setAwaitingFirstPoint, dispatch, editorState.isDirty]);

  const onCancelDrawing = useCallback(() => {
    setTempPoints([]);
    setAwaitingFirstPoint(false);
    dispatch({ type: "SET_MODE", mode: { type: "zone_editing", zoneId: editingZone?.id ?? 0 } });
  }, [editingZone?.id, dispatch]);

  const onCancelAwaitingFirst = useCallback(() => {
    if (editingZone && editingZone.id === 0 && editingZone.polygons.length === 0) {
      setSelectedZoneId(null);
      setEditingZone(null);
      setOriginalSnapshot(null);
      dispatch({ type: "MARK_CLEAN" });
      dispatch({ type: "DESELECT_ALL" });
    } else if (editingZone) {
      setTempPoints([]);
      setAwaitingFirstPoint(false);
      dispatch({ type: "SET_MODE", mode: { type: "zone_editing", zoneId: editingZone.id } });
    }
  }, [editingZone, dispatch]);

  const onStartDrawing = useCallback(() => {
    if (editingZone) {
      dispatch({ type: "START_DRAWING", zoneId: editingZone.id, awaitingFirst: true });
    }
    setTempPoints([]);
    setAwaitingFirstPoint(true);
  }, [editingZone, dispatch]);

  const onStartNewLoF = useCallback(() => {
    setDrawingLineOfFire(true);
    setTempConePoints([]);
    setExtendingConeLofId(null);
    dispatch({ type: "CREATE_NEW_LOF" });
  }, [dispatch]);

  const onContextMenuAction = useCallback(
      async (action: string) => {
        switch (action) {
          case "CREATE_NEW_ZONE":
            await handleContextNewZone();
            break;
          case "START_DRAWING":
            onStartDrawing();
            break;
          case "CREATE_NEW_LOF":
            onStartNewLoF();
            break;
          case "PASTE_POLYGONS":
            handlePaste();
            break;
          case "FINISH_DRAWING":
            await handlePolygonCompleted();
            break;
          case "REMOVE_LAST_POINT":
            handlePolygonLastPointRemoved();
            break;
          case "SET_MODE":
            onCancelDrawing();
            break;
          case "DESELECT_ALL":
            onCancelAwaitingFirst();
            break;
          case "COPY_POLYGONS":
            handleCopy();
            break;
          case "SET_FLOOR_haut":
            handleSetFloor("haut");
            break;
          case "SET_FLOOR_bas":
            handleSetFloor("bas");
            break;
          case "SET_FLOOR_clear":
            handleSetFloor(null);
            break;
          case "FLIP_H":
            handleFlipH();
            break;
          case "FLIP_V":
            handleFlipV();
            break;
          case "RENAME_LOF":
            if (contextLofId !== null) handleLineOfFireRenameAction(contextLofId);
            break;
          case "DELETE_LOF":
            if (contextLofId !== null) handleLineOfFireDelete(contextLofId);
            break;
          case "ADD_CONE":
            if (contextLofId !== null) handleLineOfFireOriginDoubleClick(contextLofId, contextConeIndex);
            break;
          case "DELETE_CONE":
            if (contextLofId !== null && contextConeIndex !== undefined) handleLineOfFireDeleteCone(contextLofId, contextConeIndex);
            break;
        }
      },
      [handleContextNewZone, onStartDrawing, onStartNewLoF, handlePaste, onCancelDrawing, onCancelAwaitingFirst, handleCopy, handleSetFloor, handleLineOfFireRenameAction, handleLineOfFireDelete, handleLineOfFireOriginDoubleClick, handleLineOfFireDeleteCone, handlePolygonCompleted, handlePolygonLastPointRemoved, handleFlipH, handleFlipV, contextLofId, contextConeIndex],
  );

  const handleMapRightClick = useCallback(
      (clientX: number, clientY: number) => {
        const currentEditorState = editorStateRef.current;
        const currentTempPoints = tempPointsRef.current;
        const currentEditingZone = editingZoneRef.current;
        const currentDrawingLineOfFire = drawingLineOfFireRef.current;
        const modeType = currentEditorState.mode.type;
        const menu = getMenuOptions(modeType);
        const fromMenu = (m: (typeof menu)[number]): ContextMenuActionItem => ({
          label: m.label,
          action: m.action,
          target: m.target,
        });

        const items: ContextMenuActionItem[] = [];

        if (modeType === "zone_drawing" && currentTempPoints.length > 0) {
          for (const opt of menu) {
            if (opt.action === "DESELECT_ALL") continue;
            if (opt.action === "FINISH_DRAWING" && currentTempPoints.length < 3) continue;
            items.push(fromMenu(opt));
          }
        } else if (modeType === "zone_drawing" && currentTempPoints.length === 0 && currentEditingZone) {
          for (const opt of menu) {
            if (opt.action === "DESELECT_ALL") items.push(fromMenu(opt));
          }
        } else {
          for (const opt of menu) {
            if (opt.action === "CREATE_NEW_LOF" && currentDrawingLineOfFire) continue;
            items.push(fromMenu(opt));
          }
        }

        dispatch({ type: "OPEN_CONTEXT_MENU", x: clientX, y: clientY, itemsLabel: "map", items });
      },
      [dispatch],
  );

  const handleZoneRightClick = useCallback(
      (_zoneId: number, clientX: number, clientY: number) => {
        const items: ContextMenuActionItem[] = [];

        if (selectedPolygonIndexes.size > 0) {
          items.push({
            label: "Copier",
            action: "COPY_POLYGONS",
          });
        }

        for (const opt of getMenuOptions(editorState.mode.type)) {
          if (opt.action === "PASTE_POLYGONS") {
            items.push({ label: opt.label, action: opt.action, target: opt.target });
          }
        }

        if (selectedPolygonIndexes.size > 0 && editingZone) {
          const hasFloor = Array.from(selectedPolygonIndexes).some(
              (i) => editingZone.polygons[i]?.floor != null,
          );

          const floorChildren: ContextMenuActionItem[] = [
            { label: "Haut", action: "SET_FLOOR_haut" },
            { label: "Bas", action: "SET_FLOOR_bas" },
          ];
          if (hasFloor) {
            floorChildren.push({
              label: "Effacer",
              action: "SET_FLOOR_clear",
            });
          }

          items.push({
            label: "Étage",
            action: "",
            children: floorChildren,
          });
        }

        items.push(
            { label: "Flip H", action: "FLIP_H" },
            { label: "Flip V", action: "FLIP_V" },
        );

        dispatch({ type: "OPEN_CONTEXT_MENU", x: clientX, y: clientY, itemsLabel: "zone", items });
      },
      [selectedPolygonIndexes, editingZone, dispatch, editorState.mode.type],
  );

  const handleNewZoneNameConfirm = (name: string) => {
    setEditingZone((prev) => {
      if (!prev) return prev;
      const updated = { ...prev, name };
      setZones((z) => {
        const exists = z.some((zone) => zone.id === prev.id);
        if (exists) {
          return z.map((zone) => (zone.id === prev.id ? updated : zone));
        }
        return [...z, updated];
      });
      return updated;
    });
    dispatch({ type: "MARK_DIRTY" });
    void handleSaveNewZone(name);
  };

  const handleNewZoneNameCancel = useCallback(() => {
    dispatch({ type: "HIDE_ZONE_NAME_DIALOG" });
  }, [dispatch]);

  const closeContextMenu = useCallback(() => {
    dispatch({ type: "CLOSE_CONTEXT_MENU" });
  }, [dispatch]);

  const handleMapChange = useCallback((targetMapId: number) => {
    if (readOnly) {
      navigate(`/teams/${teamId}/zones/${targetMapId}`);
    } else if (editorState.isDirty) {
      pendingZoneSwitch.current = { action: "navigate", mapId: targetMapId };
      dispatch({ type: "SHOW_DISCARD_DIALOG" });
    } else {
      releaseLockSync(teamId, mapId);
      navigate(`/teams/${teamId}/editor/${targetMapId}`);
    }
  }, [readOnly, navigate, teamId, editorState.isDirty, dispatch, mapId]);

  const handleRefreshLock = useCallback(async () => {
    try {
      const refreshed = await refreshLock(teamId, mapId);
      lockExpiresAtRef.current = refreshed.lock?.expires_at ?? "";
      if (refreshed.lock?.expires_at) startCountdown(refreshed.lock.expires_at);
    } catch {
      showError("Impossible de prolonger le verrou.");
    }
  }, [teamId, mapId, startCountdown]);

  const handleReacquireLock = useCallback(async () => {
    try {
      const result = await acquireLock(teamId, mapId);
      setCanEdit(true);
      setIsLocked(true);
      setLockedByMe(true);
      setShowReacquireButton(false);
      lockExpiresAtRef.current = result.lock?.expires_at ?? "";
      if (result.lock?.expires_at) startCountdown(result.lock.expires_at);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        const body = err.body as any;
        setEditorName(body.editor_name ?? `éditeur #${body.locked_by}`);
        setIsLocked(true);
        setLockedByMe(false);
        showError("Le verrou est déjà détenu par un autre éditeur.");
      } else {
        showError("Impossible de ré-acquérir le verrou.");
      }
    }
  }, [teamId, mapId, startCountdown]);

  const selectionSet = selectedZoneId != null ? new Set([selectedZoneId]) : new Set<number>();

  if (isLoading || !imageSize) {
    return <div style={{ padding: "50px", textAlign: "center" }}>Chargement de l'éditeur...</div>;
  }

  return (
      <div className="editor-container" onClick={closeContextMenu}>
        {readOnly ? (
            <div className="toolbar" style={{ justifyContent: "space-between", padding: "8px 12px" }}>
              <span style={{ fontWeight: "bold" }}>{mapsList.find(m => m.id === mapId)?.name || `Carte #${mapId}`}</span>
              <span style={{ color: "#888", fontSize: "13px" }}>(lecture seule)</span>
              <Link
                  to={`/teams/${teamId}`}
                  className="btn-secondary btn-small"
                  style={{ textDecoration: "none" }}
              >
                Retour
              </Link>
            </div>
        ) : (
            <Toolbar
                canEdit={canEdit}
                isLocked={isLocked}
                lockedByMe={lockedByMe}
                editorName={editorName}
                remainingSeconds={remainingSeconds}
                onRefreshLock={handleRefreshLock}
                onReacquireLock={handleReacquireLock}
                showReacquireButton={showReacquireButton}
                teamId={teamId}
                mapId={mapId}
                onHelpClick={() => dispatch({ type: "SHOW_HELP" })}
            />
        )}

        <div className="main-content">
          <div className="canvas-area">
            <MapCanvas
                zones={editingZone ? [editingZone] : []}
                selectedZoneIds={selectionSet}
                selectedPolygonIndexes={selectedPolygonIndexes}
                tempPoints={readOnly ? [] : tempPoints}
                imageUrl={getMapImageUrl(mapId)}
                imageWidth={imageSize.width}
                imageHeight={imageSize.height}
                onPolygonPointAdded={readOnly ? () => {} : handlePolygonPointAdded}
                onPolygonLastPointRemoved={readOnly ? () => {} : handlePolygonLastPointRemoved}
                onPolygonClick={readOnly ? () => {} : handlePolygonClick}
                onBackgroundClick={readOnly ? () => {} : handleBackgroundClick}
                onZoneDrag={readOnly ? (_ids: number[], _dx: number, _dy: number) => {} : handleZoneDrag}
                onPointDrag={readOnly ? (_zoneId: number, _polyIdx: number, _ptIdx: number, _pt: Point) => {} : handlePointDrag}
                onPointDelete={readOnly ? () => {} : handlePointDelete}
                onPointInsert={readOnly ? () => {} : handlePointInsert}
                onMapRightClick={readOnly ? () => {} : handleMapRightClick}
                onZoneRightClick={readOnly ? () => {} : handleZoneRightClick}
                awaitingFirstPoint={readOnly ? false : awaitingFirstPoint}
                linesOfFire={readOnly && !browsingLoFVersion ? [] : linesOfFire}
                selectedLineOfFireId={selectedLineOfFireId}
                tempConePoints={tempConePoints}
                drawingLineOfFire={drawingLineOfFire}
                onLineOfFirePointAdd={readOnly || browsingLoFVersion ? undefined : handleLineOfFirePointAdd}
                onLineOfFireRightClick={readOnly || browsingLoFVersion ? undefined : handleLineOfFireRightClick}
                onOriginDoubleClick={readOnly ? undefined : handleLineOfFireOriginDoubleClick}
                onConeOriginDrag={readOnly || browsingLoFVersion ? undefined : handleConeOriginDrag}
                onConeTranslateDrag={readOnly || browsingLoFVersion ? undefined : handleConeBodyDrag}
                onConeEndDrag={readOnly || browsingLoFVersion ? undefined : handleConeEndDrag}
                onConeArcDrag={readOnly || browsingLoFVersion ? undefined : handleConeArcDrag}
                onLineOfFireDragEnd={readOnly || browsingLoFVersion ? undefined : handleLineOfFireDragEnd}
            />
          </div>

          <div className="side-panel">
            <MapSelector
                mapsList={mapsList}
                mapId={mapId}
                onChange={handleMapChange}
            />
            <ZoneList
                zones={zones}
                linesOfFire={linesOfFire}
                selectedLineOfFireId={selectedLineOfFireId}
                editingZoneId={selectedZoneId}
                availableActions={editorState.availableActions}
                onSelectZone={readOnly ? (id: number) => {
                  const zone = zones.find((z) => z.id === id);
                  if (zone) {
                    setEditingZone(zone);
                    setSelectedZoneId(id);
                    setSelectedPolygonIndexes(new Set());
                  }
                } : handleSelectZone}
                onDeleteZone={readOnly ? undefined : handleDeleteZone}
                onCreateZone={readOnly ? undefined : handleCreateZone}
                onRenameZone={readOnly ? undefined : handleRenameZone}
                onSelectLineOfFire={readOnly ? undefined : handleLineOfFireClick}
                onDeleteLineOfFire={readOnly ? undefined : handleLineOfFireDelete}
                onRenameLineOfFire={readOnly ? undefined : handleLineOfFireRename}
                readOnly={readOnly}
                hideNewZoneInput={!hasSidePanelControl(editorState, "New zone input")}
            />
            <VersionManager
                ref={versionManagerRef}
                teamId={teamId}
                mapId={mapId}
                readOnly={readOnly}
                editingZone={editingZone}
                setEditingZone={setEditingZone}
                setZones={setZones}
                setOriginalSnapshot={setOriginalSnapshot}
                linesOfFire={linesOfFire}
                setLinesOfFire={setLinesOfFire}
                selectedLineOfFireId={selectedLineOfFireId}
                dispatch={dispatch}
                ensureLock={ensureLock}
                versionRefreshKey={versionRefreshKey}
                lofVersionRefreshKey={lofVersionRefreshKey}
                onBrowsingChange={handleBrowsingChange}
                onRestoreComplete={handleRestoreComplete}
            />
          </div>
        </div>

        {!readOnly && (
            <EditorContextMenu
                editorState={editorState}
                onAction={onContextMenuAction}
                onClose={closeContextMenu}
            />
        )}

        {!readOnly && (
            <ConfirmDialog
                open={editorState.dialogs.discardConfirm}
                title="Modifications non sauvegardées"
                message="Vous avez des modifications non sauvegardées. Que souhaitez-vous faire ?"
                confirmLabel="Annuler les modifications"
                danger
                onConfirm={handleDiscardConfirm}
                onCancel={handleDiscardCancel}
                altButton={{ label: "Sauvegarder", onClick: handleDiscardSave, className: "btn-primary" }}
            />
        )}

        <LineOfFireNameDialog
            open={pendingLineOfFireData !== null}
            onConfirm={handleLineOfFireNameConfirm}
            onCancel={handleLineOfFireNameCancel}
        />
        <LineOfFireNameDialog
            open={renamingLineOfFireId !== null}
            initialName={renamingLineOfFireName}
            onConfirm={handleRenameLineOfFireConfirm}
            onCancel={handleRenameLineOfFireCancel}
        />
        <ZoneNameDialog
            open={editorState.mode.type === "zone_naming"}
            initialName={editingZone?.name ?? "Nouvelle zone"}
            onConfirm={handleNewZoneNameConfirm}
            onCancel={handleNewZoneNameCancel}
        />
        <EditorHelpPopup open={editorState.dialogs.helpVisible} onClose={() => dispatch({ type: "HIDE_HELP" })} />
      </div>
  );
};
