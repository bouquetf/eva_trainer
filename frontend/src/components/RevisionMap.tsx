import { useState, useEffect, useRef, useMemo } from 'react'
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
    const updateScale = () => {
      if (!containerRef.current || !image) return
      const containerWidth = containerRef.current.clientWidth
      const s = Math.min(1, containerWidth / dimensions.width)
      setScale(s)
    }
    updateScale()
    window.addEventListener('resize', updateScale)
    return () => window.removeEventListener('resize', updateScale)
  }, [image, dimensions])

  if (!image) {
    return <div className="spinner" />
  }

  return (
    <div ref={containerRef} style={{ width: '100%', overflow: 'hidden' }}>
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
        </Layer>
      </Stage>
    </div>
  )
}
