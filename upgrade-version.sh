#!/bin/bash

# Definisikan array dengan direktori semua service
services=("dokumen-service" "informasi-service" "kegiatan-service" "user-service" "weeklyTimeline-service" "project-service")

# Versi baru yang ingin diupgrade
new_version="6599982313ed"

# Loop melalui semua service dan upgrade package
for service in "${services[@]}"
do
  echo "Upgrading $service to $new_version"
  cd $service
  go get github.com/arkaramadhan/its-vo/common@$new_version
  go get github.com/arkaramadhan/its-vo/informasi-service@$new_version
  go get github.com/arkaramadhan/its-vo/kegiatan-service@$new_version
  go get github.com/arkaramadhan/its-vo/project-service@$new_version
  go get github.com/arkaramadhan/its-vo/user-service@$new_version
  go get github.com/arkaramadhan/its-vo/weeklyTimeline-service@$new_version
  cd ..
done