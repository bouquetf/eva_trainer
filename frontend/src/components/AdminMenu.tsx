export type AdminTab = 'users' | 'permissions' | 'maps' | 'splits' | 'catalog' | 'import-export' | 'ligues' | 'archives'

type AdminMenuProps = {
  activeTab: AdminTab
  onTabChange: (tab: AdminTab) => void
  usersCount?: number
  mapsCount?: number
  splitsCount?: number
  catalogCount?: number
}

const TABS: { key: AdminTab; label: string; countKey?: 'usersCount' | 'mapsCount' | 'splitsCount' | 'catalogCount' }[] = [
  { key: 'users', label: 'Utilisateurs', countKey: 'usersCount' },
  { key: 'permissions', label: 'Permissions' },
  { key: 'maps', label: 'Cartes', countKey: 'mapsCount' },
  { key: 'splits', label: 'Splits', countKey: 'splitsCount' },
  { key: 'catalog', label: 'Catalogue', countKey: 'catalogCount' },
  { key: 'import-export', label: 'Import/Export' },
  { key: 'ligues', label: 'Ligues' },
  { key: 'archives', label: 'Archives' },
]

function AdminMenu({ activeTab, onTabChange, usersCount, mapsCount, splitsCount, catalogCount }: AdminMenuProps) {
  const countMap = { usersCount, mapsCount, splitsCount, catalogCount }

  return (
    <div className="admin-tabs">
      {TABS.map(({ key, label, countKey }) => (
        <button
          key={key}
          className={activeTab === key ? 'active' : ''}
          onClick={() => onTabChange(key)}
        >
          {label}{countKey !== undefined && countMap[countKey] !== undefined ? ` (${countMap[countKey]})` : ''}
        </button>
      ))}
    </div>
  )
}

export default AdminMenu
