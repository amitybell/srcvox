import './Spinner.css'
import { PiSpinnerThin as SpinnerIcon } from 'react-icons/pi'

export default function Spinner() {
  return (
    <span className="spinner-ctr">
      <SpinnerIcon className="spinner" />
    </span>
  )
}
