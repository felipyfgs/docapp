export function useFilterControls() {
  const open = useState('filter-controls-open', () => false)

  function toggle() {
    open.value = !open.value
  }

  function close() {
    open.value = false
  }

  return { open, toggle, close }
}
