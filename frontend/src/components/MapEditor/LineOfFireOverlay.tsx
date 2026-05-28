import React from "react";
import type { Point } from "../../lib/api/types";
import type { LineOfFire, Cone } from "../../lib/lines_of_fire/types";

interface LineOfFireOverlayProps {
  linesOfFire: LineOfFire[];
  selectedLineOfFireId: number | null;
  tempConePoints: Point[];
  onOriginDoubleClick: (lineOfFireId: number, coneIndex: number) => void;
  onConeOriginDrag: (lofId: number, coneIndex: number, newOrigin: Point) => void;
  onConeTranslateDrag: (lofId: number, coneIndex: number, newOrigin: Point) => void;
  onConeEndDrag: (lofId: number, coneIndex: number, which: "left_end" | "right_end", newPos: Point) => void;
  onConeArcDrag: (lofId: number, coneIndex: number, arcBulge: number) => void;
  onLineOfFireDragEnd: (lofId: number, coneIndex: number) => void;
  screenPointToImagePoint: (clientX: number, clientY: number) => Point;
}

function ConeShape({
  cone,
  lofId,
  coneIndex,
  isSelected,
  onDoubleClick,
  onConeOriginDrag,
  onConeTranslateDrag,
  onConeEndDrag,
  onConeArcDrag,
  onLineOfFireDragEnd,
  screenPointToImagePoint,
}: {
  cone: Cone;
  lofId: number;
  coneIndex: number;
  isSelected: boolean;
  onDoubleClick: (coneIndex: number) => void;
  onConeOriginDrag: (lofId: number, coneIndex: number, newOrigin: Point) => void;
  onConeTranslateDrag: (lofId: number, coneIndex: number, newOrigin: Point) => void;
  onConeEndDrag: (lofId: number, coneIndex: number, which: "left_end" | "right_end", newPos: Point) => void;
  onConeArcDrag: (lofId: number, coneIndex: number, arcBulge: number) => void;
  onLineOfFireDragEnd: (lofId: number, coneIndex: number) => void;
  screenPointToImagePoint: (clientX: number, clientY: number) => Point;
}) {
  const { origin, left_end, right_end } = cone;
  const samePoint = left_end.x === right_end.x && left_end.y === right_end.y;

  const midX = (left_end.x + right_end.x) / 2;
  const midY = (left_end.y + right_end.y) / 2;
  const dx = midX - origin.x;
  const dy = midY - origin.y;
  const len = Math.sqrt(dx * dx + dy * dy);
  let conePath = "";
  let arcPath = "";
  if (!samePoint && len > 0) {
    const nx = dx / len;
    const ny = dy / len;
    const bulge = len * (cone.arc_bulge ?? 0.3);
    const cpx = midX + nx * bulge;
    const cpy = midY + ny * bulge;
    conePath = `M ${origin.x} ${origin.y} L ${left_end.x} ${left_end.y} Q ${cpx} ${cpy} ${right_end.x} ${right_end.y} Z`;
    arcPath = `M ${left_end.x} ${left_end.y} Q ${cpx} ${cpy} ${right_end.x} ${right_end.y}`;
  }

  function makeDragHandler(handle: "origin" | "translate" | "left_end" | "right_end" | "arc") {
    return (e: React.MouseEvent) => {
      e.stopPropagation();
      const startScreen = screenPointToImagePoint(e.clientX, e.clientY);
      const initialOrigin = { ...origin };
      const initialLeft = { ...left_end };
      const initialRight = { ...right_end };

      const onMove = (ev: MouseEvent) => {
        const currentImg = screenPointToImagePoint(ev.clientX, ev.clientY);
        const dx = currentImg.x - startScreen.x;
        const dy = currentImg.y - startScreen.y;

        const newOrigin = {
          x: initialOrigin.x + dx,
          y: initialOrigin.y + dy,
        };
        if (handle === "origin") {
          onConeOriginDrag(lofId, coneIndex, newOrigin);
        } else if (handle === "translate") {
          onConeTranslateDrag(lofId, coneIndex, newOrigin);
        } else if (handle === "arc") {
          const midX = (initialLeft.x + initialRight.x) / 2;
          const midY = (initialLeft.y + initialRight.y) / 2;
          const ax = midX - initialOrigin.x;
          const ay = midY - initialOrigin.y;
          const aLen = Math.sqrt(ax * ax + ay * ay);
          if (aLen > 0) {
            const dirX = ax / aLen;
            const dirY = ay / aLen;
            const deltaAlongAxis = dx * dirX + dy * dirY;
            const initialBulge = cone.arc_bulge ?? 0.3;
            const newBulge = Math.max(0.01, initialBulge + deltaAlongAxis / aLen);
            onConeArcDrag(lofId, coneIndex, newBulge);
          }
        } else {
          const base = handle === "left_end" ? initialLeft : initialRight;
          onConeEndDrag(lofId, coneIndex, handle, {
            x: base.x + dx,
            y: base.y + dy,
          });
        }
      };

      const onUp = () => {
        window.removeEventListener("mousemove", onMove);
        window.removeEventListener("mouseup", onUp);
        onLineOfFireDragEnd(lofId, coneIndex);
      };

      window.addEventListener("mousemove", onMove);
      window.addEventListener("mouseup", onUp);
    };
  }

  return (
    <g>
      {samePoint || conePath === "" ? (
        <line
          x1={origin.x}
          y1={origin.y}
          x2={left_end.x}
          y2={left_end.y}
          stroke="#ff6600"
          strokeWidth={2}
          style={{ cursor: "grab" }}
          onMouseDown={makeDragHandler("translate")}
        />
      ) : (
        <path
          d={conePath}
          fill="rgba(255, 100, 0, 0.15)"
          stroke="#ff6600"
          strokeWidth={isSelected ? 3 : 2}
          strokeLinejoin="round"
          style={{ cursor: "grab" }}
          onMouseDown={makeDragHandler("translate")}
        />
      )}
      {isSelected && (
        <>
          <circle
            cx={origin.x}
            cy={origin.y}
            r="6"
            fill="#ff6600"
            stroke="#fff"
            strokeWidth={2}
            style={{ cursor: "move" }}
            onMouseDown={makeDragHandler("origin")}
          />
          <circle
            cx={left_end.x}
            cy={left_end.y}
            r="5"
            fill="#ff8800"
            stroke="#fff"
            strokeWidth={1.5}
            style={{ cursor: "move" }}
            onMouseDown={makeDragHandler("left_end")}
          />
          <circle
            cx={right_end.x}
            cy={right_end.y}
            r="5"
            fill="#ff8800"
            stroke="#fff"
            strokeWidth={1.5}
            style={{ cursor: "move" }}
            onMouseDown={makeDragHandler("right_end")}
          />
          {arcPath && (
            <path
              d={arcPath}
              fill="none"
              stroke="transparent"
              strokeWidth={14}
              style={{ cursor: "ew-resize" }}
              onMouseDown={makeDragHandler("arc")}
            />
          )}
        </>
      )}
      <circle
        cx={origin.x}
        cy={origin.y}
        r="10"
        fill="transparent"
        style={{ cursor: "grab" }}
        onMouseDown={makeDragHandler("origin")}
        onDoubleClick={(e) => {
          e.stopPropagation();
          onDoubleClick(coneIndex);
        }}
      />
    </g>
  );
}

export const LineOfFireOverlay: React.FC<LineOfFireOverlayProps> = ({
  linesOfFire,
  selectedLineOfFireId,
  tempConePoints,
  onOriginDoubleClick,
  onConeOriginDrag,
  onConeTranslateDrag,
  onConeEndDrag,
  onConeArcDrag,
  onLineOfFireDragEnd,
  screenPointToImagePoint,
}) => {
  return (
    <g className="line-of-fire-layer">
      {linesOfFire.map((lof) => {
        const isSelected = lof.id === selectedLineOfFireId;
        return (
          <g
            key={lof.id}
            data-line-of-fire-id={lof.id}
          >
            {lof.polygon_data.cones.map((cone, idx) => (
              <g key={idx} data-cone-index={idx}>
                <ConeShape
                  cone={cone}
                  lofId={lof.id}
                  coneIndex={idx}
                  isSelected={isSelected}
                  onDoubleClick={(ci) => onOriginDoubleClick(lof.id, ci)}
                  onConeOriginDrag={onConeOriginDrag}
                  onConeTranslateDrag={onConeTranslateDrag}
                  onConeEndDrag={onConeEndDrag}
                  onConeArcDrag={onConeArcDrag}
                  onLineOfFireDragEnd={onLineOfFireDragEnd}
                  screenPointToImagePoint={screenPointToImagePoint}
                />
              </g>
            ))}
            {isSelected && lof.polygon_data.cones.length === 0 && (
              <text
                x={0}
                y={-10}
                fill="#ff6600"
                fontSize="12"
                style={{ pointerEvents: "none" }}
              >
                Aucun cône
              </text>
            )}
          </g>
        );
      })}
      {tempConePoints.length > 0 && (
        <g className="temp-cone" style={{ pointerEvents: "none" }}>
          {tempConePoints.map((p, i) => (
            <circle key={i} cx={p.x} cy={p.y} r="5" fill="#ff6600" stroke="#fff" strokeWidth={2} />
          ))}
          {tempConePoints.length === 2 && (
            <line
              x1={tempConePoints[0].x}
              y1={tempConePoints[0].y}
              x2={tempConePoints[1].x}
              y2={tempConePoints[1].y}
              stroke="#ff6600"
              strokeWidth={2}
              strokeDasharray="5,5"
            />
          )}
        </g>
      )}
    </g>
  );
};
