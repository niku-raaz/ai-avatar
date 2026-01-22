import { useState } from 'react'
import { supabase } from '../supabaseClient'
import axios from 'axios'

export default function ImageUpload({ session }) {
  const [file, setFile] = useState(null)
  const [uploading, setUploading] = useState(false)
  const [status, setStatus] = useState('')

  const uploadImage = async () => {
    try {
      if (!file) {
        alert('You must select an image to upload.')
        return
      }
      setUploading(true)
      setStatus('Uploading image...')

      // 1. Upload to Supabase Storage
      // Ensure you created a bucket named 'avatars' in Supabase Dashboard
      const fileExt = file.name.split('.').pop()
      const fileName = `${Math.random()}.${fileExt}`
      const filePath = `${session.user.id}/${fileName}`

      const { error: uploadError } = await supabase.storage
        .from('avatars')
        .upload(filePath, file)

      if (uploadError) throw uploadError

      // 2. Get the Public URL
      const { data } = supabase.storage
        .from('avatars')
        .getPublicUrl(filePath)
      
      const publicUrl = data.publicUrl
      setStatus('Image uploaded! sending to backend...')

      // 3. Send to Go Backend
      await axios.post(`${import.meta.env.VITE_API_URL}/generate`, {
        user_id: session.user.id,
        user_image_url: publicUrl
      })

      setStatus('Success! AI is generating your avatar.')
      
    } catch (error) {
      alert(error.message)
      setStatus('Error occurred.')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="upload-section">
      <h3>Upload your selfie</h3>
      <input
        type="file"
        accept="image/*"
        onChange={(e) => setFile(e.target.files[0])}
        disabled={uploading}
      />
      <button onClick={uploadImage} disabled={uploading}>
        {uploading ? 'Processing...' : 'Generate Avatar'}
      </button>
      <p>{status}</p>
    </div>
  )
}