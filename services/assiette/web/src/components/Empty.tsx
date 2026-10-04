import type { ReactNode } from "react"
import { Button } from "@/components/ui/button"
import { setView, useT } from "@/lib/store"

export function Empty({
  children,
  discover,
}: {
  children: ReactNode
  /** Shows a button leading to the Discover view. */
  discover?: boolean
}) {
  const t = useT()
  return (
    <div className="rounded-2xl border border-dashed bg-card px-5 py-6 text-center">
      <p className="mt-1.5 mb-3.5 text-muted-foreground">{children}</p>
      {discover && (
        <Button onClick={() => setView("discover")}>{t.goDiscover}</Button>
      )}
    </div>
  )
}
