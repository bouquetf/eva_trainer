import React, { useState, useEffect, useCallback, useRef } from "react";
import { useNavigate, Link } from "react-router-dom";
import type { Zone, Point, ZoneVersion, ZonePolygon } from "../../lib/api/types";
import {
  listZones,
  batchActions,
  type BatchAction,
} from "../../api/zones";
import { acquireLock, releaseLock, refreshLock, releaseLockSync } from "../../api/locks";
import { getMapImageUrl, listTeamMaps } from "../../api/maps";
import type { Map as MapInfo } from "../../api/maps";
import { ApiError } from "../../api/client";
import { restoreVersion } from "../../api/versions";
import { Toolbar } from "./Toolbar";
import { MapCanvas } from "./MapCanvas";
import { ZoneList } from "./ZoneList";
import { ContextMenu } from "./ContextMenu";
import VersionBrowser from "../VersionBrowser";
import ConfirmDialog from "../ConfirmDialog";
import RestoreConfirmModal from "../RestoreConfirmModal";
import {
  translateSelectedPolygons,
  moveZonePoint,
  removeZonePoint,
  removePolygonsByIndex,
  addZonePointAt,
  mirrorZoneHorizontally,
  mirrorZoneVertically,
} from "../../lib/geometry/transforms";

interface ContextMenuState {
  x: number;
  y: number;
  items: { label: string; action: () => void; disabled?: boolean }[];
}

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
}

const LOCK_REFRESH_MS = 600000;

function showError(message: string) {
  alert(message);
}

export const MapEditor: React.FC<MapEditorProps> = ({ teamId, mapId }) => {
  const navigate = useNavigate();
  const [mapsList, setMapsList] = useState<MapInfo[]>([]);
  const [zones, setZones] = useState<Zone[]>([]);
  const [imageSize, setImageSize] = useState<{ width: number; height: number } | null>(null);
  const [selectedZoneId, setSelectedZoneId] = useState<number | null>(null);
  const [editingZone, setEditingZone] = useState<Zone | null>(null);
  const [originalSnapshot, setOriginalSnapshot] = useState<string | null>(null);
  const [isDirty, setIsDirty] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [tempPoints, setTempPoints] = useState<Point[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedPolygonIndexes, setSelectedPolygonIndexes] = useState<Set<number>>(new Set());
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const [canEdit, setCanEdit] = useState(false);
  const [isLocked, setIsLocked] = useState(false);
  const [lockedByMe, setLockedByMe] = useState(false);
  const [editorName, setEditorName] = useState<string | undefined>();
  const [remainingSeconds, setRemainingSeconds] = useState<number | undefined>();
  const [showReacquireButton, setShowReacquireButton] = useState(false);
  // (showNameDialog removed - polygon completion adds directly)
  const [showDiscardDialog, setShowDiscardDialog] = useState(false);
  const [undoStack, setUndoStack] = useState<string[]>([]);
  const [versionRefreshKey, setVersionRefreshKey] = useState(0);
  const [browsingVersion, setBrowsingVersion] = useState<ZoneVersion | null>(null);
  const [currentVersionSnapshot, setCurrentVersionSnapshot] = useState<string | null>(null);
  const [restoreTarget, setRestoreTarget] = useState<ZoneVersion | null>(null);
  const [restoring, setRestoring] = useState(false);
  const pendingZoneSwitch = useRef<PendingSwitch | null>(null);

  const lockRefreshRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const mountedRef = useRef(true);

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

    const beforeUnloadHandler = () => {
      releaseLockSync(teamId, mapId);
    };
    window.addEventListener("beforeunload", beforeUnloadHandler);

    return () => {
      mountedRef.current = false;
      window.removeEventListener("beforeunload", beforeUnloadHandler);
      if (lockRefreshRef.current) {
        clearInterval(lockRefreshRef.current);
        lockRefreshRef.current = null;
      }
      if (countdownRef.current) {
        clearInterval(countdownRef.current);
        countdownRef.current = null;
      }
      releaseLock(teamId, mapId).catch(() => {});
    };
  }, [teamId, mapId, startCountdown]);

  useEffect(() => {
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

      if (e.key === "Delete" || e.key === "Suppr") {
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
        actions.push({ type: "add_polygon", zone_id: original.id, points: poly.points });
      } else {
        const orig = origPolys.get(id);
        if (orig) {
          const changed = orig.points.length !== poly.points.length ||
            orig.points.some((p, i) => p.x !== poly.points[i]?.x || p.y !== poly.points[i]?.y);
          if (changed) {
            actions.push({ type: "update_points", zone_id: original.id, polygon_id: id, points: poly.points });
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
      setIsDirty(false);
      return true;
    }

    setIsSaving(true);
    try {
      await batchActions(teamId, mapId, actions);
      const fresh = JSON.stringify(editingZone);
      setOriginalSnapshot(fresh);
      setIsDirty(false);
      setUndoStack([]);
      setZones((prev) => prev.map((z) => (z.id === editingZone.id ? { ...editingZone } : z)));
      setVersionRefreshKey(k => k + 1);
      return true;
    } catch (err) {
      showError("Erreur lors de la sauvegarde.");
      return false;
    } finally {
      setIsSaving(false);
    }
  };

  const handleDiscard = () => {
    if (!originalSnapshot) return;
    const original = JSON.parse(originalSnapshot) as Zone;
    setEditingZone(original);
    setZones((prev) => prev.map((z) => (z.id === original.id ? original : z)));
    setIsDirty(false);
    setUndoStack([]);
  };

  function startEditZone(zone: Zone) {
    setEditingZone(zone);
    setOriginalSnapshot(JSON.stringify(zone));
    setIsDirty(false);
    setSelectedZoneId(zone.id);
    setSelectedPolygonIndexes(new Set(zone.polygons.map((_, i) => i)));
    setTempPoints([]);
    setUndoStack([]);
  }

  function handleSelectZone(zoneId: number) {
    if (isDirty) {
      pendingZoneSwitch.current = { action: 'select', zoneId };
      setShowDiscardDialog(true);
      return;
    }
    const zone = zones.find((z) => z.id === zoneId) || editingZone;
    if (zone) startEditZone(zone);
  }

  function handleCreateZone(name: string) {
    if (!name.trim()) return;
    if (isDirty) {
      pendingZoneSwitch.current = { action: 'create', name: name.trim() };
      setShowDiscardDialog(true);
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

  async function handleSaveNewZone(): Promise<boolean> {
    if (!editingZone) return false;
    if (!(await ensureLock())) return false;
    if (editingZone.id !== 0) {
      return handleSave();
    }
    if (editingZone.polygons.length === 0) {
      showError("Ajoutez au moins une forme avant de sauvegarder.");
      return false;
    }

    setIsSaving(true);
    try {
      const actions: BatchAction[] = [
        {
          type: "create_zone",
          temp_id: "new",
          polygons: editingZone.polygons.map(p => ({ points: p.points })),
          name: editingZone.name,
        },
      ];
      const result = await batchActions(teamId, mapId, actions);
      const firstResult = result.results?.[0] as { zone_id?: number; polygons?: ZonePolygon[] } | undefined;
      if (firstResult?.zone_id) {
        const newId = firstResult.zone_id;
        const serverPolys = firstResult.polygons;
        const saved = {
          ...editingZone,
          id: newId,
          polygons: serverPolys && serverPolys.length > 0
            ? editingZone.polygons.map((local, i) => ({
                ...local,
                id: serverPolys[i]?.id ?? local.id,
                zone_id: newId,
              }))
            : editingZone.polygons,
        };
        setZones((prev) => prev.map((z) => (z.id === editingZone.id && z.id === 0 ? saved : z)));
        setEditingZone(saved);
        setOriginalSnapshot(JSON.stringify(saved));
        setIsDirty(false);
        setSelectedZoneId(newId);
        setVersionRefreshKey(k => k + 1);
        return true;
      }
      return false;
    } catch {
      showError("Erreur lors de la sauvegarde.");
      return false;
    } finally {
      setIsSaving(false);
    }
  }

  const handleDeleteZone = async (id: number) => {
    if (id === 0) {
      setZones((prev) => prev.filter((z) => z.id !== id));
      if (selectedZoneId === id) {
        setSelectedZoneId(null);
        setEditingZone(null);
        setOriginalSnapshot(null);
        setIsDirty(false);
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
        setIsDirty(false);
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
      setZones((prev) => [...prev, newZone]);
      startEditZone(newZone);
    }
  };

  const handleDiscardConfirm = () => {
    setShowDiscardDialog(false);
    handleDiscard();
    executePending();
  };

  const handleDiscardSave = async () => {
    setShowDiscardDialog(false);
    const ok = editingZone && editingZone.id === 0
      ? await handleSaveNewZone()
      : await handleSave();
    if (ok) {
      executePending();
    }
  };

  const handleDiscardCancel = () => {
    setShowDiscardDialog(false);
    pendingZoneSwitch.current = null;
  };

  const handleRestoreConfirm = async () => {
    if (!restoreTarget || !editingZone) return;
    if (!(await ensureLock())) {
      setRestoreTarget(null);
      return;
    }
    setRestoring(true);
    try {
      await restoreVersion(teamId, editingZone.id, restoreTarget.id);
      setRestoreTarget(null);
      setBrowsingVersion(null);
      setCurrentVersionSnapshot(null);
      const freshZones = await listZones(teamId, mapId);
      setZones(freshZones.zones);
      const updated = freshZones.zones.find(z => z.id === editingZone.id);
      if (updated) {
        setEditingZone(updated);
        setOriginalSnapshot(JSON.stringify(updated));
      }
      setIsDirty(false);
      setVersionRefreshKey(k => k + 1);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 423) {
          showError("Vous devez avoir un verrou d'édition actif pour restaurer.");
        } else {
          showError(String(err.body?.message ?? 'Erreur lors de la restauration.'));
        }
      } else {
        showError("Erreur lors de la restauration.");
      }
    } finally {
      setRestoring(false);
    }
  };

  const handleBrowseVersion = (version: ZoneVersion | null) => {
    if (!editingZone) return
    if (version && browsingVersion === null) {
      setCurrentVersionSnapshot(JSON.stringify(editingZone))
      const z: Zone = zoneFromVersion(editingZone, version)
      setEditingZone(z)
      setBrowsingVersion(version)
      setZones((prev) => prev.map((zone) =>
        zone.id === editingZone.id ? { ...zone, name: version.version_data.name } : zone
      ))
    } else if (version && browsingVersion !== null) {
      const z: Zone = zoneFromVersion(editingZone, version)
      setEditingZone(z)
      setBrowsingVersion(version)
      setZones((prev) => prev.map((zone) =>
        zone.id === editingZone.id ? { ...zone, name: version.version_data.name } : zone
      ))
    } else if (version === null && browsingVersion !== null && currentVersionSnapshot) {
      const restored = JSON.parse(currentVersionSnapshot) as Zone
      setEditingZone(restored)
      setBrowsingVersion(null)
      setCurrentVersionSnapshot(null)
      setZones((prev) => prev.map((zone) =>
        zone.id === restored.id ? { ...zone, name: restored.name } : zone
      ))
    }
  };

  function zoneFromVersion(base: Zone, version: ZoneVersion): Zone {
    return {
      ...base,
      name: version.version_data.name,
      polygons: version.version_data.polygons.map((p, i) => ({
        id: -(i + 1),
        zone_id: base.id,
        points: p.points,
        order_index: p.order_index,
        created_at: new Date().toISOString(),
      })),
    }
  }

  function updateEditingPolygon(updater: (zone: Zone) => Zone | null) {
    pushUndo();
    setEditingZone((prev) => {
      if (!prev) return prev;
      return updater(prev) || prev;
    });
    setIsDirty(true);
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
    setIsDirty(originalSnapshot !== null && prevSnapshot !== originalSnapshot);
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
        setSelectedPolygonIndexes(new Set(zone.polygons.map((_, i) => i)));
        setTempPoints([]);
        setUndoStack([]);
      }
    }
    setIsDirty(true);
  };

  const handlePolygonClick = (zoneId: number, polygonIndex: number, e: React.MouseEvent) => {
    if (zoneId !== editingZone?.id) {
      handleSelectZone(zoneId);
      return;
    }
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

  const handleBackgroundClick = () => {
    setSelectedPolygonIndexes(new Set());
  };

  const handleFlipH = async () => {
    if (!(await ensureLock()) || !editingZone) return;
    updateEditingPolygon((z) => mirrorZoneHorizontally(z));
  };

  const handleFlipV = async () => {
    if (!(await ensureLock()) || !editingZone) return;
    updateEditingPolygon((z) => mirrorZoneVertically(z));
  };

  const handlePolygonCompleted = async () => {
    if (!(await ensureLock())) return;
    if (tempPoints.length < 3) return;
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
    setIsDirty(true);
    setSelectedPolygonIndexes(new Set([newIndex]));
  };

  const handlePolygonPointAdded = (p: Point) => {
    setTempPoints((prev) => [...prev, p]);
  };

  const handlePolygonLastPointRemoved = () => {
    setTempPoints((prev) => prev.slice(0, -1));
  };

  const handleMapDoubleClick = useCallback(
    async (point: Point) => {
      if (!(await ensureLock())) return;
      if (!editingZone) {
        showError("Sélectionnez ou créez d'abord une zone.");
        return;
      }
      handlePolygonPointAdded(point);
    },
    [editingZone, ensureLock],
  );

  const handleMapRightClick = useCallback(
    (clientX: number, clientY: number) => {
      const items = [];

      if (tempPoints.length > 0) {
        items.push({
          label: "Annuler le point",
          action: () => handlePolygonLastPointRemoved(),
        });
        items.push({
          label: "Terminer la forme",
          action: () => handlePolygonCompleted(),
        });
      }

      items.push({
        label: "Nouvelle forme",
        action: () => {
          setTempPoints([]);
        },
      });

      if (items.length > 0) {
        setContextMenu({ x: clientX, y: clientY, items });
      }
    },
    [tempPoints],
  );

  const handleZoneRightClick = useCallback(
    (_zoneId: number, clientX: number, clientY: number) => {
      const items = [
        { label: "Flip H", action: () => handleFlipH() },
        { label: "Flip V", action: () => handleFlipV() },
      ];
      setContextMenu({ x: clientX, y: clientY, items });
    },
    [],
  );

  const closeContextMenu = useCallback(() => setContextMenu(null), []);

  const handleRefreshLock = useCallback(async () => {
    try {
      const refreshed = await refreshLock(teamId, mapId);
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
      />

      <div className="main-content">
        <div className="canvas-area">
          <MapCanvas
            zones={editingZone ? [editingZone] : []}
            selectedZoneIds={selectionSet}
            selectedPolygonIndexes={selectedPolygonIndexes}
            tempPoints={tempPoints}
            imageUrl={getMapImageUrl(mapId)}
            imageWidth={imageSize.width}
            imageHeight={imageSize.height}
            onPolygonPointAdded={handlePolygonPointAdded}
            onPolygonLastPointRemoved={handlePolygonLastPointRemoved}
            onPolygonClick={handlePolygonClick}
            onBackgroundClick={handleBackgroundClick}
            onZoneDrag={handleZoneDrag}
            onPointDrag={handlePointDrag}
            onPointDelete={handlePointDelete}
            onPointInsert={handlePointInsert}
            onMapRightClick={handleMapRightClick}
            onZoneRightClick={handleZoneRightClick}
            onMapDoubleClick={handleMapDoubleClick}

          />
        </div>

        <div className="side-panel">
          <div style={{ marginBottom: "10px", display: "flex", gap: 5 }}>
            <select
              style={{ flex: 1 }}
              value={mapId}
              onChange={(e) => {
                const targetMapId = Number(e.target.value);
                if (isDirty) {
                  pendingZoneSwitch.current = { action: "navigate", mapId: targetMapId };
                  setShowDiscardDialog(true);
                } else {
                  navigate(`/teams/${teamId}/editor/${targetMapId}`);
                }
              }}
            >
              {mapsList.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.name}
                </option>
              ))}
            </select>
            <Link
              to={`/teams/${teamId}/maps/${mapId}/history`}
              className="btn-secondary btn-small"
              style={{ textDecoration: 'none', whiteSpace: 'nowrap', display: 'flex', alignItems: 'center' }}
            >
              Historique
            </Link>
          </div>
          {editingZone && (
            <div style={{ display: "flex", gap: "5px", marginBottom: "10px" }}>
              <button
                className="btn-primary btn-small"
                onClick={editingZone.id === 0 ? handleSaveNewZone : handleSave}
                disabled={!isDirty || isSaving || browsingVersion !== null}
                style={{ flex: 1 }}
              >
                {isSaving ? "Sauvegarde..." : "Sauvegarder"}
              </button>
              <button
                className="btn-secondary btn-small"
                onClick={() => {
                  if (isDirty) {
                    setShowDiscardDialog(true);
                    pendingZoneSwitch.current = null;
                  } else {
                    handleDiscard();
                  }
                }}
                disabled={!isDirty}
                style={{ flex: 1 }}
              >
                Annuler
              </button>
            </div>
          )}
          <ZoneList
            zones={zones}
            editingZoneId={selectedZoneId}
            onSelectZone={handleSelectZone}
            onDeleteZone={handleDeleteZone}
            onCreateZone={handleCreateZone}
            onRenameZone={handleRenameZone}
          />
          {editingZone && editingZone.id > 0 && (
            <VersionBrowser
              teamId={teamId}
              zoneId={editingZone.id}
              refreshKey={versionRefreshKey}
              browsingVersionId={browsingVersion?.id}
              onBrowse={handleBrowseVersion}
            />
          )}
        </div>
      </div>

      {browsingVersion && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          zIndex: 1000,
          background: '#1565c0',
          color: '#fff',
          textAlign: 'center',
          padding: '8px 16px',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          gap: 12,
          fontSize: 14,
        }}>
          <span>
            Affichage de la version du{' '}
            <strong>{new Intl.DateTimeFormat('fr-FR', { dateStyle: 'long', timeStyle: 'short' }).format(new Date(browsingVersion.created_at))}</strong>
            {' '}(lecture seule)
          </span>
          <button
            className="btn-small"
            style={{ background: '#fff', color: '#1565c0', border: 'none', fontWeight: 'bold' }}
            onClick={() => setRestoreTarget(browsingVersion)}
          >
            Revenir à cette version
          </button>
          <button
            className="btn-small"
            style={{ background: 'rgba(255,255,255,0.2)', color: '#fff', border: '1px solid rgba(255,255,255,0.4)' }}
            onClick={() => handleBrowseVersion(null)}
          >
            Retour à l'édition
          </button>
        </div>
      )}

      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          items={contextMenu.items}
          onClose={closeContextMenu}
        />
      )}

      <ConfirmDialog
        open={showDiscardDialog}
        title="Modifications non sauvegardées"
        message="Vous avez des modifications non sauvegardées. Que souhaitez-vous faire ?"
        confirmLabel="Annuler les modifications"
        danger
        onConfirm={handleDiscardConfirm}
        onCancel={handleDiscardCancel}
        altButton={{ label: "Sauvegarder", onClick: handleDiscardSave, className: "btn-primary" }}
      />

      <RestoreConfirmModal
        open={restoreTarget !== null}
        versionDate={restoreTarget ? new Intl.DateTimeFormat('fr-FR', { dateStyle: 'long', timeStyle: 'short' }).format(new Date(restoreTarget.created_at)) : ''}
        authorName={restoreTarget?.editor_name || 'Utilisateur supprimé'}
        onConfirm={handleRestoreConfirm}
        onCancel={() => { setRestoreTarget(null); setRestoring(false) }}
        loading={restoring}
      />
    </div>
  );
};
