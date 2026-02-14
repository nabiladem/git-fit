import React, { useState, useRef, useEffect } from 'react'

// ComparisonSlider() - comparison slider component
/* before (string) - path to before image; after (string) - path to after image; labelBefore (string) - label for before image; labelAfter (string) - label for after image */
export default function ComparisonSlider({
  before,
  after,
  labelBefore,
  labelAfter,
}) {
  const [sliderPosition, setSliderPosition] = useState(50)
  const [isDragging, setIsDragging] = useState(false)
  const [containerWidth, setContainerWidth] = useState(0)
  const containerRef = useRef(null)

  // handle container width
  useEffect(() => {
    if (containerRef.current) {
      setContainerWidth(containerRef.current.offsetWidth)
    }

    const handleResize = () => {
      if (containerRef.current) {
        setContainerWidth(containerRef.current.offsetWidth)
      }
    }
    window.addEventListener('resize', handleResize)

    return () => window.removeEventListener('resize', handleResize)
  }, [])

  // reset slider position when images change
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setSliderPosition(50)
  }, [before, after])

  // handle slider movement
  const handleMove = (event) => {
    if (!containerRef.current) return

    const containerRect = containerRef.current.getBoundingClientRect()
    const x = (event.clientX || event.touches[0].clientX) - containerRect.left
    const position = (x / containerRect.width) * 100

    const knobRadius = 28
    const offsetPercent = (knobRadius / containerRect.width) * 100

    setSliderPosition(
      Math.min(100 + offsetPercent, Math.max(-offsetPercent, position))
    )
  }

  useEffect(() => {
    const handleWindowMove = (e) => {
      if (isDragging) handleMove(e)
    }

    const handleWindowUp = () => setIsDragging(false)

    window.addEventListener('mousemove', handleWindowMove)
    window.addEventListener('mouseup', handleWindowUp)
    window.addEventListener('touchmove', handleWindowMove)
    window.addEventListener('touchend', handleWindowUp)

    return () => {
      window.removeEventListener('mousemove', handleWindowMove)
      window.removeEventListener('mouseup', handleWindowUp)
      window.removeEventListener('touchmove', handleWindowMove)
      window.removeEventListener('touchend', handleWindowUp)
    }
  }, [isDragging])

  return (
    <div
      ref={containerRef}
      className="relative w-full h-[300px] sm:h-[400px] rounded-2xl overflow-hidden cursor-ew-resize select-none shadow-[var(--shadow-color)] border border-[var(--glass-border)] bg-[var(--glass-bg)] backdrop-blur-2xl backdrop-saturate-200 ring-1 ring-[var(--glass-border)]"
      onMouseDown={(e) => {
        setIsDragging(true)
        handleMove(e)
      }}
      onTouchStart={(e) => {
        setIsDragging(true)
        handleMove(e)
      }}
    >
      {/* After Image (Background) */}
      <img
        src={after}
        alt="After"
        className="absolute inset-0 w-full h-full object-cover"
      />
      <div className="absolute top-2 sm:top-4 right-2 sm:right-4 bg-black/40 backdrop-blur-xl border border-[var(--glass-border)] text-white px-2 sm:px-4 py-1 sm:py-1.5 rounded-full text-xs sm:text-sm font-medium whitespace-nowrap shadow-lg ring-1 ring-[var(--glass-border)]">
        {labelAfter}
      </div>

      {/* Before Image (Foreground - Clipped) */}
      <div
        className="absolute inset-0 w-full h-full overflow-hidden"
        style={{ width: `${sliderPosition}%` }}
      >
        <img
          src={before}
          alt="Before"
          className="absolute inset-0 w-full h-full object-cover max-w-none"
          style={{ width: containerWidth || '100%' }}
        />
        <div className="absolute top-2 sm:top-4 left-2 sm:left-4 bg-black/40 backdrop-blur-xl border border-[var(--glass-border)] text-white px-2 sm:px-4 py-1 sm:py-1.5 rounded-full text-xs sm:text-sm font-medium whitespace-nowrap shadow-lg ring-1 ring-[var(--glass-border)]">
          {labelBefore}
        </div>
      </div>

      {/* Slider Handle */}
      <div
        className="absolute top-0 bottom-0 w-1.5 cursor-ew-resize z-30 -translate-x-1/2"
        style={{ left: `${sliderPosition}%` }}
      >
        {/* The "Rod" - Refracting Light */}
        <div className="absolute inset-0 bg-[var(--glass-highlight)] backdrop-blur-md backdrop-saturate-200 backdrop-contrast-125 border-x border-[var(--glass-border)] shadow-[0_0_20px_var(--glass-highlight)]"></div>

        {/* The "Shine" - Reflecting Light */}
        <div className="absolute inset-0 bg-gradient-to-b from-white/90 via-transparent to-white/90 opacity-60 mix-blend-overlay"></div>

        {/* The "Knob" - Liquid Drop */}
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-14 h-14 rounded-full flex items-center justify-center group hover:scale-110 transition-transform duration-300">
          {/* Knob Background (Adaptive) */}
          <div className="absolute inset-0 rounded-full bg-[var(--glass-bg)] backdrop-blur-sm backdrop-brightness-110 shadow-[inset_0_0_12px_rgba(255,255,255,0.3),0_8px_20px_rgba(0,0,0,0.2)] border border-[var(--glass-border)] ring-1 ring-[var(--glass-border)]"></div>

          {/* Knob Reflection (Gloss) */}
          <div className="absolute inset-0 rounded-full bg-gradient-to-br from-white/60 to-transparent opacity-40 mix-blend-overlay"></div>

          <svg
            className="relative w-6 h-6 text-white drop-shadow-[0_2px_4px_rgba(0,0,0,0.3)]"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2.5}
              d="M8 9l4-4 4 4m0 6l-4 4-4-4"
              transform="rotate(90 12 12)"
            />
          </svg>
        </div>
      </div>
    </div>
  )
}
