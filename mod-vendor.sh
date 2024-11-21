#!/bin/bash

# Definisikan array dengan direktori semua service
services=("dokumen-service" "informasi-service" "kegiatan-service" "user-service" "weeklyTimeline-service" "project-service")

# Loop melalui semua service dan upgrade package
for service in "${services[@]}"
do
  if [ -d "$service" ]; then
    echo "Processing $service..."
    cd "$service" || exit
    
    # Hapus folder vendor
    echo "Removing vendor folder..."
    rm -rf vendor
    
    # Update dan vendor dependencies
    echo "Running go mod vendor..."
    go mod tidy
    go mod vendor
    
    # Kembali ke direktori utama
    cd - || exit
    
    echo "Completed $service"
    echo "------------------------"
  else
    echo "Warning: Directory $service not found"
  fi
done

echo "All services processed successfully"