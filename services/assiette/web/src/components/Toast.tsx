import { dismissToast, useApp, useT } from "@/lib/store"

export function Toast() {
  const { toast } = useApp()
  const t = useT()
  return (
    <div
      role="status"
      aria-live="polite"
      className="fixed bottom-[calc(76px+env(safe-area-inset-bottom,0px))] left-1/2 z-30 flex max-w-[90vw] -translate-x-1/2 items-center gap-3 rounded-full bg-foreground px-4 py-2 text-sm text-background empty:hidden"
    >
      {toast && (
        <>
          {toast.msg}
          {toast.undo && (
            <button
              type="button"
              className="font-semibold underline"
              onClick={() => {
                toast.undo?.()
                dismissToast()
              }}
            >
              {t.undo}
            </button>
          )}
        </>
      )}
    </div>
  )
}
