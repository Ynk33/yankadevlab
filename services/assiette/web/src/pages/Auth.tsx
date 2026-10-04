import { useState, type FormEvent } from "react"
import { PageHeader } from "@/components/PageHeader"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { setAuthMode, submitAuth, useApp, useT } from "@/lib/store"

export function Auth() {
  const { authMode, authErr, invite } = useApp()
  const t = useT()
  const [busy, setBusy] = useState(false)
  const signup = authMode === "signup"

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const f = new FormData(e.currentTarget)
    setBusy(true)
    await submitAuth(
      String(f.get("email") || "").trim(),
      String(f.get("password") || ""),
    )
    setBusy(false)
  }

  return (
    <>
      <PageHeader
        title={
          <>
            <img
              className="mb-3 block rounded-2xl"
              src="/logo.svg"
              alt=""
              width={64}
              height={64}
            />
            <h1 className="text-4xl">Assiette</h1>
          </>
        }
        sub={signup ? t.signupSub : t.loginSub}
      />
      <form className="grid gap-3.5" onSubmit={onSubmit}>
        <Label className="grid gap-1.5 font-semibold">
          {t.email}
          <Input type="email" name="email" required autoComplete="email" />
        </Label>
        <Label className="grid gap-1.5 font-semibold">
          {t.password}
          {signup && (
            <span className="font-normal text-muted-foreground">
              {t.passwordSub}
            </span>
          )}
          <Input
            type="password"
            name="password"
            required
            minLength={signup ? 8 : undefined}
            autoComplete={signup ? "new-password" : "current-password"}
          />
        </Label>
        {authErr && (
          <div
            role="alert"
            className="rounded-xl border border-destructive bg-card px-3.5 py-2.5 text-sm text-destructive"
          >
            {authErr}
          </div>
        )}
        <div>
          <Button type="submit" disabled={busy}>
            {signup ? t.signup : t.login}
          </Button>
        </div>
        {invite && (
          <div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setAuthMode(signup ? "login" : "signup")}
            >
              {signup ? t.haveAccount : t.noAccount}
            </Button>
          </div>
        )}
      </form>
    </>
  )
}
