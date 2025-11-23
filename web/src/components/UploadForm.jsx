import React, { useState, useEffect } from 'react'
import Spinner from './Spinner'

// UploadForm() - image upload and compression form component
export default function UploadForm({ file, onFileChange }) {
  const [preview, setPreview] = useState(null)
  const [maxSize, setMaxSize] = useState(1048576) // 1MB
  const [format, setFormat] = useState('jpeg')
  const [quality, setQuality] = useState(85)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [result, setResult] = useState(null)
  const [isDragging, setIsDragging] = useState(false)

  // effect to update the preview when a new file is selected
  useEffect(() => {
    if (!file) {
      setPreview(null)
      return
    }

    const url = URL.createObjectURL(file)
    setPreview(url)
    return () => URL.revokeObjectURL(url)
  }, [file])

  const handleDragOver = (e) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const handleDragLeave = (e) => {
    e.preventDefault()
    setIsDragging(false)
  }

  const handleDrop = (e) => {
    e.preventDefault()
    setIsDragging(false)

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const syntheticEvent = {
        target: {
          files: e.dataTransfer.files
        }
      }
      onFileChange(syntheticEvent)
    }
  }

  // onSubmit() - handles form submission
  async function onSubmit(e) {
    e.preventDefault()
    setError(null)
    setResult(null)

    if (!file) {
      setError('Please choose a file to upload')
      return
    }

    // prepare form data to send to the backend
    const fd = new FormData()
    fd.append('avatar', file, file.name)
    fd.append('maxsize', String(maxSize))
    fd.append('format', format)
    fd.append('quality', String(quality))

    // API Base URL (for dev mode)
    const apiBase =
      import.meta.env && import.meta.env.DEV ? 'http://localhost:8080' : ''

    setLoading(true)
    try {
      // call backend API to compress the image
      const res = await fetch(apiBase + '/api/compress', {
        method: 'POST',
        body: fd,
      })
      const data = await res.json()

      if (!res.ok) {
        setError(data.error || data.message || 'Compression failed')
      } else {
        setResult(data)
      }
    } catch (err) {
      setError(err.message || String(err))
    } finally {
      setLoading(false)
    }
  }

  // onDownload() - handle file download
  function onDownload() {
    if (!result || !result.download_url) return

    // use fetch to get the file as a blob
    fetch(result.download_url)
      .then((response) => {
        if (!response.ok) throw new Error('Download failed')
        return response.blob()
      })
      .then((blob) => {
        const url = URL.createObjectURL(blob)

        // create a download link
        const a = document.createElement('a')
        a.href = url
        a.target = '_blank'
        a.download = result.filename
        document.body.appendChild(a)

        a.click()

        document.body.removeChild(a)
        URL.revokeObjectURL(url)
      })
      .catch((err) => {
        setError('Download failed: ' + err.message)
        console.error('Download error:', err)
      })
  }

  return (
    <form onSubmit={onSubmit} className="space-y-6">
      <div>
        <label className="block text-sm font-medium text-white/90 mb-2">
          Image
        </label>
        <div
          className={`relative group border-2 border-dashed rounded-xl transition-all duration-200 ease-in-out
            ${isDragging
              ? 'border-white bg-white/10 scale-[1.02]'
              : 'border-white/20 hover:border-white/40 bg-white/5'
            }
            ${preview ? 'p-0 overflow-hidden' : 'p-8'}
          `}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
        >
          <input
            type="file"
            accept="image/*"
            onChange={onFileChange}
            className="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
          />

          {preview ? (
            <div className="relative w-full h-64 bg-black/20 group-hover:bg-black/30 transition-colors">
              <img
                src={preview}
                alt="preview"
                className="w-full h-full object-contain"
              />
              <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black/40 backdrop-blur-sm">
                <p className="text-white font-medium">Click or drop to replace</p>
              </div>
              <div className="absolute bottom-0 left-0 right-0 p-3 bg-gradient-to-t from-black/80 to-transparent text-white text-sm truncate">
                {file && file.name}
              </div>
            </div>
          ) : (
            <div className="text-center space-y-4 pointer-events-none">
              <div className="w-16 h-16 mx-auto bg-white/10 rounded-full flex items-center justify-center mb-4">
                <svg className="w-8 h-8 text-white/80" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                </svg>
              </div>
              <div>
                <p className="text-lg font-medium text-white">
                  {isDragging ? 'Drop image here' : 'Click to upload or drag and drop'}
                </p>
                <p className="text-sm text-white/60 mt-1">
                  SVG, PNG, JPG or GIF (max 5MB)
                </p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Form Controls */}
      <div className="grid grid-cols-2 gap-6">
        <label className="block text-sm font-medium text-white/90">
          Max size (bytes)
          <input
            type="number"
            value={maxSize}
            onChange={(e) => setMaxSize(Number(e.target.value))}
            className="mt-2 block w-full bg-white/5 border border-white/20 rounded-lg px-3 py-2 text-white placeholder-white/50 focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent"
          />
        </label>
        <label className="block text-sm font-medium text-white/90">
          Format
          <select
            value={format}
            onChange={(e) => setFormat(e.target.value)}
            className="mt-2 block w-full bg-white/5 border border-white/20 rounded-lg px-3 py-2 text-white focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent appearance-none cursor-pointer"
            style={{
              backgroundImage: `url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='white' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e")`,
              backgroundPosition: `right 0.5rem center`,
              backgroundRepeat: `no-repeat`,
              backgroundSize: `1.5em 1.5em`,
            }}
          >
            <option value="jpeg" className="bg-gray-800 text-white">
              jpeg
            </option>
            <option value="png" className="bg-gray-800 text-white">
              png
            </option>
            <option value="gif" className="bg-gray-800 text-white">
              gif
            </option>
          </select>
        </label>
      </div>

      <div className="grid grid-cols-2 gap-6">
        <label className="block text-sm font-medium text-white/90">
          Quality
          <input
            type="number"
            value={quality}
            onChange={(e) => setQuality(Number(e.target.value))}
            min={1}
            max={100}
            className="mt-2 block w-full bg-white/5 border border-white/20 rounded-lg px-3 py-2 text-white placeholder-white/50 focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent"
          />
        </label>
        <div />
      </div>

      {/* Submit Button */}
      <div>
        <button
          type="submit"
          disabled={loading}
          className="w-full inline-flex justify-center items-center gap-2 px-6 py-3 bg-white/20 hover:bg-white/30 text-white font-bold rounded-xl transition-all duration-200 border border-white/30 shadow-lg hover:shadow-xl active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? (
            <>
              <Spinner size={24} />
              <span>Compressing...</span>
            </>
          ) : (
            'Compress Image'
          )}
        </button>
      </div>

      {/* Error Messages */}
      {error && (
        <div className="text-red-100 bg-red-500/50 border border-red-500/50 rounded-lg p-3 text-center">
          Error: {error}
        </div>
      )}

      {/* Display the compression result */}
      {result && (
        <div className="p-4 border border-white/20 border-t-white/50 border-l-white/50 rounded-xl bg-white/10 backdrop-blur-xl shadow-lg text-white">
          <div className="space-y-2 text-sm">
            <p>
              <strong className="font-semibold">Filename:</strong>{' '}
              {result.filename}
            </p>
            <p>
              <strong className="font-semibold">Size:</strong> {result.size}{' '}
              bytes
            </p>
            <p>
              <strong className="font-semibold">Type:</strong> {result.mime}
            </p>
          </div>
          {result.download_url && (
            <div className="mt-4 flex items-center gap-4">
              <button
                onClick={onDownload}
                className="px-4 py-2 bg-white/20 hover:bg-white/30 text-white rounded-lg border border-white/30 transition-colors"
                type="button"
              >
                Download
              </button>
              <a
                href={result.download_url}
                className="text-white/80 hover:text-white underline text-sm"
                target="_blank"
                rel="noreferrer"
              >
                Open in new tab
              </a>
            </div>
          )}
        </div>
      )}
    </form>
  )
}
