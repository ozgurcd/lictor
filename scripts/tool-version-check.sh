#!/bin/sh
# Local development compatibility policy; CI downloads remain exactly pinned.
set -eu
tool=${1:-unknown} min=${2:-unknown} below=${3:-unknown} installed=unknown
fail() {
    printf 'tool-version: FAIL %s installed %s; accepted >= %s and < %s; %s\n' "${tool##*/}" "${installed:-unknown}" "$min" "$below" "$1" >&2
    exit 1
}
[ "$#" = 3 ] || fail 'usage: TOOL MIN BELOW'
json=$("$tool" version --json 2>/dev/null) || fail 'version command failed'
installed=$(printf '%s' "$json" | jq -ers '
    if length == 1 and (.[0] | type) == "object" and (.[0].version | type) == "string"
    then .[0].version | @json else error("version") end' 2>/dev/null) || fail 'invalid JSON or missing string version'
printf '%s' "$json" | jq -e --arg min "$min" --arg below "$below" '
    [.version, $min, $below] | all(.[]; test("\\Av(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\z"))
' >/dev/null 2>&1 || fail 'requires stable vMAJOR.MINOR.PATCH versions without suffixes'
printf '%s' "$json" | jq -e '.version_agreement == "pass"' >/dev/null 2>&1 || fail 'version_agreement must be pass'
printf '%s' "$json" | jq -e --arg min "$min" --arg below "$below" '
    def parts: ltrimstr("v") | split(".") | map(tonumber);
    (.version | parts) as $v | ($min | parts) as $lo | ($below | parts) as $hi |
    $lo <= $v and $v < $hi
' >/dev/null 2>&1 || fail 'version outside compatible range'
printf 'tool-version: PASS %s installed %s; accepted >= %s and < %s; version_agreement pass\n' "${tool##*/}" "$installed" "$min" "$below"
