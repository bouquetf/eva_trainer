import { useState, useEffect, useRef, useMemo, Fragment } from 'react'
import { Stage, Layer, Path, Line, Image as KonvaImage } from 'react-konva'
import useImage from 'use-image'
import { getMapImageUrl } from '../api/maps'
import type { ZonePolygon } from '../lib/revision/types'

function useHatchPattern(): HTMLImageElement | undefined {
  return useMemo(() => {
    const canvas = document.createElement('canvas')
    canvas.width = 8
    canvas.height = 8
    const ctx = canvas.getContext('2d')
    if (!ctx) return undefined
    ctx.strokeStyle = '#ff5252'
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(0, 0)
    ctx.lineTo(8, 8)
    ctx.stroke()
    const img = new window.Image()
    img.src = canvas.toDataURL()
    return img
  }, [])
}

function pointsToSvgPath(points: { x: number; y: number }[]): string {
  if (points.length === 0) return ''
  return 'M ' + points.map(p => `${p.x} ${p.y}`).join(' L ') + ' Z'
}

export function calcAdaptiveScale(
  containerWidth: number,
  containerHeight: number,
  imageWidth: number,
  imageHeight: number
): number {
  if (containerWidth <= 0 || containerHeight <= 0 || imageWidth <= 0 || imageHeight <= 0) {
    return 1
  }
  return Math.min(containerWidth / imageWidth, containerHeight / imageHeight)
}

interface RevisionMapProps {
  mapId: number
  polygons: ZonePolygon[]
}

export function RevisionMap({ mapId, polygons }: RevisionMapProps) {
  const [image] = useImage(getMapImageUrl(mapId))
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 })
  const containerRef = useRef<HTMLDivElement>(null)
  const [scale, setScale] = useState(1)
  const hatchCanvas = useHatchPattern()

  useEffect(() => {
    if (image) {
      setDimensions({ width: image.width, height: image.height })
    }
  }, [image])

  useEffect(() => {
    const container = containerRef.current
    if (!container || !image) return

    const updateScale = () => {
      const { clientWidth: cw, clientHeight: ch } = container
      if (cw === 0 || ch === 0 || dimensions.width === 0 || dimensions.height === 0) return
      const s = Math.min(cw / dimensions.width, ch / dimensions.height)
      setScale(s)
    }

    updateScale()
    const observer = new ResizeObserver(updateScale)
    observer.observe(container)
    return () => observer.disconnect()
  }, [image, dimensions])

  if (!image) {
    return <div className="spinner" />
  }

  return (
    <div ref={containerRef} style={{ width: '100%', height: '100%', overflow: 'hidden' }}>
      <Stage
        width={dimensions.width * scale}
        height={dimensions.height * scale}
        scaleX={scale}
        scaleY={scale}
      >
        <Layer>
          <KonvaImage image={image} />
          {polygons.map((polygon) => (
            <Line
              key={polygon.id}
              points={polygon.points.flatMap(p => [p.x, p.y])}
              fill="rgba(255,80,80,0.15)"
              stroke="#ff5252"
              strokeWidth={4}
              closed
            />
          ))}
          {hatchCanvas && polygons.map((polygon) => (
            <Path
              key={`hatch-${polygon.id}`}
              data={pointsToSvgPath(polygon.points)}
              fillPatternImage={hatchCanvas}
              fillPatternRepeat="repeat"
              listening={false}
            />
          ))}
          {polygons.map((polygon) => {
            if (!polygon.floor) return null;
            const pts = polygon.points;
            const cx = pts.reduce((s, p) => s + p.x, 0) / pts.length;
            const cy = pts.reduce((s, p) => s + p.y, 0) / pts.length;
            const iconSize = 30;
            const arrowTip = polygon.floor === 'haut'
              ? cy - iconSize * 0.6
              : cy + iconSize * 0.6;
            return (
              <Fragment key={`floor-${polygon.id}`}>
                <Line
                  points={[cx - iconSize, cy, cx + iconSize, cy]}
                  stroke="white"
                  strokeWidth={2}
                  listening={false}
                />
                <Line
                  points={[cx, arrowTip, cx - iconSize * 0.4, cy, cx + iconSize * 0.4, cy]}
                  closed
                  fill="white"
                  stroke="white"
                  strokeWidth={1}
                  listening={false}
                />
              </Fragment>
            );
          })}
        </Layer>
      </Stage>
    </div>
  )
}
