import { useT } from "@/lib/store"

/** − / + control for a number of servings. */
export function Stepper({
  persons,
  onChange,
}: {
  persons: number
  onChange: (delta: number) => void
}) {
  const t = useT()
  return (
    <span
      role="group"
      aria-label={t.persGroup}
      className="inline-flex items-center overflow-hidden rounded-full border bg-card font-sans"
    >
      <button
        type="button"
        className="h-8 w-[34px] text-lg leading-none"
        aria-label={t.lessP}
        onClick={() => onChange(-1)}
      >
        −
      </button>
      <span className="min-w-14 text-center text-sm tabular-nums">
        {t.pers(persons)}
      </span>
      <button
        type="button"
        className="h-8 w-[34px] text-lg leading-none"
        aria-label={t.moreP}
        onClick={() => onChange(1)}
      >
        +
      </button>
    </span>
  )
}
