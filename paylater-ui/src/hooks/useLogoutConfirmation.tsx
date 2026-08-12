import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { clearAuth } from '../auth/authStorage'
import LogoutConfirmModal from '../components/LogoutConfirmModal'

export function useLogoutConfirmation() {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)

  const requestLogout = useCallback(() => {
    setOpen(true)
  }, [])

  const cancelLogout = useCallback(() => {
    setOpen(false)
  }, [])

  const confirmLogout = useCallback(() => {
    setOpen(false)
    clearAuth()
    navigate('/login', { replace: true })
  }, [navigate])

  const logoutModal = (
    <LogoutConfirmModal
      open={open}
      onCancel={cancelLogout}
      onConfirm={confirmLogout}
    />
  )

  return {
    requestLogout,
    logoutModal,
  }
}
