import { useEffect, useState } from 'react'
import './App.css'
import type { Food } from './types'
import MealLogger from './MealLogger'
import Dashboard from './Dashboard'
import ChatMeal from './ChatMeal'

type DisplayNutrition = {
  label: string,
  calories: number,
  protein: number,
  carbs: number,
  fat: number,
  fiber: number,
}

function displayNutrition(food: Food): DisplayNutrition {
  if (food.unit === 'grams') {
    const scale = 100 / food.unitquantity
    return {
      label: 'per 100g',
      calories: food.calories * scale,
      protein: food.protein * scale,
      carbs: food.carbs * scale,
      fat: food.fat * scale,
      fiber: food.fiber * scale,
    }
  }
  return {
    label: `per ${food.unitquantity} ${food.unit}`,
    calories: food.calories,
    protein: food.protein,
    carbs: food.carbs,
    fat: food.fat,
    fiber: food.fiber,
  }
}

function FoodSearch() {
  const [query, setQuery] = useState('')
  const [foods, setFoods] = useState<Food[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (query.trim().length === 0) {
      setFoods([])
      setError(null)
      setLoading(false)
      return
    }
    let cancelled = false
    setLoading(true)
    setError(null)
    fetch(`/api/foods/search?q=${encodeURIComponent(query)}`)
      .then((res) => {
        if (!res.ok) throw new Error('search failed')
        return res.json()
      })
      .then((data) => {
        if (cancelled) return
        setFoods(data)
      })
      .catch(() => {
        if (cancelled) return
        setError('Could not search foods. Try again.')
      })
      .finally(() => {
        if (cancelled) return
        setLoading(false)
      })
    return () => { cancelled = true }
  }, [query])

  return (
    <div className="food-search">
      <h2>Food Search</h2>
      <input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="search foods…" />
      {loading && <p>Searching…</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && query.trim().length > 0 && foods.length === 0 && (
        <p>No foods found for "{query}".</p>
      )}
      {foods.map((food) => {
        const nutrition = displayNutrition(food)
        return (
          <div key={food.id}>
            <strong>{food.name}</strong> ({nutrition.label}): {nutrition.calories.toFixed(0)} cal,{' '}
            {nutrition.protein.toFixed(1)}g protein, {nutrition.carbs.toFixed(1)}g carbs,{' '}
            {nutrition.fat.toFixed(1)}g fat, {nutrition.fiber.toFixed(1)}g fiber
          </div>
        )
      })}
    </div>
  )
}

function App() {
  const [refreshKey, setRefreshKey] = useState(0)

  return (
    <section id="center">
      <h1>Nutrition Assistant</h1>
      <FoodSearch />
      <MealLogger onSaved={() => setRefreshKey(k => k + 1)} />
      <ChatMeal onLogged={() => setRefreshKey(k => k + 1)} />
      <Dashboard refreshKey={refreshKey} />
    </section>
  )
}

export default App
