import React from 'react'
import UploadForm from './components/UploadForm'

// main application component
export default function App() {
  return (
    <div style={{minHeight:'100vh', padding:20, fontFamily:'sans-serif', display:'flex', alignItems:'center', justifyContent:'center', background:'#f3f4f6'}}>
      <div style={{width:'100%', maxWidth:720, background:'#fff', borderRadius:8, boxShadow:'0 6px 18px rgba(0,0,0,0.08)', padding:24}}>
        <header style={{marginBottom:16, textAlign:'left'}}>
          <h1 style={{margin:0, fontSize:24}}>git fit</h1>
          <p style={{marginTop:6, color:'#374151'}}>Compress images to GitHub avatar limits (1MB). Upload an image and download the compressed avatar.</p>
        </header>

        <main>
          <UploadForm />
        </main>
      </div>
    </div>
  )
}
