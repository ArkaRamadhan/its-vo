services=("dokumen-service" "informasi-service" "kegiatan-service" "weeklyTimeline-service" "project-service")

# Versi baru yang ingin diupgrade
new_version="f42a54ea18d8"

# Loop melalui semua service dan upgrade package
for service in "${services[@]}"
do
  echo "Upgrading $service to $new_version"
  go get github.com/arkaramadhan/its-vo/$service@$new_version
done