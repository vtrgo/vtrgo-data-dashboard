import { useEffect, useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

type StatsResponse = {
  boolean_percentages: Record<string, number>
  fault_counts: Record<string, number>
  float_averages: Record<string, number>
}

function App() {
  const [data, setData] = useState<StatsResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const groupedFloats = data ? groupFloatsBySection(data.float_averages) : {}

  useEffect(() => {
    const fetchStats = () => {
      fetch('/api/stats?start=-3h')
        .then((res) => {
          if (!res.ok) throw new Error('Failed to fetch stats')
          return res.json()
        })
        .then(setData)
        .catch((err) => setError(err.message))
    }

    fetchStats() // initial load
    const interval = setInterval(fetchStats, 5000) // refresh every 5s

    return () => clearInterval(interval)
  }, [])

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-3xl font-bold">VTR Dashboard</h1>
      {error && <p className="text-red-500">Error: {error}</p>}
      {!data && !error && <Skeleton className="h-32 w-full" />}

      {data && (
        <>
          <Card>
            <CardContent>
              <h2 className="text-2xl font-semibold mb-2">Boolean Percentages</h2>
              <ul className="list-disc ml-6">
                {Object.entries(data.boolean_percentages).map(([k, v]) => (
                  <li key={k}>{k}: {v.toFixed(1)}%</li>
                ))}
              </ul>
            </CardContent>
          </Card>

          <Card>
            <CardContent>
              <h2 className="text-2xl font-semibold mb-2">Fault Counts</h2>
              <ul className="list-disc ml-6">
                {Object.entries(data.fault_counts).map(([k, v]) => (
                  <li key={k}>{k}: {v}</li>
                ))}
              </ul>
            </CardContent>
          </Card>

          <h2 className="text-2xl font-semibold">Float Averages</h2>
          {Object.entries(groupedFloats).map(([sectionKey, { title, values }]) => (
            <Card key={sectionKey} className="mb-6">
              <CardContent>
                <h3 className="text-xl font-medium mb-2">{title}</h3>
                <ul className="list-none space-y-1 mt-2">
                  {Object.entries(values).map(([field, val]) => (
                    <li key={field} className="flex justify-between">
                      <span className="text-muted-foreground">{field}</span>
                      <span className="font-medium">{val.toFixed(2)}</span>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          ))}
        </>
      )}
    </div>
  )
}

function formatSectionTitle(name: string) {
  return name.replace(/([a-z])([A-Z])/g, '$1 $2')
             .replace(/([A-Z])([A-Z][a-z])/g, '$1 $2')
             .trim()
}

function groupFloatsBySection(floats: Record<string, number>) {
  const grouped: Record<string, { title: string; values: Record<string, number> }> = {}

  for (const fullKey in floats) {
    const parts = fullKey.split('.')
    if (parts.length !== 3) continue

    const [, section, subField] = parts
    if (!grouped[section]) {
      grouped[section] = {
        title: formatSectionTitle(section),
        values: {},
      }
    }
    grouped[section].values[subField] = floats[fullKey]
  }

  return grouped
}

export default App
