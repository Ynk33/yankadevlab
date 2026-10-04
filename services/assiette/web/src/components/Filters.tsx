import { CAT_DOT } from "@/components/categories"
import { Seg } from "@/components/Seg"
import { Input } from "@/components/ui/input"
import { setCat, setTab } from "@/lib/actions"
import { CAT_KEYS, TAB_KEYS } from "@/lib/catalog"
import { setQuery, useApp, useT } from "@/lib/store"
import { cn } from "@/lib/utils"

export function Filters() {
  const { st, query } = useApp()
  const t = useT()
  return (
    <>
      <Input
        type="search"
        className="mb-3.5 h-11 px-3.5"
        placeholder={t.search}
        aria-label={t.search}
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />
      <div className="mb-3 flex flex-wrap items-center gap-2.5">
        <Seg
          options={TAB_KEYS.map((k) => ({ key: k, label: t.tabs[k][0] }))}
          value={st.tab}
          onChange={setTab}
        />
        <Seg
          className="max-w-full flex-nowrap overflow-x-auto [scrollbar-width:none]"
          itemClassName="flex-none px-2 text-[0.85rem]"
          options={CAT_KEYS.map((k) => ({
            key: k,
            label: (
              <span
                className={cn(
                  k !== "all" &&
                    "before:mr-1.5 before:inline-block before:size-2 before:rounded-full before:align-[1px] before:content-['']",
                  k !== "all" && CAT_DOT[k],
                )}
              >
                {t.cats[k]}
              </span>
            ),
          }))}
          value={st.cat}
          onChange={setCat}
        />
      </div>
      <p className="-mt-1 mb-3 text-sm text-muted-foreground">
        {t.tabs[st.tab][1]}
      </p>
    </>
  )
}
