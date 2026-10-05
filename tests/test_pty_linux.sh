#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/apply.sh

fail=0

# issue #5: Linux laporn ? (single) untuk daemon, ?? (macOS) untuk yang sama.
# Filter harus skip keduanya dan ambil proses client pertama.
case_body=$' ?      herdr\npts/4    herdr\n'
got=$(printf '%s\n' "$case_body" | awk '$2=="herdr" && $1!="?" && $1!="??" {print "/dev/"$1; exit}')
[ "$got" = "/dev/pts/4" ] || { echo "FAIL linux skipped daemon ? → got '$got'"; fail=1; }

# macOS: daemon lapor ?? , harus tetap lewati
macos_body=$'??      herdr\npts/4    herdr\n'
got=$(printf '%s\n' "$macos_body" | awk '$2=="herdr" && $1!="?" && $1!="??" {print "/dev/"$1; exit}')
[ "$got" = "/dev/pts/4" ] || { echo "FAIL macos skipped daemon ?? → got '$got'"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_pty_linux" || exit 1
