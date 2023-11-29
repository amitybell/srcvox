import './Credits.css'

export default function Credits() {
  return (
    <div className="credits-ctr">
      <h2>Amity Bell</h2>
      <p>
        SrcVox is developed by{' '}
        <a rel="noreferrer" target="_blank" href="https://amitybell.com/">
          https://amitybell.com/
        </a>
      </p>

      <h2>Piper</h2>
      <p>
        The text-to-speech functionality is provided by{' '}
        <a rel="noreferrer" target="_blank" href="https://github.com/rhasspy/piper">
          https://github.com/rhasspy/piper
        </a>
      </p>

      <h2>Wails</h2>
      <p>
        UI and App is created with{' '}
        <a rel="noreferrer" target="_blank" href="https://wails.io/">
          https://wails.io/
        </a>{' '}
        and{' '}
        <a rel="noreferrer" target="_blank" href="https://react.dev/">
          https://react.dev/
        </a>
      </p>

      <h2>IP to country mapping</h2>
      <p>
        Server IP address to contry mapping is provided by{' '}
        <a rel="noreferrer" target="_blank" href="https://www.maxmind.com/">
          https://www.maxmind.com/
        </a>
      </p>

      <h2>Other Dependencies</h2>
      <p>See links below for the full list of dependencies:</p>
      <ul>
        <li>
          <a
            rel="noreferrer"
            target="_blank"
            href="https://github.com/amitybell/srcvox/blob/master/go.mod"
          >
            https://github.com/amitybell/srcvox/blob/master/go.mod
          </a>
        </li>
        <li>
          <a
            rel="noreferrer"
            target="_blank"
            href="https://github.com/amitybell/srcvox/blob/master/package.json"
          >
            https://github.com/amitybell/srcvox/blob/master/package.json
          </a>
        </li>
      </ul>
    </div>
  )
}
