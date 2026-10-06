#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p .cache/tmp
scratch=$(mktemp -d "$PWD/.cache/tmp/tool-version.XXXXXX")
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
cat > "$scratch/stub" <<'STUB'
#!/bin/sh
test "$*" = 'version --json' || exit 9
printf '%s\n' "$STUB_JSON"
exit "${STUB_EXIT:-0}"
STUB
chmod +x "$scratch/stub"
total=0 failed=0
check() {
    label=$1 expected=$2 STUB_JSON=$3
    export STUB_JSON
    total=$((total + 1))
    code=0
    output=$(sh scripts/tool-version-check.sh "$scratch/stub" "${4:-v0.5.15}" "${5:-v0.6.0}" 2>&1) || code=$?
    if [ "$code" = "$expected" ] && [ "$(printf '%s\n' "$output" | wc -l | tr -d ' ')" = 1 ] &&
       printf '%s\n' "$output" | grep -Eq '^tool-version: (PASS|FAIL) stub installed .*; accepted >= .* and < .*; .+'; then
        printf 'PASS %s: %s\n' "$label" "$output"
    else
        printf 'FAIL %s: expected exit %s, got %s: %s\n' "$label" "$expected" "$code" "$output"
        failed=$((failed + 1))
    fi
}
for version in v0.5.15 v0.5.16 v0.5.99; do
    check "$version" 0 "{\"version\":\"$version\",\"version_agreement\":\"pass\"}"
done
for version in v0.5.14 v0.6.0 v0.5.15-rc.1 v0.5.15+build v0.05.15; do
    check "$version" 1 "{\"version\":\"$version\",\"version_agreement\":\"pass\"}"
done
check disagreement 1 '{"version":"v0.5.15","version_agreement":"fail"}'
check non-json 1 'not JSON'
check missing-version 1 '{"version_agreement":"pass"}'
check missing-agreement 1 '{"version":"v0.5.15"}'
check non-string-version 1 '{"version":15,"version_agreement":"pass"}'
check multiple-documents 1 '{"version":"v0.5.15","version_agreement":"pass"} {}'
check empty 1 ''
check rulefloor-min 0 '{"version":"v0.9.1","version_agreement":"pass"}' v0.9.1 v0.10.0
check rulefloor-next 0 '{"version":"v0.9.2","version_agreement":"pass"}' v0.9.1 v0.10.0
check rulefloor-old 1 '{"version":"v0.9.0","version_agreement":"pass"}' v0.9.1 v0.10.0
check rulefloor-upper 1 '{"version":"v0.10.0","version_agreement":"pass"}' v0.9.1 v0.10.0
STUB_EXIT=7; export STUB_EXIT
check tool-failed 1 '{"version":"v0.5.15","version_agreement":"pass"}'
printf 'SELFTEST: %s cases, %s failures\n' "$total" "$failed"
test "$failed" = 0
