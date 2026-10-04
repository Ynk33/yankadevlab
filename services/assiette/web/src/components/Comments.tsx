import { Pencil, Trash2 } from "lucide-react"
import { useEffect, useState, type FormEvent } from "react"
import { IconButton } from "@/components/IconButton"
import { Button } from "@/components/ui/button"
import { api } from "@/lib/api"
import { fmtDate } from "@/lib/format"
import { toast, useApp, useT } from "@/lib/store"

interface Comment {
  id: string
  author: string
  body: string
  createdAt: string
  updatedAt: string
  mine: boolean
}

function CommentForm({
  draft,
  setDraft,
  busy,
  editing,
  onSubmit,
  onCancel,
}: {
  draft: string
  setDraft: (s: string) => void
  busy: boolean
  editing: boolean
  onSubmit: () => void
  onCancel: () => void
}) {
  const t = useT()
  return (
    <form
      className="grid gap-2"
      onSubmit={(e: FormEvent) => {
        e.preventDefault()
        onSubmit()
      }}
    >
      <textarea
        name="body"
        required
        maxLength={2000}
        placeholder={t.commentPh}
        aria-label={t.comments}
        value={draft}
        autoFocus={editing}
        onChange={(e) => setDraft(e.target.value)}
        className="min-h-16 resize-y rounded-lg border bg-card px-3 py-2.5"
      />
      <div className="flex gap-2">
        <Button type="submit" size="sm" disabled={busy}>
          {editing ? t.save : t.commentPost}
        </Button>
        {editing && (
          <Button type="button" variant="outline" size="sm" onClick={onCancel}>
            {t.commentCancel}
          </Button>
        )}
      </div>
    </form>
  )
}

/** Team comments on a recipe: list, post, and edit or delete your own. */
export function Comments({ recipeId }: { recipeId: string }) {
  const { lang } = useApp()
  const t = useT()
  const [list, setList] = useState<Comment[] | undefined>(undefined)
  const [err, setErr] = useState(false)
  const [editing, setEditing] = useState<string | undefined>(undefined)
  const [draft, setDraft] = useState("")
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    let live = true
    api<Comment[]>("/api/recipes/" + encodeURIComponent(recipeId) + "/comments")
      .then((l) => live && setList(l ?? []))
      .catch((e) => {
        console.error(e)
        if (!live) return
        setList([])
        setErr(true)
      })
    return () => {
      live = false
    }
  }, [recipeId])

  async function submit() {
    const body = draft.trim()
    if (!body || busy) return
    setBusy(true)
    try {
      if (editing) {
        const u = await api<Comment>(
          "/api/comments/" + encodeURIComponent(editing),
          { method: "PUT", body: { body } },
        )
        setList((l) => l?.map((m) => (u && m.id === u.id ? u : m)))
        setEditing(undefined)
      } else {
        const c = await api<Comment>(
          "/api/recipes/" + encodeURIComponent(recipeId) + "/comments",
          { method: "POST", body: { body } },
        )
        if (c) setList((l) => [...(l ?? []), c])
      }
      setDraft("")
    } catch (e) {
      console.error(e)
      toast(t.commentFail)
    } finally {
      setBusy(false)
    }
  }

  async function remove(id: string) {
    if (!confirm(t.confirmDelComment)) return
    try {
      await api("/api/comments/" + encodeURIComponent(id), {
        method: "DELETE",
      })
      setList((l) => l?.filter((m) => m.id !== id))
    } catch (e) {
      console.error(e)
      toast(t.commentFail)
    }
  }

  const form = (
    <CommentForm
      draft={draft}
      setDraft={setDraft}
      busy={busy}
      editing={!!editing}
      onSubmit={submit}
      onCancel={() => {
        setEditing(undefined)
        setDraft("")
      }}
    />
  )

  return (
    <>
      <h3 className="mt-5 mb-2 text-lg">{t.comments}</h3>
      {!list ? (
        <p className="text-sm text-muted-foreground">{t.commentLoading}</p>
      ) : err ? (
        <p className="text-sm text-muted-foreground">{t.commentLoadFail}</p>
      ) : !list.length ? (
        <p className="text-sm text-muted-foreground">{t.commentNone}</p>
      ) : (
        <ul className="mb-3">
          {list.map((m) => (
            <li key={m.id} className="border-b py-2">
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm text-muted-foreground">
                  {m.author} · {fmtDate(m.createdAt, lang)}
                  {new Date(m.updatedAt) > new Date(m.createdAt) &&
                    t.commentEdited}
                </span>
                {m.mine && !editing && (
                  <div className="flex gap-1.5">
                    <IconButton
                      label={t.commentEdit}
                      variant="ghost"
                      onClick={() => {
                        setEditing(m.id)
                        setDraft(m.body)
                      }}
                    >
                      <Pencil />
                    </IconButton>
                    <IconButton
                      label={t.commentDel}
                      variant="ghost"
                      onClick={() => remove(m.id)}
                    >
                      <Trash2 />
                    </IconButton>
                  </div>
                )}
              </div>
              {editing === m.id ? (
                form
              ) : (
                <p className="mt-1 [overflow-wrap:anywhere] whitespace-pre-wrap">
                  {m.body}
                </p>
              )}
            </li>
          ))}
        </ul>
      )}
      {list && !err && !editing && form}
    </>
  )
}
