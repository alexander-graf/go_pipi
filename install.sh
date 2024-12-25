#!/bin/bash

# Kompiliere das Programm
go build -o newpipi

# Erstelle Verzeichnisse
sudo mkdir -p /usr/local/bin
sudo mkdir -p /usr/share/icons/newpipi

# Kopiere die ausführbare Datei
sudo cp newpipi /usr/local/bin/
sudo chmod +x /usr/local/bin/newpipi

# Kopiere das Icon
sudo cp icon/ceferino.ico /usr/share/icons/newpipi/

# Kopiere und registriere die Desktop-Datei
sudo cp newpipi.desktop /usr/share/applications/
sudo update-desktop-database

echo "Installation abgeschlossen!" 