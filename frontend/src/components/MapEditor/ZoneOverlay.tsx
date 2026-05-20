import React from "react";
import type { Point, Zone } from "../../lib/api/types";

interface PreviewPolygon {
  points: Point[];
  order_index: number;
}

interface ZoneOverlayProps {
  zones: Zone[];
  selectedZoneIds: Set<number>;
  selectedPolygonIndexes: Set<number>;
  tempPoints: Point[];
  imageWidth: number;
  imageHeight: number;
  onPolygonClick: (zoneId: number, polygonIndex: number, event: React.MouseEvent) => void;
  onZoneRightClick: (zoneId: number, clientX: number, clientY: number) => void;
  onPointInsert: (zoneId: number, polygonIndex: number, pointIndexAfter: number, newPoint: Point) => void;
  previewPolygons?: PreviewPolygon[];
}

export const ZoneOverlay: React.FC<ZoneOverlayProps> = ({
  zones,
  selectedZoneIds,
  selectedPolygonIndexes,
  tempPoints,
  onPolygonClick,
  onZoneRightClick,
  onPointInsert,
  previewPolygons,
}) => {
  const handleSegmentDoubleClick = (
    e: React.MouseEvent,
    zoneId: number,
    polygonIndex: number,
    pointIndexAfter: number,
  ) => {
    e.stopPropagation();
    const svg = e.currentTarget.closest("svg");
    if (!svg) return;
    const rect = svg.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const scaleX = parseFloat(svg.getAttribute("viewBox")?.split(" ")[2] || "1") / rect.width;
    const scaleY = parseFloat(svg.getAttribute("viewBox")?.split(" ")[3] || "1") / rect.height;
    onPointInsert(zoneId, polygonIndex, pointIndexAfter, {
      x: Math.round(x * scaleX),
      y: Math.round(y * scaleY),
    });
  };

  return (
    <g className="zone-layer">
      <defs>
        <pattern id="hatch" patternUnits="userSpaceOnUse" width="8" height="8" patternTransform="rotate(45)">
          <line x1="0" y1="0" x2="0" y2="8" stroke="#ff5252" strokeWidth="1" />
        </pattern>
      </defs>
      {zones.map((zone) => {
        const isZoneSelected = selectedZoneIds.has(zone.id);

        return (
          <g key={zone.id} data-zone-id={zone.id}>
            {zone.polygons.map((polygon, pIdx) => {
              const pointsString = polygon.points.map((p) => `${p.x},${p.y}`).join(" ");
              const segmentCount = polygon.points.length;
              const isPolySelected = isZoneSelected && selectedPolygonIndexes.has(pIdx);

              return (
                <g key={polygon.id || `${zone.id}-poly-${pIdx}`}>
                  <polygon
                    points={pointsString}
                    fill={isZoneSelected ? "rgba(255,80,80,0.15)" : "rgba(255,80,80,0.08)"}
                    stroke="none"
                  />
                  <polygon
                    points={pointsString}
                    fill="url(#hatch)"
                    stroke="#ff5252"
                    strokeWidth={isZoneSelected ? 6 : 4}
                    style={{ cursor: "pointer" }}
                    onClick={(e) => onPolygonClick(zone.id, pIdx, e)}
                    onContextMenu={(e) => {
                      e.preventDefault();
                      onZoneRightClick(zone.id, e.clientX, e.clientY);
                    }}
                  />
                  {isPolySelected && polygon.points.map((_, i) => {
                    const next = (i + 1) % segmentCount;
                    const p1 = polygon.points[i];
                    const p2 = polygon.points[next];
                    return (
                      <line
                        key={`seg-${pIdx}-${i}`}
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
                  {polygon.floor && (() => {
                    const pts = polygon.points;
                    const cx = pts.reduce((s, p) => s + p.x, 0) / pts.length;
                    const cy = pts.reduce((s, p) => s + p.y, 0) / pts.length;
                    const iconSize = 30;

                    if (polygon.floor === 'haut') {
                      const arrowTip = cy - iconSize * 0.6;
                      return (
                        <g key={`floor-${polygon.id}-${pIdx}`} style={{ pointerEvents: 'none' }}>
                          <line x1={cx - iconSize} y1={cy} x2={cx + iconSize} y2={cy} stroke="white" strokeWidth={2} />
                          <polygon
                            points={`${cx},${arrowTip} ${cx - iconSize * 0.4},${cy} ${cx + iconSize * 0.4},${cy}`}
                            fill="white"
                          />
                        </g>
                      );
                    }

                    if (polygon.floor === 'bas') {
                      const arrowTip = cy + iconSize * 0.6;
                      return (
                        <g key={`floor-${polygon.id}-${pIdx}`} style={{ pointerEvents: 'none' }}>
                          <line x1={cx - iconSize} y1={cy} x2={cx + iconSize} y2={cy} stroke="white" strokeWidth={2} />
                          <polygon
                            points={`${cx},${arrowTip} ${cx - iconSize * 0.4},${cy} ${cx + iconSize * 0.4},${cy}`}
                            fill="white"
                          />
                        </g>
                      );
                    }

                    return null;
                  })()}
                </g>
              );
            })}

          </g>
        );
      })}

      {previewPolygons && previewPolygons.length > 0 && (
        <g className="preview-zone-layer" style={{ pointerEvents: "none" }}>
          {previewPolygons.map((poly, pIdx) => {
            const pointsString = poly.points.map((p) => `${p.x},${p.y}`).join(" ");
            return (
              <polygon
                key={`preview-${pIdx}`}
                points={pointsString}
                fill="rgba(136,136,136,0.1)"
                stroke="#888"
                strokeWidth={2}
                strokeDasharray="5,5"
              />
            );
          })}
        </g>
      )}

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
