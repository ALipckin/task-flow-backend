sudo chmod -R 777 ./

cp -n .env.example .env || echo ".env already exists, skipping."

command -v docker >/dev/null 2>&1 || { echo >&2 "Docker is not installed."; exit 1; }

if [ ! -f docker-compose.yml ]; then
  echo "docker-compose.yml not found!"
  exit 1
fi

docker compose up -d --build