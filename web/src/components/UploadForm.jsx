import React, { useState, useEffect } from 'react'

export default function UploadForm() {
  const [file, setFile] = useState(null)
  const [preview, setPreview] = useState(null)
  const [maxSize, setMaxSize] = useState(1000000)
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

  async function onSubmit(e) {
    e.preventDefault()
    setError(null)
    setResult(null)

    if (!file) {
      setError('Please choose a file to upload')
      return
    }

    const fd = new FormData()
    fd.append('avatar', file, file.name)
    fd.append('maxsize', String(maxSize))
    fd.append('format', format)
    fd.append('quality', String(quality))

    setLoading(true)
    try {
      const res = await fetch('/api/compress', { method: 'POST', body: fd })
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

  function onFileChange(e) {
    const f = e.target.files && e.target.files[0]
    setFile(f || null)
  }

  function onDownload() {
    if (!result || !result.download_url) return
    // navigate to the signed download URL
    window.location.href = result.download_url
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700">Image</label>
        <input type="file" accept="image/*" onChange={onFileChange} className="mt-1" />
      </div>

      {preview && (
        <div className="flex items-center gap-4">
          <img src={preview} alt="preview" className="w-24 h-24 object-cover rounded" />
          <div className="text-sm text-gray-600">Selected: {file && file.name}</div>
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <label className="text-sm">Max size (bytes)
          <input type="number" value={maxSize} onChange={(e) => setMaxSize(Number(e.target.value))} className="mt-1 block w-full border rounded px-2 py-1" />
        </label>
        <label className="text-sm">Format
          <select value={format} onChange={(e) => setFormat(e.target.value)} className="mt-1 block w-full border rounded px-2 py-1">
            <option value="jpeg">jpeg</option>
            <option value="png">png</option>
            <option value="gif">gif</option>
          </select>
        </label>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <label className="text-sm">Quality
          <input type="number" value={quality} onChange={(e) => setQuality(Number(e.target.value))} min={1} max={100} className="mt-1 block w-full border rounded px-2 py-1" />
        </label>
        <div />
      </div>

      <div>
        <button type="submit" disabled={loading} className="inline-flex items-center px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">{loading ? 'Compressing…' : 'Compress'}</button>
      </div>

      {error && <div className="text-red-600">Error: {error}</div>}

      {result && (
        <div className="p-3 border rounded bg-green-50">
          <div><strong>Filename:</strong> {result.filename}</div>
          <div><strong>Size:</strong> {result.size} bytes</div>
          <div><strong>Type:</strong> {result.mime}</div>
          {result.download_url && (
            <div className="mt-2">
              <button onClick={onDownload} className="px-3 py-1 bg-blue-600 text-white rounded">Download</button>
              <a href={result.download_url} className="ml-3 text-sm text-blue-700" target="_blank" rel="noreferrer">Open in new tab</a>
            </div>
          )}
        </div>
      )}
    </form>
  )
}
