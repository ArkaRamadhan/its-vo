#!/bin/bash

# Warna untuk output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Memulai Proses Push ke GitHub ===${NC}"

# Menambahkan semua perubahan
echo -e "\n${GREEN}1. Menambahkan file yang berubah...${NC}"
git add .

# Meminta input commit message
echo -e "\n${GREEN}2. Masukkan pesan commit:${NC}"
read fix

# Melakukan commit
echo -e "\n${GREEN}3. Melakukan commit...${NC}"
git commit -m "$commit_message"

# Push ke repository
echo -e "\n${GREEN}4. Melakukan push ke repository...${NC}"
git push origin main

# Menampilkan git log
echo -e "\n${BLUE}=== Git Log Terakhir ===${NC}"
git log --oneline -n 5

echo -e "\n${GREEN}=== Proses Selesai ===${NC}"