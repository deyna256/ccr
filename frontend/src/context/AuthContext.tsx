import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { getAccessToken } from '../api/client'

interface AuthContextType {
  isAuthenticated: boolean
  loading: boolean
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  loading: true,
})

export function useAuth() {
  return useContext(AuthContext)
}

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setIsAuthenticated(!!getAccessToken())
    setLoading(false)
  }, [])

  return (
    <AuthContext.Provider value={{ isAuthenticated, loading }}>
      {children}
    </AuthContext.Provider>
  )
}