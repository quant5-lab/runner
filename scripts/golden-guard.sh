#!/bin/sh
# Exits 1 when protected golden/data artifacts are staged without operator
# authorization (ALLOW_GOLDEN_CHANGE=1), exits 0 otherwise.
#
# Testability: set GOLDEN_GUARD_STAGED_FILES to a newline-separated list of
# repo-relative paths to bypass the live git diff call.
set -e

if [ "${GOLDEN_GUARD_STAGED_FILES+set}" = "set" ]; then
    staged="$GOLDEN_GUARD_STAGED_FILES"
else
    staged="$(git diff --cached --name-only)"
fi

is_protected() {
    case "$1" in
        tests/golden/fixtures/expected/*.json)       return 0 ;;
        tests/golden/fixtures/data/*.json)           return 0 ;;
        tests/golden/testutil/direction_reference.go) return 0 ;;
    esac
    return 1
}

touched=""
while IFS= read -r path; do
    [ -z "$path" ] && continue
    if is_protected "$path"; then
        touched="${touched}    ${path}\n"
    fi
done << STAGED
$staged
STAGED

[ -z "$touched" ] && exit 0
[ "${ALLOW_GOLDEN_CHANGE:-0}" = "1" ] && exit 0

printf "✗ Trusted golden/data artifacts staged without authorization:\n"
printf "%b" "$touched"
printf "  TV-verified; must not change silently.\n"
printf "  If intentional: ALLOW_GOLDEN_CHANGE=1 git commit ...\n"
exit 1
