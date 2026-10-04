import type { ReactNode } from "react"
import { cn } from "@/lib/utils"

/** Segmented control: a pill group of toggle buttons, one of them pressed. */
export function Seg<K extends string>({
  options,
  value,
  onChange,
  label,
  className,
  itemClassName,
}: {
  options: readonly { key: K; label: ReactNode }[]
  value: K
  onChange: (key: K) => void
  label?: string
  className?: string
  itemClassName?: string
}) {
  return (
    <div
      role="group"
      aria-label={label}
      className={cn(
        "inline-flex flex-wrap gap-0.5 rounded-full border bg-card p-[3px]",
        className,
      )}
    >
      {options.map((o) => (
        <button
          key={o.key}
          type="button"
          aria-pressed={value === o.key}
          onClick={() => onChange(o.key)}
          className={cn(
            "rounded-full px-3.5 py-1.5 text-[0.9rem] text-muted-foreground aria-pressed:bg-primary aria-pressed:font-semibold aria-pressed:text-primary-foreground",
            itemClassName,
          )}
        >
          {o.label}
        </button>
      ))}
    </div>
  )
}
