import { ReactNode } from 'react'
import './Modal.css'
import { Modal as MModal } from '@mantine/core'

export interface ModalProps {
  children?: ReactNode
  onClose: () => void
  title?: string
}

export default function Modal({ children, onClose, title }: ModalProps) {
  return (
    <MModal onClose={onClose} opened title={title} centered size={'xl'} className="modal">
      {children}
    </MModal>
  )
}
