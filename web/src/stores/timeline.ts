import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { TimelineEvent, TimelinePerson, PersonLocation } from '@/types'
import { useAuthStore } from '@/stores/auth'

const API_URL = import.meta.env.VITE_API_URL || '/api'

export const useTimelineStore = defineStore('timeline', () => {
  const events = ref<TimelineEvent[]>([])
  const eventsTotal = ref(0)
  const currentEvent = ref<TimelineEvent | null>(null)

  const persons = ref<TimelinePerson[]>([])
  const personsTotal = ref(0)
  const currentPerson = ref<TimelinePerson | null>(null)

  const loading = ref(false)
  const error = ref<string | null>(null)

  const authHeaders = () => {
    const authStore = useAuthStore()
    return { Authorization: `Bearer ${authStore.token}` }
  }

  // ====== Events ======

  const fetchEvents = async (params: {
    category?: string
    yearStart?: number
    yearEnd?: number
    page?: number
    limit?: number
  } = {}) => {
    loading.value = true
    error.value = null
    try {
      const query = new URLSearchParams()
      if (params.category) query.set('category', params.category)
      if (params.yearStart) query.set('year_start', String(params.yearStart))
      if (params.yearEnd) query.set('year_end', String(params.yearEnd))
      if (params.page) query.set('page', String(params.page))
      if (params.limit) query.set('limit', String(params.limit))

      const res = await fetch(`${API_URL}/timeline/events?${query}`)
      if (res.ok) {
        const data = await res.json()
        events.value = data.data || []
        eventsTotal.value = data.total || 0
      }
    } catch (e) {
      error.value = 'Failed to fetch events'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const fetchEvent = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      const res = await fetch(`${API_URL}/timeline/events/${id}`)
      if (res.ok) {
        const data = await res.json()
        currentEvent.value = data.data
      }
    } catch (e) {
      error.value = 'Failed to fetch event'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const createEvent = async (input: {
    title: string
    description?: string
    content?: string
    event_date: string
    end_date?: string
    location: string
    latitude?: number | null
    longitude?: number | null
    source: string
    category?: string
    tags?: string[]
    is_public?: boolean
  }) => {
    try {
      const res = await fetch(`${API_URL}/timeline/events`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify(input),
      })
      if (res.ok) {
        const data = await res.json()
        events.value.unshift(data.data)
        return data.data as TimelineEvent
      }
      const err = await res.json()
      throw new Error(err.error || 'Failed to create event')
    } catch (e) {
      throw e
    }
  }

  const updateEvent = async (id: string, input: Partial<{
    title: string
    description: string
    content: string
    event_date: string
    end_date: string
    location: string
    latitude: number | null
    longitude: number | null
    source: string
    category: string
    tags: string[]
    is_public: boolean
  }>) => {
    try {
      const res = await fetch(`${API_URL}/timeline/events/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify(input),
      })
      if (res.ok) {
        const data = await res.json()
        const idx = events.value.findIndex((e) => e.id === id)
        if (idx !== -1) events.value[idx] = data.data
        if (currentEvent.value?.id === id) currentEvent.value = data.data
        return data.data as TimelineEvent
      }
      const err = await res.json()
      throw new Error(err.error || 'Failed to update event')
    } catch (e) {
      throw e
    }
  }

  const deleteEvent = async (id: string) => {
    try {
      const res = await fetch(`${API_URL}/timeline/events/${id}`, {
        method: 'DELETE',
        headers: authHeaders(),
      })
      if (res.ok) {
        events.value = events.value.filter((e) => e.id !== id)
      } else {
        const err = await res.json()
        throw new Error(err.error || 'Failed to delete event')
      }
    } catch (e) {
      throw e
    }
  }

  // ====== Persons ======

  const fetchPersons = async (params: {
    search?: string
    page?: number
    limit?: number
  } = {}) => {
    loading.value = true
    error.value = null
    try {
      const query = new URLSearchParams()
      if (params.search) query.set('search', params.search)
      if (params.page) query.set('page', String(params.page))
      if (params.limit) query.set('limit', String(params.limit))

      const res = await fetch(`${API_URL}/timeline/persons?${query}`)
      if (res.ok) {
        const data = await res.json()
        persons.value = data.data || []
        personsTotal.value = data.total || 0
      }
    } catch (e) {
      error.value = 'Failed to fetch persons'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const fetchPerson = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      const res = await fetch(`${API_URL}/timeline/persons/${id}`)
      if (res.ok) {
        const data = await res.json()
        currentPerson.value = data.data
      }
    } catch (e) {
      error.value = 'Failed to fetch person'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  const createPerson = async (input: {
    name: string
    bio?: string
    birth_date?: string
    death_date?: string
    tags?: string[]
    is_public?: boolean
  }) => {
    try {
      const res = await fetch(`${API_URL}/timeline/persons`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify(input),
      })
      if (res.ok) {
        const data = await res.json()
        persons.value.unshift(data.data)
        return data.data as TimelinePerson
      }
      const err = await res.json()
      throw new Error(err.error || 'Failed to create person')
    } catch (e) {
      throw e
    }
  }

  const updatePerson = async (id: string, input: Partial<{
    name: string
    bio: string
    birth_date: string
    death_date: string
    tags: string[]
    is_public: boolean
  }>) => {
    try {
      const res = await fetch(`${API_URL}/timeline/persons/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify(input),
      })
      if (res.ok) {
        const data = await res.json()
        const idx = persons.value.findIndex((p) => p.id === id)
        if (idx !== -1) persons.value[idx] = data.data
        if (currentPerson.value?.id === id) currentPerson.value = data.data
        return data.data as TimelinePerson
      }
      const err = await res.json()
      throw new Error(err.error || 'Failed to update person')
    } catch (e) {
      throw e
    }
  }

  const deletePerson = async (id: string) => {
    try {
      const res = await fetch(`${API_URL}/timeline/persons/${id}`, {
        method: 'DELETE',
        headers: authHeaders(),
      })
      if (res.ok) {
        persons.value = persons.value.filter((p) => p.id !== id)
      } else {
        const err = await res.json()
        throw new Error(err.error || 'Failed to delete person')
      }
    } catch (e) {
      throw e
    }
  }

  // ====== Locations ======

  const fetchPersonLocations = async (personId: string) => {
    try {
      const res = await fetch(`${API_URL}/timeline/persons/${personId}/locations`)
      if (res.ok) {
        const data = await res.json()
        return data.data as PersonLocation[]
      }
      return []
    } catch (e) {
      console.error(e)
      return []
    }
  }

  const addLocation = async (personId: string, input: {
    date: string
    end_date?: string
    place_name: string
    latitude: number
    longitude: number
    note?: string
  }) => {
    try {
      const res = await fetch(`${API_URL}/timeline/persons/${personId}/locations`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify(input),
      })
      if (res.ok) {
        const data = await res.json()
        if (currentPerson.value?.id === personId) {
          if (!currentPerson.value.locations) currentPerson.value.locations = []
          currentPerson.value.locations.push(data.data)
          currentPerson.value.locations.sort(
            (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
          )
        }
        return data.data as PersonLocation
      }
      const err = await res.json()
      throw new Error(err.error || 'Failed to add location')
    } catch (e) {
      throw e
    }
  }

  const updateLocation = async (locationId: string, input: {
    date: string
    end_date?: string
    place_name: string
    latitude: number
    longitude: number
    note?: string
  }) => {
    try {
      const res = await fetch(`${API_URL}/timeline/locations/${locationId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify(input),
      })
      if (res.ok) {
        const data = await res.json()
        if (currentPerson.value?.locations) {
          const idx = currentPerson.value.locations.findIndex((l) => l.id === locationId)
          if (idx !== -1) {
            currentPerson.value.locations[idx] = data.data
            currentPerson.value.locations.sort(
              (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
            )
          }
        }
        return data.data as PersonLocation
      }
      const err = await res.json()
      throw new Error(err.error || 'Failed to update location')
    } catch (e) {
      throw e
    }
  }

  const deleteLocation = async (locationId: string) => {
    try {
      const res = await fetch(`${API_URL}/timeline/locations/${locationId}`, {
        method: 'DELETE',
        headers: authHeaders(),
      })
      if (res.ok) {
        if (currentPerson.value?.locations) {
          currentPerson.value.locations = currentPerson.value.locations.filter(
            (l) => l.id !== locationId
          )
        }
      } else {
        const err = await res.json()
        throw new Error(err.error || 'Failed to delete location')
      }
    } catch (e) {
      throw e
    }
  }

  return {
    events,
    eventsTotal,
    currentEvent,
    persons,
    personsTotal,
    currentPerson,
    loading,
    error,
    fetchEvents,
    fetchEvent,
    createEvent,
    updateEvent,
    deleteEvent,
    fetchPersons,
    fetchPerson,
    createPerson,
    updatePerson,
    deletePerson,
    fetchPersonLocations,
    addLocation,
    updateLocation,
    deleteLocation,
  }
})
