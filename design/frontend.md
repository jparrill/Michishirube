# Frontend Design

This document describes the user interface design and user experience for Michishirube.

## Technology Stack

- **HTML Templates**: Go's `html/template` for server-side rendering
- **CSS**: Custom CSS with minimal external dependencies
- **JavaScript**: Vanilla JavaScript for interactive features
- **Icons**: Simple Unicode/emoji icons or lightweight icon font

## Design Principles

### Simplicity First
- Clean, minimal interface focused on productivity
- No complex frameworks or build steps
- Fast loading and responsive design

### Developer-Focused UX
- Keyboard shortcuts for common actions
- Quick access to frequently used filters
- Minimal clicks to complete common tasks

### Information Density
- Show maximum relevant information without clutter
- Use visual hierarchy to highlight important elements
- Collapsible sections for detailed information

## Page Structure

### Main Dashboard (`/`)

```
┌─────────────────────────────────────────────────────────────┐
│ 🗯️ Michishirube                                [+ New Task] │
├─────────────────────────────────────────────────────────────┤
│ [Search: JIRA ID, title, tags...] [🔍] [☐ Include archived] │
├─────────────────────────────────────────────────────────────┤
│ [All][New][In Progress][Blocked][Done] Priority: [All ▼]    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ┌─[OCPBUGS-1234] Fix memory leak in pod controller ──────┐  │
│ │ 🔴 High | 🟡 Blocked | 📋 k8s, memory               │  │
│ │ 🔗 PR #456 (Draft) | 💬 Slack thread | 📝 3 comments    │  │
│ │ ⚠️ Blocked by: Waiting for review from @team-lead      │  │
│ │ 📅 Created: 2 days ago | 📝 Updated: 1 hour ago        │  │
│ └─────────────────────────────────────────────────────────┘  │
│                                                             │
│ ┌─[NO-JIRA] Investigate flaky test in CI ──────────────────┐  │
│ │ 🟡 Normal | 🔵 In Progress | 📋 ci, test               │  │
│ │ 🔗 GitHub Issue #789 | 📝 2 comments                    │  │
│ │ 📅 Created: 1 day ago | 📝 Updated: 30 min ago         │  │
│ └─────────────────────────────────────────────────────────┘  │
│                                                             │
│ ┌─[OCPBUGS-5678] Update documentation ──────────────────────┐  │
│ │ ⚪ Minor | 🟢 Done | 📋 docs                           │  │
│ │ 🔗 PR #123 (Merged)                                     │  │
│ │ 📅 Created: 3 days ago | 📝 Completed: 2 hours ago     │  │
│ └─────────────────────────────────────────────────────────┘  │
│                                                             │
│ [Load More...] or [Showing 1-20 of 45 tasks]               │
└─────────────────────────────────────────────────────────────┘
```

### Task Detail View (`/task/{id}`)

```
┌─────────────────────────────────────────────────────────────┐
│ ← Back to Dashboard                           [Edit] [Delete] │
├─────────────────────────────────────────────────────────────┤
│ [OCPBUGS-1234] Fix memory leak in pod controller            │
│ Priority: [High ▼] Status: [Blocked ▼]                     │
│ Tags: [k8s] [memory] [+ Add tag]                           │
│                                                             │
│ 📋 Blockers:                                                │
│ • Waiting for review from @team-lead                       │
│ • Need performance test results                            │
│ [+ Add blocker]                                            │
│                                                             │
│ 🔗 Related Links:                                           │
│ ┌─ Pull Request ─────────────────────────────────────────┐   │
│ │ PR #456: Fix memory leak                               │   │
│ │ Status: Draft | https://github.com/org/repo/pull/456  │   │
│ └───────────────────────────────────────────────────────┘   │
│ ┌─ Slack Thread ─────────────────────────────────────────┐   │
│ │ Memory leak discussion                                 │   │
│ │ Status: Active | https://company.slack.com/...        │   │
│ └───────────────────────────────────────────────────────┘   │
│ [+ Add link]                                               │
│                                                             │
│ 💬 Comments:                                                │
│ ┌─ 2 hours ago ──────────────────────────────────────────┐   │
│ │ Found the root cause in the controller reconcile loop │   │
│ └─────────────────────────────────────────────────────────┘   │
│ ┌─ 1 day ago ────────────────────────────────────────────┐   │
│ │ Need to investigate heap allocation patterns          │   │
│ └─────────────────────────────────────────────────────────┘   │
│ [Add comment...]                                           │
│                                                             │
│ Created: Jan 15, 2024 10:30 | Updated: Jan 15, 2024 14:20 │
└─────────────────────────────────────────────────────────────┘
```

### New Task Form (`/new`)

```
┌─────────────────────────────────────────────────────────────┐
│ Create New Task                                   [Cancel]   │
├─────────────────────────────────────────────────────────────┤
│ Jira ID: [OCPBUGS-     ] or [NO-JIRA]                      │
│ Title:   [_____________________________________________]     │
│ Priority: [Normal ▼]                                        │
│ Tags:    [frontend] [x] [backend] [x] [+ Add tag]          │
│                                                             │
│ Initial Links (optional):                                   │
│ [+ Add PR] [+ Add Slack] [+ Add Jira] [+ Add Other]        │
│                                                             │
│ Notes (optional):                                           │
│ [____________________________________________]              │
│ [____________________________________________]              │
│ [____________________________________________]              │
│                                                             │
│                              [Create Task] [Save as Draft] │
└─────────────────────────────────────────────────────────────┘
```

## Visual Design Elements

### Color Coding

**Status Colors:**
- 🔴 Critical priority / Blocked status
- 🟡 High priority / In Progress status  
- 🟢 Normal priority / Done status
- ⚪ Minor priority / New status
- ⚫ Archived status

**Link Type Icons:**
- 🔗 Pull Request
- 💬 Slack Thread
- 📋 Jira Ticket
- 📚 Documentation
- 🌐 Other/External

### Typography
- **Headers**: Clean sans-serif, adequate spacing
- **Body**: High contrast, readable font
- **Code/IDs**: Monospace font for Jira IDs and URLs
- **Status text**: Bold or colored for emphasis

### Interactive Elements
- **Hover effects**: Subtle highlighting on task cards
- **Click targets**: Large enough for easy interaction
- **Form validation**: Real-time feedback on inputs
- **Loading states**: Indicators for async operations

## Responsive Design

### Desktop (1200px+)
- Full-width layout with sidebar potential
- Task cards in single column with full details
- Keyboard shortcuts displayed

### Tablet (768px - 1199px)
- Condensed task cards
- Collapsible filters section
- Touch-friendly buttons

### Mobile (< 768px)
- Stacked layout
- Hamburger menu for filters
- Simplified task view
- Touch-optimized forms

## Keyboard Shortcuts

- `n` - New task
- `/` - Focus search
- `f` - Toggle filters
- `a` - Show all tasks
- `o` - Show open tasks only
- `↑/↓` - Navigate task list
- `Enter` - Open selected task
- `Esc` - Close modals/forms

## JavaScript Functionality

### Core Features
- **Search as you type**: Instant filtering without page reload
- **Tag management**: Add/remove tags with UI feedback
- **Status updates**: Quick status changes via dropdown
- **Form validation**: Client-side validation before submission

### Progressive Enhancement
- Basic functionality works without JavaScript
- JavaScript enhances UX with faster interactions
- Graceful degradation for older browsers

## Accessibility

- **Semantic HTML**: Proper heading structure and landmarks
- **ARIA labels**: For interactive elements and status indicators
- **Keyboard navigation**: Full keyboard accessibility
- **Screen reader support**: Proper labeling and descriptions
- **Color independence**: Status indicated by icons and text, not just color