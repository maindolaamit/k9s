#!/bin/bash
# Quick skin switcher for k9s

SKIN=${1:-tokyo-night}

cat > ~/.k9s/config.yaml << EOF
k9s:
  ui:
    skin: $SKIN
EOF

echo "Switched to skin: $SKIN"
echo "Available skins:"
ls -1 ~/.k9s/skins/ | sed 's/.yaml$//' | sed 's/^/  - /'
