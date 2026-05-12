import { createContext, useContext } from 'react'

export const AuthContext = createContext({
  role: 'admin',
  setRole: () => {},
  unitID: 'unit001',
  setUnitID: () => {},
})

export function useAuth() {
  return useContext(AuthContext)
}
