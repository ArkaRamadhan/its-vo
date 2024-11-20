#!/bin/bash

# Definisikan array dengan direktori semua service
services=("dokumen-service" "informasi-service" "kegiatan-service" "user-service" "weeklyTimeline-service" "project-service")

# Versi baru yang ingin diupgrade
new_version="a6462fd209c1"

# Loop melalui semua service dan upgrade package
for service in "${services[@]}"
do
  echo "Upgrading $service to $new_version"
  cd $service
  go get github.com/arkaramadhan/its-vo/common@$new_version
  go mod tidy
  go mod vendor
  echo "Finished upgrading $service"
  cd ..
done