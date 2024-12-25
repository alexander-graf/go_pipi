#!/bin/bash

# Entferne die Dateien
sudo rm /usr/local/bin/newpipi
sudo rm -r /usr/share/icons/newpipi
sudo rm /usr/share/applications/newpipi.desktop
sudo update-desktop-database

echo "Deinstallation abgeschlossen!" 