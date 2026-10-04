import { Plus } from "lucide-react"
import type { FormEvent } from "react"
import { Empty } from "@/components/Empty"
import { IconButton } from "@/components/IconButton"
import { PageHeader } from "@/components/PageHeader"
import { RefreshButton } from "@/components/RefreshButton"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { addExtra, copyText, removeExtra, setBought } from "@/lib/actions"
import { CATALOG, RAYONS } from "@/lib/catalog"
import { ingName, rayonName } from "@/lib/format"
import { fmtQty, shoppingGroups, type ShoppingItem } from "@/lib/shopping"
import { useApp, useT } from "@/lib/store"
import { cn } from "@/lib/utils"

/** Delay before a checked item moves to the basket, so the tick is visible (ms). */
const BUY_DELAY = 250

function Item({ it, done }: { it: ShoppingItem; done?: boolean }) {
  const { lang } = useApp()
  const t = useT()
  const qty = done ? it.qty : it.rem
  return (
    <label className="flex items-center gap-3 border-b px-3 py-2.5 last:border-b-0">
      <input
        type="checkbox"
        className="size-5 flex-none accent-primary"
        defaultChecked={done}
        onChange={(e) => {
          const bought = e.target.checked
          setTimeout(() => setBought(it.key, bought), bought ? BUY_DELAY : 0)
        }}
      />
      <span className={cn(done && "line-through opacity-55")}>
        {ingName(it.name, lang)}
      </span>
      {it.extra ? (
        <button
          type="button"
          className="ml-auto px-1 text-lg leading-none text-muted-foreground"
          aria-label={t.extraRm + " " + it.name}
          onClick={() => removeExtra(it.key.slice(2))}
        >
          ×
        </button>
      ) : (
        <span
          className={cn(
            "ml-auto whitespace-nowrap text-muted-foreground tabular-nums",
            done && "line-through opacity-55",
          )}
        >
          {!done && it.partial && t.still}
          {fmtQty(qty ?? 0, it.unit ?? "", lang, t)}
        </span>
      )}
    </label>
  )
}

function ExtraForm() {
  const t = useT()

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = e.currentTarget
    const f = new FormData(form)
    const name = String(f.get("name") || "").trim()
    if (!name) return
    addExtra(name, String(f.get("rayon")))
    form.reset()
    form.querySelector("input")?.focus()
  }

  return (
    <form className="mb-4.5 flex flex-wrap gap-2" onSubmit={onSubmit}>
      <Input
        name="name"
        required
        maxLength={80}
        placeholder={t.extraPh}
        aria-label={t.extraPh}
        className="min-w-0 flex-[3_1_220px]"
      />
      <div className="flex flex-[1_0_auto] items-center gap-2">
        <select
          name="rayon"
          aria-label={t.extraRayon}
          defaultValue="Épicerie"
          className="h-10 flex-1 rounded-lg border bg-card px-3"
        >
          {RAYONS.map((r) => (
            <option key={r} value={r}>
              {rayonName(r, t)}
            </option>
          ))}
        </select>
        <IconButton label={t.add} type="submit">
          <Plus />
        </IconButton>
      </div>
    </form>
  )
}

export function Shopping() {
  const { st, lang } = useApp()
  const t = useT()
  const { groups, done, count } = shoppingGroups(
    st,
    CATALOG,
    (n) => ingName(n, lang),
    lang,
  )
  const left = groups.reduce((n, g) => n + g.items.length, 0)

  function copy() {
    const text = groups
      .map(
        (g) =>
          rayonName(g.rayon, t) +
          "\n" +
          g.items
            .map((it) => {
              const q = fmtQty(it.rem ?? 0, it.unit ?? "", lang, t)
              return "- " + ingName(it.name, lang) + (q ? " : " + q : "")
            })
            .join("\n"),
      )
      .join("\n\n")
    copyText(text)
  }

  return (
    <>
      <PageHeader
        title={<h1 className="text-4xl">{t.nav.shop}</h1>}
        sub={count ? t.shopLeft(left) : t.shopNone}
      >
        {groups.length > 0 && (
          <Button variant="outline" size="sm" onClick={copy}>
            {t.copy}
          </Button>
        )}
        <RefreshButton />
      </PageHeader>
      <ExtraForm />
      {!count && <Empty>{t.shopEmpty}</Empty>}
      {count > 0 && !groups.length && <Empty>{t.allBasket}</Empty>}
      {groups.map((g) => (
        <section key={g.rayon} className="mb-4.5">
          <h3 className="mb-1.5 text-lg">{rayonName(g.rayon, t)}</h3>
          <div className="overflow-hidden rounded-lg border bg-card">
            {g.items.map((it) => (
              <Item key={it.key} it={it} />
            ))}
          </div>
        </section>
      ))}
      {done.length > 0 && (
        <details className="mt-5">
          <summary className="mb-2 cursor-pointer text-muted-foreground">
            {t.basket(done.length)}
          </summary>
          <div className="overflow-hidden rounded-lg border bg-card">
            {done.map((it) => (
              <Item key={it.key} it={it} done />
            ))}
          </div>
        </details>
      )}
    </>
  )
}
