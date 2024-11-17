#!/bin/bash

# Definisikan array dengan direktori semua service
tidys=("dokumen-service" "informasi-service" "kegiatan-service" "user-service" "weeklyTimeline-service" "project-service")

# Loop melalui semua service dan upgrade package
for tidy in "${tidys[@]}"
do
  echo "Upgrading $tidy"
  cd $tidy
  go mod tidy
  cd ..
done