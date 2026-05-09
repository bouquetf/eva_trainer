import React, { useRef } from "react";
import type { Point, Zone } from "../../lib/api/types";

interface PointHandlesProps {
  zones: Zone[];
  selectedZoneIds: Set<number>;
  selectedPolygonIndexes: Set<number>;
  onPointDrag: (zoneId: number, polygonIndex: number, pointIndex: number, newPoint: Point) => void;
  onPointDelete: (zoneId: number, polygonIndex: number, pointIndex: number) => void;
  onPointInsert: (zoneId: number, polygonIndex: number, pointIndexAfter: number, newPoint: Point) => void;
  screenPointToImagePoint: (x: number, y: number) => Point;
}

interface DragInfo {
  zoneId: number;
  polygonIndex: number;
  pointIndex: number;
}

export const PointHandles: React.FC<PointHandlesProps> = ({
  zones,
  selectedZoneIds,
  selectedPolygonIndexes,
  onPointDrag,
  onPointDelete,
  onPointInsert,
  screenPointToImagePoint,
}) => {
  const dragInfo = useRef<DragInfo | null>(null);

  const handleMouseDown = (
    e: React.MouseEvent,
    zoneId: number,
    polygonIndex: number,
    pointIndex: number,
  ) => {
    e.stopPropagation();
    if (e.button === 0) {
      if (e.ctrlKey || e.metaKey) {
        onPointDelete(zoneId, polygonIndex, pointIndex);
        return;
      }
      dragInfo.current = { zoneId, polygonIndex, pointIndex };
      window.addEventListener("mousemove", handleMouseMove);
      window.addEventListener("mouseup", handleMouseUp);
    }
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (!dragInfo.current) return;
    const newPoint = screenPointToImagePoint(e.clientX, e.clientY);
    onPointDrag(
      dragInfo.current.zoneId,
      dragInfo.current.polygonIndex,
      dragInfo.current.pointIndex,
      newPoint,
    );
  };

  const handleMouseUp = () => {
    dragInfo.current = null;
    window.removeEventListener("mousemove", handleMouseMove);
    window.removeEventListener("mouseup", handleMouseUp);
  };

  const handleSegmentDoubleClick = (
    e: React.MouseEvent,
    zoneId: number,
    polygonIndex: number,
    pointIndexAfter: number,
  ) => {
    e.stopPropagation();
    const imgPoint = screenPointToImagePoint(e.clientX, e.clientY);
    onPointInsert(zoneId, polygonIndex, pointIndexAfter, imgPoint);
  };

  return (
    <g className="point-handles-layer">
      {zones
        .filter((z) => selectedZoneIds.has(z.id))
        .map((zone) =>
          zone.polygons.map((polygon, pIdx) => {
            if (!selectedPolygonIndexes.has(pIdx)) return null;
            const points = polygon.points;
            const segmentCount = points.length;

            return (
              <g key={`handles-${zone.id}-poly-${pIdx}`}>
                {points.map((_, i) => {
                  const next = (i + 1) % segmentCount;
                  const p1 = points[i];
                  const p2 = points[next];

                  return (
                    <line
                      key={`seg-${zone.id}-${pIdx}-${i}`}
                      x1={p1.x}
                      y1={p1.y}
                      x2={p2.x}
                      y2={p2.y}
                      stroke="transparent"
                      strokeWidth="12"
                      style={{ cursor: "pointer" }}
                      onDoubleClick={(e) =>
                        handleSegmentDoubleClick(e, zone.id, pIdx, i)
                      }
                    />
                  );
                })}
                {points.map((p, i) => (
                  <circle
                    key={`${zone.id}-${pIdx}-${i}`}
                    cx={p.x}
                    cy={p.y}
                    r="6"
                    className="point-handle"
                    onMouseDown={(e) => handleMouseDown(e, zone.id, pIdx, i)}
                    onClick={(e) => e.stopPropagation()}
                    onContextMenu={(e) => { e.preventDefault(); e.stopPropagation(); }}
                  />
                ))}
              </g>
            );
          }),
        )}
    </g>
  );
};
