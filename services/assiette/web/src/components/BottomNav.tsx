import { BookOpen, Compass, Ellipsis, Heart, ShoppingCart } from "lucide-react"
import { setView, useApp, useT, type View } from "@/lib/store"

const ITEMS: { view: View; Icon: typeof Compass }[] = [
  { view: "discover", Icon: Compass },
  { view: "fav", Icon: Heart },
  { view: "mine", Icon: BookOpen },
  { view: "shop", Icon: ShoppingCart },
  { view: "plus", Icon: Ellipsis },
]

export function BottomNav() {
  const { view, st } = useApp()
  const t = useT()
  const nMine = Object.keys(st.mine).length
  return (
    <nav
      aria-label="Sections"
      className="fixed inset-x-0 bottom-0 z-10 border-t bg-card pb-[env(safe-area-inset-bottom,0px)]"
    >
      <div className="mx-auto grid max-w-[680px] grid-cols-5">
        {ITEMS.map(({ view: v, Icon }) => {
          const n = v === "mine" ? nMine : 0
          return (
            <button
              key={v}
              type="button"
              aria-current={view === v ? "page" : undefined}
              aria-label={t.nav[v] + (n ? " (" + n + ")" : "")}
              title={t.nav[v]}
              onClick={() => setView(v)}
              className="flex items-center justify-center px-1 pt-3 pb-3.5 text-muted-foreground aria-[current=page]:text-foreground aria-[current=page]:shadow-[inset_0_3px_0_var(--primary)]"
            >
              <Icon className="size-[22px]" />
              {n > 0 && (
                <b className="ml-1 inline-block min-w-[18px] rounded-[9px] bg-primary px-1.25 text-[0.72rem] leading-[18px] text-primary-foreground">
                  {n}
                </b>
              )}
            </button>
          )
        })}
      </div>
    </nav>
  )
}
