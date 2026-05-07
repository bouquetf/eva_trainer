import React, { useRef } from "react";
import type { Point, ZoneDto } from "../../lib/api/types";

interface PointHandlesProps {
  zones: ZoneDto[];
  selectedZoneIds: Set<string>;
  onPointDrag: (zoneId: string, pointIndex: number, newPoint: Point) => void;
  onPointDelete: (zoneId: string, pointIndex: number) => void;
  onPointInsert: (zoneId: string, pointIndexAfter: number, newPoint: Point) => void;
  screenPointToImagePoint: (x: number, y: number) => Point;
}

export const PointHandles: React.FC<PointHandlesProps> = ({
  zones,
  selectedZoneIds,
  onPointDrag,
  onPointDelete,
  onPointInsert,
  screenPointToImagePoint,
}) => {
  const dragInfo = useRef<{ zoneId: string; pointIndex: number } | null>(null);

  const handleMouseDown = (e: React.MouseEvent, zoneId: string, pointIndex: number) => {
    e.stopPropagation();
    if (e.button === 2) {
      e.preventDefault();
      onPointDelete(zoneId, pointIndex);
      return;
    }
    if (e.button === 0) {
      dragInfo.current = { zoneId, pointIndex };
      window.addEventListener("mousemove", handleMouseMove);
      window.addEventListener("mouseup", handleMouseUp);
    }
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (!dragInfo.current) return;
    const newPoint = screenPointToImagePoint(e.clientX, e.clientY);
    onPointDrag(dragInfo.current.zoneId, dragInfo.current.pointIndex, newPoint);
  };

  const handleMouseUp = () => {
    dragInfo.current = null;
    window.removeEventListener("mousemove", handleMouseMove);
    window.removeEventListener("mouseup", handleMouseUp);
  };

  const handleSegmentDoubleClick = (e: React.MouseEvent, zoneId: string, pointIndexAfter: number) => {
    e.stopPropagation();
    const imgPoint = screenPointToImagePoint(e.clientX, e.clientY);
    onPointInsert(zoneId, pointIndexAfter, imgPoint);
  };

  return (
    <g className="point-handles-layer">
      {zones
        .filter((z) => selectedZoneIds.has(z.id))
        .map((zone) => {
          const points = zone.points;
          const segmentCount = points.length;

          return (
            <g key={`handles-${zone.id}`}>
              {points.map((_, i) => {
                const next = (i + 1) % segmentCount;
                const p1 = points[i];
                const p2 = points[next];

                return (
                  <line
                    key={`${zone.id}-seg-${i}`}
                    x1={p1.x}
                    y1={p1.y}
                    x2={p2.x}
                    y2={p2.y}
                    stroke="transparent"
                    strokeWidth="12"
                    style={{ cursor: "pointer" }}
                    onDoubleClick={(e) => handleSegmentDoubleClick(e, zone.id, i)}
                  />
                );
              })}
              {points.map((p, i) => (
                <circle
                  key={`${zone.id}-${i}`}
                  cx={p.x}
                  cy={p.y}
                  r="6"
                  className="point-handle"
                  onMouseDown={(e) => handleMouseDown(e, zone.id, i)}
                  onContextMenu={(e) => e.preventDefault()}
                />
              ))}
            </g>
          );
        })}
    </g>
  );
};
