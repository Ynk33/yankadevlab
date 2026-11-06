import { useState } from "react"
import { RecipeCard } from "@/components/RecipeCard"
import { Seg } from "@/components/Seg"
import { Button } from "@/components/ui/button"
import { CATALOG } from "@/lib/catalog"
import { clearGeneration, generate, MAX_RECIPES } from "@/lib/generate"
import { useApp, useT } from "@/lib/store"

const COUNTS = Array.from({ length: MAX_RECIPES }, (_, i) => {
  const key = String(i + 1)
  return { key, label: key }
})

function Group({
  title,
  sub,
  ids,
}: {
  title: string
  sub?: string
  ids: string[]
}) {
  if (!ids.length) return null
  return (
    <section className="mt-4">
      <h2 className="text-lg">{title}</h2>
      {sub && <p className="mb-2 text-sm text-muted-foreground">{sub}</p>}
      {ids.map((id) => (
        <RecipeCard key={id} r={CATALOG[id]} mode="discover" />
      ))}
    </section>
  )
}

/** Free-text recipe request: catalog recipes that fit, completed with generated ones. */
export function GeneratePanel() {
  const { gen } = useApp()
  const t = useT()
  const [prompt, setPrompt] = useState("")
  const [count, setCount] = useState("3")
  const hasResults = gen.matches.length > 0 || gen.generated.length > 0
  return (
    <section className="mb-5 rounded-xl border bg-card p-3.5">
      <h2 className="mb-2 text-lg">{t.genTitle}</h2>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          if (prompt.trim() && !gen.busy)
            void generate(prompt.trim(), Number(count))
        }}
      >
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder={t.genPh}
          maxLength={1000}
          rows={2}
          className="w-full resize-y rounded-lg border border-input bg-card px-3 py-2 text-base outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 md:text-sm"
        />
        <div className="mt-2 flex flex-wrap items-center gap-2">
          <Seg
            options={COUNTS}
            value={count}
            onChange={setCount}
            label={t.genCount}
            itemClassName="px-3 py-1"
          />
          <Button
            type="submit"
            className="ml-auto"
            disabled={gen.busy || !prompt.trim()}
          >
            {t.genGo}
          </Button>
        </div>
      </form>
      {gen.busy && (
        <p className="mt-3 text-sm text-muted-foreground">{t.genBusy}</p>
      )}
      {gen.error && (
        <p className="mt-3 text-sm text-destructive">{gen.error}</p>
      )}
      <Group title={t.genFromCatalog} ids={gen.matches} />
      <Group title={t.genNew} sub={t.genNewSub} ids={gen.generated} />
      {hasResults && (
        <div className="mt-2 flex justify-end">
          <Button variant="ghost" size="sm" onClick={clearGeneration}>
            {t.genClear}
          </Button>
        </div>
      )}
    </section>
  )
}
