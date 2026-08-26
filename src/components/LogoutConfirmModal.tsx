import { useEffect, useId, useRef } from 'react'
import './LogoutConfirmModal.css'

type LogoutConfirmModalProps = {
  open: boolean
  onCancel: () => void
  onConfirm: () => void
}

function LogoutConfirmModal({
  open,
  onCancel,
  onConfirm,
}: LogoutConfirmModalProps) {
  const titleId = useId()
  const descriptionId = useId()
  const cancelRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (!open) {
      return
    }

    cancelRef.current?.focus()

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        onCancel()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [open, onCancel])

  if (!open) {
    return null
  }

  return (
    <div className="logout-modal" role="presentation">
      <button
        type="button"
        className="logout-modal__backdrop"
        aria-label="Close logout confirmation"
        onClick={onCancel}
      />
      <div
        className="logout-modal__dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descriptionId}
      >
        <h2 id={titleId} className="logout-modal__title">
          Logout
        </h2>
        <p id={descriptionId} className="logout-modal__message">
          Are you sure you want to logout?
        </p>
        <div className="logout-modal__actions">
          <button
            ref={cancelRef}
            type="button"
            className="logout-modal__button"
            onClick={onCancel}
          >
            Cancel
          </button>
          <button
            type="button"
            className="logout-modal__button logout-modal__button--danger"
            onClick={onConfirm}
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  )
}

export default LogoutConfirmModal
