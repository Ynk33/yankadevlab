import { Empty } from "@/components/Empty"
import { Filters } from "@/components/Filters"
import { PageHeader } from "@/components/PageHeader"
import { RecipeCard } from "@/components/RecipeCard"
import { RefreshButton } from "@/components/RefreshButton"
import { Button } from "@/components/ui/button"
import { CATALOG, matches } from "@/lib/catalog"
import { deckKey, excluded, visible } from "@/lib/deck"
import { reshuffle, useApp, useT } from "@/lib/store"
import { slug } from "@/lib/utils"

const MAX_RESULTS = 40

function Results() {
  const { st, query } = useApp()
  const t = useT()
  const q = slug(query)
  if (q) {
    const res = Object.values(CATALOG).filter(
      (r) => visible(r, st) && matches(r, q),
    )
    if (!res.length) return <p className="text-muted-foreground">{t.noMatch}</p>
    return (
      <>
        <p className="mb-2 text-sm text-muted-foreground">
          {t.results(res.length)}
        </p>
        {res.slice(0, MAX_RESULTS).map((r) => (
          <RecipeCard key={r.id} r={r} mode="discover" />
        ))}
      </>
    )
  }
  const deck = (st.decks[deckKey(st.tab, st.cat)] || [])
    .map((id) => CATALOG[id])
    .filter((r) => r && !st.dislikes[r.id] && !excluded(r, st.prefs.exclusions))
  return (
    <>
      {deck.length ? (
        deck.map((r) => <RecipeCard key={r.id} r={r} mode="discover" />)
      ) : (
        <Empty>{t.emptyDeck}</Empty>
      )}
      <div className="mt-4 mb-2 flex justify-center">
        <Button variant="outline" onClick={reshuffle}>
          {t.other}
        </Button>
      </div>
    </>
  )
}

export function Discover() {
  const t = useT()
  return (
    <>
      <PageHeader
        title={<h1 className="text-4xl">{t.nav.discover}</h1>}
        sub={t.catalog(Object.keys(CATALOG).length)}
      >
        <RefreshButton />
      </PageHeader>
      <Filters />
      <Results />
    </>
  )
}
