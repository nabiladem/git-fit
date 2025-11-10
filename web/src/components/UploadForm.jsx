import React, { useState, useEffect } from 'react'

// UploadForm() - image upload and compression form component
export default function UploadForm({ file, onFileChange }) {
  const [preview, setPreview] = useState(null)
  const [maxSize, setMaxSize] = useState(1048576) // 1MB
  const [format, setFormat] = useState('jpeg')
  const [quality, setQuality] = useState(85)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [result, setResult] = useState(null)

  useEffect(() => {
    if (!file) {
      setPreview(null)
      return
    }

    const url = URL.createObjectURL(file)
    setPreview(url)
    return () => URL.revokeObjectURL(url)
  }, [file])

  // onSubmit() - handles form submission
  async function onSubmit(e) {
    e.preventDefault()
    setError(null)
    setResult(null)

    if (!file) {
      setError('Please choose a file to upload')
      return
    }

    // prepare form data
    const fd = new FormData()
    fd.append('avatar', file, file.name)
    fd.append('maxsize', String(maxSize))
    fd.append('format', format)
    fd.append('quality', String(quality))

    // API Base URL (for dev mode)
    const apiBase = import.meta.env && import.meta.env.DEV ? 'http://localhost:8080' : ''

    setLoading(true)
    try {
      // call backend API to compress the image
      const res = await fetch(apiBase + '/api/compress', { method: 'POST', body: fd })
      const data = await res.json()

      if (!res.ok) {
        setError(data.error || data.message || 'Compression failed')
      } else {
        // store result on successful compression
        setResult(data)
      }
    } catch (err) {
      setError(err.message || String(err)) // handle any fetch errors
    } finally {
      setLoading(false)
    }
  }

  // onDownload() - handle file download
  function onDownload() {
    if (!result || !result.download_url) return

    // use fetch to get the file
    fetch(result.download_url)
      .then((response) => {
        if (!response.ok) throw new Error('Download failed')
        return response.blob() // get the file as a blob
      })
      .then((blob) => {
        // create a temporary URL for the blob
        const url = URL.createObjectURL(blob)

        // create a temporary <a> element for downloading
        const a = document.createElement('a')
        a.href = url
        a.target = '_blank'
        a.download = result.filename // use the filename from the result
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
    <form onSubmit={onSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700">Image</label>
        <input type="file" accept="image/*" onChange={onFileChange} className="mt-1" />
      </div>

      {/* Image Preview */}
      {preview && (
        <div className="flex items-center gap-4">
          <img src={preview} alt="preview" className="w-24 h-24 object-cover rounded" />
          <div className="text-sm text-gray-600">Selected: {file && file.name}</div>
        </div>
      )}

      {/* Form Controls */}
      <div className="grid grid-cols-2 gap-4">
        <label className="text-sm">Max size (bytes)
          <input
            type="number"
            value={maxSize}
            onChange={(e) => setMaxSize(Number(e.target.value))}
            className="mt-1 block w-full border rounded px-2 py-1"
          />
        </label>
        <label className="text-sm">Format
          <select
            value={format}
            onChange={(e) => setFormat(e.target.value)}
            className="mt-1 block w-full border rounded px-2 py-1"
          >
            <option value="jpeg">jpeg</option>
            <option value="png">png</option>
            <option value="gif">gif</option>
          </select>
        </label>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <label className="text-sm">Quality
          <input
            type="number"
            value={quality}
            onChange={(e) => setQuality(Number(e.target.value))}
            min={1}
            max={100}
            className="mt-1 block w-full border rounded px-2 py-1"
          />
        </label>
        <div />
      </div>

      {/* Submit Button */}
      <div>
        <button
          type="submit"
          disabled={loading}
          className="inline-flex items-center px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
        >
          {loading ? 'Compressing…' : 'Compress'}
        </button>
      </div>

      {/* Error Messages */}
      {error && <div className="text-red-600">Error: {error}</div>}

      {/* Display the compression result */}
      {result && (
        <div className="p-3 border rounded bg-green-50">
          <div><strong>Filename:</strong> {result.filename}</div>
          <div><strong>Size:</strong> {result.size} bytes</div>
          <div><strong>Type:</strong> {result.mime}</div>
          {result.download_url && (
            <div className="mt-2">
              <button
                onClick={onDownload}
                className="px-3 py-1 bg-blue-600 text-white rounded"
              >
                Download
              </button>
              <a
                href={result.download_url}
                className="ml-3 text-sm text-blue-700"
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
