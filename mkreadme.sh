#!/bin/sh
cat <<EOF
gohacks
=======

Various Go hacks.

EOF

go list -f '- **{{ .Name }}:** {{ .Doc }}' github.com/npat-efault/gohacks/...

echo
