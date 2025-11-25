import React from 'react'

// Spinner() - a simple loading spinner component
export default function Spinner({ size = 28 }) {
  return (
    <div
      className={`animate-spin inline-block`}
      style={{
        width: size,
        height: size,
        borderWidth: '3px',
        borderStyle: 'solid',
        borderColor: `rgba(255, 255, 255, 0.1)`,
        borderTopColor: 'rgba(255, 255, 255, 0.9)',
        borderRadius: '50%',
        filter: 'drop-shadow(0 0 4px rgba(255, 255, 255, 0.3))',
      }}
    ></div>
  )
}
