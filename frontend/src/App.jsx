import { useState, useEffect } from 'react'
import { supabase } from './supabaseClient'
import Auth from './components/Auth'
import ImageUpload from './components/ImageUpload'
import './App.css' // You can add basic styling here

function App() {
  const [session, setSession] = useState(null)

  useEffect(() => {
    // Check active session on load
    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(session)
    })

    // Listen for changes (login/logout)
    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, session) => {
      setSession(session)
    })

    return () => subscription.unsubscribe()
  }, [])

  return (
    <div className="container">
      <h1>AI Avatar Gen 🚀</h1>
      {!session ? (
        <Auth />
      ) : (
        <div className="dashboard">
          <p>Welcome, {session.user.email}!</p>
          <ImageUpload session={session} />
          <button onClick={() => supabase.auth.signOut()}>Sign Out</button>
        </div>
      )}
    </div>
  )
}

export default App