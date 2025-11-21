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
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="w-full max-w-3xl bg-white/30 backdrop-blur-lg rounded-2xl shadow-xl border border-white/20 p-8">
        <header className="mb-8 text-center">
          <h1 className="text-4xl font-bold text-white drop-shadow-md mb-2">
            git fit
          </h1>
          <p className="text-white/90 text-lg font-medium">
            Compress images to GitHub avatar limits (1MB).
          </p>
        </header>

        <main>
          <UploadForm file={file} onFileChange={handleFileChange} />

          {/* File validation errors */}
          {fileError && (
            <p className="text-red-100 bg-red-500/50 border border-red-500/50 rounded-lg p-3 font-medium mt-4 text-center">
              {fileError}
            </p>
          )}

          {/* Status + Result */}
          <div className="mt-6 space-y-4">
            {loading && (
              <div className="flex justify-center items-center">
                <Spinner size={40} />
              </div>
            )}
            {error && (
              <p className="text-red-100 bg-red-500/50 border border-red-500/50 rounded-lg p-3 font-medium text-center">
                {error}
              </p>
            )}

            {result && (
              <div className="p-6 border border-white/30 rounded-xl bg-white/20 backdrop-blur-md shadow-inner text-white">
                <h2 className="text-xl font-bold mb-4 text-center">
                  Compression Successful!
                </h2>
                <div className="space-y-2 text-sm">
                  <p>
                    <strong className="font-semibold">Filename:</strong>{' '}
                    {result.filename}
                  </p>
                  <p>
                    <strong className="font-semibold">Size:</strong>{' '}
                    {result.size} bytes
                  </p>
                  <p>
                    <strong className="font-semibold">MIME Type:</strong>{' '}
                    {result.mime}
                  </p>
                  <p className="pt-2">
                    <a
                      href={result.download_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-block bg-white/20 hover:bg-white/30 text-white font-bold py-2 px-4 rounded-lg transition-colors duration-200 border border-white/30"
                    >
                      Open in new tab
                    </a>
                  </p>
                  <p className="text-white/70 text-xs mt-2">
                    Expires in: {result.expires_in} seconds
                  </p>
                </div>
              </div>
            )}
          </div>
        </main>
      </div>
    </div>
  )
}
