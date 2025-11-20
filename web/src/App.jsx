import React, { useState } from 'react'
import UploadForm from './components/UploadForm'

// App() - main application component
export default function App() {
  // state for tracking file and validation states
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
        </main>
      </div>
    </div>
  )
}
