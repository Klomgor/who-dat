import './styles/colors.scss'
import './styles/typography.scss'
import './styles/document.scss'
import './styles/layout.scss'
import './styles/form.scss'
import './styles/content.scss'
import './styles/results.scss'

import Alpine from 'alpinejs'

// Type definitions
interface WhoisData {
  domain?: Record<string, any>
  registrar?: Record<string, any>
  registrant?: Record<string, any>
  administrative?: Record<string, any>
  technical?: Record<string, any>
  billing?: Record<string, any>
}

interface WhoisResponse {
  domain: string
  data: WhoisData
  cached: boolean
  timestamp: string
  error?: {
    code: string
    message: string
  }
}

interface FormattedSection {
  label: string
  items: Array<{ key: string; value: any }>
}

interface AppState {
  domain: string
  loading: boolean
  error: string
  response: WhoisResponse | null
  formattedSections: FormattedSection[]
  showRaw: boolean
  showMetadata: boolean
}

// Utility functions
const formatLabel = (str: string): string =>
  str.replace(/_/g, ' ')
    .replace(/\b\w/g, c => c.toUpperCase())

const formatValue = (val: any): string => {
  if (val === null || val === undefined) return 'N/A'
  if (Array.isArray(val)) return val.join(', ')
  if (typeof val === 'object') return JSON.stringify(val, null, 2)
  if (typeof val === 'string' && val.startsWith('http')) {
    return `<a href="${val}" target="_blank" rel="noopener">${val}</a>`
  }
  return String(val)
}

const shouldSkipField = (key: string, value: any): boolean => {
  if (value === null || value === undefined || value === '') return true
  if (value === 'REDACTED' || value === 'DATA REDACTED') return true
  if (key === 'id' && value.includes('REDACTED')) return true
  return false
}

const formatSection = (sectionName: string, data: Record<string, any>): FormattedSection => {
  const items = Object.entries(data)
    .filter(([key, val]) => !shouldSkipField(key, val))
    .map(([key, val]) => ({
      key: formatLabel(key),
      value: formatValue(val)
    }))

  return {
    label: formatLabel(sectionName),
    items
  }
}

const formatWhoisData = (data: WhoisData): FormattedSection[] => {
  const sections: FormattedSection[] = []
  const sectionOrder = ['domain', 'registrar', 'registrant', 'administrative', 'technical', 'billing']

  sectionOrder.forEach(key => {
    const sectionData = data[key as keyof WhoisData]
    if (sectionData && Object.keys(sectionData).length > 0) {
      sections.push(formatSection(key, sectionData))
    }
  })

  return sections
}

// Main Alpine component
function whoisLookup(): AppState & { lookup: () => Promise<void>; toggleRaw: () => void; toggleMetadata: () => void } {
  return {
    domain: '',
    loading: false,
    error: '',
    response: null,
    formattedSections: [],
    showRaw: false,
    showMetadata: true,

    toggleRaw() {
      this.showRaw = !this.showRaw
    },

    toggleMetadata() {
      this.showMetadata = !this.showMetadata
    },

    async lookup() {
      if (!this.domain.trim()) {
        this.error = 'Please enter a domain name'
        return
      }

      this.loading = true
      this.error = ''
      this.response = null
      this.formattedSections = []

      try {
        const res = await fetch(`/${this.domain.trim()}`)
        const data: WhoisResponse = await res.json()

        if (!res.ok) {
          throw new Error(data.error?.message || 'Failed to fetch WHOIS data')
        }

        if (data.error) {
          this.error = `${data.error.code}: ${data.error.message}`
          return
        }

        this.response = data
        this.formattedSections = formatWhoisData(data.data)
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'An unknown error occurred'
        console.error('WHOIS lookup failed:', err)
      } finally {
        this.loading = false
      }
    }
  }
}

// Register Alpine component
Alpine.data('whoisLookup', whoisLookup)

// Make Alpine available globally
window.Alpine = Alpine

// Start Alpine
Alpine.start()
