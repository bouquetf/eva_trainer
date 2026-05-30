import { useState, useRef, useCallback, forwardRef, useImperativeHandle, useEffect } from "react";
import type { Zone, ZoneVersion, LineOfFireVersion, LineOfFire } from "../../lib/api/types";
import { restoreVersion } from "../../api/versions";
import { listZones } from "../../api/zones";
import { ApiError } from "../../api/client";
import type { EditorAction } from "../../lib/editor";
import VersionBrowser from "../VersionBrowser";
import LineOfFireVersionBrowser from "../LineOfFireVersionBrowser";
import RestoreConfirmModal from "../RestoreConfirmModal";

export interface VersionManagerHandle {
  clearBrowsing(): void;
}

interface VersionManagerProps {
  teamId: number;
  mapId: number;
  readOnly?: boolean;

  editingZone: Zone | null;
  setEditingZone: React.Dispatch<React.SetStateAction<Zone | null>>;
  setZones: React.Dispatch<React.SetStateAction<Zone[]>>;
  setOriginalSnapshot: React.Dispatch<React.SetStateAction<string | null>>;

  linesOfFire: LineOfFire[];
  setLinesOfFire: React.Dispatch<React.SetStateAction<LineOfFire[]>>;
  selectedLineOfFireId: number | null;

  dispatch: React.Dispatch<EditorAction>;
  ensureLock: () => Promise<boolean>;

  versionRefreshKey?: number;
  lofVersionRefreshKey?: number;

  onBrowsingChange?: (browsingZone: boolean, browsingLoF: boolean) => void;
  onRestoreComplete?: () => void;
}

function lofFromVersion(base: LineOfFire, version: LineOfFireVersion): LineOfFire {
  const vd = version.version_data as { name?: string; polygon_data?: unknown };
  return {
    ...base,
    name: (vd.name ?? base.name) as string,
    polygon_data: (vd.polygon_data ?? base.polygon_data) as LineOfFire["polygon_data"],
  };
}

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
  };
}

const bannerStyle: React.CSSProperties = {
  position: "fixed",
  top: 0,
  left: 0,
  right: 0,
  zIndex: 1000,
  background: "#1565c0",
  color: "#fff",
  textAlign: "center",
  padding: "8px 16px",
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
  gap: 12,
  fontSize: 14,
};

const VersionManager = forwardRef<VersionManagerHandle, VersionManagerProps>(function VersionManager(
  {
    teamId,
    mapId,
    readOnly = false,
    editingZone,
    setEditingZone,
    setZones,
    setOriginalSnapshot,
    linesOfFire,
    setLinesOfFire,
    selectedLineOfFireId,
    dispatch,
    ensureLock,
    versionRefreshKey = 0,
    lofVersionRefreshKey = 0,
    onBrowsingChange,
    onRestoreComplete,
  },
  ref,
) {
  const [browsingVersion, setBrowsingVersion] = useState<ZoneVersion | null>(null);
  const [browsingLoFVersion, setBrowsingLoFVersion] = useState<LineOfFireVersion | null>(null);
  const [currentVersionSnapshot, setCurrentVersionSnapshot] = useState<string | null>(null);
  const lofCurrentSnapshotRef = useRef<string | null>(null);
  const [restoreTarget, setRestoreTarget] = useState<ZoneVersion | null>(null);
  const [restoring, setRestoring] = useState(false);

  const notify = useCallback(
    (zv: ZoneVersion | null, lv: LineOfFireVersion | null) => {
      onBrowsingChange?.(zv !== null, lv !== null);
    },
    [onBrowsingChange],
  );

  const handleBrowseVersion = useCallback(
    (version: ZoneVersion | null) => {
      if (!editingZone) return;
      if (version && browsingVersion === null) {
        setCurrentVersionSnapshot(JSON.stringify(editingZone));
        setEditingZone(zoneFromVersion(editingZone, version));
        setBrowsingVersion(version);
        dispatch({ type: "ENTER_VERSION_BROWSING" });
        setZones((prev) =>
          prev.map((zone) =>
            zone.id === editingZone.id ? { ...zone, name: version.version_data.name } : zone,
          ),
        );
        notify(version, browsingLoFVersion);
      } else if (version && browsingVersion !== null) {
        setEditingZone(zoneFromVersion(editingZone, version));
        setBrowsingVersion(version);
        setZones((prev) =>
          prev.map((zone) =>
            zone.id === editingZone.id ? { ...zone, name: version.version_data.name } : zone,
          ),
        );
        notify(version, browsingLoFVersion);
      } else if (version === null && browsingVersion !== null && currentVersionSnapshot) {
        const restored = JSON.parse(currentVersionSnapshot) as Zone;
        setEditingZone(restored);
        setBrowsingVersion(null);
        setCurrentVersionSnapshot(null);
        dispatch({ type: "EXIT_VERSION_BROWSING" });
        setZones((prev) =>
          prev.map((zone) =>
            zone.id === restored.id ? { ...zone, name: restored.name } : zone,
          ),
        );
        notify(null, browsingLoFVersion);
      }
    },
    [editingZone, browsingVersion, browsingLoFVersion, currentVersionSnapshot, setEditingZone, dispatch, setZones, notify],
  );

  const handleBrowseLoFVersion = useCallback(
    (version: LineOfFireVersion | null) => {
      if (version && browsingLoFVersion === null) {
        lofCurrentSnapshotRef.current = JSON.stringify(linesOfFire);
        const target = linesOfFire.find((l) => l.id === version.line_of_fire_id);
        if (!target) return;
        setLinesOfFire((prev) =>
          prev.map((l) => (l.id === version.line_of_fire_id ? lofFromVersion(target, version) : l)),
        );
        setBrowsingLoFVersion(version);
        dispatch({ type: "ENTER_VERSION_BROWSING" });
        notify(browsingVersion, version);
      } else if (version && browsingLoFVersion !== null) {
        const target = linesOfFire.find((l) => l.id === version.line_of_fire_id);
        if (!target) return;
        setLinesOfFire((prev) =>
          prev.map((l) => (l.id === version.line_of_fire_id ? lofFromVersion(target, version) : l)),
        );
        setBrowsingLoFVersion(version);
        notify(browsingVersion, version);
      } else if (version === null && browsingLoFVersion !== null && lofCurrentSnapshotRef.current) {
        const restored = JSON.parse(lofCurrentSnapshotRef.current) as LineOfFire[];
        setLinesOfFire(restored);
        setBrowsingLoFVersion(null);
        lofCurrentSnapshotRef.current = null;
        dispatch({ type: "EXIT_VERSION_BROWSING" });
        notify(browsingVersion, null);
      }
    },
    [linesOfFire, browsingLoFVersion, browsingVersion, setLinesOfFire, dispatch, notify],
  );

  const handleRestoreConfirm = useCallback(async () => {
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
      const updated = freshZones.zones.find((z) => z.id === editingZone.id);
      if (updated) {
        setEditingZone(updated);
        setOriginalSnapshot(JSON.stringify(updated));
      }
      dispatch({ type: "MARK_CLEAN" });
      onRestoreComplete?.();
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 423) {
          alert("Vous devez avoir un verrou d'édition actif pour restaurer.");
        } else {
          alert(String(err.body?.message ?? "Erreur lors de la restauration."));
        }
      } else {
        alert("Erreur lors de la restauration.");
      }
    } finally {
      setRestoring(false);
    }
  }, [restoreTarget, editingZone, teamId, mapId, ensureLock, setZones, setEditingZone, setOriginalSnapshot, dispatch, onRestoreComplete]);

  const clearBrowsing = useCallback(() => {
    if (lofCurrentSnapshotRef.current) {
      const restored = JSON.parse(lofCurrentSnapshotRef.current) as LineOfFire[];
      setLinesOfFire(restored);
      lofCurrentSnapshotRef.current = null;
    }
    setBrowsingLoFVersion(null);
    setBrowsingVersion(null);
    setCurrentVersionSnapshot(null);
    notify(null, null);
  }, [setLinesOfFire, notify]);

  useImperativeHandle(ref, () => ({ clearBrowsing }), [clearBrowsing]);

  useEffect(() => {
    notify(browsingVersion, browsingLoFVersion);
  }, [browsingVersion, browsingLoFVersion, notify]);

  return (
    <>
      {!readOnly && selectedLineOfFireId != null && !browsingVersion && (
        <LineOfFireVersionBrowser
          teamId={teamId}
          lineOfFireId={selectedLineOfFireId}
          refreshKey={lofVersionRefreshKey}
          browsingVersionId={browsingLoFVersion?.id}
          onBrowse={handleBrowseLoFVersion}
        />
      )}
      {!readOnly && editingZone && editingZone.id > 0 && (
        <VersionBrowser
          teamId={teamId}
          zoneId={editingZone.id}
          refreshKey={versionRefreshKey}
          browsingVersionId={browsingVersion?.id}
          onBrowse={handleBrowseVersion}
        />
      )}
      {!readOnly && browsingLoFVersion && (
        <div style={bannerStyle}>
          <span>
            Affichage de la version du{" "}
            <strong>
              {new Intl.DateTimeFormat("fr-FR", { dateStyle: "long", timeStyle: "short" }).format(
                new Date(browsingLoFVersion.created_at),
              )}
            </strong>{" "}
            (lecture seule)
          </span>
          <button
            className="btn-small"
            style={{ background: "#fff", color: "#1565c0", border: "none", fontWeight: "bold" }}
            onClick={() => handleBrowseLoFVersion(null)}
          >
            Retour à l'édition
          </button>
        </div>
      )}
      {!readOnly && browsingVersion && (
        <div style={bannerStyle}>
          <span>
            Affichage de la version du{" "}
            <strong>
              {new Intl.DateTimeFormat("fr-FR", { dateStyle: "long", timeStyle: "short" }).format(
                new Date(browsingVersion.created_at),
              )}
            </strong>{" "}
            (lecture seule)
          </span>
          <button
            className="btn-small"
            style={{ background: "#fff", color: "#1565c0", border: "none", fontWeight: "bold" }}
            onClick={() => setRestoreTarget(browsingVersion)}
          >
            Revenir à cette version
          </button>
          <button
            className="btn-small"
            style={{ background: "rgba(255,255,255,0.2)", color: "#fff", border: "1px solid rgba(255,255,255,0.4)" }}
            onClick={() => handleBrowseVersion(null)}
          >
            Retour à l'édition
          </button>
        </div>
      )}
      <RestoreConfirmModal
        open={restoreTarget !== null}
        versionDate={
          restoreTarget
            ? new Intl.DateTimeFormat("fr-FR", { dateStyle: "long", timeStyle: "short" }).format(
                new Date(restoreTarget.created_at),
              )
            : ""
        }
        authorName={restoreTarget?.editor_name || "Utilisateur supprimé"}
        onConfirm={handleRestoreConfirm}
        onCancel={() => {
          setRestoreTarget(null);
          setRestoring(false);
        }}
        loading={restoring}
      />
    </>
  );
});

export default VersionManager;
