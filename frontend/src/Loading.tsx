import './Loading.css'
import { PiSpinner as LoadingIcon } from 'react-icons/pi'

export default function Loading() {
  return (
    <div className="loading-ctr">
      <LoadingIcon />
    </div>
  )
}
