import { Empty } from "@/components/Empty"
import { PageHeader } from "@/components/PageHeader"
import { RecipeCard } from "@/components/RecipeCard"
import { RefreshButton } from "@/components/RefreshButton"
import { CATALOG } from "@/lib/catalog"
import { useApp, useT } from "@/lib/store"

export function Mine() {
  const { st } = useApp()
  const t = useT()
  const items = Object.entries(st.mine)
    .filter(([id]) => CATALOG[id])
    .sort((a, b) => a[1].addedAt - b[1].addedAt)
    .map(([id]) => CATALOG[id])
  return (
    <>
      <PageHeader
        title={<h1 className="text-4xl">{t.nav.mine}</h1>}
        sub={t.mineSub(items.length)}
      >
        <RefreshButton />
      </PageHeader>
      {items.length ? (
        items.map((r) => <RecipeCard key={r.id} r={r} mode="mine" />)
      ) : (
        <Empty discover>{t.mineEmpty}</Empty>
      )}
    </>
  )
}
