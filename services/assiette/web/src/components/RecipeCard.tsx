import { Check, CircleCheck, Plus } from "lucide-react"
import { CAT_BORDER } from "@/components/categories"
import { IconButton } from "@/components/IconButton"
import { Stepper } from "@/components/Stepper"
import { addMine, markDone, removeMine, setPersons } from "@/lib/actions"
import { offSeason, PROT, type Recipe } from "@/lib/catalog"
import { metaLine, recipeText } from "@/lib/format"
import { MONTH, openSheet, useApp, useT } from "@/lib/store"
import { cn } from "@/lib/utils"

function Tags({ r }: { r: Recipe }) {
  const t = useT()
  const main = [r.minutes <= 20 && t.tQuick, r.batch && t.tBatch].filter(
    (x) => typeof x === "string",
  )
  const other = [
    r.world && t.tWorld,
    offSeason(r, MONTH).length > 0 && t.tOff,
  ].filter((x) => typeof x === "string")
  if (!main.length && !other.length) return null
  return (
    <span className="mt-1.5 flex flex-wrap gap-1.5">
      {main.map((s) => (
        <span
          key={s}
          className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground"
        >
          {s}
        </span>
      ))}
      {other.map((s) => (
        <span
          key={s}
          className="rounded-full border px-2 py-0.5 text-xs text-muted-foreground"
        >
          {s}
        </span>
      ))}
    </span>
  )
}

export function RecipeCard({
  r,
  mode,
}: {
  r: Recipe
  mode: "discover" | "fav" | "mine"
}) {
  const { st, lang } = useApp()
  const t = useT()
  const text = recipeText(r, lang)
  const mine = st.mine[r.id]
  const cat = PROT[r.protein]
  return (
    <article
      className={cn(
        "mb-2 grid grid-cols-[1fr_auto] items-center gap-x-3 gap-y-2 rounded-xl border border-l-[5px] bg-card py-3 pr-3 pl-3.5",
        cat && CAT_BORDER[cat],
      )}
    >
      <button
        type="button"
        className="block cursor-pointer rounded text-left focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-ring"
        onClick={() => openSheet(r.id)}
      >
        <span className="block leading-snug font-semibold">
          {text.title}
          {st.favs[r.id] && (
            <span className="text-fav" aria-label={t.favorite}>
              {" "}
              ♥
            </span>
          )}
        </span>
        <span className="mt-0.5 block text-[0.85rem] text-muted-foreground">
          {metaLine(r, t)}
        </span>
        {mode === "discover" && text.highlights.length > 0 && (
          <span className="mt-1 block text-[0.82rem] text-muted-foreground">
            {text.highlights.join(", ")}
          </span>
        )}
        <Tags r={r} />
      </button>
      {mode !== "mine" &&
        (mine ? (
          <IconButton
            label={t.add}
            variant="outline"
            className="border-primary bg-muted text-foreground"
            aria-pressed
            onClick={() => removeMine(r.id)}
          >
            <Check />
          </IconButton>
        ) : (
          <IconButton
            label={t.add}
            aria-pressed={false}
            onClick={() => addMine(r.id)}
          >
            <Plus />
          </IconButton>
        ))}
      {mode === "mine" && mine && (
        <div className="col-span-full flex flex-wrap items-center gap-2">
          <Stepper
            persons={mine.persons}
            onChange={(d) => setPersons(r.id, d)}
          />
          <IconButton
            label={t.done}
            className="ml-auto"
            onClick={() => markDone(r.id)}
          >
            <CircleCheck />
          </IconButton>
        </div>
      )}
    </article>
  )
}
