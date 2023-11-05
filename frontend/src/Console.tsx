export interface ConsoleProps {
  lines: string[]
}

export default function Console({ lines }: ConsoleProps) {
  return (
    <div>
      {lines.map((ln, i) => (
        <pre key={i}>{ln}</pre>
      ))}
    </div>
  )
}
