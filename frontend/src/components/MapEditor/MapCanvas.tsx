import React, { useRef } from "react";
import type { Point, ZoneDto, MapDto, EditorTool } from "../../lib/api/types";
import { ZoneOverlay } from "./ZoneOverlay";
import { PointHandles } from "./PointHandles";
import { screenPointToImagePoint, screenDeltaToImageDelta } from "../../lib/geometry/coordinates";

interface MapCanvasProps {
  map: MapDto;
  zones: ZoneDto[];
  selectedZoneIds: Set<string>;
  activeTool: EditorTool;
  tempPoints: Point[];
  onPolygonPointAdded: (p: Point) => void;
  onPolygonLastPointRemoved: () => void;
  onPolygonCompleted: () => void;
  onZoneClick: (id: string, event: React.MouseEvent) => void;
  onBackgroundClick: () => void;
  onZoneDrag: (zoneIds: string[], dx: number, dy: number) => void;
  onPointDrag: (zoneId: string, pointIndex: number, newPoint: Point) => void;
  onPointDelete: (zoneId: string, pointIndex: number) => void;
  onPointInsert: (zoneId: string, pointIndexAfter: number, newPoint: Point) => void;
  onMapRightClick: (clientX: number, clientY: number) => void;
  onZoneRightClick: (zoneId: string, clientX: number, clientY: number) => void;
  onMapDoubleClick: (point: Point) => void;
}

export const MapCanvas: React.FC<MapCanvasProps> = ({
  map,
  zones,
  selectedZoneIds,
  tempPoints,
  onPolygonPointAdded,
  onPolygonLastPointRemoved,
  onZoneClick,
  onBackgroundClick,
  onZoneDrag,
  onPointDrag,
  onPointDelete,
  onPointInsert,
  onMapRightClick,
  onZoneRightClick,
  onMapDoubleClick,
}) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const dragStart = useRef<{ x: number; y: number } | null>(null);
  const clickTimer = useRef<number | null>(null);
  const lastClickTime = useRef(0);

  const getImgPoint = (clientX: number, clientY: number): Point => {
    if (!svgRef.current) return { x: 0, y: 0 };
    return screenPointToImagePoint(clientX, clientY, svgRef.current, map.width, map.height);
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    if (e.button === 2) return; // géré par onContextMenu

    if (e.button === 0) {
      const imgPoint = getImgPoint(e.clientX, e.clientY);

      // Détection de double-clic
      const now = Date.now();
      if (now - lastClickTime.current < 300) {
        if (clickTimer.current) {
          clearTimeout(clickTimer.current);
          clickTimer.current = null;
        }
        lastClickTime.current = 0;
        onMapDoubleClick(imgPoint);
        return;
      }
      lastClickTime.current = now;

      // Simple clic - on démarre un timer pour différencier clic et drag
      clickTimer.current = window.setTimeout(() => {
        clickTimer.current = null;
      }, 200);

      // Si on a des tempPoints (dessin en cours), on ajoute le point
      if (tempPoints.length > 0) {
        if (e.ctrlKey || e.metaKey) {
          onPolygonLastPointRemoved();
        } else {
          onPolygonPointAdded(imgPoint);
        }
        return;
      }

      // Démarrer un drag potentiel pour déplacer les zones sélectionnées
      if (selectedZoneIds.size > 0) {
        dragStart.current = { x: e.clientX, y: e.clientY };
        window.addEventListener("mousemove", handleMouseMove);
        window.addEventListener("mouseup", handleMouseUp);
      }
    }
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (!dragStart.current || !svgRef.current) return;
    const dxScreen = e.clientX - dragStart.current.x;
    const dyScreen = e.clientY - dragStart.current.y;

    // Si on a bougé suffisamment, c'est un drag
    if (Math.abs(dxScreen) > 3 || Math.abs(dyScreen) > 3) {
      if (clickTimer.current) {
        clearTimeout(clickTimer.current);
        clickTimer.current = null;
      }
      const { dx, dy } = screenDeltaToImageDelta(dxScreen, dyScreen, svgRef.current, map.width, map.height);
      if (Math.abs(dx) > 0.5 || Math.abs(dy) > 0.5) {
        onZoneDrag(Array.from(selectedZoneIds), dx, dy);
        dragStart.current = { x: e.clientX, y: e.clientY };
      }
    }
  };

  const handleMouseUp = () => {
    dragStart.current = null;
    window.removeEventListener("mousemove", handleMouseMove);
    window.removeEventListener("mouseup", handleMouseUp);
  };

  const handleBackgroundClick = (e: React.MouseEvent) => {
    if (e.target === svgRef.current) {
      onBackgroundClick();
    }
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const target = e.target as SVGElement;

    // Vérifier si on a cliqué sur une zone
    const zoneElement = target.closest("[data-zone-id]");
    if (zoneElement) {
      const zoneId = zoneElement.getAttribute("data-zone-id");
      if (zoneId) {
        onZoneRightClick(zoneId, e.clientX, e.clientY);
        return;
      }
    }

    // Sinon, c'est un clic sur le fond
    onMapRightClick(e.clientX, e.clientY);
  };

  return (
    <div className="map-wrapper">
      <img
        src={map.imageUrl}
        alt={map.name}
        className="map-image"
        style={{ width: "auto", height: "auto" }}
      />
      <svg
        ref={svgRef}
        viewBox={`0 0 ${map.width} ${map.height}`}
        className="svg-overlay"
        onMouseDown={handleMouseDown}
        onClick={handleBackgroundClick}
        onContextMenu={handleContextMenu}
      >
        <ZoneOverlay
          zones={zones}
          selectedZoneIds={selectedZoneIds}
          tempPoints={tempPoints}
          imageWidth={map.width}
          imageHeight={map.height}
          onZoneClick={onZoneClick}
          onZoneRightClick={onZoneRightClick}
          onPointInsert={onPointInsert}
        />

        {selectedZoneIds.size > 0 && (
          <PointHandles
            zones={zones}
            selectedZoneIds={selectedZoneIds}
            onPointDrag={onPointDrag}
            onPointDelete={onPointDelete}
            onPointInsert={onPointInsert}
            screenPointToImagePoint={(x, y) => getImgPoint(x, y)}
          />
        )}
      </svg>
    </div>
  );
};
