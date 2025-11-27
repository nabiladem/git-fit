import React, { useState, useEffect } from 'react'
import UploadForm from './components/UploadForm'

// App() - main application component
export default function App() {
  // state for tracking file, API response, errors, and validation states
  const [file, setFile] = useState(null)
  const [fileError, setFileError] = useState(null)
  const [theme, setTheme] = useState('dark')

  // Handle theme change
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
  }, [theme])

  const toggleTheme = () => {
    setTheme((prev) => (prev === 'dark' ? 'light' : 'dark'))
  }

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
    <div className="min-h-screen flex items-center justify-center p-6 transition-colors duration-500">
      <div className="relative w-full max-w-3xl bg-[var(--glass-bg)] backdrop-blur-2xl backdrop-saturate-200 rounded-3xl shadow-2xl border border-[var(--glass-border)] border-t-[var(--glass-highlight)] border-l-[var(--glass-highlight)] p-10 shadow-[var(--shadow-color)] ring-1 ring-[var(--glass-border)] transition-all duration-500">
        <header className="mb-8 text-center relative">
          {/* Theme Toggle */}
          <button
            onClick={toggleTheme}
            className="absolute right-0 top-0 p-2 rounded-full bg-[var(--glass-bg)] border border-[var(--glass-border)] text-[var(--text-primary)] hover:bg-[var(--glass-highlight)] transition-all duration-300 z-50 group"
            aria-label="Toggle theme"
          >
            {theme === 'dark' ? (
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
                  d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
                />
              </svg>
            ) : (
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
                  d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
                />
              </svg>
            )}
          </button>

          <h1 className="text-4xl font-bold text-[var(--text-primary)] drop-shadow-md mb-2 transition-colors duration-300">
            git fit
          </h1>

          <p className="text-[var(--text-secondary)] text-lg font-medium transition-colors duration-300">
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
