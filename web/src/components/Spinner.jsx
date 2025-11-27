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
        borderColor: `var(--glass-border)`,
        borderTopColor: 'var(--accent-color)',
        borderRadius: '50%',
        filter: 'drop-shadow(0 0 4px var(--glass-highlight))',
      }}
    ></div>
  )
}
