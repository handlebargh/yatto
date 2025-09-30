## v0.18.2 (2025-09-30)

### Fix

- tell jj where to fetch and rebase from (#42)
- run rebase command prior to push

## v0.18.1 (2025-09-27)

### Fix

- **deps**: update dependencies

## v0.18.0 (2025-09-27)

### Feat

- handle remotes automatically at initialization (#39)

## v0.17.0 (2025-09-27)

### Feat

- add jj backend support (#37)

### Fix

- **deps**: update dependencies

### Refactor

- rename storage config struct and fix other linting errors
- pass a config struct to CreateStorageDir function
- align form update methods

## v0.16.0 (2025-09-19)

### Feat

- add Homebrew formula for yatto (#32)

### Fix

- **deps**: update dependencies

## v0.15.0 (2025-08-12)

### Feat

- add multi-select for choosing existing labels (#25)

## v0.14.0 (2025-08-06)

### Feat

- allow editing task from within task view (#23)
- add vim-like keybinds (#20)
- create keybinds for quit/go back

### Fix

- use viewport model on task preview box (#22)
- require upper case D to delete a project

## v0.13.2 (2025-08-03)

### Fix

- allow printing after pull; do not pull if repo is missing INIT file

## v0.13.1 (2025-08-03)

### Fix

- must commit before rebase

### Refactor

- don't use hardcoded color

## v0.13.0 (2025-08-03)

### Feat

- make colors customizable and fix wrong instructions in the readme

## v0.12.0 (2025-08-03)

### Feat

- allow filtering labels by regexp
- sort static task list and print with labels
- add the printer package to print task lists in a non-interactive way to stdout

### Fix

- **deps**: update dependencies

### Refactor

- create separate packages for colors and helpers

## v0.11.0 (2025-07-30)

### Feat

- print version information when run with -version flag (#9)

### Fix

- allow task titles and labels of arbitrary length

### Refactor

- show correct version info when installed by go toolchain
- get rid of an ambiguity that could make believe that this question could lead to another form (#1)

## v0.10.4 (2025-07-27)

### Fix

- use visible list instead of GlobalIndex
- correct due date format description

## v0.10.3 (2025-07-27)

### Fix

- use local time zone for due dates

## v0.10.2 (2025-07-26)

### Fix

- fix sorting by using multi-sort
- show label when task is due today

## v0.10.1 (2025-07-26)

### Fix

- redo task details view and display it in a pager

## v0.10.0 (2025-07-26)

### Feat

- add new state in progress and redo task sorting

## v0.9.2 (2025-07-24)

### Fix

- **deps**: update dependencies

## v0.9.1 (2025-07-24)

### Fix

- allow past dates when editing task
- show done label even when task was due at the same day

### Refactor

- remove obsolete config variable

## v0.9.0 (2025-07-22)

### Feat

- add task labels
- add sorting by task state

## v0.8.0 (2025-07-21)

### Feat

- sort tasks by due date
- add due dates to tasks

### Fix

- sort tasks with no due date last
- resolve issues with project and task status messages

### Refactor

- add tasks due status message
- use receiver functions where possible
- make PriorityValue a receiver function

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
