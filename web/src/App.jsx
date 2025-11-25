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
      <div className="w-full max-w-3xl bg-white/10 backdrop-blur-2xl backdrop-saturate-200 rounded-3xl shadow-2xl border border-white/20 border-t-white/40 border-l-white/30 p-10 shadow-[0_8px_32px_0_rgba(0,0,0,0.36)] ring-1 ring-white/10">
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
            <p className="text-red-100 bg-red-500/10 backdrop-blur-xl border border-red-500/20 rounded-xl p-4 font-medium mt-4 text-center shadow-[0_4px_16px_0_rgba(220,38,38,0.2)] animate-fade-in">
              {fileError}
            </p>
          )}
        </main>
      </div>
    </div>
  )
}
