import type { ComponentProps } from "react"
import { Button } from "@/components/ui/button"

/** Small round button showing only an icon; `label` is used for both aria-label and the tooltip. */
export function IconButton({
  label,
  variant = "default",
  ...props
}: { label: string } & ComponentProps<typeof Button>) {
  return (
    <Button
      size="icon-sm"
      variant={variant}
      aria-label={label}
      title={label}
      {...props}
    />
  )
}
