import React from "react";
import type { Point, ZoneDto } from "../../lib/api/types";
import { getPolygonCentroid } from "../../lib/geometry/transforms";

interface ZoneOverlayProps {
  zones: ZoneDto[];
  selectedZoneIds: Set<string>;
  tempPoints: Point[];
  imageWidth: number;
  imageHeight: number;
  onZoneClick: (zoneId: string, event: React.MouseEvent) => void;
  onZoneRightClick: (zoneId: string, clientX: number, clientY: number) => void;
  onPointInsert: (zoneId: string, pointIndexAfter: number, newPoint: Point) => void;
}

export const ZoneOverlay: React.FC<ZoneOverlayProps> = ({
  zones,
  selectedZoneIds,
  tempPoints,
  onZoneClick,
  onZoneRightClick,
  onPointInsert,
}) => {
  const handleSegmentDoubleClick = (e: React.MouseEvent, zoneId: string, pointIndexAfter: number) => {
    e.stopPropagation();
    const svg = e.currentTarget.closest("svg");
    if (!svg) return;
    const rect = svg.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const scaleX = parseFloat(svg.getAttribute("viewBox")?.split(" ")[2] || "1") / rect.width;
    const scaleY = parseFloat(svg.getAttribute("viewBox")?.split(" ")[3] || "1") / rect.height;
    onPointInsert(zoneId, pointIndexAfter, { x: Math.round(x * scaleX), y: Math.round(y * scaleY) });
  };

  return (
    <g className="zone-layer">
      {/* Zones existantes */}
      {zones.map((zone) => {
        const isSelected = selectedZoneIds.has(zone.id);
        const pointsString = zone.points.map((p) => `${p.x},${p.y}`).join(" ");
        const style = zone.style || {
          fill: "rgba(0, 255, 150, 0.25)",
          stroke: "#00e676",
          strokeWidth: 4,
        };

        const centroid = getPolygonCentroid(zone.points);
        const segmentCount = zone.points.length;

        return (
          <g key={zone.id} data-zone-id={zone.id}>
            <polygon
              points={pointsString}
              fill={isSelected ? "rgba(0, 255, 150, 0.5)" : style.fill}
              stroke={style.stroke}
              strokeWidth={isSelected ? (style.strokeWidth || 4) + 2 : style.strokeWidth}
              strokeDasharray={style.strokeDasharray}
              style={{ transition: "fill 0.2s, stroke-width 0.2s", cursor: "pointer" }}
              onClick={(e) => onZoneClick(zone.id, e)}
              onContextMenu={(e) => {
                e.preventDefault();
                onZoneRightClick(zone.id, e.clientX, e.clientY);
              }}
            />
            {/* Lignes invisibles pour la détection de double-clic sur les segments */}
            {isSelected && zone.points.map((_, i) => {
              const next = (i + 1) % segmentCount;
              const p1 = zone.points[i];
              const p2 = zone.points[next];
              return (
                <line
                  key={`seg-${i}`}
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
            {/* Label de la zone */}
            <text
              x={centroid.x}
              y={centroid.y}
              className="zone-label"
              textAnchor="middle"
              dominantBaseline="middle"
            >
              {zone.label}
              {isSelected ? " (Sel)" : ""}
            </text>
          </g>
        );
      })}

      {/* Zone temporaire en cours de dessin */}
      {tempPoints.length > 0 && (
        <g className="temp-zone">
          {tempPoints.length > 1 ? (
            <polyline
              points={tempPoints.map((p) => `${p.x},${p.y}`).join(" ")}
              fill="none"
              stroke="white"
              strokeWidth="2"
              strokeDasharray="5,5"
            />
          ) : null}
          {tempPoints.map((p, i) => (
            <circle key={i} cx={p.x} cy={p.y} r="5" fill="white" />
          ))}
        </g>
      )}
    </g>
  );
};
