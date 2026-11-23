import { ref } from 'vue'

export function useGlobalPage() {
  const page = ref<any>()
  const refreshing = ref(false)
  const loadingMore = ref(false)
  const hasMore = ref(true)

  return {
    page,
    refreshing,
    loadingMore,
    hasMore
  }
}
