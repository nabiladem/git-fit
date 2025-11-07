import React, { useState } from 'react'
import UploadForm from './components/UploadForm'

// App() - main application component
export default function App() {
  // state for tracking file, API response, and errors
  const [file, setFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [result, setResult] = useState(null)

  // handle file selection in UploadForm
  const handleFileChange = (e) => {
    setFile(e.target.files[0])
    setResult(null)  // Reset previous result
  }

  // handle form submission to call backend API for compression
  const handleCompress = async () => {
    if (!file) {
      setError('Please select a file to upload.')
      return
    }

    setLoading(true)
    setError(null)

    const formData = new FormData()
    formData.append('avatar', file)

    try {
      // call backend to compress the image
      const response = await fetch('http://localhost:8080/api/compress', {
        method: 'POST',
        body: formData,
      })

      // check if the response is okay
      if (!response.ok) {
        throw new Error('Compression failed. Please try again.')
      }

      const data = await response.json()

      if (data.error) {
        throw new Error(data.error)
      }

      setResult(data)  // set the result of the compression
    } catch (err) {
      setError(err.message)  // handle error
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', padding: 20, fontFamily: 'sans-serif', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f3f4f6' }}>
      <div style={{ width: '100%', maxWidth: 720, background: '#fff', borderRadius: 8, boxShadow: '0 6px 18px rgba(0,0,0,0.08)', padding: 24 }}>
        <header style={{ marginBottom: 16, textAlign: 'left' }}>
          <h1 style={{ margin: 0, fontSize: 24 }}>git fit</h1>
          <p style={{ marginTop: 6, color: '#374151' }}>
            Compress images to GitHub avatar limits (1MB). Upload an image and download the compressed avatar.
          </p>
        </header>

        <main>
          <UploadForm file={file} onFileChange={handleFileChange} />
          <div style={{ marginTop: 20 }}>
            {loading && <p>Compressing image...</p>}
            {error && <p style={{ color: 'red' }}>{error}</p>}
            {result && (
              <div>
                <h2>Compression Successful!</h2>
                <p>Filename: {result.filename}</p>
                <p>Size: {result.size} bytes</p>
                <p>MIME Type: {result.mime}</p>
                <p>
                  <a href={result.download_url} download={result.filename}>
                    Download the compressed image
                  </a>
                </p>
                <p>Expires in: {result.expires_in} seconds</p>
              </div>
            )}
          </div>

          <button onClick={handleCompress} style={{ padding: '10px 20px', fontSize: 16, background: '#4CAF50', color: 'white', border: 'none', borderRadius: 4, cursor: 'pointer' }}>
            Compress Image
          </button>
        </main>
      </div>
    </div>
  )
}
