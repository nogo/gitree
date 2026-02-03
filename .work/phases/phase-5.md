# Phase 5: Heatmap Rendering & Layout

## Goal
Implement calendar heatmap grid rendering and the responsive layout that combines stats and heatmap.

## Changes
- `internal/tui/insights/render_heatmap.go`: New file with calendar rendering
- `internal/tui/insights/render.go`: New file with View() and layout logic

## Implementation

### render_heatmap.go
1. Implement `renderCalendar(cal CalendarData, width, height int) string`:
   - Month labels on top row: `Jan  Feb  Mar  Apr ...`
   - Weekday labels on left: `Mon`, `Wed`, `Fri` (or single letter)
   - Grid of cells using block characters: `█` or `▓` or `░`
   - Each cell colored by heat level (Heat0-Heat4 styles)
   - Handle viewport if calendar wider than available width
2. Helper: `cellChar(level int) string` - returns appropriate block char
3. Helper: `monthLabels(cal CalendarData) string` - positioned month names

### render.go
1. Implement `(v InsightsView) View() string`:
   - Check if wide (>= 100) or narrow
   - Wide layout:
     ```
     lipgloss.JoinHorizontal(
       statsPanel,   // authors + files stacked
       heatmapPanel, // calendar
     )
     ```
   - Narrow layout:
     ```
     lipgloss.JoinVertical(
       statsPanel,
       heatmapPanel,
     )
     ```
   - Add summary line at bottom
2. Helper: `renderStatsPanel(width, height int) string`:
   - Stack authors (top half) and files (bottom half)
3. Calculate panel dimensions based on total width/height

## Testing
- Wide terminal shows side-by-side layout
- Narrow terminal stacks vertically
- Calendar cells show correct heat colors
- Month labels align with weeks

## Estimated LOC
~120
