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
  const [copied, setCopied] = useState(false)
  const fileInputRef = React.useRef(null)

  // fetch NASA APOD with caching
  useEffect(() => {
    const fetchAPOD = async () => {
      try {
        // check cache first
        const today = new Date().toISOString().split('T')[0]
        const cachedData = localStorage.getItem('apod_cache')
        const cachedDate = localStorage.getItem('apod_date')

        if (cachedData && cachedDate === today) {
          console.log('Using cached APOD')
          setComparisonData(JSON.parse(cachedData))
          return
        }

        const apiKey = 'NASA_API_KEY'
        const response = await fetch(
          `https://api.nasa.gov/planetary/apod?api_key=${apiKey}`
        )

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }

        const data = await response.json()

        if (data.media_type !== 'image' || !data.url) {
          throw new Error('APOD is not an image today')
        }

        const apodData = {
          before: data.url,
          after: data.url,
          isDemo: true,
          beforeLabel: 'Original (3.2 MB)',
          afterLabel: 'Compressed (1 MB)',
        }

        setComparisonData(apodData)

        // cache for today
        localStorage.setItem('apod_cache', JSON.stringify(apodData))
        localStorage.setItem('apod_date', today)

        console.log('Loaded fresh APOD:', data.title)
      } catch (err) {
        console.error('Failed to fetch NASA APOD:', err)

        const cachedData = localStorage.getItem('apod_cache')
        if (cachedData) {
          console.log('Using previous APOD from cache')
          setComparisonData(JSON.parse(cachedData))
          return
        }

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

    fetchAPOD()
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
        <label className="block text-sm font-semibold text-white drop-shadow-sm mb-2 ml-1">
          Image
        </label>
        <div
          className={`relative group border-2 border-dashed rounded-2xl transition-all duration-500 ease-out
            ${isDragging
              ? 'border-white bg-white/10 backdrop-blur-xl scale-[1.02] shadow-[0_8px_32px_0_rgba(31,38,135,0.37)]'
              : 'border-white/20 hover:border-white/40 bg-white/5 backdrop-blur-md hover:bg-white/10 shadow-[inset_0_2px_4px_0_rgba(0,0,0,0.1)]'
            }
            ${preview ? 'p-0 overflow-hidden border-white/10' : 'p-10'}
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
        <label className="block text-sm font-semibold text-white drop-shadow-sm ml-1">
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
              className="block w-full bg-white/5 backdrop-blur-xl border border-white/30 rounded-xl px-4 py-3 text-white placeholder-white/50 focus:outline-none focus:ring-2 focus:ring-white/30 focus:border-white/40 transition-all duration-300 ease-out hover:bg-white/10 focus:bg-white/15 shadow-[inset_0_2px_4px_rgba(0,0,0,0.1)]"
            />
            <div className="relative flex bg-white/5 p-1 rounded-xl backdrop-blur-md border border-white/20 shadow-inner transition-all duration-300 ease-out">
              {/* Sliding background */}
              <div
                className="absolute top-1 bottom-1 bg-white/20 rounded-lg backdrop-blur-md shadow-[0_4px_16px_0_rgba(31,38,135,0.37)] ring-1 ring-white/20 transition-transform duration-500 ease-[cubic-bezier(0.34,1.56,0.64,1)]"
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
                    // Adjust value if it exceeds new limit
                    const val = parseFloat(sizeValue)
                    if (unit === 'MB' && val > 1) {
                      setSizeValue('1')
                    }
                  }}
                  className={`
                    relative z-10 px-3 py-3 rounded-lg text-sm font-bold transition-colors duration-200 uppercase tracking-wide
                    ${sizeUnit === unit ? 'text-white' : 'text-white/60 hover:text-white'}
                  `}
                >
                  {unit}
                </button>
              ))}
            </div>
          </div>
        </label>
        <label className="block text-sm font-semibold text-white drop-shadow-sm ml-1">
          Format
          <div className="relative mt-2 flex bg-white/5 p-1 rounded-xl backdrop-blur-md border border-white/20 shadow-inner transition-all duration-300 ease-out">
            {/* Sliding background */}
            <div
              className="absolute top-1 bottom-1 bg-white/20 rounded-lg backdrop-blur-md shadow-[0_4px_16px_0_rgba(31,38,135,0.37)] ring-1 ring-white/20 transition-transform duration-500 ease-[cubic-bezier(0.34,1.56,0.64,1)]"
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
                  ${format === fmt ? 'text-white' : 'text-white/60 hover:text-white'}
                `}
              >
                {fmt}
              </button>
            ))}
          </div>
        </label>
      </div>

      {format === 'jpeg' && (
        <div className="space-y-2">
          <div className="flex justify-between items-center ml-1">
            <label className="block text-sm font-semibold text-white drop-shadow-sm">
              Quality
            </label>
            <span className="text-sm font-medium text-white/80 bg-black/20 px-2 py-0.5 rounded-md border border-white/10">
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
          className="w-full inline-flex justify-center items-center gap-2 px-6 py-4 bg-gradient-to-br from-white/20 to-white/5 hover:from-white/30 hover:to-white/10 text-white font-bold rounded-2xl transition-all duration-300 border border-white/30 shadow-[0_8px_32px_0_rgba(31,38,135,0.37)] hover:shadow-[0_8px_32px_0_rgba(31,38,135,0.5)] hover:scale-[1.02] active:scale-95 disabled:cursor-not-allowed backdrop-blur-xl backdrop-saturate-150 shadow-[inset_0_1px_0_rgba(255,255,255,0.4)]"
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
            labelBefore={
              comparisonData.beforeLabel ||
              (comparisonData.isDemo ? 'Original' : 'Original')
            }
            labelAfter={
              comparisonData.afterLabel ||
              (comparisonData.isDemo ? 'Compressed (Simulated)' : 'Compressed')
            }
          />
        </div>
      )}

      {/* Display the compression result */}
      {result && (
        <div className="p-6 border border-white/20 border-t-white/40 border-l-white/40 rounded-2xl bg-gradient-to-br from-white/15 to-white/5 backdrop-blur-xl shadow-[0_8px_32px_0_rgba(0,0,0,0.36)] text-white animate-scale-in ring-1 ring-white/10">
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
                type="button"
                onClick={onDownload}
                className="flex-1 bg-gradient-to-br from-white/20 to-white/5 hover:from-white/30 hover:to-white/10 text-white py-3 px-6 rounded-xl font-bold border border-white/30 hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 backdrop-blur-md shadow-[0_4px_16px_0_rgba(31,38,135,0.37)] shadow-[inset_0_1px_0_rgba(255,255,255,0.4)]"
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
                className="bg-gradient-to-br from-white/20 to-white/5 hover:from-white/30 hover:to-white/10 text-white py-3 px-6 rounded-xl font-bold border border-white/30 hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 flex items-center justify-center gap-2 backdrop-blur-md shadow-[0_4px_16px_0_rgba(31,38,135,0.37)] shadow-[inset_0_1px_0_rgba(255,255,255,0.4)]"
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
