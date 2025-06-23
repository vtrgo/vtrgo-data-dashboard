import { useEffect, useState } from 'react'
import { Bar } from 'react-chartjs-2'
import { Chart as ChartJS, BarElement, CategoryScale, LinearScale, Tooltip, Legend } from 'chart.js'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Tooltip,
  Legend
)


type StatsResponse = {
  boolean_percentages: Record<string, number>
  fault_counts: Record<string, number>
  float_averages: Record<string, number>
}

function App() {
  const [data, setData] = useState<StatsResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchStats = () => {
      fetch('/api/stats?start=-5m')
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
    <div style={{ padding: '2rem' }}>
      <h1>VTR Dashboard</h1>
      {error && <p style={{ color: 'red' }}>Error: {error}</p>}
      {!data && !error && <p>Loading...</p>}

      {data && (
        <div>
          <h2>Boolean Percentages</h2>
          <ul>
            {Object.entries(data.boolean_percentages).map(([k, v]) => (
              <li key={k}>{k}: {v.toFixed(1)}%</li>
            ))}
          </ul>

          <h2>Fault Counts</h2>
          <ul>
            {Object.entries(data.fault_counts).map(([k, v]) => (
              <li key={k}>{k}: {v}</li>
            ))}
          </ul>

          <h2>Float Averages</h2>
            <Bar
              data={{
                labels: Object.keys(data.float_averages),
                datasets: [
                  {
                    label: 'Float Averages',
                    data: Object.values(data.float_averages),
                  },
                ],
              }}
              options={{
                responsive: true,
                plugins: {
                  legend: { display: false },
                },
                scales: {
                  y: {
                    beginAtZero: true,
                  },
                },
              }}
            />

        </div>
      )}
    </div>
  )
}

export default App
