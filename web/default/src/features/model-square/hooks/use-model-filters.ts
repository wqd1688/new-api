/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useCallback, useMemo, useState } from 'react'
import { useSearch } from '@tanstack/react-router'
import { FILTER_ALL, SORT_OPTIONS } from '@/features/pricing/constants'
import {
  filterBySearch,
  filterByVendor,
  sortModels,
} from '@/features/pricing/lib/filters'
import type { PricingModel } from '@/features/pricing/types'

type ModelFilterState = {
  search?: string
  sort?: string
  vendor?: string
}

export function useModelFilters(models: PricingModel[]) {
  const search = useSearch({ from: '/models/' })
  const [filterState, setFilterState] = useState<ModelFilterState>(() => ({
    search: search.search,
    sort: search.sort,
    vendor: search.vendor,
  }))

  const searchInput = filterState.search || ''
  const sortBy = filterState.sort || SORT_OPTIONS.NAME
  const vendorFilter = filterState.vendor || FILTER_ALL

  const updateFilters = useCallback((updates: Record<string, unknown>) => {
    setFilterState((prev) => {
      const next: Record<string, unknown> = { ...prev, ...updates }
      for (const key of Object.keys(next)) {
        if (next[key] === undefined || next[key] === null) {
          delete next[key]
        }
      }
      return next as ModelFilterState
    })
  }, [])

  const setSearchInput = useCallback(
    (value: string) => updateFilters({ search: value || undefined }),
    [updateFilters]
  )

  const setSortBy = useCallback(
    (value: string) =>
      updateFilters({ sort: value === SORT_OPTIONS.NAME ? undefined : value }),
    [updateFilters]
  )

  const setVendorFilter = useCallback(
    (value: string) =>
      updateFilters({ vendor: value === FILTER_ALL ? undefined : value }),
    [updateFilters]
  )

  const filteredModels = useMemo(() => {
    if (!models || models.length === 0) return []

    let result = filterBySearch(models, searchInput)
    result = filterByVendor(result, vendorFilter)
    return sortModels(result, sortBy)
  }, [models, searchInput, sortBy, vendorFilter])

  const hasActiveFilters = vendorFilter !== FILTER_ALL
  const activeFilterCount = hasActiveFilters ? 1 : 0

  const clearFilters = useCallback(() => {
    updateFilters({ vendor: undefined })
  }, [updateFilters])

  const clearSearch = useCallback(() => {
    updateFilters({ search: undefined })
  }, [updateFilters])

  return {
    searchInput,
    sortBy,
    vendorFilter,
    setSearchInput,
    setSortBy,
    setVendorFilter,
    filteredModels,
    hasActiveFilters,
    activeFilterCount,
    clearFilters,
    clearSearch,
  }
}
