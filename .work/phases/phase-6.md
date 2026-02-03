# Phase 6: App Integration

## Goal
Wire the InsightsView into the main app with mode switching and filter integration.

## Changes
- `internal/tui/app.go`: Add insights mode, keybinding, view routing, filter wiring

## Implementation
1. Add to Model struct:
   - `insights insights.InsightsView`
   - `showInsights bool`
2. Update `NewModel()`:
   - Initialize insights: `insights: insights.New()`
3. Add keybinding in `Update()` switch:
   ```go
   case "s":
       m.showInsights = !m.showInsights
       if m.showInsights {
           m.recalculateInsights()
       }
       return m, nil
   ```
4. Implement `recalculateInsights()`:
   - Get current filtered commits from list
   - Load file changes for commits (batch or cached)
   - Call `m.insights.Recalculate(commits, fileChanges)`
5. Update `View()`:
   - Add check before other views:
     ```go
     if m.showInsights {
         return m.renderInsightsLayout()
     }
     ```
6. Implement `renderInsightsLayout()`:
   - Same structure as main layout but with insights view instead of list
   - Keep histogram visible at bottom
   - Update header to show "Insights" mode indicator
7. Update `applyAllFilters()`:
   - If showInsights, recalculate insights after filter change
8. Update help overlay:
   - Add `s` keybinding: "Toggle insights view"

## Testing
- Press `s` → switches to insights view
- Press `s` again → returns to graph view
- Apply branch filter → insights update to show filtered data
- Histogram selection → insights recalculate
- Header shows mode indicator

## Estimated LOC
~80
