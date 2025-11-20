import React, { useState } from 'react'
import UploadForm from './components/UploadForm'

// App() - main application component
export default function App() {
  // state for tracking file, API response, errors, and validation states
  const [file, setFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [result, setResult] = useState(null)
  const [fileError, setFileError] = useState(null)

  // handle file selection in UploadForm
  // e - event object from file input change
  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0]
    setFile(selectedFile)
    setResult(null)
    setFileError(null)

    // file type validation
    if (selectedFile && !selectedFile.type.startsWith('image/')) {
      setFileError('Invalid file type. Please upload an image.')
    }
  }

  // handle form submission to call backend API for compression
  const handleCompress = async () => {
    if (!file) {
      setError('Please select a file to upload.')
      return
    }

    if (fileError) {
      setError(fileError)
      return
    }

    // reset states
    setLoading(true)
    setError(null)

    // prepare form data
    const formData = new FormData()
    formData.append('avatar', file)

    try {
      // use the API base URL from the .env file
      const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'

      // call backend to compress the image
      const response = await fetch(`${apiUrl}/api/compress`, {
        method: 'POST',
        body: formData,
      })

      if (!response.ok) {
        throw new Error('Compression failed. Please try again.')
      }

      // parse JSON response
      const data = await response.json()
      if (data.error) throw new Error(data.error)

      setResult(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100 font-sans p-6">
      <div className="w-full max-w-3xl bg-white rounded-lg shadow-lg p-6">
        <header className="mb-6 text-left">
          <h1 className="text-2xl font-semibold text-gray-900">git fit</h1>
          <p className="mt-2 text-gray-600">
            Compress images to GitHub avatar limits (1MB). Upload an image and
            download the compressed avatar.
          </p>
        </header>

        <main>
          <UploadForm file={file} onFileChange={handleFileChange} />

          {/* File validation errors */}
          {fileError && (
            <p className="text-red-600 font-medium mt-4">{fileError}</p>
          )}

          {/* Status + Result */}
          <div className="mt-6 space-y-4">
            {loading && (
              <div className="flex justify-center items-center">
                <div className="w-8 h-8 border-4 border-blue-500 border-dashed rounded-full animate-spin"></div>
              </div>
            )}
            {error && <p className="text-red-600 font-medium">{error}</p>}

            {result && (
              <div className="p-4 border rounded bg-green-50">
                <h2 className="text-lg font-semibold text-green-800">
                  Compression Successful!
                </h2>
                <p className="text-sm text-gray-700 mt-1">
                  <strong>Filename:</strong> {result.filename}
                </p>
                <p className="text-sm text-gray-700">
                  <strong>Size:</strong> {result.size} bytes
                </p>
                <p className="text-sm text-gray-700">
                  <strong>MIME Type:</strong> {result.mime}
                </p>
                <p className="text-sm text-gray-700 mt-1">
                  <a
                    href={result.download_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:text-blue-800 underline"
                  >
                    Open in new tab
                  </a>
                </p>
                <p className="text-sm text-gray-500 mt-1">
                  Expires in: {result.expires_in} seconds
                </p>
              </div>
            )}
          </div>
        </main>
      </div>
    </div>
  )
}
