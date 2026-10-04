import { Empty } from "@/components/Empty"
import { Filters } from "@/components/Filters"
import { PageHeader } from "@/components/PageHeader"
import { RecipeCard } from "@/components/RecipeCard"
import { RefreshButton } from "@/components/RefreshButton"
import { CATALOG, matches } from "@/lib/catalog"
import { visible } from "@/lib/deck"
import { useApp, useT } from "@/lib/store"
import { slug } from "@/lib/utils"

export function Favorites() {
  const { st, query } = useApp()
  const t = useT()
  const favs = Object.keys(st.favs)
    .map((id) => CATALOG[id])
    .filter(Boolean)
    .sort((a, b) => st.favs[b.id] - st.favs[a.id])
  const q = slug(query)
  const res = favs.filter((r) => visible(r, st) && matches(r, q))
  return (
    <>
      <PageHeader
        title={<h1 className="text-4xl">{t.nav.fav}</h1>}
        sub={t.favsSub}
      >
        <RefreshButton />
      </PageHeader>
      {favs.length ? (
        <>
          <Filters />
          {res.length ? (
            res.map((r) => <RecipeCard key={r.id} r={r} mode="fav" />)
          ) : (
            <p className="text-muted-foreground">{t.noMatch}</p>
          )}
        </>
      ) : (
        <Empty discover>{t.favsEmpty}</Empty>
      )}
    </>
  )
}
