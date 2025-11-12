import React from 'react'

export default function Spinner({ size = 6, color = 'border-green-600' }) {
  return (
    <div
      className={`w-${size} h-${size} border-4 border-gray-200 ${color} border-t-transparent rounded-full animate-spin`}
      role="status"
      aria-label="Loading"
    ></div>
  )
}
