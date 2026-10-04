import { useSyncExternalStore } from "react"
import { api, HttpError, setUnauthorizedHandler } from "@/lib/api"
import { CATALOG, registerCustom, type CustomRecipeJSON } from "@/lib/catalog"
import { deckKey, newDeck } from "@/lib/deck"
import { I18N, type Dict, type Lang } from "@/lib/i18n"
import { mergeDiff } from "@/lib/merge"
import {
  DEFAULT_STATE,
  normalizeState,
  type AppState,
  type StoredState,
} from "@/lib/state"

export type View = "discover" | "fav" | "mine" | "shop" | "plus"
export type AuthMode = "login" | "signup"

export interface Toast {
  id: number
  msg: string
  undo?: () => void
}

export interface Snapshot {
  status: "loading" | "auth" | "ready"
  authMode: AuthMode
  authErr: string
  /** Invite token read from `#invite=` and kept until it's used. */
  invite?: string
  st: AppState
  lang: Lang
  team: string[]
  view: View
  query: string
  /** Id of the recipe shown in the sheet. */
  sheet?: string
  refreshing: boolean
  toast?: Toast
}

/** Current month, 1-12. */
export const MONTH = new Date().getMonth() + 1
const SAVE_DELAY = 500
const AUTO_REFRESH_MS = 10_000

let snap: Snapshot = {
  status: "loading",
  authMode: "login",
  authErr: "",
  st: structuredClone(DEFAULT_STATE),
  lang: "fr",
  team: [],
  view: "discover",
  query: "",
  refreshing: false,
}
const listeners = new Set<() => void>()

let synced: unknown = {}
let saveTimer: ReturnType<typeof setTimeout> | undefined
let queue: Promise<unknown> = Promise.resolve()
let edits = 0
let lastLoad = 0
let reloading = false
let toastTimer: ReturnType<typeof setTimeout> | undefined

function set(patch: Partial<Snapshot>) {
  snap = { ...snap, ...patch }
  for (const l of listeners) l()
}

function subscribe(l: () => void) {
  listeners.add(l)
  return () => listeners.delete(l)
}

export function useApp(): Snapshot {
  return useSyncExternalStore(subscribe, () => snap)
}

export function getSnapshot(): Snapshot {
  return snap
}

export function useT(): Dict {
  return I18N[useApp().lang]
}

/** Shows a toast for 2.6 s, or 5 s when it offers an undo. */
export function toast(msg: string, undo?: () => void) {
  clearTimeout(toastTimer)
  const id = (snap.toast?.id ?? 0) + 1
  set({ toast: { id, msg, undo } })
  toastTimer = setTimeout(
    () => {
      if (snap.toast?.id === id) set({ toast: undefined })
    },
    undo ? 5000 : 2600,
  )
}

export function dismissToast() {
  clearTimeout(toastTimer)
  set({ toast: undefined })
}

const isAuth = () => snap.status === "auth"

function t(): Dict {
  return I18N[snap.lang]
}

/** Runs server writes one after the other; a failure shows a toast. */
function enqueue(fn: () => Promise<unknown>) {
  queue = queue.then(fn).catch((e) => {
    console.error(e)
    toast(t().saveFail)
  })
  return queue
}

function saveState(st: AppState) {
  const d = structuredClone(st)
  return enqueue(async () => {
    const patch = mergeDiff(synced, d)
    if (!patch) return
    await api("/api/state", { method: "PATCH", body: patch, keepalive: true })
    synced = d
  })
}

/** Saves the team state, debounced unless `now`. */
function save(now = false) {
  edits++
  clearTimeout(saveTimer)
  saveTimer = undefined
  if (now) return saveState(snap.st)
  saveTimer = setTimeout(() => {
    saveTimer = undefined
    saveState(snap.st).then(autoRefresh)
  }, SAVE_DELAY)
}

/** Draws a new deck for the current tab and category when it has no known recipe left. */
function withDeck(st: AppState, force = false): AppState {
  const key = deckKey(st.tab, st.cat)
  if (!force && (st.decks[key] || []).some((id) => CATALOG[id])) return st
  const ids = newDeck(Object.values(CATALOG), st, MONTH, Date.now())
  return { ...st, decks: { ...st.decks, [key]: ids } }
}

/** Applies a change to the team state, makes sure the current deck exists and saves. */
export function update(
  fn: (st: AppState) => AppState,
  opt?: { now?: boolean },
) {
  const before = snap.st
  const st = fn(before)
  const next =
    st.tab !== before.tab || st.cat !== before.cat ? withDeck(st) : st
  set({ st: next })
  save(opt?.now)
}

export function reshuffle() {
  set({ st: withDeck(snap.st, true) })
  save()
  window.scrollTo(0, 0)
}

export function setView(view: View) {
  set({ view, sheet: undefined })
  window.scrollTo(0, 0)
}

export function setQuery(query: string) {
  set({ query })
}

export function openSheet(id: string) {
  set({ sheet: id })
}

export function closeSheet() {
  set({ sheet: undefined })
}

export function setLang(lang: Lang) {
  set({ lang })
  document.documentElement.lang = lang
  enqueue(() => api("/api/me", { method: "PUT", body: { lang } }))
}

function showAuth(mode: AuthMode) {
  set({ status: "auth", authMode: mode })
}

export function setAuthMode(mode: AuthMode) {
  set({ authMode: mode, authErr: "" })
}

async function loadRemote() {
  lastLoad = Date.now()
  const [st, custom, team, me] = await Promise.all([
    api<StoredState>("/api/state"),
    api<CustomRecipeJSON[]>("/api/custom"),
    api<{ members: { email: string }[] }>("/api/team"),
    api<{ lang: string }>("/api/me"),
  ])
  for (const r of custom || []) registerCustom(r)
  return {
    st,
    team: (team?.members || []).map((m) => m.email),
    lang: me?.lang === "en" ? "en" : "fr",
  } satisfies { st?: StoredState; team: string[]; lang: Lang }
}

function applyRemote(r: Awaited<ReturnType<typeof loadRemote>>) {
  synced = structuredClone(r.st ?? {})
  const loaded = normalizeState(r.st)
  const st = withDeck(loaded)
  document.documentElement.lang = r.lang
  set({ st, team: r.team, lang: r.lang })
  if (st !== loaded) save()
}

/**
 * Reloads the team state from the server, unless a local edit is pending or happens meanwhile.
 * `manual` shows the spinner and a toast.
 */
export async function reload(manual: boolean) {
  if (reloading) return
  reloading = true
  if (manual) set({ refreshing: true })
  try {
    if (saveTimer) save(true)
    await queue
    const e0 = edits
    const r = await loadRemote()
    if (saveTimer || edits !== e0) return
    applyRemote(r)
    if (manual) toast(t().refreshed)
  } catch (e) {
    console.error(e)
    if (manual && !isAuth()) toast(t().loadFail)
  } finally {
    reloading = false
    if (snap.refreshing) set({ refreshing: false })
  }
}

function autoRefresh() {
  if (
    snap.status === "ready" &&
    !saveTimer &&
    Date.now() - lastLoad >= AUTO_REFRESH_MS
  ) {
    reload(false)
  }
}

let listening = false

function listen() {
  if (listening) return
  listening = true
  document.addEventListener("click", autoRefresh)
  document.addEventListener("keydown", autoRefresh)
  document.addEventListener("visibilitychange", () => {
    if (snap.status !== "ready") return
    if (document.hidden) {
      if (saveTimer) save(true)
      return
    }
    autoRefresh()
  })
}

async function joinInvite(token: string) {
  set({ invite: undefined })
  if (!confirm(t().confirmJoin)) return
  try {
    await api("/api/invites/" + encodeURIComponent(token) + "/join", {
      method: "POST",
    })
    toast(t().joined)
  } catch (e) {
    console.error(e)
    if (isAuth()) return
    toast(
      e instanceof HttpError && e.status === 410
        ? t().inviteGone
        : t().joinFail,
    )
  }
}

/** Starts the app: reads the invite from the URL, checks the session, joins the invite and loads the team. */
export async function boot() {
  setUnauthorizedHandler(() => showAuth(snap.invite ? "signup" : "login"))
  listen()
  const m = location.hash.match(/invite=([\w-]+)/)
  if (m) {
    set({ invite: m[1] })
    history.replaceState(null, "", location.pathname + location.search)
  }
  set({ status: "loading" })
  try {
    await api("/api/me")
  } catch (e) {
    if (isAuth()) return
    console.error(e)
  }
  if (snap.invite) await joinInvite(snap.invite)
  if (isAuth()) return
  try {
    applyRemote(await loadRemote())
  } catch (e) {
    console.error(e)
    if (isAuth()) return
    toast(t().loadFail)
  }
  set({ status: "ready" })
}

const AUTH_ERRORS: Record<
  number,
  "badInput" | "badCreds" | "emailTaken" | "inviteGone" | "tooMany"
> = {
  400: "badInput",
  401: "badCreds",
  409: "emailTaken",
  410: "inviteGone",
  429: "tooMany",
}

export async function submitAuth(email: string, password: string) {
  const signup = snap.authMode === "signup"
  set({ authErr: "" })
  try {
    await api(signup ? "/signup" : "/login", {
      method: "POST",
      body: signup
        ? { token: snap.invite, email, password }
        : { email, password },
    })
    if (signup) set({ invite: undefined })
  } catch (e) {
    console.error(e)
    const key = e instanceof HttpError ? AUTH_ERRORS[e.status] : undefined
    set({ authErr: t()[key ?? "authFail"] })
    return
  }
  await boot()
}

export async function logout() {
  await fetch("/logout", { method: "POST" }).catch(() => undefined)
  location.reload()
}

export async function createInvite() {
  try {
    const r = await api<{ token: string }>("/api/invites", { method: "POST" })
    const url = location.origin + "/#invite=" + r?.token
    try {
      await navigator.clipboard.writeText(url)
      toast(t().inviteCopied)
    } catch {
      prompt(t().inviteLink, url)
    }
  } catch (e) {
    console.error(e)
    toast(t().inviteFail)
  }
}
