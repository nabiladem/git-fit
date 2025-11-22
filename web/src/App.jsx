import React, { useState } from 'react'
import UploadForm from './components/UploadForm'

// App() - main application component
export default function App() {
  // state for tracking file, API response, errors, and validation states
  const [file, setFile] = useState(null)
  const [fileError, setFileError] = useState(null)

  // handle file selection in UploadForm
  // e - event object from file input change
  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0]
    setFile(selectedFile)
    setFileError(null)

    // file type validation
    if (selectedFile && !selectedFile.type.startsWith('image/')) {
      setFileError('Invalid file type. Please upload an image.')
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
        </main>
      </div>
    </div>
  )
}
