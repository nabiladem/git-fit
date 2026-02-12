import React, { useState, useEffect } from 'react'
import Spinner from './Spinner'
import ComparisonSlider from './ComparisonSlider'

// UploadForm() - image upload and compression form component
/* file - the selected file to compress
   onFileChange - callback function to handle file selection */
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
  const [copied, setCopied] = useState(false)
  const fileInputRef = React.useRef(null)
  const intervalRef = React.useRef(null)

  // startChanging() - starts continuous value update
  // direction: 1 for increase, -1 for decrease
  const startChanging = (direction) => {
    updateValue(direction)

    if (intervalRef.current) window.clearInterval(intervalRef.current)

    intervalRef.current = window.setInterval(() => {
      updateValue(direction)
    }, 250)
  }

  // stopChanging() - stops continuous value update
  const stopChanging = () => {
    if (intervalRef.current) {
      window.clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }

  // updateValue() - helper to update size value
  // direction: 1 for increase, -1 for decrease
  const updateValue = (direction) => {
    setSizeValue((prev) => {
      let val = parseFloat(prev) || 0
      const isMB = sizeUnit === 'MB'
      const step = isMB ? 0.1 : 10
      const max = isMB ? 1 : 1024
      const min = isMB ? 0.1 : 10

      let newVal = val + direction * step
      if (newVal > max) newVal = max
      if (newVal < min) newVal = min

      return isMB ? newVal.toFixed(1) : String(Math.round(newVal))
    })
  }

  // loadAPOD() - fetches or loads cached APOD data
  const loadAPOD = async () => {
    try {
      // check cache first
      const today = new Date().toLocaleDateString('en-CA', {
        timeZone: 'America/New_York',
      })
      const cachedData = localStorage.getItem('apod_cache')
      const cachedDate = localStorage.getItem('apod_date')
      const cachedTimestamp = localStorage.getItem('apod_timestamp')

      // use cache if it's from today and less than 1 hour old
      if (cachedData && cachedDate === today && cachedTimestamp) {
        const cacheAge = Date.now() - parseInt(cachedTimestamp)
        const oneHour = 60 * 60 * 1000

        if (cacheAge < oneHour) {
          setComparisonData(JSON.parse(cachedData))
          return
        }
      }

      const apiKey =
        import.meta.env.NASA_API_KEY ||
        import.meta.env.VITE_NASA_API_KEY ||
        'DEMO_KEY'
      const response = await fetch(
        `https://api.nasa.gov/planetary/apod?api_key=${apiKey}`
      )

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()

      // validate APOD data
      if (data.media_type !== 'image' || !data.url) {
        throw new Error('APOD is not an image today')
      }

      // prepare APOD data
      const apodData = {
        before: data.url,
        after: data.url,
        isDemo: true,
        beforeLabel: 'Original (3.2 MB)',
        afterLabel: 'Compressed (1 MB)',
      }

      setComparisonData(apodData)

      // cache for today with timestamp
      localStorage.setItem('apod_cache', JSON.stringify(apodData))
      localStorage.setItem('apod_date', today)
      localStorage.setItem('apod_timestamp', Date.now().toString())
    } catch (err) {
      console.error('Failed to fetch NASA APOD:', err)

      const cachedData = localStorage.getItem('apod_cache')
      if (cachedData) {
        setComparisonData(JSON.parse(cachedData))
        return
      }

      // fallback image
      setComparisonData({
        before:
          'https://images.unsplash.com/photo-1451187580459-43490279c0fa?q=80&w=2072&auto=format&fit=crop',
        after:
          'https://images.unsplash.com/photo-1451187580459-43490279c0fa?q=80&w=2072&auto=format&fit=crop',
        isDemo: true,
        beforeLabel: 'Original (3.2 MB)',
        afterLabel: 'Compressed (1 MB)',
      })
    }
  }

  // fetch NASA APOD with caching
  useEffect(() => {
    loadAPOD()
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
    setMaxSize(Math.floor(val * multiplier))
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
  // e - event object from form submission
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
    // sanitize filename to prevent "The string did not match the expected pattern" error
    // remove control characters, newlines, and other invalid characters
    const sanitizedFilename = file.name.replace(/[\x00-\x1F\x7F]/g, '')
    fd.append('avatar', file, sanitizedFilename || 'image')
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
            isDemo: false,
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

  // copyToClipboard() - copy link to clipboard
  const copyToClipboard = async () => {
    if (!result?.download_url) return

    try {
      await navigator.clipboard.writeText(result.download_url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  return (
    <form onSubmit={onSubmit} className="space-y-6 animate-slide-up">
      <div>
        <label className="block text-sm font-semibold text-[var(--text-primary)] drop-shadow-sm mb-2 ml-1">
          Image
        </label>
        <div
          className={`relative group border-2 border-dashed rounded-2xl transition-all duration-500 ease-out
            ${isDragging
              ? 'border-[var(--glass-border)] bg-[var(--glass-bg)] backdrop-blur-xl scale-[1.02] shadow-[var(--shadow-color)]'
              : 'border-[var(--glass-border)] hover:border-[var(--glass-highlight)] bg-[var(--glass-bg)] backdrop-blur-md hover:bg-[var(--glass-highlight)] shadow-[inset_0_2px_4px_0_rgba(0,0,0,0.1)]'
            }
            ${preview ? 'p-0 overflow-hidden border-[var(--glass-border)]' : 'p-10'}
          `}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
        >
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            onChange={onFileChange}
            className="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
          />

          {preview ? (
            <div className="relative w-full h-64 bg-black/20 group-hover:bg-black/30 transition-all duration-300">
              <img
                src={preview}
                alt="preview"
                className="w-full h-full object-contain"
              />
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation()
                  if (fileInputRef.current) {
                    fileInputRef.current.value = ''
                  }
                  onFileChange({ target: { files: [] } })
                }}
                className="absolute top-2 right-2 z-20 w-6 h-6 flex items-center justify-center rounded-full bg-black/50 hover:bg-black/70 border border-white/20 backdrop-blur-sm transition-all duration-200"
                aria-label="Remove image"
              >
                <svg
                  className="w-3.5 h-3.5 text-white"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
              <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-all duration-300 bg-black/40 backdrop-blur-sm">
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
              <div className="w-16 h-16 mx-auto bg-[var(--glass-bg)] rounded-full flex items-center justify-center mb-4">
                <svg
                  className="w-8 h-8 text-[var(--text-secondary)]"
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
                <p className="text-lg font-medium text-[var(--text-primary)]">
                  {isDragging
                    ? 'Drop image here'
                    : 'Click to upload or drag and drop'}
                </p>
                <p className="text-sm text-[var(--text-secondary)] mt-1">
                  SVG, PNG, JPG or GIF (max 5MB)
                </p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Form Controls */}
      <div className="grid grid-cols-2 gap-6">
        <label className="block text-sm font-semibold text-white drop-shadow-sm ml-1">
          Max Size
          <div className="flex gap-2 mt-2">
            <div className="relative flex-1 group">
              <input
                type="number"
                value={sizeValue}
                onChange={(e) => setSizeValue(e.target.value)}
                onBlur={() => {
                  let val = parseFloat(sizeValue)
                  const isMB = sizeUnit === 'MB'
                  const max = isMB ? 1 : 1024
                  const min = isMB ? 0.1 : 10

                  if (isNaN(val)) val = max
                  if (val > max) val = max
                  if (val < min) val = min

                  // Format to remove unnecessary decimals for KB
                  setSizeValue(isMB ? String(val) : String(Math.round(val)))
                }}
                className="block w-full h-full bg-[var(--input-bg)] backdrop-blur-xl border border-[var(--glass-border)] rounded-xl pl-4 pr-10 py-3 text-[var(--text-primary)] placeholder-[var(--text-secondary)] focus:outline-none focus:ring-2 focus:ring-[var(--glass-border)] focus:border-[var(--glass-highlight)] transition-all duration-300 ease-out hover:bg-[var(--glass-highlight)] focus:bg-[var(--glass-highlight)] shadow-[inset_0_2px_4px_rgba(0,0,0,0.1)] no-spinner"
              />
              {/* Custom Spin Buttons */}
              <div className="absolute right-1 top-1 bottom-1 flex flex-col w-8 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                <button
                  type="button"
                  onMouseDown={() => startChanging(1)}
                  onMouseUp={stopChanging}
                  onMouseLeave={stopChanging}
                  className="flex-1 flex items-center justify-center hover:bg-[var(--glass-highlight)] rounded-t-lg text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
                >
                  <svg
                    className="w-3 h-3"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2.5}
                      d="M5 15l7-7 7 7"
                    />
                  </svg>
                </button>
                <button
                  type="button"
                  onMouseDown={() => startChanging(-1)}
                  onMouseUp={stopChanging}
                  onMouseLeave={stopChanging}
                  className="flex-1 flex items-center justify-center hover:bg-[var(--glass-highlight)] rounded-b-lg text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
                >
                  <svg
                    className="w-3 h-3"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2.5}
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <div className="relative flex bg-[var(--input-bg)] p-1 rounded-xl backdrop-blur-md border border-[var(--glass-border)] shadow-inner transition-all duration-300 ease-out">
              {/* Sliding background */}
              <div
                className="absolute top-1 bottom-1 bg-[var(--glass-highlight)] rounded-lg backdrop-blur-md ring-1 ring-[var(--glass-border)] transition-transform duration-500 ease-[cubic-bezier(0.34,1.56,0.64,1)]"
                style={{
                  width: 'calc(50% - 4px)',
                  transform: `translateX(${sizeUnit === 'MB' ? '100%' : '0%'})`,
                }}
              />
              {['KB', 'MB'].map((unit) => (
                <button
                  key={unit}
                  type="button"
                  onClick={() => {
                    setSizeUnit(unit)
                    const val = parseFloat(sizeValue)
                    // Auto-adjust value when switching units if out of bounds
                    if (unit === 'MB') {
                      if (val > 1) setSizeValue('1')
                      if (val <= 0) setSizeValue('0.1')
                    } else {
                      if (val > 1024) setSizeValue('1024')
                      if (val < 10) setSizeValue('10')
                    }
                  }}
                  className={`
                    relative z-10 px-3 py-3 rounded-lg text-sm font-bold transition-colors duration-200 uppercase tracking-wide
                    ${sizeUnit === unit ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}
                  `}
                >
                  {unit}
                </button>
              ))}
            </div>
          </div>
        </label>
        <label className="block text-sm font-semibold text-[var(--text-primary)] drop-shadow-sm ml-1">
          Format
          <div className="relative mt-2 flex bg-[var(--input-bg)] p-1 rounded-xl backdrop-blur-md border border-[var(--glass-border)] shadow-inner transition-all duration-300 ease-out">
            {/* Sliding background */}
            <div
              className="absolute top-1 bottom-1 bg-[var(--glass-highlight)] rounded-lg backdrop-blur-md ring-1 ring-[var(--glass-border)] transition-transform duration-500 ease-[cubic-bezier(0.34,1.56,0.64,1)]"
              style={{
                width: 'calc(33.333% - 4px)',
                transform: `translateX(${format === 'jpeg' ? '0%' : format === 'png' ? '100%' : '200%'})`,
              }}
            />
            {['jpeg', 'png', 'gif'].map((fmt) => (
              <button
                key={fmt}
                type="button"
                onClick={() => setFormat(fmt)}
                className={`
                  relative z-10 flex-1 py-3 rounded-lg text-sm font-bold transition-colors duration-200 uppercase tracking-wide
                  ${format === fmt ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}
                `}
              >
                {fmt}
              </button>
            ))}
          </div>
        </label>
      </div>

      {format === 'jpeg' && (
        <div className="space-y-2 animate-fade-in">
          <div className="flex justify-between items-center ml-1">
            <label className="block text-sm font-semibold text-[var(--text-primary)] drop-shadow-sm">
              Quality
            </label>
            <span className="text-sm font-medium text-[var(--text-secondary)] bg-[var(--input-bg)] px-2 py-0.5 rounded-md border border-[var(--glass-border)]">
              {quality}%
            </span>
          </div>
          <div className="relative h-6 flex items-center">
            <input
              type="range"
              value={quality}
              onChange={(e) => setQuality(Number(e.target.value))}
              min={1}
              max={100}
              className="w-full"
            />
          </div>
        </div>
      )}

      {/* Submit Button */}
      <div>
        <button
          type="submit"
          disabled={loading}
          className="w-full inline-flex justify-center items-center gap-2 px-6 py-4 bg-gradient-to-br from-[var(--glass-highlight)] to-[var(--glass-bg)] hover:from-[var(--glass-border)] hover:to-[var(--glass-highlight)] text-[var(--text-primary)] font-bold rounded-2xl transition-all duration-300 border border-[var(--glass-border)] hover:scale-[1.02] active:scale-95 disabled:cursor-not-allowed backdrop-blur-xl backdrop-saturate-150 shadow-[inset_0_1px_0_var(--glass-highlight)]"
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
        <div className="text-red-100 bg-red-500/10 backdrop-blur-xl border border-red-500/20 rounded-xl p-4 text-center shadow-[0_4px_16px_0_rgba(220,38,38,0.2)] animate-fade-in">
          Error: {error}
        </div>
      )}

      {/* Comparison Slider */}
      {comparisonData && (
        <div className="space-y-4 animate-fade-in">
          <ComparisonSlider
            before={comparisonData.before}
            after={comparisonData.after}
            labelBefore={comparisonData.beforeLabel || 'Original'}
            labelAfter={comparisonData.afterLabel || 'Compressed'}
          />
        </div>
      )}

      {/* Display the compression result */}
      {result && (
        <div className="p-6 border border-[var(--glass-border)] border-t-[var(--glass-highlight)] border-l-[var(--glass-highlight)] rounded-2xl bg-[var(--glass-bg)] backdrop-blur-xl shadow-[var(--shadow-color)] text-[var(--text-primary)] animate-scale-in ring-1 ring-[var(--glass-border)]">
          {/* Success Header */}
          <div className="flex items-center justify-between gap-3 mb-5 pb-4 border-b border-[var(--glass-border)]">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-green-500/10 backdrop-blur-sm flex items-center justify-center ring-1 ring-green-400/30 animate-pulse">
                <svg
                  className="w-6 h-6 text-green-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  style={{
                    filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.8))',
                  }}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2.5}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
              <div>
                <h3
                  className="font-bold text-[var(--text-primary)] text-lg"
                  style={{
                    textShadow:
                      '0 2px 8px rgba(0,0,0,0.8), 0 0 2px rgba(0,0,0,0.8)',
                  }}
                >
                  Compression Complete
                </h3>
                <p
                  className="text-xs text-[var(--text-primary)] mt-0.5"
                  style={{ textShadow: '0 1px 4px rgba(0,0,0,0.8)' }}
                >
                  Your image is ready to download
                </p>
              </div>
            </div>
            <button
              onClick={() => {
                setResult(null)
                if (fileInputRef.current) {
                  fileInputRef.current.value = ''
                }
                loadAPOD()
              }}
              className="p-2 rounded-full hover:bg-[var(--glass-highlight)] transition-colors duration-200 text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
              aria-label="Close"
            >
              <svg
                className="w-5 h-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          {/* Compression Stats Banner */}
          {file && (
            <div className="mb-5 p-4 rounded-xl bg-gradient-to-r from-green-500/5 via-emerald-500/5 to-blue-500/5 border border-[var(--glass-border)] backdrop-blur-2xl bg-[var(--stats-bg)]">
              <div className="grid grid-cols-3 gap-4 text-center">
                <div>
                  <div
                    className="text-2xl font-bold text-green-400"
                    style={{
                      textShadow:
                        '0 2px 8px rgba(0,0,0,0.8), 0 0 2px rgba(0,0,0,0.8)',
                    }}
                  >
                    {(((file.size - result.size) / file.size) * 100).toFixed(1)}
                    %
                  </div>
                  <div
                    className="text-xs text-[var(--text-primary)] uppercase tracking-wide font-medium mt-1"
                    style={{ textShadow: '0 1px 4px rgba(0,0,0,0.8)' }}
                  >
                    Size Reduced
                  </div>
                </div>
                <div>
                  <div
                    className="text-2xl font-bold text-[var(--text-primary)]"
                    style={{
                      textShadow:
                        '0 2px 8px rgba(0,0,0,0.8), 0 0 2px rgba(0,0,0,0.8)',
                    }}
                  >
                    {formatBytes(file.size - result.size)}
                  </div>
                  <div
                    className="text-xs text-[var(--text-primary)] uppercase tracking-wide font-medium mt-1"
                    style={{ textShadow: '0 1px 4px rgba(0,0,0,0.8)' }}
                  >
                    Space Saved
                  </div>
                </div>
                <div>
                  <div
                    className="text-2xl font-bold text-blue-400"
                    style={{
                      textShadow:
                        '0 2px 8px rgba(0,0,0,0.8), 0 0 2px rgba(0,0,0,0.8)',
                    }}
                  >
                    {(file.size / result.size).toFixed(1)}:1
                  </div>
                  <div
                    className="text-xs text-[var(--text-primary)] uppercase tracking-wide font-medium mt-1"
                    style={{ textShadow: '0 1px 4px rgba(0,0,0,0.8)' }}
                  >
                    Compression
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Stats Grid */}
          <div className="grid grid-cols-2 gap-4 mb-5">
            <div className="flex flex-col p-3 rounded-lg bg-[var(--input-bg)] border border-[var(--glass-border)]">
              <span className="text-xs text-[var(--text-secondary)] uppercase tracking-wide font-medium mb-1">
                Filename
              </span>
              <span
                className="text-sm font-semibold text-[var(--text-primary)] truncate"
                title={result.filename}
              >
                {result.filename}
              </span>
            </div>
            <div className="flex flex-col p-3 rounded-lg bg-[var(--input-bg)] border border-[var(--glass-border)]">
              <span className="text-xs text-[var(--text-secondary)] uppercase tracking-wide font-medium mb-1">
                File Size
              </span>
              <span className="text-sm font-semibold text-[var(--text-primary)]">
                {result.size} bytes
              </span>
            </div>
            <div className="flex flex-col p-3 rounded-lg bg-[var(--input-bg)] border border-[var(--glass-border)] col-span-2">
              <span className="text-xs text-[var(--text-secondary)] uppercase tracking-wide font-medium mb-1">
                Format
              </span>
              <span className="text-sm font-semibold text-[var(--text-primary)] uppercase">
                {result.mime}
              </span>
            </div>
          </div>

          {/* Action Buttons */}
          {result.download_url && (
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={onDownload}
                className="flex-1 bg-gradient-to-br from-[var(--glass-highlight)] to-[var(--glass-bg)] hover:from-[var(--glass-border)] hover:to-[var(--glass-highlight)] text-[var(--text-primary)] py-3 px-6 rounded-xl font-bold border border-[var(--glass-border)] hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 backdrop-blur-md shadow-[inset_0_1px_0_var(--glass-highlight)]"
              >
                <svg
                  className="w-5 h-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                  />
                </svg>
                Download
              </button>
              <button
                type="button"
                onClick={copyToClipboard}
                className="bg-gradient-to-br from-[var(--glass-highlight)] to-[var(--glass-bg)] hover:from-[var(--glass-border)] hover:to-[var(--glass-highlight)] text-[var(--text-primary)] py-3 px-6 rounded-xl font-bold border border-[var(--glass-border)] hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 backdrop-blur-md shadow-[inset_0_1px_0_var(--glass-highlight)]"
              >
                {copied ? (
                  <>
                    <svg
                      className="w-5 h-5 text-green-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                    Copied!
                  </>
                ) : (
                  <>
                    <svg
                      className="w-5 h-5"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                      />
                    </svg>
                    Copy Link
                  </>
                )}
              </button>
            </div>
          )}
        </div>
      )}
    </form>
  )
}
