import React, { useRef } from "react";
import type { Point, Zone } from "../../lib/api/types";
import { ZoneOverlay } from "./ZoneOverlay";
import { PointHandles } from "./PointHandles";
import { screenPointToImagePoint, screenDeltaToImageDelta } from "../../lib/geometry/coordinates";

interface PreviewPolygon {
  points: Point[];
  order_index: number;
}

interface MapCanvasProps {
  zones: Zone[];
  selectedZoneIds: Set<number>;
  selectedPolygonIndexes: Set<number>;
  tempPoints: Point[];
  imageUrl: string;
  imageWidth: number;
  imageHeight: number;
  onPolygonPointAdded: (p: Point) => void;
  onPolygonLastPointRemoved: () => void;
  onPolygonClick: (zoneId: number, polygonIndex: number, event: React.MouseEvent) => void;
  onBackgroundClick: () => void;
  onZoneDrag: (zoneIds: number[], dx: number, dy: number) => void;
  onPointDrag: (zoneId: number, polygonIndex: number, pointIndex: number, newPoint: Point) => void;
  onPointDelete: (zoneId: number, polygonIndex: number, pointIndex: number) => void;
  onPointInsert: (zoneId: number, polygonIndex: number, pointIndexAfter: number, newPoint: Point) => void;
  onMapRightClick: (clientX: number, clientY: number) => void;
  onZoneRightClick: (zoneId: number, clientX: number, clientY: number) => void;
  onMapDoubleClick: (point: Point) => void;
  previewPolygons?: PreviewPolygon[];
}

export const MapCanvas: React.FC<MapCanvasProps> = ({
  zones,
  selectedZoneIds,
  selectedPolygonIndexes,
  tempPoints,
  imageUrl,
  imageWidth,
  imageHeight,
  onPolygonPointAdded,
  onPolygonLastPointRemoved,
  onPolygonClick,
  onBackgroundClick,
  onZoneDrag,
  onPointDrag,
  onPointDelete,
  onPointInsert,
  onMapRightClick,
  onZoneRightClick,
  onMapDoubleClick,
  previewPolygons,
}) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const dragStart = useRef<{ x: number; y: number } | null>(null);

  const getImgPoint = (clientX: number, clientY: number): Point => {
    if (!svgRef.current) return { x: 0, y: 0 };
    return screenPointToImagePoint(clientX, clientY, svgRef.current, imageWidth, imageHeight);
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    if (e.button === 2) return;

    if (e.button === 0) {
      if (tempPoints.length > 0) {
        const imgPoint = getImgPoint(e.clientX, e.clientY);
        if (e.ctrlKey || e.metaKey) {
          onPolygonLastPointRemoved();
        } else {
          onPolygonPointAdded(imgPoint);
        }
        return;
      }

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

    if (Math.abs(dxScreen) > 3 || Math.abs(dyScreen) > 3) {
      const { dx, dy } = screenDeltaToImageDelta(dxScreen, dyScreen, svgRef.current, imageWidth, imageHeight);
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

  const handleSvgClick = (e: React.MouseEvent) => {
    const target = e.target as SVGElement;
    if (target === svgRef.current || target.tagName.toLowerCase() === "image") {
      onBackgroundClick();
    }
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const target = e.target as SVGElement;

    const zoneElement = target.closest("[data-zone-id]");
    if (zoneElement) {
      const zoneIdStr = zoneElement.getAttribute("data-zone-id");
      if (zoneIdStr) {
        const zoneId = parseInt(zoneIdStr, 10);
        if (!isNaN(zoneId)) {
          onZoneRightClick(zoneId, e.clientX, e.clientY);
          return;
        }
      }
    }

    onMapRightClick(e.clientX, e.clientY);
  };

  return (
    <div className="map-wrapper">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${imageWidth} ${imageHeight}`}
        className="svg-overlay"
        onMouseDown={handleMouseDown}
        onClick={handleSvgClick}
        onDoubleClick={(e) => {
          const imgPoint = getImgPoint(e.clientX, e.clientY);
          onMapDoubleClick(imgPoint);
        }}
        onContextMenu={handleContextMenu}
      >
        <image href={imageUrl} width={imageWidth} height={imageHeight} />
        <ZoneOverlay
          zones={zones}
          selectedZoneIds={selectedZoneIds}
          selectedPolygonIndexes={selectedPolygonIndexes}
          tempPoints={tempPoints}
          imageWidth={imageWidth}
          imageHeight={imageHeight}
          onPolygonClick={onPolygonClick}
          onZoneRightClick={onZoneRightClick}
          onPointInsert={onPointInsert}
          previewPolygons={previewPolygons}
        />

        {selectedZoneIds.size > 0 && selectedPolygonIndexes.size > 0 && (
          <PointHandles
            zones={zones}
            selectedZoneIds={selectedZoneIds}
            selectedPolygonIndexes={selectedPolygonIndexes}
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
