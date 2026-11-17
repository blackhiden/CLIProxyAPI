#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Create isolated feature worktrees from a shared main commit.

Usage: parallel-feature-worktrees.sh <repository> <base-commit> <worktree-parent> <task> [<task>...]

Each task is a branch/worktree suffix. Existing branches and paths are never
overwritten. The script only creates worktrees; it never removes or merges them.
USAGE
}

if [[ $# -lt 4 ]]; then
  usage >&2
  exit 2
fi

repository=$(cd "$1" && pwd)
base_commit=$2
worktree_parent=$3
shift 3

git -C "$repository" cat-file -e "${base_commit}^{commit}"
mkdir -p "$worktree_parent"
worktree_parent=$(cd "$worktree_parent" && pwd)

for task in "$@"; do
  if [[ ! "$task" =~ ^[a-z0-9]+([a-z0-9-]*[a-z0-9])?$ ]]; then
    printf 'Invalid task suffix: %s\n' "$task" >&2
    exit 2
  fi
  branch="feature/$task"
  path="$worktree_parent/$task"

  if git -C "$repository" show-ref --verify --quiet "refs/heads/$branch"; then
    printf 'Branch already exists; refusing to overwrite: %s\n' "$branch" >&2
    exit 1
  fi
  if [[ -e "$path" ]]; then
    printf 'Worktree path already exists; refusing to overwrite: %s\n' "$path" >&2
    exit 1
  fi

  git -C "$repository" worktree add -b "$branch" "$path" "$base_commit"
done