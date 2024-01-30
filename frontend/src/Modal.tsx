import cls from './Modal.module.css'

import { Modal as MModal } from '@mantine/core'
import { ReactNode } from 'react'

export interface ModalProps {
  children?: ReactNode
  onClose: () => void
  title?: string
}

export default function Modal({ children, onClose, title }: ModalProps) {
  return (
    <MModal onClose={onClose} opened title={title} centered size={'xl'} className={cls.root}>
      {children}
    </MModal>
  )
}
