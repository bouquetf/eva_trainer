import React, { useState, useEffect, useCallback } from "react";
import type { MapDto, ZoneDto, EditorTool, Point } from "../../lib/api/types";
import { mockMapsApi } from "../../lib/api/mockMapsApi";
import { Toolbar } from "./Toolbar";
import { MapCanvas } from "./MapCanvas";
import { ZoneList } from "./ZoneList";
import { ZoneDetails } from "./ZoneDetails";
import { ContextMenu } from "./ContextMenu";
import { createZoneFromPolygonPoints, cloneZone } from "../../lib/geometry/zoneFactory";
import {
  translateZone,
  moveZonePoint,
  removeZonePoint,
  addZonePointAt,
  mirrorZoneHorizontally,
  mirrorZoneVertically,
} from "../../lib/geometry/transforms";

interface ContextMenuState {
  x: number;
  y: number;
  items: { label: string; action: () => void; disabled?: boolean }[];
}

export const MapEditor: React.FC = () => {
  const [map, setMap] = useState<MapDto | null>(null);
  const [zones, setZones] = useState<ZoneDto[]>([]);
  const [selectedZoneIds, setSelectedZoneIds] = useState<Set<string>>(new Set());
  const [activeTool, setActiveTool] = useState<EditorTool>("select");
  const [tempPoints, setTempPoints] = useState<Point[]>([]);
  const [clipboard, setClipboard] = useState<ZoneDto[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  const mapId = "atlantis";

  useEffect(() => {
    const init = async () => {
      setIsLoading(true);
      try {
        const mapData = await mockMapsApi.getMap(mapId);
        const zonesData = await mockMapsApi.getZones(mapId);
        setMap(mapData);
        setZones(zonesData);
      } catch (error) {
        console.error("Failed to load data", error);
      } finally {
        setIsLoading(false);
      }
    };
    init();
  }, []);

  const closeContextMenu = useCallback(() => setContextMenu(null), []);

  // Actions Globales
  const handleSave = async () => {
    if (!map) return;
    try {
      await mockMapsApi.saveZones(map.id, zones);
      alert("Zones sauvegardées !");
    } catch (error) {
      alert("Erreur lors de la sauvegarde.");
    }
  };

  const handleReload = async () => {
    if (!map) return;
    const zonesData = await mockMapsApi.getZones(map.id);
    setZones(zonesData);
    setSelectedZoneIds(new Set());
    setTempPoints([]);
  };

  const handleReset = () => {
    if (window.confirm("Voulez-vous vraiment supprimer TOUTES les zones ?")) {
      setZones([]);
      setSelectedZoneIds(new Set());
      setTempPoints([]);
    }
  };

  // Gestion du dessin
  const handlePolygonPointAdded = (p: Point) => {
    setTempPoints((prev) => [...prev, p]);
  };

  const handlePolygonLastPointRemoved = () => {
    setTempPoints((prev) => prev.slice(0, -1));
  };

  const handlePolygonCompleted = () => {
    if (tempPoints.length < 3) return;
    if (!map) return;

    const newZone = createZoneFromPolygonPoints(
      map.id,
      `Zone ${zones.length + 1}`,
      tempPoints
    );
    setZones((prev) => [...prev, newZone]);
    setTempPoints([]);
    setSelectedZoneIds(new Set([newZone.id]));
  };

  // Gestion de la sélection
  const handleZoneClick = useCallback((id: string, event: React.MouseEvent) => {
    const additive = event.ctrlKey || event.metaKey || event.shiftKey;
    setSelectedZoneIds((prev) => {
      const next = new Set(prev);
      if (additive) {
        if (next.has(id)) {
          next.delete(id);
        } else {
          next.add(id);
        }
      } else {
        next.clear();
        next.add(id);
      }
      return next;
    });
  }, []);

  const handleBackgroundClick = useCallback(() => {
    setSelectedZoneIds(new Set());
  }, []);

  const handleSelectAll = () => {
    setSelectedZoneIds(new Set(zones.map((z) => z.id)));
  };

  const handleClearSelection = () => {
    setSelectedZoneIds(new Set());
  };

  const handleDeleteSelection = () => {
    const remaining = zones.filter((z) => !selectedZoneIds.has(z.id));
    setZones(remaining);
    setSelectedZoneIds(new Set());
  };

  const handleDeleteZone = (id: string) => {
    setZones((prev) => prev.filter((z) => z.id !== id));
    setSelectedZoneIds((prev) => {
      const next = new Set(prev);
      next.delete(id);
      return next;
    });
  };

  // Transformations
  const handleZoneDrag = (ids: string[], dx: number, dy: number) => {
    setZones((prev) =>
      prev.map((z) => (ids.includes(z.id) ? translateZone(z, dx, dy) : z))
    );
  };

  const handlePointDrag = (zoneId: string, pointIndex: number, newPoint: Point) => {
    setZones((prev) =>
      prev.map((z) => (z.id === zoneId ? moveZonePoint(z, pointIndex, newPoint) : z))
    );
  };

  const handlePointDelete = (zoneId: string, pointIndex: number) => {
    setZones((prev) =>
      prev.map((z) => {
        if (z.id !== zoneId) return z;
        const updated = removeZonePoint(z, pointIndex);
        return updated === null ? z : updated;
      })
    );
  };

  const handlePointInsert = (zoneId: string, pointIndexAfter: number, newPoint: Point) => {
    setZones((prev) =>
      prev.map((z) => (z.id === zoneId ? addZonePointAt(z, pointIndexAfter, newPoint) : z))
    );
  };

  const handleCopy = () => {
    const selected = zones.filter((z) => selectedZoneIds.has(z.id));
    if (selected.length > 0) {
      setClipboard(selected);
    }
  };

  const handlePaste = () => {
    if (clipboard.length === 0) return;
    const offset = 30;
    const newZones = clipboard.map((z) => {
      const cloned = cloneZone(z);
      return translateZone(cloned, offset, offset);
    });
    setZones((prev) => [...prev, ...newZones]);
    setSelectedZoneIds(new Set(newZones.map((z) => z.id)));
  };

  const handleFlipH = () => {
    setZones((prev) =>
      prev.map((z) => (selectedZoneIds.has(z.id) ? mirrorZoneHorizontally(z) : z))
    );
  };

  const handleFlipV = () => {
    setZones((prev) =>
      prev.map((z) => (selectedZoneIds.has(z.id) ? mirrorZoneVertically(z) : z))
    );
  };

  // Gestion du clic droit sur la carte (vide)
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
          setActiveTool("select");
          setTempPoints([]);
          setSelectedZoneIds(new Set());
        },
      });

      if (clipboard.length > 0) {
        items.push({ label: "Coller", action: handlePaste });
      }

      if (items.length > 0) {
        setContextMenu({ x: clientX, y: clientY, items });
      }
    },
    [clipboard, tempPoints, handlePolygonCompleted, handlePaste]
  );

  // Gestion du clic droit sur une zone
  const handleZoneRightClick = useCallback(
    (zoneId: string, clientX: number, clientY: number) => {
      const items = [
        { label: "Copier", action: handleCopy },
        { label: "Flip H", action: handleFlipH },
        { label: "Flip V", action: handleFlipV },
        { label: "Supprimer", action: () => handleDeleteZone(zoneId) },
      ];
      setContextMenu({ x: clientX, y: clientY, items });
    },
    [handleCopy, handleFlipH, handleFlipV, handleDeleteZone]
  );

  // Gestion du double-clic sur la carte (vide) -> commence nouvelle forme
  const handleMapDoubleClick = useCallback(
    (point: Point) => {
      if (tempPoints.length === 0) {
        setSelectedZoneIds(new Set());
      }
      handlePolygonPointAdded(point);
    },
    [tempPoints, handlePolygonPointAdded]
  );

  const selectedZones = zones.filter((z) => selectedZoneIds.has(z.id));

  if (isLoading || !map) {
    return <div style={{ padding: "50px", textAlign: "center" }}>Chargement de l'éditeur...</div>;
  }

  return (
    <div className="editor-container" onClick={closeContextMenu}>
      <Toolbar
        activeTool={activeTool}
        setActiveTool={setActiveTool}
        onSave={handleSave}
        onReload={handleReload}
        onReset={handleReset}
        onSelectAll={handleSelectAll}
        onClearSelection={handleClearSelection}
        onDeleteSelection={handleDeleteSelection}
        onCopy={handleCopy}
        onPaste={handlePaste}
        onFlipH={handleFlipH}
        onFlipV={handleFlipV}
        selectedCount={selectedZoneIds.size}
        tempPointsCount={tempPoints.length}
      />

      <div className="main-content">
        <div className="canvas-area">
          <MapCanvas
            map={map}
            zones={zones}
            selectedZoneIds={selectedZoneIds}
            activeTool={activeTool}
            tempPoints={tempPoints}
            onPolygonPointAdded={handlePolygonPointAdded}
            onPolygonLastPointRemoved={handlePolygonLastPointRemoved}
            onPolygonCompleted={handlePolygonCompleted}
            onZoneClick={handleZoneClick}
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
          <ZoneList
            zones={zones}
            selectedZoneIds={selectedZoneIds}
            onZoneClick={handleZoneClick}
            onDeleteZone={handleDeleteZone}
          />
          <ZoneDetails
            selectedZones={selectedZones}
            onDeleteSelection={handleDeleteSelection}
          />
        </div>
      </div>

      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          items={contextMenu.items}
          onClose={closeContextMenu}
        />
      )}
    </div>
  );
};
