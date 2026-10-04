import { RefreshCw } from "lucide-react"
import { IconButton } from "@/components/IconButton"
import { reload, useApp, useT } from "@/lib/store"
import { cn } from "@/lib/utils"

export function RefreshButton() {
  const { refreshing } = useApp()
  const t = useT()
  return (
    <IconButton
      label={t.refresh}
      variant="outline"
      size="icon"
      disabled={refreshing}
      onClick={() => reload(true)}
    >
      <RefreshCw
        className={cn(refreshing && "animate-spin motion-reduce:animate-none")}
      />
    </IconButton>
  )
}
