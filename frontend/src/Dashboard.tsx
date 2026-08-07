import { useEffect, useState } from 'react'
import type { DailySummary } from './types'

const CURRENT_USER_ID = 1

function todayDate(): string {
  return new Date().toISOString().split('T')[0]
}

export default function Dashboard({ refreshKey }: { refreshKey: number }) {
  const [summary, setSummary] = useState<DailySummary | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const date = todayDate()
    setError(null)
    fetch(`/api/summary/daily?user_id=${CURRENT_USER_ID}&date=${date}`)
      .then(res => {
        if (!res.ok) throw new Error('summary failed')
        return res.json()
      })
      .then(data => setSummary(data))
      .catch(() => setError('Could not load today\'s summary.'))
  }, [refreshKey])

  if (error) return <div className="dashboard"><p className="error">{error}</p></div>
  if (!summary) return <div className="dashboard">Loading…</div>

  return (
    <div className="dashboard">
      <h2>Today's Summary — {summary.date}</h2>
      <div className="dashboard-totals">
        <span>{summary.total.calories.toFixed(0)} cal</span>
        <span>{summary.total.protein.toFixed(1)}g protein</span>
        <span>{summary.total.carbs.toFixed(1)}g carbs</span>
        <span>{summary.total.fat.toFixed(1)}g fat</span>
        <span>{summary.total.fiber.toFixed(1)}g fiber</span>
      </div>

      <h3>Meal History</h3>
      {summary.meals.length === 0 && <p>No meals logged today.</p>}
      {summary.meals.map(meal => (
        <div className="meal-history-entry" key={meal.id}>
          <div className="meal-history-time">
            {new Date(meal.logged_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </div>
          <ul>
            {meal.items.map(item => (
              <li key={item.id}>
                {item.quantity} {item.unit} {item.food?.name ?? `food #${item.food_id}`}
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  )
}
