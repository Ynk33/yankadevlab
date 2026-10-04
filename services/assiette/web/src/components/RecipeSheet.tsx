import { Dialog } from "@base-ui/react/dialog"
import { Ban, CircleCheck, Heart, Plus, X } from "lucide-react"
import { useState } from "react"
import { Comments } from "@/components/Comments"
import { IconButton } from "@/components/IconButton"
import { Stepper } from "@/components/Stepper"
import {
  addMine,
  dislike,
  markDone,
  removeMine,
  setPersons,
  toggleFav,
} from "@/lib/actions"
import { CATALOG, offSeason, type Recipe } from "@/lib/catalog"
import { ingName, metaLine, recipeText } from "@/lib/format"
import { fmtQty, scale } from "@/lib/shopping"
import { closeSheet, MONTH, useApp, useT } from "@/lib/store"
import { cn } from "@/lib/utils"

function SeasonLine({ r }: { r: Recipe }) {
  const { lang } = useApp()
  const t = useT()
  const off = offSeason(r, MONTH)
  const month = t.months[MONTH - 1]
  return (
    <p className="mt-1.5 text-sm">
      {off.length
        ? t.offSeason(month, off.map((n) => ingName(n, lang)).join(", "))
        : t.inSeason(month)}
      {r.world && t.worldS}
      {r.batch && t.keeps}
    </p>
  )
}

function Nutri({ value, label }: { value: string; label: string }) {
  return (
    <div className="rounded-lg bg-muted p-2.5 text-center">
      <b className="block font-heading text-xl font-normal">{value}</b>
      <span className="text-[0.8rem] text-muted-foreground">{label}</span>
    </div>
  )
}

function Content({ r }: { r: Recipe }) {
  const { st, lang } = useApp()
  const t = useT()
  const [draftPersons, setDraftPersons] = useState<number | undefined>(
    undefined,
  )
  const text = recipeText(r, lang)
  const mine = st.mine[r.id]
  const persons = mine?.persons ?? draftPersons ?? (st.prefs.persons || 2)
  const k = scale(r, persons)
  const fav = !!st.favs[r.id]

  function changePersons(d: number) {
    if (mine) setPersons(r.id, d)
    else setDraftPersons(Math.min(12, Math.max(1, persons + d)))
  }

  return (
    <>
      <div className="flex justify-end">
        <Dialog.Close className="p-1.5 text-muted-foreground">
          {t.close}
        </Dialog.Close>
      </div>
      <Dialog.Title className="mt-0.5 mb-1.5 text-2xl">
        {text.title}
      </Dialog.Title>
      <p className="text-muted-foreground">{metaLine(r, t)}</p>
      <SeasonLine r={r} />
      <div className="mt-3 mb-1 grid grid-cols-3 gap-2">
        <Nutri value={String(r.kcal ?? "–")} label={t.kcal} />
        <Nutri value={(r.proteines ?? "–") + " g"} label={t.prot_} />
        <Nutri value={(r.fibres ?? "–") + " g"} label={t.fibre} />
      </div>
      <p className="text-sm text-muted-foreground">{t.estim}</p>
      {text.highlights.length > 0 && (
        <ul className="mt-3 mb-1 flex flex-wrap gap-1.5">
          {text.highlights.map((x) => (
            <li key={x} className="rounded-full bg-muted px-3 py-1 text-sm">
              {x}
            </li>
          ))}
        </ul>
      )}
      <h3 className="mt-5 mb-2 flex items-center justify-between gap-2.5 text-lg">
        {t.ings}
        <Stepper persons={persons} onChange={changePersons} />
      </h3>
      <ul>
        {r.ingredients.map((i) => (
          <li
            key={i.name + i.unit}
            className="flex justify-between gap-3 border-b py-1.5"
          >
            <span>{ingName(i.name, lang)}</span>
            <span className="whitespace-nowrap text-muted-foreground tabular-nums">
              {fmtQty(i.qty * k, i.unit, lang, t)}
            </span>
          </li>
        ))}
      </ul>
      <h3 className="mt-5 mb-2 text-lg">{t.prep}</h3>
      <ol className="list-decimal pl-5.5">
        {(text.steps.length ? text.steps : r.steps).map((s, n) => (
          <li key={n} className="mb-2">
            {s}
          </li>
        ))}
      </ol>
      <Comments recipeId={r.id} />
      <div className="mt-5.5 flex flex-wrap gap-2 border-t pt-4">
        {mine ? (
          <>
            <IconButton label={t.done} onClick={() => markDone(r.id)}>
              <CircleCheck />
            </IconButton>
            <IconButton
              label={t.remove}
              variant="ghost"
              onClick={() => removeMine(r.id)}
            >
              <X />
            </IconButton>
          </>
        ) : (
          <IconButton label={t.addMine} onClick={() => addMine(r.id, persons)}>
            <Plus />
          </IconButton>
        )}
        <IconButton
          label={t.fav}
          variant="ghost"
          aria-pressed={fav}
          className={cn(fav && "border-fav text-fav")}
          onClick={() => toggleFav(r.id)}
        >
          <Heart fill={fav ? "currentColor" : "none"} />
        </IconButton>
        <IconButton
          label={t.never}
          variant="outline"
          className="border-destructive text-destructive"
          onClick={() => dislike(r.id)}
        >
          <Ban />
        </IconButton>
      </div>
    </>
  )
}

/** Recipe details: a bottom sheet on mobile, a centered dialog from 720px. */
export function RecipeSheet() {
  const { sheet } = useApp()
  const r = sheet ? CATALOG[sheet] : undefined
  return (
    <Dialog.Root open={!!r} onOpenChange={(open) => !open && closeSheet()}>
      <Dialog.Portal>
        <Dialog.Backdrop className="fixed inset-0 z-20 bg-[rgba(10,18,13,.45)]" />
        <Dialog.Popup className="fixed inset-x-0 bottom-0 z-21 mx-auto max-h-[92%] max-w-[680px] overflow-auto rounded-t-[18px] bg-card px-4.5 pt-3.5 pb-[calc(24px+env(safe-area-inset-bottom,0px))] animate-in fade-in slide-in-from-bottom-6 duration-200 motion-reduce:animate-none min-[720px]:top-0 min-[720px]:m-auto min-[720px]:h-fit min-[720px]:max-h-[calc(100%-64px)] min-[720px]:rounded-[18px] min-[720px]:pb-6">
          {r && <Content key={r.id} r={r} />}
        </Dialog.Popup>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
