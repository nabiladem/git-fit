import React from 'react'

// Spinner() - a simple loading spinner component
export default function Spinner({ size = 28, color = 'border-white' }) {
  return (
    <div
      className={`animate-spin inline-block`}
      style={{
        width: size,
        height: size,
        borderWidth: '3px',
        borderStyle: 'solid',
        borderColor: `rgba(255, 255, 255, 0.3)`,
        borderTopColor: 'white',
        borderRadius: '50%',
      }}
    ></div>
  )
}
