import { BottomNav } from "@/components/BottomNav"
import { RecipeSheet } from "@/components/RecipeSheet"
import { Toast } from "@/components/Toast"
import { useApp, type View } from "@/lib/store"
import { Auth } from "@/pages/Auth"
import { Discover } from "@/pages/Discover"
import { Favorites } from "@/pages/Favorites"
import { Mine } from "@/pages/Mine"
import { More } from "@/pages/More"
import { Shopping } from "@/pages/Shopping"

const PAGES: Record<View, () => React.JSX.Element> = {
  discover: Discover,
  fav: Favorites,
  mine: Mine,
  shop: Shopping,
  plus: More,
}

export default function App() {
  const { status, view } = useApp()
  const Page = PAGES[view]
  return (
    <>
      <main className="mx-auto max-w-[680px] px-4 pt-5 pb-[calc(96px+env(safe-area-inset-bottom,0px))]">
        {status === "loading" && (
          <p className="text-muted-foreground">Chargement…</p>
        )}
        {status === "auth" && <Auth />}
        {status === "ready" && <Page />}
      </main>
      {status === "ready" && (
        <>
          <BottomNav />
          <RecipeSheet />
        </>
      )}
      <Toast />
    </>
  )
}
