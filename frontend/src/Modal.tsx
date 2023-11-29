import { ReactNode, useEffect, useRef } from 'react'
import './Modal.css'
import { createPortal } from 'react-dom'
import { PiXCircle as CloseIcon } from 'react-icons/pi'

interface DialogProps {
  children?: ReactNode
  onClose: () => void
}

function Dialog({ children, onClose }: DialogProps) {
  const ref = useRef<HTMLDialogElement>(null)
  useEffect(() => {
    ref.current?.showModal()
  }, [])
  return (
    <dialog ref={ref} onClose={onClose} className="modal page-ctr">
      {children}
    </dialog>
  )
}

export interface ModalProps extends DialogProps {}

export default function Modal({ children, onClose }: ModalProps) {
  return (
    <div className="modal-ctr">
      {createPortal(
        <Dialog onClose={onClose}>
          <form className="modal-hdr" method="dialog">
            <button type="submit" className="modal-close-btn">
              <CloseIcon />
            </button>
          </form>
          {children}
        </Dialog>,
        document.body,
      )}
    </div>
  )
}
