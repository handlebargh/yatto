## v0.7.1 (2025-07-19)

### Fix

- supply full path to task file to recognize updates
- ensure err is nil if project directory already exists

### Refactor

- set form confirm to true by default

## v0.7.0 (2025-07-19)

### Feat

- show some project information in initial view
- allow users to choose project colors
- separate tasks in projects

### Fix

- set the right item object
- initialize renderer at startup and set list title colors

### Refactor

- set file extension on task files

## v0.6.2 (2025-07-16)

### Fix

- pull and rebase before committing if remote is enabled

### Refactor

- reduce time to show filled progress bar

## v0.6.1 (2025-07-15)

### Fix

- allow task titles and descriptions with the letter q

## v0.6.0 (2025-07-15)

### Feat

- render tasks as markdown
- show spinner while pulling on init
- use animated progress bar instead of spinner
- extensive refactoring that also introduces the branches view

### Fix

- add some form improvements
- pull only if remote is enabled and check for errors
- remove background and padding from title in preview
- remove branch item
- remove branch view for adding too much complexity with too little benefit
- get branch view working

### Refactor

- **deps**: update dependencies
- show deletion prompts and git errors centered on their own
- show done tasks in list with green, strikethrough title
- create setter methods
- change how list items are displayed
- make git a hard dependency
- rewrite git commands for the most part

## v0.5.0 (2025-07-10)

### Feat

- allow task completion to be toggled by shortcut

### Refactor

- redo list view
- remove confirm in edit view in favour of hotkey

## v0.4.0 (2025-07-09)

### Feat

- show spinner while synchronization is in progress
- run git commands in background

### Fix

- set margin to vertically align with task view
- revert conditional
- synchronize tasks at startup

## v0.3.0 (2025-07-09)

### Feat

- let yatto use remote repositories
- add status messages on delete/update/create tasks
- display completed tasks differently

### Refactor

- activate git by default
- show task priority right next to task state

## v0.2.0 (2025-07-08)

### Feat

- add task completion status

### Refactor

- redo task view

## v0.1.0 (2025-07-08)

### Feat

- show task view right next to task list

### Fix

- show the correct confirmation text for both situations
- init git in existing storage dirs and fix FileExists function
- check for git remote instead of git use at all

## v0.0.3 (2025-07-08)

## v0.0.2 (2025-07-08)

### Refactor

- fix module path
