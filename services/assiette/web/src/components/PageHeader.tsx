import type { ReactNode } from "react"

export function PageHeader({
  title,
  sub,
  children,
}: {
  title: ReactNode
  sub?: ReactNode
  children?: ReactNode
}) {
  return (
    <header className="mb-4 flex items-end justify-between gap-3">
      <div>
        {title}
        {sub && <p className="mt-0.5 text-muted-foreground">{sub}</p>}
      </div>
      {children && <div className="flex items-center gap-2">{children}</div>}
    </header>
  )
}
