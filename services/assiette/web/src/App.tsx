import { Toast } from "@/components/Toast"
import { Auth } from "@/pages/Auth"
import { useApp } from "@/lib/store"

export default function App() {
  const { status } = useApp()
  return (
    <>
      <main className="mx-auto max-w-[680px] px-4 pt-5 pb-[calc(96px+env(safe-area-inset-bottom,0px))]">
        {status === "loading" && (
          <p className="text-muted-foreground">Chargement…</p>
        )}
        {status === "auth" && <Auth />}
        {status === "ready" && <h1 className="text-4xl">Assiette</h1>}
      </main>
      <Toast />
    </>
  )
}
