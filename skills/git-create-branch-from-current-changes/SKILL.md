---
name: git-create-branch-from-current-changes
description: Safely create a new Git branch from the current working directory with uncommitted changes, without staging, committing, stashing, or modifying files.
---

# Create Git Branch From Current Uncommitted Changes

## When To Use
- Use when uncommitted working directory changes should live on a separate branch.
- The goal is to keep all changes untouched, avoid staging or committing anything, and create a short conventional branch name.
- Show all Git commands before executing them.

## Inputs To Ask For
- Ask for a short English description of the change, 2-6 words, if the user did not provide one.
- Ask for confirmation that it is OK to inspect with `git status` and later run `git switch -c <branch>`.
- If no description is available, ask: `Give me a short English description for these changes (2-6 words), used to name the branch.`

## Branch Type Rules
- Use `fix/` if the description contains `fix`, `bug`, `error`, `crash`, or `fail`.
- Use `refactor/` if it contains `refactor`, `cleanup`, or `restructure`.
- Use `chore/` if it contains `chore`, `deps`, `dependency`, `ci`, or `build`.
- Otherwise use `feature/`.

## Branch Name Format
- Lowercase the description.
- Replace spaces and non-alphanumeric characters with `-`.
- Collapse repeated `-` characters.
- Trim leading and trailing `-`.
- Optionally truncate the slug to about 40-50 characters.
- Final format is `<type>/<slug>`, such as `feature/add-user-search`, `fix/login-bug`, `refactor/auth-middleware`, or `chore/update-ci-deps`.

## Safety Requirements
- Never run `git add`, `git commit`, or `git stash`.
- Never modify files; only change the current branch with Git commands.
- Never discard, reset, or revert changes.
- Always show exact Git commands before executing them and ask for confirmation.
- If any command fails, stop and show the error. Do not attempt automatic fixes.

## Workflow
1. Explain that you will inspect the repository state, then propose a branch name and commands.
2. Run `git status` only after permission, unless the user pasted current status output.
3. Summarize the current branch and whether files are modified, added, deleted, or untracked.
4. If the directory is not a Git repo or there are no changes, stop and explain.
5. Build and show the proposed branch name and inferred type.
6. Ask whether to use the branch name. If the user says no, ask for a custom full branch name and validate that it is non-empty and contains no whitespace.
7. Show the planned commands before executing:

```bash
git status
git switch -c <new-branch-name>
```

8. Explain that the command creates and switches to a new branch from current `HEAD` while keeping uncommitted changes as-is.
9. Ask explicitly: `Confirm that I should run the above command(s) now. (y/N)`
10. If confirmed, run exactly `git switch -c <new-branch-name>`.
11. Use `git checkout -b <new-branch-name>` only if `git switch` is unavailable or fails for that reason.
12. If branch creation fails, show the exact error and ask the user to choose a different name or stop.
13. After success, run `git status` and confirm the new branch is active and the uncommitted changes remain listed.

## Scope Boundaries
- If the user asks to stage or commit, that is outside this skill; ask for confirmation before doing anything beyond branch creation.
- Keep explanations terse and focused.
