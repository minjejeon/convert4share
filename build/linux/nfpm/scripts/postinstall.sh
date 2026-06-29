#!/bin/sh
# Refresh the desktop database and icon cache after install.
update-desktop-database -q || true
gtk-update-icon-cache -q -t -f /usr/share/icons/hicolor || true
