import React, { useState, useEffect } from 'react'
import Spinner from './Spinner'
import ComparisonSlider from './ComparisonSlider'

// UploadForm() - image upload and compression form component
export default function UploadForm({ file, onFileChange }) {
  // formatBytes() - formats bytes to appropriate unit
  const formatBytes = (bytes, decimals = 2, forceUnit = null) => {
    if (!+bytes) return '0 Bytes'

    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KB', 'MB', 'GB']

    if (forceUnit && sizes.includes(forceUnit)) {
      const i = sizes.indexOf(forceUnit)
      return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
    }

    if (bytes >= 1000000) {
      return `${(bytes / (1024 * 1024)).toFixed(decimals)} MB`
    }

    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
  }

  const [preview, setPreview] = useState(null)

  const [sizeValue, setSizeValue] = useState('1')
  const [sizeUnit, setSizeUnit] = useState('MB')
  const [maxSize, setMaxSize] = useState(1048576) // 1MB
  const [format, setFormat] = useState('jpeg')
  const [quality, setQuality] = useState(85)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [result, setResult] = useState(null)
  const [isDragging, setIsDragging] = useState(false)
  const [comparisonData, setComparisonData] = useState(null)

  // fetch NASA APOD
  useEffect(() => {
    fetch('https://api.nasa.gov/planetary/apod?api_key=DEMO_KEY')
      .then(res => res.json())
      .then(data => {
        if (data.url) {
          setComparisonData({
            before: data.url,
            after: data.url,
            isDemo: true,
            beforeLabel: 'Original (3.2 MB)',
            afterLabel: 'Compressed (1 MB)'
          })
        } else {
          throw new Error('No URL in response')
        }
      })
      .catch(err => {
        console.error('Failed to fetch NASA APOD:', err)

        // fallback image
        setComparisonData({
          before: 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?q=80&w=2072&auto=format&fit=crop',
          after: 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?q=80&w=2072&auto=format&fit=crop',
          isDemo: true,
          beforeLabel: 'Original (3.2 MB)',
          afterLabel: 'Compressed (1 MB)'
        })
      })
  }, [])

  // update preview when file changes
  useEffect(() => {
    if (!file) {
      setPreview(null)
      return
    }

    // create a URL for the file
    const url = URL.createObjectURL(file)
    setPreview(url)
    return () => URL.revokeObjectURL(url)
  }, [file])

  // update maxSize when sizeValue or sizeUnit changes
  useEffect(() => {
    const multiplier = sizeUnit === 'MB' ? 1024 * 1024 : 1024
    const val = parseFloat(sizeValue) || 0
    setMaxSize(val * multiplier)
  }, [sizeValue, sizeUnit])

  // handle drag and drop events
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

    // handle dropped file
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const syntheticEvent = {
        target: {
          files: e.dataTransfer.files,
        },
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

        // update comparison slider image
        if (file) {
          const objectUrl = URL.createObjectURL(file)
          setComparisonData({
            before: objectUrl,
            after: data.download_url,
            beforeLabel: `Original (${formatBytes(file.size)})`,
            afterLabel: `Compressed (${formatBytes(data.size)})`,
            isDemo: false
          })
        }
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

    // fetch to get the file as a blob
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
          className={`relative group border-2 border-dashed rounded-2xl transition-all duration-300 ease-out
            ${isDragging
              ? 'border-white bg-white/10 backdrop-blur-xl scale-[1.02] shadow-xl'
              : 'border-white/20 hover:border-white/40 bg-white/5 backdrop-blur-md hover:bg-white/10'
            }
            ${preview ? 'p-0 overflow-hidden' : 'p-10'}
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
                <p className="text-white font-medium">
                  Click or drop to replace
                </p>
              </div>
              <div className="absolute bottom-0 left-0 right-0 p-3 bg-gradient-to-t from-black/80 to-transparent text-white text-sm truncate">
                {file && file.name}
              </div>
            </div>
          ) : (
            <div className="text-center space-y-4 pointer-events-none">
              <div className="w-16 h-16 mx-auto bg-white/10 rounded-full flex items-center justify-center mb-4">
                <svg
                  className="w-8 h-8 text-white/80"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                  />
                </svg>
              </div>
              <div>
                <p className="text-lg font-medium text-white">
                  {isDragging
                    ? 'Drop image here'
                    : 'Click to upload or drag and drop'}
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
          Max size
          <div className="flex gap-2 mt-2">
            <input
              type="number"
              value={sizeValue}
              onChange={(e) => setSizeValue(e.target.value)}
              onBlur={() => {
                let val = parseFloat(sizeValue)
                const limit = sizeUnit === 'MB' ? 1 : 1024

                if (isNaN(val)) val = limit // Default to limit if invalid

                // Clamp value
                if (val > limit) val = limit
                if (val < 0.1) val = 0.1

                setSizeValue(String(val))
              }}
              className="block w-full bg-white/5 backdrop-blur-xl border border-white/10 rounded-xl px-4 py-3 text-white placeholder-white/50 focus:outline-none focus:ring-2 focus:ring-white/30 focus:border-white/30 transition-all hover:bg-white/10"
            />
            <select
              value={sizeUnit}
              onChange={(e) => {
                const newUnit = e.target.value
                setSizeUnit(newUnit)
                // Adjust value if it exceeds new limit
                const val = parseFloat(sizeValue)
                if (newUnit === 'MB' && val > 1) {
                  setSizeValue('1')
                }
              }}
              className="bg-white/5 backdrop-blur-xl border border-white/10 rounded-xl px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-white/30 focus:border-white/30 appearance-none cursor-pointer w-28 transition-all hover:bg-white/10"
            >
              <option value="KB" className="bg-gray-800 text-white">
                KB
              </option>
              <option value="MB" className="bg-gray-800 text-white">
                MB
              </option>
            </select>
          </div>
        </label>
        <label className="block text-sm font-medium text-white/90">
          Format
          <select
            value={format}
            onChange={(e) => setFormat(e.target.value)}
            className="mt-2 block w-full bg-white/5 backdrop-blur-xl border border-white/10 rounded-xl px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-white/30 focus:border-white/30 appearance-none cursor-pointer transition-all hover:bg-white/10"
            style={{
              backgroundImage: `url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='white' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e")`,
              backgroundPosition: `right 1rem center`,
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
          Quality: {quality}%
          <input
            type="range"
            value={quality}
            onChange={(e) => setQuality(Number(e.target.value))}
            min={1}
            max={100}
            className="mt-2 block w-full h-2 bg-white/20 rounded-lg appearance-none cursor-pointer accent-white hover:bg-white/30 transition-colors"
          />
        </label>
        <div />
      </div>

      {/* Submit Button */}
      <div>
        <button
          type="submit"
          disabled={loading}
          className="w-full inline-flex justify-center items-center gap-2 px-6 py-4 bg-white/10 hover:bg-white/20 text-white font-bold rounded-2xl transition-all duration-300 border border-white/20 shadow-lg hover:shadow-xl hover:scale-[1.02] active:scale-95 disabled:cursor-not-allowed backdrop-blur-xl backdrop-saturate-150"
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

      {/* Comparison Slider */}
      {comparisonData && (
        <div className="space-y-4 animate-fade-in">

          <ComparisonSlider
            before={comparisonData.before}
            after={comparisonData.after}
            labelBefore={comparisonData.beforeLabel || (comparisonData.isDemo ? "Original" : "Original")}
            labelAfter={comparisonData.afterLabel || (comparisonData.isDemo ? "Compressed (Simulated)" : "Compressed")}
          />
        </div>
      )}

      {/* Display the compression result */}
      {result && (
        <div className="p-6 border border-white/20 border-t-white/40 border-l-white/40 rounded-2xl bg-white/10 backdrop-blur-xl shadow-xl text-white animate-fade-in">
          <div className="space-y-3 text-sm">
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
