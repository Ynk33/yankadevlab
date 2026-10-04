import type { FormEvent, ReactNode } from "react"
import { PageHeader } from "@/components/PageHeader"
import { RefreshButton } from "@/components/RefreshButton"
import { Seg } from "@/components/Seg"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { savePrefs, undislike } from "@/lib/actions"
import { CATALOG } from "@/lib/catalog"
import { recipeText } from "@/lib/format"
import { createInvite, logout, setLang, useApp, useT } from "@/lib/store"

const LANGS = [
  { key: "fr", label: "Français" },
  { key: "en", label: "English" },
] as const

function Section({
  title,
  sub,
  children,
}: {
  title: string
  sub?: string
  children: ReactNode
}) {
  return (
    <section className="mt-7">
      <h2 className="mb-1 text-2xl">{title}</h2>
      {sub && <p className="mb-3 text-sm text-muted-foreground">{sub}</p>}
      {children}
    </section>
  )
}

function Row({ children }: { children: ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-2.5 border-b py-2.5">
      {children}
    </div>
  )
}

function Prefs() {
  const { st } = useApp()
  const t = useT()
  const p = st.prefs

  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const f = new FormData(e.currentTarget)
    savePrefs({
      ...p,
      persons: Math.min(
        12,
        Math.max(1, parseInt(String(f.get("persons"))) || 2),
      ),
      exclusions: String(f.get("exclusions") || ""),
    })
  }

  return (
    <form className="grid gap-3.5" onSubmit={onSubmit}>
      <Label className="grid gap-1.5 text-[0.95rem] leading-normal font-semibold">
        {t.defP}
        <span className="text-[0.85rem] font-normal text-muted-foreground">
          {t.defPSub}
        </span>
        <Input
          type="number"
          name="persons"
          min={1}
          max={12}
          defaultValue={p.persons}
        />
      </Label>
      <Label className="grid gap-1.5 text-[0.95rem] leading-normal font-semibold">
        {t.excl}
        <span className="text-[0.85rem] font-normal text-muted-foreground">
          {t.exclSub}
        </span>
        <textarea
          name="exclusions"
          placeholder={t.exclPh}
          defaultValue={p.exclusions}
          className="min-h-16 resize-y rounded-lg border bg-card px-3 py-2.5 font-normal"
        />
      </Label>
      <div>
        <Button type="submit">{t.save}</Button>
      </div>
      <p className="text-sm text-muted-foreground">{t.savedDb}</p>
    </form>
  )
}

export function More() {
  const { st, lang, team } = useApp()
  const t = useT()
  const disliked = Object.keys(st.dislikes)
    .map((id) => CATALOG[id])
    .filter(Boolean)
  return (
    <>
      <PageHeader title={<h1 className="text-4xl">{t.nav.plus}</h1>}>
        <RefreshButton />
      </PageHeader>
      <Section title={t.neverT} sub={t.neverSub}>
        {disliked.length ? (
          disliked.map((r) => (
            <Row key={r.id}>
              <span>{recipeText(r, lang).title}</span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => undislike(r.id)}
              >
                {t.reallow}
              </Button>
            </Row>
          ))
        ) : (
          <p className="text-muted-foreground">{t.neverEmpty}</p>
        )}
      </Section>
      <Section title={t.settings}>
        <div className="mb-4">
          <p className="mb-1.5 font-semibold">{t.lang}</p>
          <Seg label={t.lang} options={LANGS} value={lang} onChange={setLang} />
        </div>
        <Prefs key={JSON.stringify(st.prefs)} />
      </Section>
      <Section title={t.team} sub={t.teamSub}>
        {team.map((email) => (
          <Row key={email}>
            <span>{email}</span>
          </Row>
        ))}
        <div className="mt-4 mb-2 flex gap-2.5">
          <Button size="sm" onClick={createInvite}>
            {t.invite}
          </Button>
          <Button variant="outline" size="sm" onClick={logout}>
            {t.logout}
          </Button>
        </div>
      </Section>
    </>
  )
}
